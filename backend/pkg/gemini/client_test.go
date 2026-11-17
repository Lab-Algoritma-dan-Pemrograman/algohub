package gemini

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type captured struct {
	method, path, key, contentType, body string
}

func newServer(t *testing.T, status int, reply string) (c *Client, reqs chan captured) {
	t.Helper()
	reqs = make(chan captured, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		reqs <- captured{r.Method, r.URL.Path, r.URL.Query().Get("key"), r.Header.Get("Content-Type"), strings.TrimSpace(string(b))}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	c = NewClient("kunci-uji")
	c.BaseURL = srv.URL + "/v1beta"
	c.HTTP = srv.Client()
	return c, reqs
}

func TestGenerateContent(t *testing.T) {
	c, reqs := newServer(t, http.StatusOK, `{"candidates":[{"content":{"parts":[{"text":"[{\"a\":1}]"},{"text":"lain"}]}}]}`)
	text, err := c.GenerateContent(context.Background(), "gemini-3-flash", "Halo <b>", "image/png", "QUJD")
	if err != nil || text != `[{"a":1}]` {
		t.Fatalf("GenerateContent = %q, %v", text, err)
	}
	want := captured{
		http.MethodPost, "/v1beta/models/gemini-3-flash:generateContent", "kunci-uji", "application/json",
		`{"contents":[{"parts":[{"text":"Halo <b>"},{"inline_data":{"mime_type":"image/png","data":"QUJD"}}]}]}`,
	}
	if got := <-reqs; got != want {
		t.Fatalf("request = %+v\ningin %+v", got, want)
	}
}

func TestGenerateContentEmptyCandidates(t *testing.T) {
	c, _ := newServer(t, http.StatusOK, `{"candidates":[]}`)
	if text, err := c.GenerateContent(context.Background(), "m", "p", "image/png", "x"); err != nil || text != "" {
		t.Fatalf("GenerateContent = %q, %v", text, err)
	}
}

func TestGenerateContentAPIError(t *testing.T) {
	c, _ := newServer(t, http.StatusTooManyRequests, `quota habis`)
	_, err := c.GenerateContent(context.Background(), "m", "p", "image/png", "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTooManyRequests || apiErr.Body != "quota habis" {
		t.Fatalf("err = %#v", err)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	if c := NewClient("k"); c.BaseURL != "https://generativelanguage.googleapis.com/v1beta" || c.APIKey != "k" {
		t.Fatalf("client = %+v", c)
	}
}
