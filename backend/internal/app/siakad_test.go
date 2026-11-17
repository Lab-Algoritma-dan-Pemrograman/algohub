package app

import "testing"

func TestNewRequiresSiakadEnv(t *testing.T) {
	const (
		siakadDSN    = "postgres://s@127.0.0.1:1/db"
		elearningDSN = "postgres://e@127.0.0.1:1/db"
	)
	cases := []struct {
		siakad, elearning, secret string
		want                      string
	}{
		{"", elearningDSN, "rahasia", "config: DATABASE_URL_SIAKAD wajib diisi"},
		{siakadDSN, " ", "rahasia", "config: DATABASE_URL_ELEARNING wajib diisi"},
		{siakadDSN, elearningDSN, "", "config: SIAKAD_JWT_SECRET wajib diisi"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Setenv("APP_ENV", "")
			t.Setenv("CORS_ORIGINS", "")
			t.Setenv("CORS_ORIGIN_PATTERNS", "")
			// Port 1 tidak pernah terbuka: kalau validasi terlewat, test gagal karena error koneksi.
			t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@127.0.0.1:1/db")
			t.Setenv("DATABASE_URL_SIAKAD", tc.siakad)
			t.Setenv("DATABASE_URL_ELEARNING", tc.elearning)
			t.Setenv("SIAKAD_JWT_SECRET", tc.secret)

			_, err := New()
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, ingin %q", err, tc.want)
			}
		})
	}
}
