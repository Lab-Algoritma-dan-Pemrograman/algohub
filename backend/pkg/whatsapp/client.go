package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	defaultFonnteURL    = "https://api.fonnte.com/send"
	defaultWablasURL    = "https://kudus.wablas.com/api/send-message"
	defaultWhacenterURL = "https://app.whacenter.com/api/send"
)

// Client mengirim pesan WhatsApp lewat gateway fonnte, wablas, whacenter, atau URL webhook kustom.
type Client struct {
	HTTP         *http.Client
	FonnteURL    string
	WablasURL    string
	WhacenterURL string
}

// NewClient membuat Client dengan URL gateway bawaan.
func NewClient() *Client {
	return &Client{
		// 50s: di bawah maxDuration Vercel Hobby (60s) supaya fungsi sempat membalas.
		HTTP:         &http.Client{Timeout: 50 * time.Second},
		FonnteURL:    defaultFonnteURL,
		WablasURL:    defaultWablasURL,
		WhacenterURL: defaultWhacenterURL,
	}
}

type fonnteBody struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

type wablasBody struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

type whacenterBody struct {
	DeviceID string `json:"device_id"`
	Number   string `json:"number"`
	Message  string `json:"message"`
}

type webhookBody struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// Send mengirim pesan dan mengembalikan teks respons gateway apa pun status HTTP-nya.
func (c *Client) Send(ctx context.Context, provider, token, phone, message string) (string, error) {
	var (
		url  string
		auth string
		body any
	)
	switch provider {
	case "fonnte":
		url, auth, body = c.FonnteURL, token, fonnteBody{phone, message}
	case "wablas":
		url, auth, body = c.WablasURL, token, wablasBody{phone, message}
	case "whacenter":
		url, body = c.WhacenterURL, whacenterBody{token, phone, message}
	default:
		url, auth, body = provider, "Bearer "+token, webhookBody{phone, message}
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	text, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
