package config

import (
	"os"
	"strings"
)

type ElearningConfig struct {
	SupabaseJWTSecret      string
	SupabaseAnonKey        string
	SupabaseURL            string
	SupabaseServiceRoleKey string
	SiakadJWTSecret        string
	LegacyJWTSecret        string
	GeminiAPIKeys          []string
	OllamaURL              string
	OllamaModel            string
	OllamaAPIKey           string
	AppURL                 string
}

// Dibaca terpisah dari Load supaya config.go tidak berubah. SIAKAD_JWT_SECRET dan JWT_SECRET di-trim
// karena api/auth.ts lama juga men-trim keduanya sebelum verifikasi.
func LoadElearning() ElearningConfig {
	return ElearningConfig{
		SupabaseJWTSecret:      getEnv("ELEARNING_SUPABASE_JWT_SECRET", ""),
		SupabaseAnonKey:        getEnv("ELEARNING_SUPABASE_ANON_KEY", ""),
		SupabaseURL:            getEnv("ELEARNING_SUPABASE_URL", ""),
		SupabaseServiceRoleKey: getEnv("ELEARNING_SUPABASE_SERVICE_ROLE_KEY", ""),
		SiakadJWTSecret:        strings.TrimSpace(os.Getenv("SIAKAD_JWT_SECRET")),
		LegacyJWTSecret:        strings.TrimSpace(getEnv("JWT_SECRET", defaultJWTSecret)),
		GeminiAPIKeys:          splitList(os.Getenv("GEMINI_API_KEY")),
		OllamaURL:              getEnv("ELEARNING_OLLAMA_URL", ""),
		OllamaModel:            getEnv("ELEARNING_OLLAMA_MODEL", "gpt-oss:120b-cloud"),
		OllamaAPIKey:           getEnv("ELEARNING_OLLAMA_API_KEY", ""),
		AppURL:                 strings.TrimRight(getEnv("ELEARNING_APP_URL", "https://elearning.algohub.web.id"), "/"),
	}
}
