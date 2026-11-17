package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

// HasUser membedakan embed users(role) yang null (tidak ada baris users) dari role yang bernilai null.
type SessionRole struct {
	NIM           string     `gorm:"column:nim"`
	LastHeartbeat *time.Time `gorm:"column:last_heartbeat"`
	Role          *string    `gorm:"column:role"`
	HasUser       bool       `gorm:"column:has_user"`
}

type SessionRepository interface {
	Upsert(ctx context.Context, s *entity.ActiveSession, setAttempt bool) error
	DeleteByNIM(ctx context.Context, nim string) (int64, error)
	DeleteStale(ctx context.Context, before time.Time) (int64, error)
	ListWithRole(ctx context.Context, page, limit int) ([]SessionRole, int64, error)
}

type sessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) SessionRepository { return &sessionRepository{db: db} }

// attempt_id hanya ikut ditulis bila dikirim, karena heartbeat Layout.tsx tidak menyertakannya dan upsert PostgREST
// hanya memperbarui kolom yang ada di payload.
func (r *sessionRepository) Upsert(ctx context.Context, s *entity.ActiveSession, setAttempt bool) error {
	cols := []string{"nim", "nama", "kelas", "last_heartbeat", "current_activity"}
	args := []any{s.NIM, s.Nama, s.Kelas, s.LastHeartbeat, s.CurrentActivity}
	if setAttempt {
		cols = append(cols, "attempt_id")
		args = append(args, s.AttemptID)
	}
	sets := make([]string, 0, len(cols)-1)
	for _, c := range cols[1:] {
		sets = append(sets, c+" = EXCLUDED."+c)
	}
	sql := "INSERT INTO active_sessions (" + strings.Join(cols, ", ") + ") VALUES " + placeholders(len(cols)) +
		" ON CONFLICT (nim) DO UPDATE SET " + strings.Join(sets, ", ")
	return r.db.WithContext(ctx).Exec(sql, args...).Error
}

func (r *sessionRepository) DeleteByNIM(ctx context.Context, nim string) (int64, error) {
	res := r.db.WithContext(ctx).Where("nim = ?", nim).Delete(&entity.ActiveSession{})
	return res.RowsAffected, res.Error
}

func (r *sessionRepository) DeleteStale(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("last_heartbeat < ?", before).Delete(&entity.ActiveSession{})
	return res.RowsAffected, res.Error
}

// Satu join menggantikan embed users(role) PostgREST.
func (r *sessionRepository) ListWithRole(ctx context.Context, page, limit int) ([]SessionRole, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.ActiveSession{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []SessionRole
	err := r.db.WithContext(ctx).Raw(`SELECT s.nim, s.last_heartbeat, u.role, (u.nim IS NOT NULL) AS has_user
FROM active_sessions s
LEFT JOIN users u ON u.nim = s.nim
ORDER BY s.last_heartbeat DESC NULLS LAST, s.nim
LIMIT ? OFFSET ?`, limit, (page-1)*limit).Scan(&rows).Error
	return rows, total, err
}
