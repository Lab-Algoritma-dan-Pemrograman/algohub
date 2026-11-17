package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "rahasia-uji"

func TestSignVerifyRoundTrip(t *testing.T) {
	token, err := SignHS256(testSecret, map[string]any{
		"id":   json.Number("9007199254740993"),
		"role": "asisten",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("SignHS256: %v", err)
	}
	claims, err := VerifyHS256(testSecret, token)
	if err != nil {
		t.Fatalf("VerifyHS256: %v", err)
	}
	if claims["id"] != json.Number("9007199254740993") {
		t.Fatalf("id = %#v, ingin json.Number utuh", claims["id"])
	}
	if claims["role"] != "asisten" {
		t.Fatalf("role = %#v", claims["role"])
	}
}

func TestVerifyRejects(t *testing.T) {
	expired, _ := SignHS256(testSecret, map[string]any{"id": 1, "exp": time.Now().Add(-time.Minute).Unix()})
	notYet, _ := SignHS256(testSecret, map[string]any{"id": 1, "nbf": time.Now().Add(time.Hour).Unix()})
	wrongSecret, _ := SignHS256("lain", map[string]any{"id": 1})
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"id": 1}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	rs256, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"id": 1}).SignedString(key)

	cases := map[string]string{
		"kedaluwarsa":   expired,
		"belum berlaku": notYet,
		"secret salah":  wrongSecret,
		"alg none":      none,
		"RS256":         rs256,
		"bukan jwt":     "abc",
	}
	for name, token := range cases {
		if _, err := VerifyHS256(testSecret, token); err == nil {
			t.Errorf("%s: token harus ditolak", name)
		}
	}
}
