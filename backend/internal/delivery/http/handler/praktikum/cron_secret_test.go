package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func cronTestEngine(t *testing.T, secret string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCronHandler(nil, nil, secret)
	r.POST("/cron/praktikum/auto-submit", h.AutoSubmit)
	r.POST("/cron/praktikum/rumpang", h.Rumpang)
	return r
}

// Use case nil: jalur otorisasi harus menolak SEBELUM menyentuh usecase.
func TestCronHandlerSecretRejected(t *testing.T) {
	for _, tc := range []struct {
		name   string
		secret string
		header string
	}{
		{"secret kosong", "", "s3cret"},
		{"header kosong", "s3cret", ""},
		{"header salah", "s3cret", "beda"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := cronTestEngine(t, tc.secret)
			for _, path := range []string{"/cron/praktikum/auto-submit", "/cron/praktikum/rumpang"} {
				req := httptest.NewRequest(http.MethodPost, path, http.NoBody)
				if tc.header != "" {
					req.Header.Set("X-Cron-Secret", tc.header)
				}
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				if w.Code != http.StatusUnauthorized {
					t.Errorf("%s %s: code = %d, want 401", tc.name, path, w.Code)
				}
			}
		})
	}
}
