package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	envProduction    = "production"
	defaultJWTSecret = "dev-secret-change-me"
)

var devOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
}

type DBConfig struct {
	Praktikum string
	Siakad    string
	Elearning string
}

type Config struct {
	AppEnv             string
	CORSOrigins        []string
	CORSOriginPatterns []*regexp.Regexp

	DB DBConfig

	Siakad SiakadConfig

	JWTSecret      string
	JWTExpireHours int

	// AuthMode: legacy (HS256 lokal saja), dual (Supabase RS256 atau HS256), supabase (hanya RS256)
	AuthMode          string
	SupabaseJWKSURL   string
	SupabaseJWTIssuer string
	SupabaseJWTAud    string
	EmailAllowItpln   bool

	// CronSecret: rahasia header pemicu endpoint cron (auto-submit) dari cron-job.org.
	CronSecret string

	OnlineTTL time.Duration

	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseBucket     string
	// SupabaseBucketLaporan: bucket privat khusus berkas laporan besar (baca via signed URL).
	SupabaseBucketLaporan string

	OllamaURL    string
	OllamaModel  string
	OllamaAPIKey string

	GlotURL   string
	GlotToken string

	// Parameter hash Firebase Auth (untuk verifikasi password lama saat migrasi).
	// Kosong = fitur nonaktif (login fallback ke bcrypt saja).
	FbScryptSignerKey     string
	FbScryptSaltSeparator string
	FbScryptRounds        int
	FbScryptMemCost       int
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("backend/.env")

	cfg := &Config{
		AppEnv:      envOr("APP_ENV", "development"),
		CORSOrigins: splitList(os.Getenv("CORS_ORIGINS")),
		DB: DBConfig{
			Praktikum: envOr("DATABASE_URL_PRAKTIKUM", ""),
			Siakad:    envOr("DATABASE_URL_SIAKAD", ""),
			Elearning: envOr("DATABASE_URL_ELEARNING", ""),
		},
		JWTSecret:             getEnv("JWT_SECRET", defaultJWTSecret),
		JWTExpireHours:        getEnvInt("JWT_EXPIRE_HOURS", 12),
		AuthMode:              getEnv("AUTH_MODE", "legacy"),
		SupabaseJWKSURL:       getEnv("SUPABASE_JWKS_URL", ""),
		SupabaseJWTIssuer:     getEnv("SUPABASE_JWT_ISSUER", ""),
		SupabaseJWTAud:        getEnv("SUPABASE_JWT_AUD", "authenticated"),
		EmailAllowItpln:       getEnv("ALLOW_ITPLN_DOMAIN", "false") == "true",
		CronSecret:            getEnv("CRON_SECRET", ""),
		OnlineTTL:             time.Duration(getEnvInt("ONLINE_TTL_SECONDS", 120)) * time.Second,
		SupabaseURL:           getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey:    getEnv("SUPABASE_SERVICE_KEY", ""),
		SupabaseBucket:        getEnv("SUPABASE_BUCKET", "public-assets"),
		SupabaseBucketLaporan: getEnv("SUPABASE_BUCKET_LAPORAN", "laporan-besar"),
		OllamaURL:             getEnv("OLLAMA_URL", "https://router.bynara.id/v1/chat/completions"),
		OllamaModel:           getEnv("OLLAMA_MODEL", "mimo-v2.5-pro-hermes"),
		OllamaAPIKey:          getEnv("OLLAMA_API_KEY", ""),
		GlotURL:               getEnv("GLOT_URL", "https://glot.io/api/run"),
		GlotToken:             getEnv("GLOT_TOKEN", ""),
		FbScryptSignerKey:     getEnv("FB_SCRYPT_SIGNER_KEY", ""),
		FbScryptSaltSeparator: getEnv("FB_SCRYPT_SALT_SEPARATOR", ""),
		FbScryptRounds:        getEnvInt("FB_SCRYPT_ROUNDS", 8),
		FbScryptMemCost:       getEnvInt("FB_SCRYPT_MEM_COST", 14),
		Siakad:                loadSiakad(),
	}

	patterns, err := compilePatterns(splitList(os.Getenv("CORS_ORIGIN_PATTERNS")))
	if err != nil {
		return nil, err
	}
	cfg.CORSOriginPatterns = patterns

	if len(cfg.CORSOrigins) == 0 {
		if cfg.IsProduction() {
			return nil, errors.New("config: CORS_ORIGINS wajib diisi di production")
		}
		cfg.CORSOrigins = devOrigins
	}
	if cfg.DB.Praktikum == "" {
		return nil, errors.New("config: DATABASE_URL_PRAKTIKUM wajib diisi")
	}
	// MEDIUM (code review): TOLAK fallback dev-secret-change-me di production.
	// Satu salah-konfigurasi deploy = siapa pun bisa forge token HS256 admin.
	if cfg.IsProduction() && (cfg.JWTSecret == "" || cfg.JWTSecret == defaultJWTSecret) {
		return nil, errors.New("config: JWT_SECRET wajib di-set (non-default) saat APP_ENV=production")
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == envProduction
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// Tanpa TrimSpace: nilai rahasia dipakai persis seperti prod lama supaya token yang sudah terbit tetap valid.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Pola selalu dibungkus ^(?:...)$ supaya regex tanpa anchor tidak meloloskan origin lain.
func compilePatterns(raw []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(raw))
	for _, p := range raw {
		re, err := regexp.Compile("^(?:" + strings.TrimSuffix(strings.TrimPrefix(p, "^"), "$") + ")$")
		if err != nil {
			return nil, fmt.Errorf("config: CORS_ORIGIN_PATTERNS %q tidak valid: %w", p, err)
		}
		out = append(out, re)
	}
	return out, nil
}
