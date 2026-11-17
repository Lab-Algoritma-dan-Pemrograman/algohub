package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type ReportRepository interface {
	Upsert(ctx context.Context, p *entity.ElearningProgress) error
}

type reportRepository struct{ db *gorm.DB }

func NewReportRepository(db *gorm.DB) ReportRepository { return &reportRepository{db: db} }

// Kolom ditulis eksplisit karena skema elearning_progress di prod belum pasti (T-24); kolom lain tidak disentuh.
func (r *reportRepository) Upsert(ctx context.Context, p *entity.ElearningProgress) error {
	return r.db.WithContext(ctx).Exec(`INSERT INTO elearning_progress
(nim, student_name, lessons_completed, total_lessons, completion_percentage, is_completed, completed_levels, current_level, last_accessed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (nim) DO UPDATE SET
student_name = EXCLUDED.student_name,
lessons_completed = EXCLUDED.lessons_completed,
total_lessons = EXCLUDED.total_lessons,
completion_percentage = EXCLUDED.completion_percentage,
is_completed = EXCLUDED.is_completed,
completed_levels = EXCLUDED.completed_levels,
current_level = EXCLUDED.current_level,
last_accessed_at = EXCLUDED.last_accessed_at`,
		p.NIM, p.StudentName, p.LessonsCompleted, p.TotalLessons, p.CompletionPercentage, p.IsCompleted, p.CompletedLevels, p.CurrentLevel, p.LastAccessedAt).Error
}
