package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// InlineData adalah berkas base64 yang dikirim bersama prompt.
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// Part adalah satu bagian konten: teks atau inlineData.
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// GenerationConfig memuat responseMimeType dan responseSchema opsional.
type GenerationConfig struct {
	ResponseMimeType string          `json:"responseMimeType,omitempty"`
	ResponseSchema   json.RawMessage `json:"responseSchema,omitempty"`
}

type roleContent struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type partsRequest struct {
	Contents         []roleContent    `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type partsResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text    *string `json:"text"`
				Thought bool    `json:"thought"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type errorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateParts mengirim parts sebagai pesan user memakai apiKey per panggilan (bukan c.APIKey, karena elearning merotasi kunci) dan mengembalikan gabungan teks kandidat pertama.
func (c *Client) GenerateParts(ctx context.Context, apiKey, model string, parts []Part, cfg GenerationConfig) (string, error) {
	body, err := json.Marshal(partsRequest{
		Contents:         []roleContent{{Role: "user", Parts: parts}},
		GenerationConfig: cfg,
	})
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/models/" + url.PathEscape(model) + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// Header (seperti SDK @google/genai), bukan ?key=, supaya kunci tidak ikut tercetak di pesan *url.Error.
	req.Header.Set("x-goog-api-key", apiKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Kode status dan error.message wajib ada di pesan: api/ai.ts berhenti merotasi kunci bila memuat "404"/"not found"/"safety".
		var e errorBody
		msg := strings.TrimSpace(string(raw))
		if json.Unmarshal(raw, &e) == nil && e.Error.Message != "" {
			msg = e.Error.Message
		}
		return "", fmt.Errorf("gemini: status %d: %s", resp.StatusCode, msg)
	}
	var out partsResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("gemini: respons bukan JSON: %w", err)
	}
	if len(out.Candidates) == 0 {
		return "", nil
	}
	var sb strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		if p.Text != nil && !p.Thought {
			sb.WriteString(*p.Text)
		}
	}
	return sb.String(), nil
}
