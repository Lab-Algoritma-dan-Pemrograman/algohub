package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type PrestigeRepository interface {
	CountSince(ctx context.Context, nim string, since time.Time) (int64, error)
	Reset(ctx context.Context, nim string) (string, error)
}

type prestigeRepository struct{ db *gorm.DB }

func NewPrestigeRepository(db *gorm.DB) PrestigeRepository { return &prestigeRepository{db: db} }

func (r *prestigeRepository) CountSince(ctx context.Context, nim string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.PrestigeHistory{}).Where("nim = ? AND reset_at >= ?", nim, since).Count(&n).Error
	return n, err
}

// RPC dipanggil apa adanya: validasi 6/6, cooldown, kuota, arsip, dan hapus progres berjalan atomik di do_prestige_reset.
// Hasil jsonb dibaca sebagai teks lalu di-parse di usecase seperti `data as any` di prestigeService.ts.
func (r *prestigeRepository) Reset(ctx context.Context, nim string) (string, error) {
	var row struct{ Result *string }
	if err := r.db.WithContext(ctx).Raw("SELECT public.do_prestige_reset(?)::text AS result", nim).Scan(&row).Error; err != nil {
		return "", err
	}
	if row.Result == nil {
		return "null", nil
	}
	return *row.Result, nil
}
