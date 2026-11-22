package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/hash"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// fakeBlocklist: blocklist jti di memori, mencatat revoke + memutuskan cek.
// Ini menggantikan DB praktikum (tabel revoked_tokens) di unit test.
type fakeBlocklist struct {
	revoked map[string]bool
	expires map[string]time.Time
	failOn  string // bila jti == failOn, lookup mengembalikan error (fail-open)
}

func newFakeBlocklist() *fakeBlocklist {
	return &fakeBlocklist{revoked: map[string]bool{}, expires: map[string]time.Time{}}
}

// fakeUserRepo: cukup FindByID untuk memenuhi resolve-role di legacyAuth.
// Metode lain dipinjam dari interface (nil) karena tak dipakai di path auth test.
type fakeUserRepo struct {
	repository.UserRepository
	user *entity.User
}

func (f *fakeUserRepo) FindByID(int) (*entity.User, error) { return f.user, nil }

func (f *fakeBlocklist) IsTokenRevoked(_ context.Context, jti string) (bool, error) {
	if jti == f.failOn {
		return false, errors.New("db down")
	}
	return f.revoked[jti], nil
}

func (f *fakeBlocklist) RevokeToken(_ context.Context, jti string, userID int, expiresAt time.Time) error {
	f.revoked[jti] = true
	f.expires[jti] = expiresAt
	return nil
}

func TestK1_Logout_MencabutJti_TokenSudahTidakDiterima(t *testing.T) {
	jm := jwt.NewManager("rahasia-uji", 12)
	bl := newFakeBlocklist()
	uc := usecase.NewAuthUsecase(nil, nil, jm, &config.Config{}, hash.FbScryptConfig{}).WithBlocklist(bl)

	token, err := jm.Generate(7, "123456", "mahasiswa")
	assert.NoError(t, err)

	claims, revoked, err := uc.Verify(token)
	assert.NoError(t, err)
	assert.False(t, revoked, "token segar belum dicabut")
	assert.NotEmpty(t, claims.ID, "klaim jti harus ada")

	// Middleware path: legacyAuth memanggil blocklist sebelum izin lewat.
	exp, ok := jwt.ExpiryTime(claims)
	assert.True(t, ok)

	uc.Logout(7, claims.ID, exp)

	// Bukti revoke tersimpan.
	assert.True(t, bl.revoked[claims.ID], "Logout harus memasukkan jti ke blocklist")

	_, revoked, err = uc.Verify(token)
	assert.NoError(t, err)
	assert.True(t, revoked, "token yang sudah logout harus dilaporkan revoked")
}

func TestK1_MiddlewareLegacyAuth_MenolakTokenRevoked(t *testing.T) {
	jm := jwt.NewManager("rahasia-uji", 12)
	bl := newFakeBlocklist()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.PraktikumAuth(&config.Config{AuthMode: "legacy"}, jm, &fakeUserRepo{user: &entity.User{ID: 1, Role: entity.RoleMahasiswa, NIM: "123"}}, bl))
	r.GET("/uji", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"jti": middleware.TokenID(c)})
	})

	token, err := jm.Generate(1, "123", "mahasiswa")
	assert.NoError(t, err)
	claims, err := jm.Verify(token)
	assert.NoError(t, err)

	uc := usecase.NewAuthUsecase(nil, nil, jm, &config.Config{}, hash.FbScryptConfig{}).WithBlocklist(bl)
	exp, _ := jwt.ExpiryTime(claims)
	uc.Logout(1, claims.ID, exp)

	// Token yang sudah logout ditolak 401.
	req := httptest.NewRequest(http.MethodGet, "/uji", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "token revoked harus 401")
}

func TestK1_TokenLegacyTanpaJti_Diterima(t *testing.T) {
	jm := jwt.NewManager("rahasia-uji", 12)
	bl := newFakeBlocklist()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.PraktikumAuth(&config.Config{AuthMode: "legacy"}, jm, &fakeUserRepo{user: &entity.User{ID: 99, Role: entity.RoleMahasiswa, NIM: "999"}}, bl))
	r.GET("/uji", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Token lama: ditandatangani langsung TANPA jti (RegisteredClaims.ID kosong).
	now := time.Now()
	old := jwt.Claims{
		UserID: 99, NIM: "999", Role: "mahasiswa",
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Hour)),
		},
	}
	legacyToken, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, old).SignedString([]byte("rahasia-uji"))
	assert.NoError(t, err)

	uc := usecase.NewAuthUsecase(nil, nil, jm, &config.Config{}, hash.FbScryptConfig{}).WithBlocklist(bl)
	_, revoked, err := uc.Verify(legacyToken)
	assert.NoError(t, err)
	assert.False(t, revoked, "token legacy tanpa jti tidak boleh ditolak")

	req := httptest.NewRequest(http.MethodGet, "/uji", nil)
	req.Header.Set("Authorization", "Bearer "+legacyToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "token legacy tanpa jti tetap sah")
}

func TestK1_LogoutTanpaJti_AmanNoOp(t *testing.T) {
	bl := newFakeBlocklist()
	uc := usecase.NewAuthUsecase(nil, nil, jwt.NewManager("s", 12), &config.Config{}, hash.FbScryptConfig{}).WithBlocklist(bl)

	uc.Logout(5, "", time.Now().Add(time.Hour)) // token legacy, jti kosong
	assert.Empty(t, bl.revoked, "logout tanpa jti tak boleh mencabut apa pun")

	// Use case tanpa blocklist (pemanggil lama) tak boleh panic.
	usecase.NewAuthUsecase(nil, nil, jwt.NewManager("s", 12), &config.Config{}, hash.FbScryptConfig{}).Logout(5, "x", time.Now())
}

func TestK1_BlocklistError_FailOpen(t *testing.T) {
	jm := jwt.NewManager("rahasia-uji", 12)
	bl := newFakeBlocklist()
	bl.failOn = "gagal"
	uc := usecase.NewAuthUsecase(nil, nil, jm, &config.Config{}, hash.FbScryptConfig{}).WithBlocklist(bl)

	// jti != "gagal" di sini, jadi fail-open tak terpicu; paksa cek dengan jti gagal:
	bl.revoked["gagal"] = true
	tk, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.MapClaims{
		"uid": 1, "nim": "1", "role": "mahasiswa",
		"jti": "gagal",
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("rahasia-uji"))
	assert.NoError(t, err)

	_, revoked, err := uc.Verify(tk)
	assert.NoError(t, err)
	assert.False(t, revoked, "DB blocklist gagal: fail-open (dianggap belum dicabut), sama seperti keputusan siakad")
}

func TestK1_LogoutTanpaExp_PakaiUmurTokenConfig(t *testing.T) {
	bl := newFakeBlocklist()
	cfg := &config.Config{JWTExpireHours: 12}
	uc := usecase.NewAuthUsecase(nil, nil, jwt.NewManager("s", 12), cfg, hash.FbScryptConfig{}).WithBlocklist(bl)

	before := time.Now().Add(11 * time.Hour)
	uc.Logout(3, "jti-tanpa-exp", time.Time{})
	after := time.Now().Add(13 * time.Hour)

	got, ok := bl.expires["jti-tanpa-exp"]
	assert.True(t, ok, "revoke tetap tersimpan walau exp token tak dikirim")
	assert.True(t, got.After(before) && got.Before(after), "TTL fallback = umur token config (12 jam), bukan permanen")
}
