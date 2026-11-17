package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type ProgressRepository interface {
	CountCompleted(ctx context.Context, nim string) (int64, error)
	CompletedLessonIDs(ctx context.Context, nim string) ([]string, error)
}

type progressRepository struct{ db *gorm.DB }

func NewProgressRepository(db *gorm.DB) ProgressRepository { return &progressRepository{db: db} }

func (r *progressRepository) completed(ctx context.Context, nim string) *gorm.DB {
	return r.db.WithContext(ctx).Model(&entity.StudentProgress{}).Where("nim = ? AND completed = ?", nim, true)
}

func (r *progressRepository) CountCompleted(ctx context.Context, nim string) (int64, error) {
	var n int64
	err := r.completed(ctx, nim).Count(&n).Error
	return n, err
}

func (r *progressRepository) CompletedLessonIDs(ctx context.Context, nim string) ([]string, error) {
	var ids []string
	err := r.completed(ctx, nim).Where("lesson_id IS NOT NULL").Pluck("lesson_id", &ids).Error
	return ids, err
}
