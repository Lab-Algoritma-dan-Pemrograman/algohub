package config

import (
	"slices"
	"testing"
)

func TestLoadElearningDefaults(t *testing.T) {
	for _, k := range []string{
		"ELEARNING_SUPABASE_JWT_SECRET", "ELEARNING_SUPABASE_ANON_KEY", "ELEARNING_SUPABASE_URL",
		"ELEARNING_SUPABASE_SERVICE_ROLE_KEY", "SIAKAD_JWT_SECRET", "JWT_SECRET", "GEMINI_API_KEY",
		"ELEARNING_OLLAMA_URL", "ELEARNING_OLLAMA_MODEL", "ELEARNING_OLLAMA_API_KEY", "ELEARNING_APP_URL",
	} {
		t.Setenv(k, "")
	}

	cfg := LoadElearning()
	if cfg.LegacyJWTSecret != defaultJWTSecret {
		t.Errorf("LegacyJWTSecret = %q, ingin %q", cfg.LegacyJWTSecret, defaultJWTSecret)
	}
	if cfg.OllamaModel != "gpt-oss:120b-cloud" {
		t.Errorf("OllamaModel = %q", cfg.OllamaModel)
	}
	if cfg.AppURL != "https://elearning.algohub.web.id" {
		t.Errorf("AppURL = %q", cfg.AppURL)
	}
	if cfg.SupabaseJWTSecret != "" || cfg.SiakadJWTSecret != "" || cfg.OllamaURL != "" || len(cfg.GeminiAPIKeys) != 0 {
		t.Errorf("nilai kosong harus tetap kosong: %+v", cfg)
	}
}

func TestLoadElearningValues(t *testing.T) {
	t.Setenv("ELEARNING_SUPABASE_JWT_SECRET", " rahasia-supabase ")
	t.Setenv("ELEARNING_SUPABASE_ANON_KEY", "anon")
	t.Setenv("ELEARNING_SUPABASE_URL", "https://x.supabase.co")
	t.Setenv("ELEARNING_SUPABASE_SERVICE_ROLE_KEY", "service")
	t.Setenv("SIAKAD_JWT_SECRET", "  rahasia-siakad\n")
	t.Setenv("JWT_SECRET", " rahasia-go ")
	t.Setenv("GEMINI_API_KEY", " k1, ,k2 ,")
	t.Setenv("ELEARNING_OLLAMA_URL", "https://ollama.example")
	t.Setenv("ELEARNING_OLLAMA_MODEL", "model-x")
	t.Setenv("ELEARNING_OLLAMA_API_KEY", "ok")
	t.Setenv("ELEARNING_APP_URL", "https://el.example.com//")

	cfg := LoadElearning()
	if cfg.SupabaseJWTSecret != " rahasia-supabase " {
		t.Errorf("SupabaseJWTSecret harus mentah, dapat %q", cfg.SupabaseJWTSecret)
	}
	if cfg.SiakadJWTSecret != "rahasia-siakad" || cfg.LegacyJWTSecret != "rahasia-go" {
		t.Errorf("secret legacy harus di-trim: %q %q", cfg.SiakadJWTSecret, cfg.LegacyJWTSecret)
	}
	if !slices.Equal(cfg.GeminiAPIKeys, []string{"k1", "k2"}) {
		t.Errorf("GeminiAPIKeys = %v", cfg.GeminiAPIKeys)
	}
	if cfg.AppURL != "https://el.example.com" {
		t.Errorf("AppURL = %q", cfg.AppURL)
	}
	if cfg.SupabaseAnonKey != "anon" || cfg.SupabaseURL != "https://x.supabase.co" || cfg.SupabaseServiceRoleKey != "service" ||
		cfg.OllamaURL != "https://ollama.example" || cfg.OllamaModel != "model-x" || cfg.OllamaAPIKey != "ok" {
		t.Errorf("nilai env tidak terbaca: %+v", cfg)
	}
}
