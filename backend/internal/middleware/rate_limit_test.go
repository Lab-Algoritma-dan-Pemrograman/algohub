package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func limitedEngine(t *testing.T, mw ...gin.HandlerFunc) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	r.GET("/", append(mw, func(c *gin.Context) { c.Status(http.StatusOK) })...)
	return r
}

func hit(r http.Handler, header map[string]string) int {
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.RemoteAddr = "10.0.0.1:5000"
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRateLimitPerForwardedIP(t *testing.T) {
	r := limitedEngine(t, RateLimit(2, time.Minute))
	a := map[string]string{"X-Forwarded-For": "203.0.113.7"}
	b := map[string]string{"X-Forwarded-For": "203.0.113.8"}

	for i := range 2 {
		if code := hit(r, a); code != http.StatusOK {
			t.Fatalf("klien A request ke-%d = %d, want 200", i+1, code)
		}
	}
	if code := hit(r, a); code != http.StatusTooManyRequests {
		t.Errorf("klien A request ke-3 = %d, want 429", code)
	}
	for i := range 2 {
		if code := hit(r, b); code != http.StatusOK {
			t.Errorf("klien B request ke-%d = %d, want 200 (RemoteAddr sama, XFF beda)", i+1, code)
		}
	}
	if code := hit(r, b); code != http.StatusTooManyRequests {
		t.Errorf("klien B request ke-3 = %d, want 429", code)
	}
}

func TestRateLimitUserKeyedByUserID(t *testing.T) {
	setUser := func(c *gin.Context) {
		if id, err := strconv.Atoi(c.GetHeader("X-Test-User")); err == nil {
			c.Set(CtxUserID, id)
		}
		c.Next()
	}
	r := limitedEngine(t, setUser, RateLimitUser(1, time.Minute))
	ip := "203.0.113.7"

	if code := hit(r, map[string]string{"X-Forwarded-For": ip, "X-Test-User": "1"}); code != http.StatusOK {
		t.Fatalf("user 1 request ke-1 = %d, want 200", code)
	}
	if code := hit(r, map[string]string{"X-Forwarded-For": ip, "X-Test-User": "1"}); code != http.StatusTooManyRequests {
		t.Errorf("user 1 request ke-2 = %d, want 429", code)
	}
	if code := hit(r, map[string]string{"X-Forwarded-For": ip, "X-Test-User": "2"}); code != http.StatusOK {
		t.Errorf("user 2 dari IP sama = %d, want 200", code)
	}
	if code := hit(r, map[string]string{"X-Forwarded-For": ip}); code != http.StatusOK {
		t.Errorf("tanpa user, fallback IP = %d, want 200", code)
	}
	if code := hit(r, map[string]string{"X-Forwarded-For": ip}); code != http.StatusTooManyRequests {
		t.Errorf("tanpa user, fallback IP request ke-2 = %d, want 429", code)
	}
}
