package usecase

import (
	"maps"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
)

type Tokens struct {
	secret string
}

func NewTokens(secret string) *Tokens {
	return &Tokens{secret: secret}
}

func (t *Tokens) Verify(token string) (map[string]any, error) {
	return jwt.VerifyHS256(t.secret, token)
}

func (t *Tokens) Sign(claims map[string]any, ttl time.Duration) (string, error) {
	now := time.Now()
	out := make(map[string]any, len(claims)+2)
	maps.Copy(out, claims)
	out["iat"] = now.Unix()
	out["exp"] = now.Add(ttl).Unix()
	return jwt.SignHS256(t.secret, out)
}
