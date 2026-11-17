package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/database"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
)

// Route praktikum dan siakad sama-sama dipasang di /api dan /api/cron; bentrok path baru
// ketahuan sebagai panic Gin saat server start, jadi rakitan lengkap app.New diuji tanpa DB.
func TestPraktikumAndSiakadRoutesCoexist(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("CRON_SECRET", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@127.0.0.1:1/db")
	t.Setenv("DATABASE_URL_SIAKAD", "postgres://s@127.0.0.1:1/db")
	t.Setenv("DATABASE_URL_ELEARNING", "postgres://e@127.0.0.1:1/db")
	t.Setenv("SIAKAD_JWT_SECRET", "rahasia")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	r, err := deliveryhttp.NewRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	deliveryhttp.RegisterPraktikum(r, newPraktikum(cfg, nil))
	registerSiakad(r, cfg, &database.DBs{})

	cases := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/api/health", http.StatusOK},
		{http.MethodPost, "/api/cron/praktikum/auto-submit", http.StatusUnauthorized},
		{http.MethodPost, "/api/cron/auto-submit", http.StatusUnauthorized},
		{http.MethodGet, "/api/cron/siakad/check-shifts", http.StatusUnauthorized},
		{http.MethodGet, "/api/cron/keep-alive", http.StatusUnauthorized},
		{http.MethodPost, "/api/siakad/auth/verify", http.StatusUnauthorized},
		{http.MethodGet, "/api/siakad/rpc", http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, http.NoBody))
		if rec.Code != tc.want {
			t.Errorf("%s %s = %d, ingin %d: %s", tc.method, tc.path, rec.Code, tc.want, rec.Body)
		}
	}
}
