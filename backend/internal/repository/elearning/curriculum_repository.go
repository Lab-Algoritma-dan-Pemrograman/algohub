package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type LevelLessons struct {
	ID        string
	Title     string
	LessonIDs []string
}

type CurriculumRepository interface {
	LessonQuiz(ctx context.Context, id string) (*string, error)
	CountLessons(ctx context.Context) (int64, error)
	LevelTree(ctx context.Context) ([]LevelLessons, error)
}

type curriculumRepository struct{ db *gorm.DB }

func NewCurriculumRepository(db *gorm.DB) CurriculumRepository {
	return &curriculumRepository{db: db}
}

// Dibaca sebagai teks supaya kolom TEXT berisi JSON maupun jsonb sama-sama di-parse di usecase seperti validate-quiz.ts.
func (r *curriculumRepository) LessonQuiz(ctx context.Context, id string) (*string, error) {
	var row struct{ Quiz *string }
	err := r.db.WithContext(ctx).Model(&entity.Lesson{}).Select("quiz::text AS quiz").Where("id = ?", id).Take(&row).Error
	if err != nil {
		return nil, err
	}
	return row.Quiz, nil
}

func (r *curriculumRepository) CountLessons(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Lesson{}).Count(&n).Error
	return n, err
}

// Satu join menggantikan embed levels(modules(lessons)) PostgREST.
func (r *curriculumRepository) LevelTree(ctx context.Context) ([]LevelLessons, error) {
	var rows []struct {
		LevelID  string
		Title    string
		LessonID *string
	}
	err := r.db.WithContext(ctx).Raw(`SELECT l.id AS level_id, COALESCE(l.title, '') AS title, ls.id AS lesson_id
FROM levels l
LEFT JOIN modules m ON m.level_id = l.id
LEFT JOIN lessons ls ON ls.module_id = m.id
ORDER BY l.sort_order, l.id, m.sort_order, m.id, ls.sort_order, ls.id`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	var out []LevelLessons
	for _, row := range rows {
		if len(out) == 0 || out[len(out)-1].ID != row.LevelID {
			out = append(out, LevelLessons{ID: row.LevelID, Title: row.Title})
		}
		if row.LessonID != nil {
			last := &out[len(out)-1]
			last.LessonIDs = append(last.LessonIDs, *row.LessonID)
		}
	}
	return out, nil
}
