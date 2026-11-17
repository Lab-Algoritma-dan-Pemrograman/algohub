package config

import (
	"slices"
	"strings"
	"testing"
)

func TestLoadDefaultsToDevOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@localhost:6543/db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppEnv != "development" || cfg.IsProduction() {
		t.Fatalf("AppEnv = %q, ingin development", cfg.AppEnv)
	}
	if !slices.Equal(cfg.CORSOrigins, devOrigins) {
		t.Fatalf("CORSOrigins = %v, ingin %v", cfg.CORSOrigins, devOrigins)
	}
}

func TestLoadProductionRequiresOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ORIGINS", "  ")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load harus gagal tanpa CORS_ORIGINS di production")
	}
}

func TestLoadParsesOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ORIGINS", " https://algohub.web.id , ,https://siakad.algohub.web.id")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@localhost:6543/db")
	t.Setenv("JWT_SECRET", "rahasia-uji")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"https://algohub.web.id", "https://siakad.algohub.web.id"}
	if !slices.Equal(cfg.CORSOrigins, want) {
		t.Fatalf("CORSOrigins = %v, ingin %v", cfg.CORSOrigins, want)
	}
}

func TestLoadRejectsInvalidPattern(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "https://(")

	if _, err := Load(); err == nil {
		t.Fatal("Load harus gagal untuk regex tidak valid")
	}
}

func TestLoadRequiresPraktikumDSN(t *testing.T) {
	cases := []struct {
		name    string
		env     string
		origins string
	}{
		{"development", "", ""},
		{"production", "production", "https://algohub.web.id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.env)
			t.Setenv("CORS_ORIGINS", tc.origins)
			t.Setenv("CORS_ORIGIN_PATTERNS", "")
			t.Setenv("JWT_SECRET", "rahasia-uji")
			t.Setenv("DATABASE_URL_PRAKTIKUM", " ")

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "DATABASE_URL_PRAKTIKUM") {
				t.Fatalf("err = %v, ingin error DATABASE_URL_PRAKTIKUM", err)
			}
		})
	}
}

func TestLoadProductionRejectsDefaultJWTSecret(t *testing.T) {
	for _, secret := range []string{"", defaultJWTSecret} {
		t.Run(secret, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("CORS_ORIGINS", "https://algohub.web.id")
			t.Setenv("CORS_ORIGIN_PATTERNS", "")
			t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@localhost:6543/db")
			t.Setenv("JWT_SECRET", secret)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
				t.Fatalf("err = %v, ingin error JWT_SECRET", err)
			}
		})
	}
}

func TestLoadKeepsPraktikumSecretVerbatim(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", "postgres://u@localhost:6543/db")
	t.Setenv("JWT_SECRET", " a b ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.JWTSecret != " a b " {
		t.Fatalf("JWTSecret = %q, ingin %q", cfg.JWTSecret, " a b ")
	}
}

func TestLoadDBConfig(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")
	t.Setenv("DATABASE_URL_PRAKTIKUM", " postgres://p@localhost:6543/praktikum ")
	t.Setenv("DATABASE_URL_SIAKAD", "postgres://s@localhost:6543/siakad")
	t.Setenv("DATABASE_URL_ELEARNING", "postgres://e@localhost:6543/elearning")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := DBConfig{
		Praktikum: "postgres://p@localhost:6543/praktikum",
		Siakad:    "postgres://s@localhost:6543/siakad",
		Elearning: "postgres://e@localhost:6543/elearning",
	}
	if cfg.DB != want {
		t.Fatalf("DB = %+v, ingin %+v", cfg.DB, want)
	}
}

func TestPatternsAreAnchored(t *testing.T) {
	patterns, err := compilePatterns([]string{
		`https://algohub-[a-z0-9-]+\.vercel\.app`,
		`^https://x\.example\.com$`,
	})
	if err != nil {
		t.Fatalf("compilePatterns: %v", err)
	}

	cases := []struct {
		origin string
		idx    int
		want   bool
	}{
		{"https://algohub-abc-123.vercel.app", 0, true},
		{"https://evil.com/https://algohub-abc.vercel.app", 0, false},
		{"https://algohub-abc.vercel.app.evil.com", 0, false},
		{"https://x.example.com", 1, true},
		{"https://x.example.com.evil.com", 1, false},
	}
	for _, tc := range cases {
		if got := patterns[tc.idx].MatchString(tc.origin); got != tc.want {
			t.Errorf("pola %d, origin %q: %v, ingin %v", tc.idx, tc.origin, got, tc.want)
		}
	}
}
