package vercelblob

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

const testToken = "vercel_blob_rw_teststore_secret"

func TestUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Query().Get("pathname") != "foto/a.png" {
			t.Fatalf("request = %s %s", r.Method, r.URL.String())
		}
		if r.Header.Get("Authorization") != "Bearer "+testToken || r.Header.Get("x-vercel-blob-access") != "public" {
			t.Fatalf("header auth/access tidak sesuai")
		}
		_ = json.NewEncoder(w).Encode(blobResult{URL: "https://teststore.public.blob.vercel-storage.com/foto/a.png"})
	}))
	defer server.Close()

	client := New(testToken, Public)
	client.apiURL = server.URL
	client.http = server.Client()
	got, err := client.Upload("foto/a.png", []byte("png"), "image/png", false)
	if err != nil || got == "" {
		t.Fatalf("Upload() = %q, %v", got, err)
	}
}

func TestSignedUploadURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/signed-token" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(signedToken{
			DelegationToken:    "delegation",
			ClientSigningToken: "signing-secret",
			ValidUntil:         time.Now().Add(time.Hour).UnixMilli(),
		})
	}))
	defer server.Close()

	client := New(testToken, Private)
	client.apiURL = server.URL
	client.http = server.Client()
	raw, err := client.SignedUploadURL("1/2/laporan.pdf", 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Query().Get("pathname") != "1/2/laporan.pdf" || u.Query().Get("vercel-blob-delegation") != "delegation" {
		t.Fatalf("signed URL tidak valid: %s", raw)
	}
	mac := hmac.New(sha256.New, []byte("signing-secret"))
	_, _ = mac.Write([]byte("operation=put\npathname=1/2/laporan.pdf"))
	want := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if got := u.Query().Get("vercel-blob-signature"); got != want {
		t.Fatalf("signature = %q, ingin %q", got, want)
	}
}

func TestStoreIDFromToken(t *testing.T) {
	if got := storeIDFromToken(testToken); got != "teststore" {
		t.Fatalf("store id = %q", got)
	}
	if New("rusak", Public).Enabled() {
		t.Fatal("token rusak tidak boleh mengaktifkan client")
	}
}
