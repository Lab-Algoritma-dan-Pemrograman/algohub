package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/hash"
)

const sessionTTL = time.Hour

// Klaim minimal tanpa PII sensitif; key yang tidak ada di baris login_user tidak ikut ditandatangani.
var sessionClaimKeys = []string{"id", "username", "role", "division", "nim", "assistant_code", "full_name"}

type AuthUsecase struct {
	auth   repository.AuthRepository
	tokens *Tokens
}

func NewAuthUsecase(auth repository.AuthRepository, tokens *Tokens) *AuthUsecase {
	return &AuthUsecase{auth: auth, tokens: tokens}
}

func (u *AuthUsecase) Login(ctx context.Context, username, password string) (body map[string]any, token string, f *Failure) {
	if username == "" || password == "" {
		return nil, "", fail(http.StatusBadRequest, "Username dan password wajib diisi.")
	}

	raw, err := u.auth.LoginUser(ctx, username, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return nil, "", fail(http.StatusInternalServerError, "Internal server error")
		}
		msg := pgErr.Message
		if systemErrorRe.MatchString(msg) {
			msg = "Terjadi kesalahan sistem. Silakan coba lagi."
		}
		return nil, "", fail(http.StatusBadRequest, msg)
	}
	if raw == nil {
		return nil, "", fail(http.StatusUnauthorized, "Username tidak ditemukan atau password salah!")
	}
	user, err := decodeObject(raw)
	if err != nil {
		return nil, "", fail(http.StatusInternalServerError, "Internal server error")
	}

	// Postgres tidak punya bcrypt native, jadi hash dicocokkan di server lalu dibuang dari respons.
	passwordHash, _ := user["password"].(string)
	delete(user, "password")
	if passwordHash == "" || !hash.Verify(passwordHash, password) {
		return nil, "", fail(http.StatusUnauthorized, "Username tidak ditemukan atau password salah!")
	}
	if !truthy(user["is_active"]) {
		return nil, "", fail(http.StatusForbidden, "Akun Anda sedang dinonaktifkan.")
	}
	if f := u.checkShift(ctx, user); f != nil {
		return nil, "", f
	}

	claims := map[string]any{}
	for _, k := range sessionClaimKeys {
		if v, ok := user[k]; ok {
			claims[k] = v
		}
	}
	// jti dipakai revoke_token saat logout.
	claims["jti"] = uuid.NewString()
	token, err = u.tokens.Sign(claims, sessionTTL)
	if err != nil {
		return nil, "", fail(http.StatusInternalServerError, "Internal server error")
	}
	return map[string]any{"user": user, "token": token}, token, nil
}

// Shift aktif hanya membatasi praktikan yang shift-nya sudah terisi.
func (u *AuthUsecase) checkShift(ctx context.Context, user map[string]any) *Failure {
	shift, hasShift := user["shift"]
	if role, _ := user["role"].(string); role != "mahasiswa" || !hasShift || shift == nil {
		return nil
	}
	active, ok := u.auth.ActiveShift(ctx)
	if !ok || active == "" {
		active = "all"
	}
	if active == "all" {
		return nil
	}
	if active == "none" {
		return failDetails(http.StatusForbidden, "Akses Ditutup", "Maaf, akses login saat ini sedang ditutup untuk semua praktikan.")
	}
	s, isText := shift.(string)
	if !truthy(shift) || s == "?" {
		return failDetails(http.StatusForbidden, "Akun Belum Aktif", "Mohon tunggu plotting shift dari Koordinator sebelum Anda bisa login.")
	}
	if !isText || s != active {
		return failDetails(http.StatusForbidden, "Akses Ditolak",
			fmt.Sprintf("Maaf, akun Anda terdaftar di Shift %s. Saat ini hanya Shift %s yang diizinkan masuk.", claimText(shift), active))
	}
	return nil
}

// C-2: token dimasukkan ke blocklist sampai kedaluwarsa. Semua error diabaikan karena
// cookie tetap dihapus dan token yang sudah invalid memang tidak perlu diblokir.
func (u *AuthUsecase) Logout(ctx context.Context, token string) {
	if token == "" {
		return
	}
	claims, err := u.tokens.Verify(token)
	if err != nil || !truthy(claims["jti"]) || !truthy(claims["exp"]) || !truthy(claims["id"]) {
		return
	}
	exp, err := claimFloat(claims["exp"])
	if err != nil {
		return
	}
	expiresAt := time.UnixMilli(int64(exp * 1000)).UTC()
	_ = u.auth.RevokeToken(ctx, claimText(claims["jti"]), claimText(claims["id"]), expiresAt)
}

// clearCookie true berarti cookie sesi harus dihapus karena token di dalamnya tidak bisa dipakai lagi.
func (u *AuthUsecase) Verify(ctx context.Context, token string) (body map[string]any, clearCookie bool, f *Failure) {
	if token == "" {
		return nil, false, fail(http.StatusUnauthorized, "Unauthorized: Missing token")
	}
	claims, err := u.tokens.Verify(token)
	if err != nil {
		return nil, true, fail(http.StatusUnauthorized, "Unauthorized: Invalid or expired token")
	}
	if isRevoked(ctx, u.auth, claims) {
		return nil, true, fail(http.StatusUnauthorized, "Unauthorized: Session telah berakhir. Silakan login kembali.")
	}
	delete(claims, "iat")
	delete(claims, "exp")
	delete(claims, "jti")
	return map[string]any{"user": claims}, false, nil
}
