package config

import (
	"slices"
	"testing"
)

func TestLoadDefaultsToDevOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("CORS_ORIGIN_PATTERNS", "")

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
