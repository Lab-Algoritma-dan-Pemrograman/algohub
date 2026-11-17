package repository

import (
	"context"

	"gorm.io/gorm"
)

// Query ringan supaya project Supabase Free tidak di-pause karena tidak aktif.
type Pinger interface {
	Ping(ctx context.Context) error
}

type pinger struct {
	db  *gorm.DB
	sql string
}

func NewSiakadPinger(db *gorm.DB) Pinger {
	return &pinger{db: db, sql: "SELECT public.get_public_settings()"}
}

func NewElearningPinger(db *gorm.DB) Pinger {
	return &pinger{db: db, sql: "SELECT id FROM lessons LIMIT 1"}
}

func (p *pinger) Ping(ctx context.Context) error {
	return p.db.WithContext(ctx).Exec(p.sql).Error
}
