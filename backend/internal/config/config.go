package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const envProduction = "production"

var devOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5176",
}

type Config struct {
	AppEnv             string
	CORSOrigins        []string
	CORSOriginPatterns []*regexp.Regexp
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      envOr("APP_ENV", "development"),
		CORSOrigins: splitList(os.Getenv("CORS_ORIGINS")),
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
