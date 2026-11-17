package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StatusError adalah balasan non-2xx dari /api/chat dengan teks yang sama seperti api/grade.ts lama.
type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("Ollama API returned status: %d", e.Code)
}

type nativeChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Format   string        `json:"format"`
}

type nativeChatResponse struct {
	Message *struct {
		Content string `json:"content"`
	} `json:"message"`
	Response string `json:"response"`
}

// NativeChatJSON mengirim prompt ke endpoint native Ollama /api/chat dengan format JSON dan mengembalikan isi balasan.
func (c *Client) NativeChatJSON(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(nativeChatRequest{
		Model:    c.cfg.Model,
		Messages: []ChatMessage{{Role: "user", Content: prompt}},
		Stream:   false,
		Format:   "json",
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.URL, "/")+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", &StatusError{Code: resp.StatusCode}
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if !json.Valid(raw) {
		return "", fmt.Errorf("respons Ollama bukan JSON: %.200s", raw)
	}
	// Urutan message.content || response || JSON.stringify(resJson) sama dengan api/grade.ts.
	var out nativeChatResponse
	if json.Unmarshal(raw, &out) == nil {
		if out.Message != nil && out.Message.Content != "" {
			return out.Message.Content, nil
		}
		if out.Response != "" {
			return out.Response, nil
		}
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", err
	}
	return compact.String(), nil
}
