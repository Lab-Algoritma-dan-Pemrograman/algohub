package whatsapp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type captured struct {
	path, auth, contentType, body string
}

func TestSend(t *testing.T) {
	reqs := make(chan captured, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		reqs <- captured{r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type"), string(b)}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":false}`))
	}))
	defer srv.Close()

	c := NewClient()
	c.HTTP = srv.Client()
	c.FonnteURL = srv.URL + "/fonnte"
	c.WablasURL = srv.URL + "/wablas"
	c.WhacenterURL = srv.URL + "/whacenter"

	cases := []struct {
		provider string
		want     captured
	}{
		{"fonnte", captured{"/fonnte", "tok", "application/json", `{"target":"628123","message":"Halo"}`}},
		{"wablas", captured{"/wablas", "tok", "application/json", `{"phone":"628123","message":"Halo"}`}},
		{"whacenter", captured{"/whacenter", "", "application/json", `{"device_id":"tok","number":"628123","message":"Halo"}`}},
		{srv.URL + "/kustom", captured{"/kustom", "Bearer tok", "application/json", `{"to":"628123","message":"Halo"}`}},
	}
	for _, tc := range cases {
		text, err := c.Send(context.Background(), tc.provider, "tok", "628123", "Halo")
		if err != nil || text != `{"status":false}` {
			t.Errorf("%s: Send = %q, %v", tc.provider, text, err)
			continue
		}
		if got := <-reqs; got != tc.want {
			t.Errorf("%s: request = %+v\ningin %+v", tc.provider, got, tc.want)
		}
	}
}

func TestDefaultURLs(t *testing.T) {
	c := NewClient()
	if c.FonnteURL != "https://api.fonnte.com/send" || c.WablasURL != "https://kudus.wablas.com/api/send-message" ||
		c.WhacenterURL != "https://app.whacenter.com/api/send" {
		t.Fatalf("URL bawaan = %+v", c)
	}
}

func TestSendInvalidProviderURL(t *testing.T) {
	if _, err := NewClient().Send(context.Background(), "bukan url", "tok", "628123", "Halo"); err == nil {
		t.Fatal("provider bukan URL harus error tanpa request")
	}
}
