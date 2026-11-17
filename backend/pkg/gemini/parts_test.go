package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newPartsClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient("kunci-bawaan")
	c.BaseURL = srv.URL
	return c
}

func TestGeneratePartsSendsRequestAndJoinsText(t *testing.T) {
	var gotPath, gotKey, gotQueryKey string
	var gotBody map[string]any
	c := newPartsClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-goog-api-key")
		gotQueryKey = r.URL.Query().Get("key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"pikir","thought":true},{"text":"hal"},{"text":"o"}]}},{"content":{"parts":[{"text":"lain"}]}}]}`))
	})

	parts := []Part{{Text: "tanya"}, {InlineData: &InlineData{MimeType: "application/pdf", Data: "QUJD"}}}
	cfg := GenerationConfig{ResponseMimeType: "application/json", ResponseSchema: json.RawMessage(`{"type":"OBJECT"}`)}
	text, err := c.GenerateParts(context.Background(), "kunci-1", "gemini-2.5-flash", parts, cfg)
	if err != nil {
		t.Fatalf("GenerateParts: %v", err)
	}
	if text != "halo" {
		t.Errorf("text = %q, want halo", text)
	}
	if gotPath != "/models/gemini-2.5-flash:generateContent" || gotKey != "kunci-1" || gotQueryKey != "" {
		t.Errorf("path %q key %q query key %q", gotPath, gotKey, gotQueryKey)
	}
	contents := gotBody["contents"].([]any)[0].(map[string]any)
	if contents["role"] != "user" || len(contents["parts"].([]any)) != 2 {
		t.Errorf("contents = %v", contents)
	}
	inline := contents["parts"].([]any)[1].(map[string]any)["inlineData"].(map[string]any)
	if inline["mimeType"] != "application/pdf" || inline["data"] != "QUJD" {
		t.Errorf("inlineData = %v", inline)
	}
	gc := gotBody["generationConfig"].(map[string]any)
	if gc["responseMimeType"] != "application/json" || gc["responseSchema"].(map[string]any)["type"] != "OBJECT" {
		t.Errorf("generationConfig = %v", gc)
	}
}

func TestGeneratePartsOmitsEmptySchema(t *testing.T) {
	var raw []byte
	c := newPartsClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	})
	text, err := c.GenerateParts(context.Background(), "k", "m", []Part{{Text: "x"}}, GenerationConfig{ResponseMimeType: "text/plain"})
	if err != nil || text != "" {
		t.Fatalf("text %q err %v", text, err)
	}
	if strings.Contains(string(raw), "responseSchema") {
		t.Errorf("responseSchema harus dihilangkan: %s", raw)
	}
}

func TestGeneratePartsErrorCarriesStatusAndMessage(t *testing.T) {
	c := newPartsClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":404,"message":"models/x is not found for API version v1beta","status":"NOT_FOUND"}}`))
	})
	_, err := c.GenerateParts(context.Background(), "rahasia", "x", []Part{{Text: "x"}}, GenerationConfig{})
	if err == nil {
		t.Fatal("harus error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "404") || !strings.Contains(msg, "not found") || strings.Contains(msg, "rahasia") {
		t.Errorf("pesan = %q", msg)
	}
}

func TestGeneratePartsErrorWithoutJSONBody(t *testing.T) {
	c := newPartsClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("gateway rusak"))
	})
	_, err := c.GenerateParts(context.Background(), "k", "m", nil, GenerationConfig{})
	if err == nil || err.Error() != "gemini: status 500: gateway rusak" {
		t.Errorf("err = %v", err)
	}
}
