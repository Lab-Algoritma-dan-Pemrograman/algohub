package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSiakadRateLimit(t *testing.T) {
	r := limitedEngine(t, SiakadRateLimit(5, time.Minute, "pesan batas", nil))
	a := map[string]string{"X-Vercel-Forwarded-For": "203.0.113.7"}
	b := map[string]string{"X-Vercel-Forwarded-For": "203.0.113.8, 10.0.0.1"}

	for i := range 5 {
		if code := hit(r, a); code != http.StatusOK {
			t.Fatalf("request ke-%d = %d, ingin 200", i+1, code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Vercel-Forwarded-For", "203.0.113.7")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request ke-6 = %d, ingin 429", w.Code)
	}
	if body := w.Body.String(); body != `{"error":"pesan batas"}` {
		t.Fatalf("body = %s", body)
	}
	if code := hit(r, b); code != http.StatusOK {
		t.Fatalf("IP lain = %d, ingin 200", code)
	}
}

func TestSiakadRateLimitApplies(t *testing.T) {
	applies := func(c *gin.Context) bool { return c.GetHeader("X-Hitung") == "1" }
	r := limitedEngine(t, SiakadRateLimit(1, time.Minute, "x", applies))
	ip := "203.0.113.9"

	for range 3 {
		if code := hit(r, map[string]string{"X-Vercel-Forwarded-For": ip}); code != http.StatusOK {
			t.Fatalf("request yang tidak dihitung = %d, ingin 200", code)
		}
	}
	if code := hit(r, map[string]string{"X-Vercel-Forwarded-For": ip, "X-Hitung": "1"}); code != http.StatusOK {
		t.Fatalf("request dihitung ke-1 = %d, ingin 200", code)
	}
	if code := hit(r, map[string]string{"X-Vercel-Forwarded-For": ip, "X-Hitung": "1"}); code != http.StatusTooManyRequests {
		t.Fatalf("request dihitung ke-2 = %d, ingin 429", code)
	}
}
