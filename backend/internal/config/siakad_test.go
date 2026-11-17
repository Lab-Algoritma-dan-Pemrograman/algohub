package config

import "testing"

func TestSiakadValidate(t *testing.T) {
	full := DBConfig{Siakad: "postgres://s@localhost/db", Elearning: "postgres://e@localhost/db"}
	cases := []struct {
		name string
		db   DBConfig
		cfg  SiakadConfig
		want string
	}{
		{"siakad kosong", DBConfig{Elearning: full.Elearning}, SiakadConfig{JWTSecret: "x"}, "config: DATABASE_URL_SIAKAD wajib diisi"},
		{"elearning kosong", DBConfig{Siakad: full.Siakad}, SiakadConfig{JWTSecret: "x"}, "config: DATABASE_URL_ELEARNING wajib diisi"},
		{"secret kosong", full, SiakadConfig{}, "config: SIAKAD_JWT_SECRET wajib diisi"},
	}
	for _, tc := range cases {
		err := tc.cfg.Validate(tc.db)
		if err == nil || err.Error() != tc.want {
			t.Errorf("%s: err = %v, ingin %q", tc.name, err, tc.want)
		}
	}
	if err := (&SiakadConfig{JWTSecret: "x"}).Validate(full); err != nil {
		t.Errorf("konfigurasi lengkap: %v", err)
	}
}

func TestLoadSiakad(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@localhost:6543/db")
	t.Setenv("SIAKAD_JWT_SECRET", " s p ")
	t.Setenv("VAPID_PUBLIC_KEY", "")
	t.Setenv("VAPID_PRIVATE_KEY", "priv")
	t.Setenv("WA_GATEWAY_TOKEN", "wa")
	t.Setenv("GEMINI_API_KEY", "gem")
	t.Setenv("BYNARA_API_KEY", "byn")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := SiakadConfig{
		JWTSecret:       " s p ",
		VAPIDPublicKey:  defaultVAPIDPublicKey,
		VAPIDPrivateKey: "priv",
		WAGatewayToken:  "wa",
		GeminiAPIKey:    "gem",
		BynaraAPIKey:    "byn",
	}
	if cfg.Siakad != want {
		t.Fatalf("Siakad = %+v, ingin %+v", cfg.Siakad, want)
	}
}
