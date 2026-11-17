package supabasejwt

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func anonKey(t *testing.T, key []byte) string {
	t.Helper()
	tok, err := SignHS256(jwt.MapClaims{"role": "anon", "iss": "supabase"}, key, time.Hour)
	require.NoError(t, err)
	return tok
}

func TestDetectSecretRaw(t *testing.T) {
	secret := "rahasia-supabase-mentah-yang-panjang"
	got := DetectSecret("  "+secret+"\n", anonKey(t, []byte(secret)))
	assert.Equal(t, []byte(secret), got)
}

func TestDetectSecretBase64(t *testing.T) {
	key := []byte{0xfb, 0xff, 0x01, 0x02, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80, 0x90, 0xa0, 0xb0, 0xc0}
	for name, enc := range map[string]*base64.Encoding{
		"std": base64.StdEncoding, "rawstd": base64.RawStdEncoding, "url": base64.URLEncoding, "rawurl": base64.RawURLEncoding,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, key, DetectSecret(enc.EncodeToString(key), anonKey(t, key)))
		})
	}
}

func TestDetectSecretFallbackRaw(t *testing.T) {
	assert.Nil(t, DetectSecret("  ", "apa-saja"))
	assert.Equal(t, []byte("abc"), DetectSecret("abc", ""))
	assert.Equal(t, []byte("abc"), DetectSecret("abc", anonKey(t, []byte("kunci-lain"))))
	assert.Equal(t, []byte("abc"), DetectSecret("abc", "bukan.jwt"))
}

func TestSignHS256SetsIatExp(t *testing.T) {
	key := []byte("k")
	in := jwt.MapClaims{"nim": "1"}
	tok, err := SignHS256(in, key, 24*time.Hour)
	require.NoError(t, err)
	assert.NotContains(t, in, "exp", "claims masukan tidak boleh diubah")

	claims, err := VerifyHMAC(tok, key, "HS256")
	require.NoError(t, err)
	iat, err := claims["iat"].(json.Number).Int64()
	require.NoError(t, err)
	exp, err := claims["exp"].(json.Number).Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(86400), exp-iat)
	assert.Equal(t, "1", claims["nim"])
}

func TestVerifyHMACErrors(t *testing.T) {
	key := []byte("kunci")
	good, err := SignHS256(jwt.MapClaims{"nim": "1"}, key, time.Hour)
	require.NoError(t, err)

	_, err = VerifyHMAC("bukan-token", key)
	assert.ErrorIs(t, err, jwt.ErrTokenMalformed)

	_, err = VerifyHMAC(good, []byte("kunci-salah"))
	assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)

	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()}).SignedString(key)
	require.NoError(t, err)
	_, err = VerifyHMAC(expired, key)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)

	hs384, err := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{"nim": "1"}).SignedString(key)
	require.NoError(t, err)
	_, err = VerifyHMAC(hs384, key)
	assert.NoError(t, err, "tanpa methods HS384 diterima")
	_, err = VerifyHMAC(hs384, key, "HS256")
	assert.Error(t, err, "HS384 ditolak bila hanya HS256")

	noExp, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"nim": "1"}).SignedString(key)
	require.NoError(t, err)
	_, err = VerifyHMAC(noExp, key)
	assert.NoError(t, err, "exp tidak wajib")
}
