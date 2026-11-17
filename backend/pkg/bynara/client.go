package bynara

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultURL = "https://router.bynara.id/v1/chat/completions"

// Client memanggil endpoint chat-completions Bynara Router dengan Bearer API key.
type Client struct {
	APIKey string
	URL    string
	HTTP   *http.Client
}

// NewClient membuat Client dengan URL Bynara bawaan.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		URL:    defaultURL,
		// 50s: di bawah maxDuration Vercel Hobby (60s) supaya fungsi sempat membalas.
		HTTP: &http.Client{Timeout: 50 * time.Second},
	}
}

// APIError adalah respons non-2xx dari Bynara beserta body mentahnya.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bynara: status %d", e.Status)
}

type textPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type imageURL struct {
	URL string `json:"url"`
}

type imagePart struct {
	Type     string   `json:"type"`
	ImageURL imageURL `json:"image_url"`
}

type message struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ChatImage mengirim prompt dan satu gambar (data URL), lalu mengembalikan isi pesan pilihan pertama sebagai teks.
func (c *Client) ChatImage(ctx context.Context, model, prompt, dataURL string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(chatRequest{Model: model, Messages: []message{{
		Role: "user",
		Content: []any{
			textPart{Type: "text", Text: prompt},
			imagePart{Type: "image_url", ImageURL: imageURL{URL: dataURL}},
		},
	}}})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return "", &APIError{Status: resp.StatusCode, Body: string(body)}
	}
	var out chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	var raw json.RawMessage
	if len(out.Choices) > 0 {
		raw = out.Choices[0].Message.Content
	}
	return contentText(raw)
}

// Model OpenAI-compatible bisa membalas content berupa string, array bagian, atau bentuk lain;
// aturan konversinya sama dengan kode lama supaya parser slot menerima teks yang sama.
func contentText(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return `""`, nil
	}
	switch raw[0] {
	case '"':
		var s string
		err := json.Unmarshal(raw, &s)
		return s, err
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return "", err
		}
		var sb strings.Builder
		for _, item := range items {
			var p struct {
				Text    any `json:"text"`
				Content any `json:"content"`
			}
			if json.Unmarshal(item, &p) != nil {
				continue
			}
			if s, ok := p.Text.(string); ok && s != "" {
				sb.WriteString(s)
			} else if s, ok := p.Content.(string); ok {
				sb.WriteString(s)
			}
		}
		return sb.String(), nil
	}
	switch string(raw) {
	case "null", "false", "0":
		return `""`, nil
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", err
	}
	return compact.String(), nil
}
