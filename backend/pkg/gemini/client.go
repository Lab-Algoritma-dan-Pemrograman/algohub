package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Client memanggil endpoint generateContent Gemini dengan API key.
type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

// NewClient membuat Client dengan URL Gemini bawaan.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: defaultBaseURL,
		// 50s: di bawah maxDuration Vercel Hobby (60s) supaya fungsi sempat membalas.
		HTTP: &http.Client{Timeout: 50 * time.Second},
	}
}

// APIError adalah respons non-2xx dari Gemini beserta body mentahnya.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gemini: status %d", e.Status)
}

type inlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type content struct {
	Parts []part `json:"parts"`
}

type request struct {
	Contents []content `json:"contents"`
}

type response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// GenerateContent mengirim prompt dan satu file inline, lalu mengembalikan teks bagian pertama kandidat pertama.
func (c *Client) GenerateContent(ctx context.Context, model, prompt, mimeType, dataBase64 string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(request{Contents: []content{{Parts: []part{
		{Text: prompt},
		{InlineData: &inlineData{MimeType: mimeType, Data: dataBase64}},
	}}}})
	if err != nil {
		return "", err
	}
	// T-21: key lewat query string dibawa dari kode lama.
	endpoint := c.BaseURL + "/models/" + url.PathEscape(model) + ":generateContent?key=" + url.QueryEscape(c.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return "", &APIError{Status: resp.StatusCode, Body: string(body)}
	}
	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}
