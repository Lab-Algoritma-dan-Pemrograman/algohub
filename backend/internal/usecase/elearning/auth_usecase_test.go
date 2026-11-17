package usecase_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/supabasejwt"
)

const (
	supabaseSecret = "rahasia-supabase-uji"
	legacySecret   = "rahasia-go-uji"
	siakadSecret   = "rahasia-siakad-uji"
)

func sign(t *testing.T, key string, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(method, claims).SignedString([]byte(key))
	require.NoError(t, err)
	return tok
}

func testConfig() *config.ElearningConfig {
	return &config.ElearningConfig{SupabaseJWTSecret: supabaseSecret, LegacyJWTSecret: legacySecret, SiakadJWTSecret: siakadSecret}
}

func newAuth(cfg *config.ElearningConfig, login usecase.PraktikumLogin) *usecase.AuthUsecase {
	return usecase.NewAuthUsecase(cfg, login, time.Now)
}

func assertUCError(t *testing.T, err error, status int, message string) (detail any) {
	t.Helper()
	var e *usecase.Error
	require.True(t, errors.As(err, &e), "error harus *usecase.Error, dapat %v", err)
	assert.Equal(t, status, e.Status)
	assert.Equal(t, message, e.Message)
	return e.Detail
}

func resignedClaims(t *testing.T, token string, key []byte) jwt.MapClaims {
	t.Helper()
	claims, err := supabasejwt.VerifyHMAC(token, key, "HS256")
	require.NoError(t, err, "token harus ditandatangani ulang dengan secret Supabase")
	return claims
}

func TestVerifyAcceptsThreeTokenFormats(t *testing.T) {
	uc := newAuth(testConfig(), nil)

	t.Run("siakad", func(t *testing.T) {
		tok := sign(t, siakadSecret, jwt.SigningMethodHS256, jwt.MapClaims{"username": "202711001", "full_name": "Budi", "role": "koordinator"})
		res, err := uc.Verify(tok)
		require.NoError(t, err)
		assert.Equal(t, dto.AuthPayload{NIM: "202711001", Nama: "Budi", Kelas: "", Role: usecase.RoleKordas}, res.Payload)
		c := resignedClaims(t, res.Token, []byte(supabaseSecret))
		assert.Equal(t, "202711001", c["sub"])
		assert.Equal(t, "authenticated", c["role"])
		assert.Equal(t, "kordas", c["user_role"])
		assert.Equal(t, "supabase", c["iss"])
		assert.Equal(t, "authenticated", c["aud"])
		assert.Contains(t, c, "uid")
		assert.Nil(t, c["uid"])
	})

	t.Run("elearning", func(t *testing.T) {
		tok := sign(t, supabaseSecret, jwt.SigningMethodHS512, jwt.MapClaims{"nim": "202711002", "nama": "Sari", "kelas": "TI-1A", "role": "asisten", "shift": 2})
		res, err := uc.Verify(tok)
		require.NoError(t, err)
		assert.Equal(t, "202711002", res.Payload.NIM)
		assert.Equal(t, "Sari", res.Payload.Nama)
		assert.Equal(t, "TI-1A", res.Payload.Kelas)
		require.NotNil(t, res.Payload.Shift)
		assert.Equal(t, "2", *res.Payload.Shift)
		assert.Equal(t, usecase.RoleAsisten, res.Payload.Role)
	})

	t.Run("go praktikum", func(t *testing.T) {
		tok := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"uid": 7, "nim": "202711003", "role": "mahasiswa"})
		res, err := uc.Verify(tok)
		require.NoError(t, err)
		assert.Equal(t, "202711003", res.Payload.NIM)
		assert.Equal(t, "202711003", res.Payload.Nama, "nama jatuh ke nim")
		assert.Equal(t, usecase.RolePraktikan, res.Payload.Role)
		c := resignedClaims(t, res.Token, []byte(supabaseSecret))
		assert.Equal(t, json.Number("7"), c["uid"])
		assert.Equal(t, "praktikan", c["user_role"])
	})
}

func TestVerifyRejectsBadTokens(t *testing.T) {
	uc := newAuth(testConfig(), nil)

	_, err := uc.Verify("")
	assertUCError(t, err, http.StatusBadRequest, "Token is required")

	_, err = uc.Verify("bukan-token")
	assertUCError(t, err, http.StatusUnauthorized, "Invalid token: Invalid Compact JWS")

	wrong := sign(t, "secret-lain", jwt.SigningMethodHS256, jwt.MapClaims{"nim": "1", "nama": "A"})
	_, err = uc.Verify(wrong)
	assertUCError(t, err, http.StatusUnauthorized, "Invalid token: signature verification failed")

	expired := sign(t, siakadSecret, jwt.SigningMethodHS256, jwt.MapClaims{"nim": "1", "nama": "A", "exp": time.Now().Add(-time.Minute).Unix()})
	_, err = uc.Verify(expired)
	assertUCError(t, err, http.StatusUnauthorized, `Invalid token: "exp" claim timestamp check failed`)

	legacyHS384 := sign(t, legacySecret, jwt.SigningMethodHS384, jwt.MapClaims{"nim": "1", "nama": "A"})
	_, err = uc.Verify(legacyHS384)
	assert.Error(t, err, "secret legacy hanya HS256")

	onlySupabase := newAuth(&config.ElearningConfig{SupabaseJWTSecret: supabaseSecret}, nil)
	_, err = onlySupabase.Verify(wrong)
	assertUCError(t, err, http.StatusUnauthorized, "Invalid token: Konfigurasi verifikasi token belum tersedia.")
}

func TestVerifyMissingFields(t *testing.T) {
	uc := newAuth(testConfig(), nil)
	tok := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"role": "admin", "uid": 1, "nim": ""})
	_, err := uc.Verify(tok)
	detail := assertUCError(t, err, http.StatusBadRequest, "Token missing required fields (nim/username, nama/full_name)")
	assert.Equal(t, map[string]any{"receivedKeys": []string{"nim", "role", "uid"}}, detail)
}

func TestVerifyWithoutSupabaseSecretReturnsOriginal(t *testing.T) {
	uc := newAuth(&config.ElearningConfig{LegacyJWTSecret: legacySecret}, nil)
	tok := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"nim": "202711001", "role": "superadmin"})
	res, err := uc.Verify(tok)
	require.NoError(t, err)
	assert.Equal(t, tok, res.Token)
	assert.Equal(t, usecase.RoleKordas, res.Payload.Role)
}

func TestResignUsesDetectedBase64Secret(t *testing.T) {
	key := []byte("kunci-biner-\x00\x01\x02-uji")
	encoded := base64.StdEncoding.EncodeToString(key)
	anon := sign(t, string(key), jwt.SigningMethodHS256, jwt.MapClaims{"role": "anon"})
	uc := newAuth(&config.ElearningConfig{SupabaseJWTSecret: encoded, SupabaseAnonKey: anon, LegacyJWTSecret: legacySecret}, nil)

	tok := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"nim": "202711001", "nama": "Budi"})
	res, err := uc.Verify(tok)
	require.NoError(t, err)
	resignedClaims(t, res.Token, key)
	_, err = supabasejwt.VerifyHMAC(res.Token, []byte(encoded))
	assert.Error(t, err, "secret mentah tidak boleh dipakai bila anon key membuktikan base64")

	elearning := sign(t, string(key), jwt.SigningMethodHS256, jwt.MapClaims{"nim": "202711002", "nama": "Sari"})
	_, err = uc.Verify(elearning)
	assert.NoError(t, err, "token elearning diverifikasi dengan kunci hasil deteksi")
}

func TestReceiveToken(t *testing.T) {
	uc := newAuth(testConfig(), nil)

	_, err := uc.ReceiveToken("")
	assertUCError(t, err, http.StatusBadRequest, `Field "token" wajib disertakan di request body.`)

	_, err = uc.ReceiveToken("x.y.z")
	assertUCError(t, err, http.StatusUnauthorized, "Token tidak valid: Invalid Compact JWS")

	goToken := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"uid": 7, "nim": "202711003", "role": "mahasiswa"})
	_, err = uc.ReceiveToken(goToken)
	detail := assertUCError(t, err, http.StatusBadRequest, "Token tidak memiliki field yang diperlukan (nim/username, nama/full_name).")
	assert.Equal(t, map[string]any{"receivedKeys": []string{"nim", "role", "uid"}}, detail)

	tok := sign(t, siakadSecret, jwt.SigningMethodHS256, jwt.MapClaims{"username": "202711001", "full_name": "Budi", "role": "Laboran", "shift": "", "major": "TI", "email": "b@x.id"})
	res, err := uc.ReceiveToken(tok)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "/?token="+res.Token, res.RedirectURL)
	assert.Nil(t, res.Payload.Shift, "shift kosong jadi null")
	assert.Equal(t, "TI", res.Payload.Jurusan)
	assert.Equal(t, "b@x.id", res.Payload.Email)
	assert.Equal(t, usecase.RoleAsisten, res.Payload.Role)
	c := resignedClaims(t, res.Token, []byte(supabaseSecret))
	assert.NotContains(t, c, "uid")
	assert.Contains(t, c, "shift")
}

type fakeLogin struct {
	res             *usecase.PraktikumLoginResult
	err             error
	calls           int
	ident, pass, ip string
	userAgent       string
}

func (f *fakeLogin) Login(_ context.Context, identifier, password, ip, userAgent string) (*usecase.PraktikumLoginResult, error) {
	f.calls++
	f.ident, f.pass, f.ip, f.userAgent = identifier, password, ip, userAgent
	return f.res, f.err
}

func TestLoginSuccess(t *testing.T) {
	shift := 2
	fake := &fakeLogin{res: &usecase.PraktikumLoginResult{Token: "token-backend", NIM: "202711001", Nama: "Budi", NamaKelas: "TI-1A", Shift: &shift, Role: "mahasiswa"}}
	uc := newAuth(testConfig(), fake)

	res, err := uc.Login(context.Background(), dto.LoginRequest{Identifier: "  202711001 ", Password: " pw "}, "203.0.113.7", "uji-agent")
	require.NoError(t, err)
	assert.Equal(t, "202711001", fake.ident, "identifier di-trim")
	assert.Equal(t, " pw ", fake.pass, "password tidak di-trim")
	assert.Equal(t, "203.0.113.7", fake.ip)
	assert.Equal(t, "uji-agent", fake.userAgent)

	assert.Equal(t, "token-backend", res.BackendToken)
	assert.Equal(t, usecase.RolePraktikan, res.Payload.Role)
	require.NotNil(t, res.Payload.Shift)
	assert.Equal(t, "2", *res.Payload.Shift)
	assert.Nil(t, res.Payload.Jurusan)
	assert.Nil(t, res.Payload.Email)
	namaKelas := "TI-1A"
	assert.Equal(t, dto.LoginUser{NIM: "202711001", Nama: "Budi", Kelas: "TI-1A", Role: "praktikan", NamaKelas: &namaKelas, Shift: &shift}, res.User)

	c := resignedClaims(t, res.Token, []byte(supabaseSecret))
	assert.Equal(t, "praktikan", c["user_role"])
	assert.Equal(t, "TI-1A", c["kelas"])
	assert.Contains(t, c, "jurusan")
	assert.Nil(t, c["jurusan"])
}

func TestLoginFallbacksAndRole(t *testing.T) {
	fake := &fakeLogin{res: &usecase.PraktikumLoginResult{Token: "token-backend", Role: "koordinator"}}
	uc := newAuth(&config.ElearningConfig{}, fake)

	res, err := uc.Login(context.Background(), dto.LoginRequest{Identifier: "a@b.id", Password: "pw"}, "", "")
	require.NoError(t, err)
	assert.Equal(t, "token-backend", res.Token, "tanpa secret Supabase token backend dipakai")
	assert.Equal(t, "a@b.id", res.User.NIM)
	assert.Equal(t, "a@b.id", res.User.Nama)
	assert.Equal(t, usecase.RoleKordas, res.User.Role)
	assert.Nil(t, res.User.NamaKelas)
	assert.Nil(t, res.User.Shift)
	assert.Nil(t, res.Payload.Shift)
}

func TestLoginFailures(t *testing.T) {
	fake := &fakeLogin{err: errors.New("akun tidak ada")}
	uc := newAuth(testConfig(), fake)

	_, err := uc.Login(context.Background(), dto.LoginRequest{Identifier: "   ", Password: "pw"}, "", "")
	assertUCError(t, err, http.StatusBadRequest, "NIM/email dan password wajib diisi.")
	_, err = uc.Login(context.Background(), dto.LoginRequest{Identifier: "202711001"}, "", "")
	assertUCError(t, err, http.StatusBadRequest, "NIM/email dan password wajib diisi.")
	assert.Equal(t, 0, fake.calls)

	_, err = uc.Login(context.Background(), dto.LoginRequest{Identifier: "202711001", Password: "salah"}, "", "")
	assertUCError(t, err, http.StatusUnauthorized, "NIM/email atau password salah.")
	assert.Equal(t, 1, fake.calls)
}

func TestAuthenticate(t *testing.T) {
	uc := newAuth(testConfig(), nil)

	tok := sign(t, supabaseSecret, jwt.SigningMethodHS256, jwt.MapClaims{
		"nim": "202711001", "nama": "Budi", "kelas": "TI-1A", "shift": "1", "jurusan": "TI", "email": "b@x.id",
		"role": "authenticated", "user_role": "kordas", "sub": "202711001",
	})
	p, err := uc.Authenticate(tok)
	require.NoError(t, err)
	assert.Equal(t, "202711001", p.NIM)
	assert.Equal(t, "202711001", p.AuthNIM)
	assert.Equal(t, "Budi", p.Nama)
	assert.Equal(t, "TI-1A", p.Kelas)
	assert.Equal(t, usecase.RoleKordas, p.Role)
	require.NotNil(t, p.Shift)
	require.NotNil(t, p.Jurusan)
	require.NotNil(t, p.Email)
	assert.Equal(t, "1", *p.Shift)
	assert.Equal(t, "TI", *p.Jurusan)
	assert.Equal(t, "b@x.id", *p.Email)
	assert.Equal(t, "202711001", p.RawClaims["sub"])

	siakad := sign(t, siakadSecret, jwt.SigningMethodHS256, jwt.MapClaims{"username": "202711002", "sub": "uuid-siakad", "role": "mahasiswa"})
	p, err = uc.Authenticate(siakad)
	require.NoError(t, err)
	assert.Equal(t, "202711002", p.NIM, "nim || username || sub")
	assert.Equal(t, "uuid-siakad", p.AuthNIM, "auth_nim(): COALESCE(nim, sub, username)")
	assert.Equal(t, "202711002", p.Nama)
	assert.Equal(t, usecase.RolePraktikan, p.Role)

	anon := sign(t, supabaseSecret, jwt.SigningMethodHS256, jwt.MapClaims{"role": "anon", "iss": "supabase"})
	_, err = uc.Authenticate(anon)
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")

	_, err = uc.Authenticate("rusak")
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")
}

func TestSupabaseToken(t *testing.T) {
	jurusan := "TI"
	p := &dto.Principal{NIM: "202711001", Nama: "Budi", Kelas: "TI-1A", Jurusan: &jurusan, Role: usecase.RoleAsisten}

	res, err := newAuth(testConfig(), nil).SupabaseToken("asli", p)
	require.NoError(t, err)
	assert.Equal(t, 86400, res.ExpiresIn)
	c := resignedClaims(t, res.Token, []byte(supabaseSecret))
	assert.Equal(t, "202711001", c["sub"])
	assert.Equal(t, "asisten", c["user_role"])
	assert.Equal(t, "TI", c["jurusan"])
	assert.Nil(t, c["email"])

	res, err = newAuth(&config.ElearningConfig{}, nil).SupabaseToken("asli", p)
	require.NoError(t, err)
	assert.Equal(t, &dto.SupabaseTokenResponse{Token: "asli"}, res)
}
