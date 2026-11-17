package usecase_test

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func siakadToken(t *testing.T) string {
	return sign(t, siakadSecret, jwt.SigningMethodHS256, jwt.MapClaims{"username": "202711001", "full_name": "Budi", "class_code": "TI-1A", "role": "asisten", "shift": 1})
}

func TestSSOCodeSingleUse(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)}
	uc := usecase.NewAuthUsecase(testConfig(), nil, clock.Now)

	code, err := uc.SSO(siakadToken(t))
	require.NoError(t, err)
	assert.NotContains(t, code, "eyJ", "token tidak boleh muncul di code")

	res, err := uc.ReceiveCode(code)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "202711001", res.Payload.NIM)
	assert.Equal(t, "Budi", res.Payload.Nama)
	assert.Equal(t, "TI-1A", res.Payload.Kelas)
	assert.Equal(t, usecase.RoleAsisten, res.Payload.Role)
	assert.Nil(t, res.Payload.Shift, "claim SSO mengikuti api/index.ts tanpa shift")
	c := resignedClaims(t, res.Token, []byte(supabaseSecret))
	assert.Equal(t, "asisten", c["user_role"])
	assert.Equal(t, "202711001", c["sub"])
	assert.NotContains(t, c, "shift")
	assert.NotContains(t, c, "uid")

	_, err = uc.ReceiveCode(code)
	assertUCError(t, err, http.StatusUnauthorized, "Kode SSO tidak valid atau kedaluwarsa.")
}

func TestSSOCodeExpiresAfter60Seconds(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)}
	uc := usecase.NewAuthUsecase(testConfig(), nil, clock.Now)

	onTime, err := uc.SSO(siakadToken(t))
	require.NoError(t, err)
	late, err := uc.SSO(siakadToken(t))
	require.NoError(t, err)

	clock.Advance(60 * time.Second)
	_, err = uc.ReceiveCode(onTime)
	assert.NoError(t, err, "tepat 60 detik masih berlaku")

	clock.Advance(time.Second)
	_, err = uc.ReceiveCode(late)
	assertUCError(t, err, http.StatusUnauthorized, "Kode SSO tidak valid atau kedaluwarsa.")
}

func TestSSORejects(t *testing.T) {
	uc := usecase.NewAuthUsecase(testConfig(), nil, time.Now)

	_, err := uc.SSO("rusak")
	assert.ErrorIs(t, err, usecase.ErrSSOInvalidToken)

	noName := sign(t, legacySecret, jwt.SigningMethodHS256, jwt.MapClaims{"uid": 1, "nim": "202711001"})
	_, err = uc.SSO(noName)
	assert.ErrorIs(t, err, usecase.ErrSSOMissingFields)

	code, err := uc.SSO(siakadToken(t))
	require.NoError(t, err)
	tampered := []byte(code)
	tampered[len(tampered)/2] ^= 1
	_, err = uc.ReceiveCode(string(tampered))
	assert.Error(t, err)

	other := usecase.NewAuthUsecase(&config.ElearningConfig{SupabaseJWTSecret: "secret-lain", SiakadJWTSecret: siakadSecret}, nil, time.Now)
	_, err = other.ReceiveCode(code)
	assert.Error(t, err, "code dari secret lain ditolak")

	noSecret := usecase.NewAuthUsecase(&config.ElearningConfig{SiakadJWTSecret: siakadSecret}, nil, time.Now)
	_, err = noSecret.SSO(siakadToken(t))
	require.Error(t, err)
	assert.NotErrorIs(t, err, usecase.ErrSSOInvalidToken)
	assert.NotErrorIs(t, err, usecase.ErrSSOMissingFields)
}

func TestSSOConcurrentExchangeOnlyOnce(t *testing.T) {
	uc := usecase.NewAuthUsecase(testConfig(), nil, time.Now)
	code, err := uc.SSO(siakadToken(t))
	require.NoError(t, err)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
		ok int
	)
	for range 8 {
		wg.Go(func() {
			if _, err := uc.ReceiveCode(code); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	assert.Equal(t, 1, ok)
}
