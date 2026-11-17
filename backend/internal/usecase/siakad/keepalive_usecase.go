package usecase

import (
	"context"
	"net/http"
	"time"

	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
)

type KeepAliveUsecase struct {
	pingers []repository.Pinger
	now     func() time.Time
}

func NewKeepAliveUsecase(siakad, elearning repository.Pinger) *KeepAliveUsecase {
	return &KeepAliveUsecase{pingers: []repository.Pinger{siakad, elearning}, now: time.Now}
}

// Semua DB tetap di-ping walau salah satu gagal, supaya project lain tidak ikut di-pause.
func (u *KeepAliveUsecase) Run(ctx context.Context) (map[string]any, *Failure) {
	var first error
	for _, p := range u.pingers {
		if err := p.Ping(ctx); err != nil && first == nil {
			first = err
		}
	}
	if first != nil {
		return nil, fail(http.StatusInternalServerError, errText(first))
	}
	return map[string]any{
		"success":   true,
		"message":   "Supabase pinged successfully",
		"timestamp": isoMillis(u.now()),
	}, nil
}
