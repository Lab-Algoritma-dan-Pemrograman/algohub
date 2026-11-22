package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Engine dipakai supaya ClientIP melihat konfigurasi TrustedPlatform yang
// sama dengan produksi (router.go).
func clientIPTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.TrustedPlatform = "X-Vercel-Forwarded-For"
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	return r
}

func TestClientIP(t *testing.T) {
	r := clientIPTestEngine(t)
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "%s", ClientIP(c))
	})

	cases := []struct {
		name   string
		header map[string]string
		want   string
	}{
		{"x-vercel-forwarded-for dipercaya", map[string]string{"X-Vercel-Forwarded-For": "203.0.113.7"}, "203.0.113.7"},
		{"x-vercel-forwarded-for dipakai apa adanya", map[string]string{"X-Vercel-Forwarded-For": "203.0.113.7, 10.0.0.1"}, "203.0.113.7, 10.0.0.1"},
		{"xff client diabaikan", map[string]string{"X-Forwarded-For": "198.51.100.99"}, "192.0.2.1"},
		{"xff berantai client diabaikan", map[string]string{"X-Forwarded-For": "198.51.100.99, 10.0.0.1"}, "192.0.2.1"},
		{"x-real-ip client diabaikan", map[string]string{"X-Real-IP": "198.51.100.99"}, "192.0.2.1"},
		{"xff menang atas x-real-ip dua-duanya diabaikan", map[string]string{"X-Forwarded-For": "198.51.100.99", "X-Real-IP": "198.51.100.98"}, "192.0.2.1"},
		{"vercel menang atas xff spoof", map[string]string{"X-Vercel-Forwarded-For": "203.0.113.7", "X-Forwarded-For": "198.51.100.99"}, "203.0.113.7"},
		{"tanpa header fallback remoteIP", nil, "192.0.2.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.RemoteAddr = "192.0.2.1:4321"
			for k, v := range tc.header {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got := w.Body.String(); got != tc.want {
				t.Errorf("ClientIP = %q, want %q", got, tc.want)
			}
		})
	}
}
