package supabasejwt

import (
	"encoding/base64"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var hmacMethods = []string{"HS256", "HS384", "HS512"}

// DetectSecret mengembalikan kunci HMAC Supabase dalam encoding (raw UTF-8 atau base64) yang lolos verifikasi anon key.
func DetectSecret(secret, anonKey string) []byte {
	clean := strings.TrimSpace(secret)
	if clean == "" {
		return nil
	}
	raw := []byte(clean)
	if anonKey == "" {
		return raw
	}
	candidates := [][]byte{raw}
	// Buffer.from(s, 'base64') di Node menerima alfabet standar maupun URL, dengan atau tanpa padding.
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(clean); err == nil && len(b) > 0 {
			candidates = append(candidates, b)
		}
	}
	for _, key := range candidates {
		if _, err := VerifyHMAC(anonKey, key); err == nil {
			return key
		}
	}
	return raw
}

// SignHS256 menandatangani salinan claims dengan HS256 serta mengisi iat dan exp = iat + ttl.
func SignHS256(claims jwt.MapClaims, key []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	c := make(jwt.MapClaims, len(claims)+2)
	maps.Copy(c, claims)
	c["iat"] = now.Unix()
	c["exp"] = now.Add(ttl).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(key)
}

// VerifyHMAC memverifikasi token HMAC dan mengembalikan claims-nya; tanpa methods, HS256/HS384/HS512 diterima.
func VerifyHMAC(token string, key []byte, methods ...string) (jwt.MapClaims, error) {
	if len(methods) == 0 {
		methods = hmacMethods
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode tanda tangan tak terduga: %v", t.Header["alg"])
		}
		return key, nil
	}, jwt.WithValidMethods(methods), jwt.WithJSONNumber())
	if err != nil {
		return nil, err
	}
	return claims, nil
}
