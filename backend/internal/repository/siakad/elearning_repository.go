package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/siakad"
)

const (
	elearningLevelsSQL    = "SELECT to_json(l) FROM public.levels AS l"
	elearningProgressSQL  = "SELECT to_json(p) FROM public.student_progress AS p WHERE p.nim = ?"
	elearningSessionSQL   = "SELECT to_json(s) FROM public.active_sessions AS s WHERE s.nim = ? LIMIT 2"
	elearningModulesSQL   = "SELECT json_build_object('id', id, 'level_id', level_id) FROM public.modules"
	elearningLessonsSQL   = "SELECT json_build_object('id', id, 'module_id', module_id) FROM public.lessons"
	elearningOverridesSQL = "SELECT to_json(level_access_overrides) FROM public.users WHERE nim = ? LIMIT 2"
)

// Hanya membaca DB elearning; hasil sinkronisasi ditulis ke DB siakad lewat ProgressRepository.
type ElearningRepository interface {
	Levels(ctx context.Context) ([]entity.ElearningLevel, error)
	StudentProgress(ctx context.Context, nim string) ([]entity.ElearningStudentProgress, error)
	ActiveSession(ctx context.Context, nim string) (*entity.ElearningSession, error)
	Modules(ctx context.Context) ([]entity.ElearningModule, error)
	Lessons(ctx context.Context) ([]entity.ElearningLesson, error)
	LevelOverrides(ctx context.Context, nim string) (json.RawMessage, error)
}

type elearningRepository struct {
	db *gorm.DB
}

func NewElearningRepository(db *gorm.DB) ElearningRepository {
	return &elearningRepository{db: db}
}

func (r *elearningRepository) Levels(ctx context.Context) ([]entity.ElearningLevel, error) {
	return queryJSON[entity.ElearningLevel](ctx, r.db, elearningLevelsSQL)
}

func (r *elearningRepository) StudentProgress(ctx context.Context, nim string) ([]entity.ElearningStudentProgress, error) {
	return queryJSON[entity.ElearningStudentProgress](ctx, r.db, elearningProgressSQL, nim)
}

// Meniru .maybeSingle(): nil kalau baris tidak ada atau lebih dari satu.
func (r *elearningRepository) ActiveSession(ctx context.Context, nim string) (*entity.ElearningSession, error) {
	rows, err := queryJSON[entity.ElearningSession](ctx, r.db, elearningSessionSQL, nim)
	if err != nil || len(rows) != 1 {
		return nil, err
	}
	return &rows[0], nil
}

func (r *elearningRepository) Modules(ctx context.Context) ([]entity.ElearningModule, error) {
	return queryJSON[entity.ElearningModule](ctx, r.db, elearningModulesSQL)
}

func (r *elearningRepository) Lessons(ctx context.Context) ([]entity.ElearningLesson, error) {
	return queryJSON[entity.ElearningLesson](ctx, r.db, elearningLessonsSQL)
}

// Meniru .maybeSingle(): nil kalau baris tidak ada atau lebih dari satu.
func (r *elearningRepository) LevelOverrides(ctx context.Context, nim string) (json.RawMessage, error) {
	rows, err := queryJSON[json.RawMessage](ctx, r.db, elearningOverridesSQL, nim)
	if err != nil || len(rows) != 1 {
		return nil, err
	}
	return rows[0], nil
}

func queryJSON[T any](ctx context.Context, db *gorm.DB, query string, args ...any) ([]T, error) {
	rows, err := db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []T
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if raw == nil {
			raw = []byte("null")
		}
		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
