package http

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
)

const (
	allowedOrigin = "https://siakad.algohub.web.id"
	previewOrigin = "https://algohub-siakad-v2-git-main-lab-algoritma-pemrograman.vercel.app"
	foreignOrigin = "https://evil.example.com"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r, err := NewRouter(&config.Config{
		AppEnv:             "test",
		CORSOrigins:        []string{allowedOrigin},
		CORSOriginPatterns: []*regexp.Regexp{regexp.MustCompile(`^https://algohub-[a-z0-9-]+-lab-algoritma-pemrograman\.vercel\.app$`)},
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return r
}

func serve(r http.Handler, method, path, origin string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, http.NoBody)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodGet, "/api/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, ingin 200", rec.Code)
	}
	if body := rec.Body.String(); body != `{"service":"algohub-api","status":"ok"}` {
		t.Fatalf("body %s", body)
	}
}

func TestCORSAllowedOrigins(t *testing.T) {
	r := newTestRouter(t)
	for _, origin := range []string{allowedOrigin, previewOrigin} {
		rec := serve(r, http.MethodGet, "/api/health", origin, nil)
		h := rec.Header()
		if h.Get("Access-Control-Allow-Origin") != origin {
			t.Errorf("%s: Allow-Origin = %q", origin, h.Get("Access-Control-Allow-Origin"))
		}
		if h.Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("%s: Allow-Credentials tidak true", origin)
		}
		if h.Get("Access-Control-Expose-Headers") != "Content-Disposition, Retry-After" {
			t.Errorf("%s: Expose-Headers = %q", origin, h.Get("Access-Control-Expose-Headers"))
		}
		if h.Get("Vary") != "Origin" {
			t.Errorf("%s: Vary = %q", origin, h.Get("Vary"))
		}
	}
}

func TestCORSForeignOriginGetsNoHeaders(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodGet, "/api/health", foreignOrigin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, ingin 200 (browser yang memblokir)", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("Allow-Origin = %q untuk origin asing", v)
	}
}

func TestCORSNoOriginIsServed(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodGet, "/api/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, ingin 200 untuk request tanpa Origin", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("Allow-Origin = %q tanpa Origin", v)
	}
}

func TestCORSPreflight(t *testing.T) {
	r := newTestRouter(t)
	reached := false
	r.POST("/api/test/guarded", func(c *gin.Context) {
		reached = true
		c.Status(http.StatusUnauthorized)
	})

	rec := serve(r, http.MethodOptions, "/api/test/guarded", allowedOrigin, map[string]string{
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "authorization, content-type",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, ingin 204", rec.Code)
	}
	if reached {
		t.Fatal("preflight tidak boleh mencapai handler")
	}
	h := rec.Header()
	want := map[string]string{
		"Access-Control-Allow-Origin":      allowedOrigin,
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Authorization, Content-Type, X-Requested-With, X-Cron-Secret",
		"Access-Control-Max-Age":           "86400",
	}
	for k, v := range want {
		if h.Get(k) != v {
			t.Errorf("%s = %q, ingin %q", k, h.Get(k), v)
		}
	}
}

func TestCORSPreflightUnknownRoute(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodOptions, "/api/tidak-ada", allowedOrigin, map[string]string{
		"Access-Control-Request-Method": "GET",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, ingin 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Fatal("Max-Age hilang")
	}
}

func TestCORSPreflightForeignOrigin(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodOptions, "/api/health", foreignOrigin, map[string]string{
		"Access-Control-Request-Method": "GET",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, ingin 204", rec.Code)
	}
	for _, k := range []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Methods", "Access-Control-Max-Age"} {
		if v := rec.Header().Get(k); v != "" {
			t.Errorf("%s = %q untuk origin asing", k, v)
		}
	}
}

func TestCORSHeadersOnErrors(t *testing.T) {
	r := newTestRouter(t)
	r.GET("/api/test/401", func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) })
	r.GET("/api/test/panic", func(*gin.Context) { panic("uji") })

	cases := []struct {
		path string
		want int
	}{
		{"/api/test/401", http.StatusUnauthorized},
		{"/api/test/panic", http.StatusInternalServerError},
		{"/api/tidak-ada", http.StatusNotFound},
	}
	for _, tc := range cases {
		rec := serve(r, http.MethodGet, tc.path, allowedOrigin, nil)
		if rec.Code != tc.want {
			t.Errorf("%s: status %d, ingin %d", tc.path, rec.Code, tc.want)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != allowedOrigin {
			t.Errorf("%s: header CORS hilang di response error", tc.path)
		}
	}
}
