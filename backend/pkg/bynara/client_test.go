package bynara

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
	method, path, auth, contentType, body string
}

func newServer(t *testing.T, status int, reply string) (c *Client, reqs chan captured) {
	t.Helper()
	reqs = make(chan captured, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		reqs <- captured{r.Method, r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type"), strings.TrimSpace(string(b))}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	c = NewClient("kunci-uji")
	c.URL = srv.URL + "/v1/chat/completions"
	c.HTTP = srv.Client()
	return c, reqs
}

func TestChatImageRequest(t *testing.T) {
	c, reqs := newServer(t, http.StatusOK, `{"choices":[{"message":{"content":"[]"}}]}`)
	text, err := c.ChatImage(context.Background(), "stepfun-3.7-flash", "Halo <b>", "data:image/png;base64,QUJD")
	if err != nil || text != "[]" {
		t.Fatalf("ChatImage = %q, %v", text, err)
	}
	want := captured{
		http.MethodPost, "/v1/chat/completions", "Bearer kunci-uji", "application/json",
		`{"model":"stepfun-3.7-flash","messages":[{"role":"user","content":[{"type":"text","text":"Halo <b>"},{"type":"image_url","image_url":{"url":"data:image/png;base64,QUJD"}}]}]}`,
	}
	if got := <-reqs; got != want {
		t.Fatalf("request = %+v\ningin %+v", got, want)
	}
}

func TestChatImageContentShapes(t *testing.T) {
	cases := []struct {
		reply, want string
	}{
		{`{"choices":[{"message":{"content":"teks"}}]}`, "teks"},
		{`{"choices":[{"message":{"content":[{"text":"a"},{"content":"b"},{"text":"","content":"c"},"x",{"type":"image"}]}}]}`, "abc"},
		{`{"choices":[{"message":{"content":null}}]}`, `""`},
		{`{"choices":[{"message":{}}]}`, `""`},
		{`{"choices":[]}`, `""`},
		{`{"choices":[{"message":{"content":0}}]}`, `""`},
		{`{"choices":[{"message":{"content":false}}]}`, `""`},
		{`{"choices":[{"message":{"content":{ "a" : 1 }}}]}`, `{"a":1}`},
		{`{"choices":[{"message":{"content":12}}]}`, `12`},
	}
	for _, tc := range cases {
		c, _ := newServer(t, http.StatusOK, tc.reply)
		if text, err := c.ChatImage(context.Background(), "m", "p", "u"); err != nil || text != tc.want {
			t.Errorf("%s: ChatImage = %q, %v; ingin %q", tc.reply, text, err, tc.want)
		}
	}
}

func TestChatImageAPIError(t *testing.T) {
	c, _ := newServer(t, http.StatusBadGateway, `upstream mati`)
	_, err := c.ChatImage(context.Background(), "m", "p", "u")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusBadGateway || apiErr.Body != "upstream mati" {
		t.Fatalf("err = %#v", err)
	}
}

func TestDefaultURL(t *testing.T) {
	if c := NewClient("k"); c.URL != "https://router.bynara.id/v1/chat/completions" || c.APIKey != "k" {
		t.Fatalf("client = %+v", c)
	}
}
