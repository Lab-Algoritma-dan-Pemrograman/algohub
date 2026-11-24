package vercelblob

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIURL = "https://vercel.com/api/blob"
	apiVersion    = "13"
	maxSignedTTL  = 7 * 24 * time.Hour
	maxUploadSize = 10 << 20
)

type Access string

const (
	Public  Access = "public"
	Private Access = "private"
)

// Client mengakses Vercel Blob lewat HTTP API agar backend Go tidak bergantung
// pada SDK JavaScript. Token hanya dibaca dari environment server.
type Client struct {
	token   string
	storeID string
	access  Access
	apiURL  string
	http    *http.Client
}

func New(token string, access Access) *Client {
	return &Client{
		token:   strings.TrimSpace(token),
		storeID: storeIDFromToken(strings.TrimSpace(token)),
		access:  access,
		apiURL:  defaultAPIURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func storeIDFromToken(token string) string {
	parts := strings.Split(token, "_")
	if len(parts) < 5 || parts[0] != "vercel" || parts[1] != "blob" {
		return ""
	}
	return parts[3]
}

func (c *Client) Enabled() bool {
	return c.token != "" && c.storeID != "" && (c.access == Public || c.access == Private)
}

type blobResult struct {
	URL         string `json:"url"`
	DownloadURL string `json:"downloadUrl"`
	Pathname    string `json:"pathname"`
}

func (c *Client) Upload(path string, content []byte, contentType string, upsert bool) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("vercel blob belum dikonfigurasi")
	}
	path, err := cleanPath(path)
	if err != nil {
		return "", err
	}
	headers := map[string]string{
		"Content-Type":            contentType,
		"x-vercel-blob-access":    string(c.access),
		"x-content-type":          contentType,
		"x-add-random-suffix":     "0",
		"x-allow-overwrite":       boolHeader(upsert),
		"x-cache-control-max-age": "2592000",
	}
	body, err := c.do(http.MethodPut, "/?pathname="+url.QueryEscape(path), content, headers)
	if err != nil {
		return "", err
	}
	var out blobResult
	if err := json.Unmarshal(body, &out); err != nil || out.URL == "" {
		return "", fmt.Errorf("respons Vercel Blob tidak valid")
	}
	return out.URL, nil
}

func (c *Client) Delete(path string) error {
	if !c.Enabled() {
		return fmt.Errorf("vercel blob belum dikonfigurasi")
	}
	path, err := cleanPath(path)
	if err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string][]string{"urls": {c.blobURL(path)}})
	_, err = c.do(http.MethodPost, "/delete", payload, map[string]string{"Content-Type": "application/json"})
	return err
}

func (c *Client) SignedUploadURL(path string, ttl time.Duration) (string, error) {
	path, err := cleanPath(path)
	if err != nil {
		return "", err
	}
	contentType := contentTypeForPath(path)
	token, err := c.issueSignedToken(path, "put", ttl, contentType)
	if err != nil {
		return "", err
	}
	return c.presignedURL(path, "put", token, false), nil
}

func (c *Client) SignedDownloadURL(path string, ttl time.Duration, downloadName string) (string, error) {
	path, err := cleanPath(path)
	if err != nil {
		return "", err
	}
	// Pastikan objek memang berada di Blob. Ini memungkinkan lapisan failover
	// mencoba Supabase untuk path lama yang belum pernah dimigrasikan.
	if _, err := c.do(http.MethodGet, "/?url="+url.QueryEscape(c.blobURL(path)), nil, nil); err != nil {
		return "", err
	}
	token, err := c.issueSignedToken(path, "get", ttl, "")
	if err != nil {
		return "", err
	}
	return c.presignedURL(path, "get", token, downloadName != ""), nil
}

type signedToken struct {
	DelegationToken    string `json:"delegationToken"`
	ClientSigningToken string `json:"clientSigningToken"`
	ValidUntil         int64  `json:"validUntil"`
}

func (c *Client) issueSignedToken(path, operation string, ttl time.Duration, contentType string) (*signedToken, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("vercel blob belum dikonfigurasi")
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > maxSignedTTL {
		ttl = maxSignedTTL
	}
	validUntil := time.Now().Add(ttl).UnixMilli()
	payload := map[string]any{
		"pathname":   path,
		"operations": []string{operation},
		"validUntil": validUntil,
	}
	if operation == "put" {
		payload["maximumSizeInBytes"] = maxUploadSize
		payload["allowedContentTypes"] = []string{contentType}
	}
	raw, _ := json.Marshal(payload)
	body, err := c.do(http.MethodPost, "/signed-token", raw, map[string]string{"Content-Type": "application/json"})
	if err != nil {
		return nil, err
	}
	var out signedToken
	if err := json.Unmarshal(body, &out); err != nil || out.DelegationToken == "" || out.ClientSigningToken == "" {
		return nil, fmt.Errorf("respons signed token Vercel Blob tidak valid")
	}
	return &out, nil
}

func (c *Client) presignedURL(path, operation string, token *signedToken, download bool) string {
	canonical := "operation=" + operation + "\npathname=" + path
	mac := hmac.New(sha256.New, []byte(token.ClientSigningToken))
	_, _ = mac.Write([]byte(canonical))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	var target string
	if operation == "put" {
		target = c.apiURL + "/?pathname=" + url.QueryEscape(path)
	} else {
		target = c.blobURL(path)
	}
	u, _ := url.Parse(target)
	q := u.Query()
	if download {
		q.Set("download", "1")
	}
	q.Set("vercel-blob-delegation", token.DelegationToken)
	q.Set("vercel-blob-signature", signature)
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) blobURL(path string) string {
	u := &url.URL{
		Scheme: "https",
		Host:   c.storeID + "." + string(c.access) + ".blob.vercel-storage.com",
		Path:   "/" + strings.TrimLeft(path, "/"),
	}
	return u.String()
}

type apiError struct {
	StatusCode int
}

func (e *apiError) Error() string { return fmt.Sprintf("vercel blob API %d", e.StatusCode) }

func (c *Client) do(method, endpoint string, body []byte, extraHeaders map[string]string) ([]byte, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("vercel blob belum dikonfigurasi")
	}
	requestID := make([]byte, 8)
	_, _ = rand.Read(requestID)
	req, err := http.NewRequest(method, c.apiURL+endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("x-api-blob-request-id", c.storeID+":"+strconv.FormatInt(time.Now().UnixMilli(), 10)+":"+hex.EncodeToString(requestID))
	req.Header.Set("x-vercel-blob-store-id", c.storeID)
	req.Header.Set("x-api-blob-request-attempt", "0")
	req.Header.Set("x-api-version", apiVersion)
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, &apiError{StatusCode: resp.StatusCode}
	}
	return raw, nil
}

func cleanPath(path string) (string, error) {
	path = strings.TrimLeft(strings.TrimSpace(path), "/")
	if path == "" || strings.Contains(path, "//") || strings.Contains(path, "..") {
		return "", fmt.Errorf("path berkas tidak valid")
	}
	return path, nil
}

func contentTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func boolHeader(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
