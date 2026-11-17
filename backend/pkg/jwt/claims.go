package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// SignHS256 menandatangani klaim bebas dengan HS256 dan header typ=JWT.
func SignHS256(secret string, claims map[string]any) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims)).SignedString([]byte(secret))
}

// VerifyHS256 memverifikasi token HMAC dan mengembalikan klaimnya dengan angka sebagai json.Number.
func VerifyHS256(secret, token string) (map[string]any, error) {
	claims := jwt.MapClaims{}
	// json.Number supaya klaim id dikirim ulang ke DB persis seperti saat ditandatangani.
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return []byte(secret), nil
	}, jwt.WithJSONNumber())
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}
