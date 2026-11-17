package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type ProgressRepository interface {
	CountCompleted(ctx context.Context, nim string) (int64, error)
	CompletedLessonIDs(ctx context.Context, nim string) ([]string, error)
	Exists(ctx context.Context, nim, lessonID string) (bool, error)
	Insert(ctx context.Context, nim, lessonID string, completedAt time.Time) error
	DeleteByNIM(ctx context.Context, nim string) (int64, error)
	DeleteByNIMLessons(ctx context.Context, nim string, lessonIDs []string) (int64, error)
	LessonIDsIn(ctx context.Context, nim string, lessonIDs []string) ([]string, error)
	ListByNIM(ctx context.Context, nim string, page, limit int) ([]entity.StudentProgress, int64, error)
	Recent(ctx context.Context, nim string, limit int) ([]entity.StudentProgress, error)
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

// Meniru .maybeSingle() di completeLesson: lebih dari satu baris dibaca sebagai data kosong, jadi pelajaran dianggap belum selesai.
func (r *progressRepository) Exists(ctx context.Context, nim, lessonID string) (bool, error) {
	var vals []*bool
	err := r.db.WithContext(ctx).Model(&entity.StudentProgress{}).Where("nim = ? AND lesson_id = ?", nim, lessonID).Limit(2).Pluck("completed", &vals).Error
	if err != nil {
		return false, err
	}
	return len(vals) == 1 && vals[0] != nil && *vals[0], nil
}

// Trigger check_progress_velocity tetap berjalan di DB dan bisa menolak dengan P0001.
func (r *progressRepository) Insert(ctx context.Context, nim, lessonID string, completedAt time.Time) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO student_progress (nim, lesson_id, completed, completed_at) VALUES (?, ?, true, ?)", nim, lessonID, completedAt).Error
}

func (r *progressRepository) DeleteByNIM(ctx context.Context, nim string) (int64, error) {
	res := r.db.WithContext(ctx).Where("nim = ?", nim).Delete(&entity.StudentProgress{})
	return res.RowsAffected, res.Error
}

func (r *progressRepository) DeleteByNIMLessons(ctx context.Context, nim string, lessonIDs []string) (int64, error) {
	if len(lessonIDs) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Where("nim = ? AND lesson_id IN ?", nim, lessonIDs).Delete(&entity.StudentProgress{})
	return res.RowsAffected, res.Error
}

// Tanpa filter completed, sama dengan resetLevelProgress (progressService.ts:311).
func (r *progressRepository) LessonIDsIn(ctx context.Context, nim string, lessonIDs []string) ([]string, error) {
	ids := []string{}
	if len(lessonIDs) == 0 {
		return ids, nil
	}
	err := r.db.WithContext(ctx).Model(&entity.StudentProgress{}).Where("nim = ? AND lesson_id IN ?", nim, lessonIDs).Pluck("lesson_id", &ids).Error
	return ids, err
}

// Urutan completed_at desc dari AdminDashboard.tsx:1271; id menjadi pemecah seri supaya halaman stabil.
func (r *progressRepository) ListByNIM(ctx context.Context, nim string, page, limit int) ([]entity.StudentProgress, int64, error) {
	filter := func(db *gorm.DB) *gorm.DB { return db.Where("nim = ? AND completed = ?", nim, true) }
	return findPage[entity.StudentProgress](ctx, filter, r.db, "completed_at DESC, id", page, limit)
}

func (r *progressRepository) Recent(ctx context.Context, nim string, limit int) ([]entity.StudentProgress, error) {
	var rows []entity.StudentProgress
	err := r.completed(ctx, nim).Order("completed_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
