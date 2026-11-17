package usecase

import (
	"cmp"
	"context"
	"net/http"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

const ssoTokenTTL = 2 * time.Hour

type SSOUsecase struct {
	sso    repository.SSORepository
	tokens *Tokens
	now    func() time.Time
}

func NewSSOUsecase(sso repository.SSORepository, tokens *Tokens) *SSOUsecase {
	return &SSOUsecase{sso: sso, tokens: tokens, now: time.Now}
}

// Endpoint tanpa sesi: handshake_code sekali pakai yang dibuat siakad adalah satu-satunya bukti identitas.
func (u *SSOUsecase) Generate(ctx context.Context, code string) (map[string]any, *Failure) {
	if code == "" {
		return nil, fail(http.StatusUnauthorized, "Missing handshake code")
	}
	userID, found, err := u.sso.HandshakeUserID(ctx, code, u.now().UTC())
	if err != nil || !found {
		return nil, fail(http.StatusUnauthorized, "Sesi verifikasi kadaluarsa. Harap coba lagi.")
	}
	notFound := fail(http.StatusNotFound, "Data pengguna tidak ditemukan di database")
	if userID == nil {
		return nil, notFound
	}
	user, found, err := u.sso.User(ctx, *userID)
	if err != nil || !found {
		return nil, notFound
	}
	_ = u.sso.DeleteHandshake(ctx, code)

	nim := cmp.Or(user.NIM, user.Username)
	nama := cmp.Or(user.FullName, user.Nama)
	if nim == "" || nama == "" {
		return nil, fail(http.StatusBadRequest, "Data pengguna tidak lengkap di database")
	}

	var sched entity.ScheduleClass
	// T-22: role dibandingkan mentah seperti kode lama, jadi kordas/admin tidak mendapat kelas dari jadwal.
	switch user.Role {
	case "mahasiswa", "praktikan":
		sched, _ = u.sso.StudentSchedule(ctx, *userID)
	case "asisten", "koordinator":
		sched, _ = u.sso.AssistantSchedule(ctx, *userID)
	}

	token, err := u.tokens.Sign(map[string]any{
		"nim":     nim,
		"nama":    nama,
		"kelas":   cmp.Or(sched.ClassCode, user.ClassCode),
		"jurusan": cmp.Or(sched.Major, user.Major),
		"role":    user.Role,
	}, ssoTokenTTL)
	if err != nil {
		return nil, fail(http.StatusInternalServerError, "Internal server error")
	}
	return map[string]any{"token": token}, nil
}
