package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func nativeServer(t *testing.T, status int, body string, seen func(r *http.Request, payload map[string]any)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(raw, &payload)
		if seen != nil {
			seen(r, payload)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewClient(Config{URL: srv.URL + "/", Model: "gpt-oss:120b-cloud", APIKey: "kunci"})
}

func TestNativeChatJSONRequestShape(t *testing.T) {
	var path, auth string
	var payload map[string]any
	c := nativeServer(t, http.StatusOK, `{"message":{"role":"assistant","content":"{\"a\":1}"}}`, func(r *http.Request, p map[string]any) {
		path, auth, payload = r.URL.Path, r.Header.Get("Authorization"), p
	})
	got, err := c.NativeChatJSON(context.Background(), "nilai ini")
	if err != nil {
		t.Fatalf("NativeChatJSON: %v", err)
	}
	if got != `{"a":1}` {
		t.Errorf("content = %q", got)
	}
	if path != "/api/chat" || auth != "Bearer kunci" {
		t.Errorf("path %q auth %q", path, auth)
	}
	msgs := payload["messages"].([]any)
	first := msgs[0].(map[string]any)
	if payload["model"] != "gpt-oss:120b-cloud" || payload["stream"] != false || payload["format"] != "json" || first["role"] != "user" || first["content"] != "nilai ini" {
		t.Errorf("payload = %v", payload)
	}
}

func TestNativeChatJSONFallbacks(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"response", `{"message":{"content":""},"response":"dari response"}`, "dari response"},
		{"json mentah", `{ "done": true,  "x": [1, 2] }`, `{"done":true,"x":[1,2]}`},
	}
	for _, tc := range cases {
		c := nativeServer(t, http.StatusOK, tc.body, nil)
		got, err := c.NativeChatJSON(context.Background(), "p")
		if err != nil || got != tc.want {
			t.Errorf("%s: %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}
}

func TestNativeChatJSONNon2xx(t *testing.T) {
	c := nativeServer(t, http.StatusBadGateway, "upstream", nil)
	_, err := c.NativeChatJSON(context.Background(), "p")
	if err == nil || err.Error() != "Ollama API returned status: 502" {
		t.Errorf("err = %v", err)
	}
}

func TestNativeChatJSONWithoutAPIKey(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"response":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(Config{URL: srv.URL, Model: "m"})
	if _, err := c.NativeChatJSON(context.Background(), "p"); err != nil || auth != "" {
		t.Errorf("err %v auth %q", err, auth)
	}
}
