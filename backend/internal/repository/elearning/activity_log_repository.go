package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type ActivityLogRepository interface {
	Insert(ctx context.Context, log *entity.ActivityLog) error
}

type activityLogRepository struct{ db *gorm.DB }

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

// Kolom persis monitoringService.addAuditLog; id dan ip_address memakai default DB.
func (r *activityLogRepository) Insert(ctx context.Context, log *entity.ActivityLog) error {
	return r.db.WithContext(ctx).Exec(`INSERT INTO activity_logs (nim, nama, event_type, "timestamp", details) VALUES (?, ?, ?, ?, ?)`,
		log.NIM, log.Nama, log.EventType, log.Timestamp, log.Details).Error
}
