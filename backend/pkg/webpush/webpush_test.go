package webpush

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	wp "github.com/SherClockHolmes/webpush-go"
)

func testSubscription(t *testing.T, endpoint string) Subscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	auth := make([]byte, 16)
	if _, err := rand.Read(auth); err != nil {
		t.Fatal(err)
	}
	return Subscription{
		Endpoint: endpoint,
		P256dh:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString(auth),
	}
}

func vapidClaims(t *testing.T, header string) map[string]any {
	t.Helper()
	rest, ok := strings.CutPrefix(header, "vapid t=")
	if !ok {
		t.Fatalf("Authorization = %q", header)
	}
	token, _, _ := strings.Cut(rest, ", k=")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT VAPID = %q", token)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		t.Fatal(err)
	}
	return claims
}

func TestSend(t *testing.T) {
	private, public, err := wp.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	type captured struct {
		header http.Header
		body   []byte
	}
	reqs := make(chan captured, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		reqs <- captured{r.Header.Clone(), b}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	payload := []byte(`{"title":"Halo","body":"Isi rahasia"}`)
	s := NewSender(public, private, "mailto:admin@lab-ap.web.id", srv.Client())
	status, err := s.Send(context.Background(), testSubscription(t, srv.URL+"/push/abc"), payload)
	if err != nil || status != http.StatusCreated {
		t.Fatalf("Send = %d, %v", status, err)
	}
	c := <-reqs
	if c.header.Get("TTL") != "2419200" {
		t.Errorf("TTL = %q", c.header.Get("TTL"))
	}
	if c.header.Get("Content-Encoding") != "aes128gcm" || len(c.body) == 0 || bytes.Contains(c.body, []byte("rahasia")) {
		t.Errorf("payload tidak terenkripsi: %q %q", c.header.Get("Content-Encoding"), c.body)
	}
	claims := vapidClaims(t, c.header.Get("Authorization"))
	if claims["sub"] != "mailto:admin@lab-ap.web.id" {
		t.Errorf("sub = %v", claims["sub"])
	}
	if claims["aud"] != srv.URL {
		t.Errorf("aud = %v", claims["aud"])
	}
}

func TestSendStatusError(t *testing.T) {
	private, public, err := wp.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer srv.Close()

	status, err := NewSender(public, private, "mailto:admin@lab-ap.web.id", srv.Client()).
		Send(context.Background(), testSubscription(t, srv.URL), []byte("x"))
	var se *StatusError
	if status != http.StatusGone || !errors.As(err, &se) || se.Code != http.StatusGone {
		t.Fatalf("Send = %d, %v", status, err)
	}
}

func TestSendWithoutPrivateKey(t *testing.T) {
	called := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called <- struct{}{} }))
	defer srv.Close()

	_, err := NewSender("pub", "", "mailto:admin@lab-ap.web.id", srv.Client()).
		Send(context.Background(), testSubscription(t, srv.URL), []byte("x"))
	if !errors.Is(err, ErrNoPrivateKey) || len(called) != 0 {
		t.Fatalf("err = %v, request terkirim = %d", err, len(called))
	}
}
