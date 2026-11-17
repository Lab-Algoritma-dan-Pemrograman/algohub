package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	ListLevels(ctx context.Context, page, limit int) ([]entity.Level, int64, error)
	ListModules(ctx context.Context, page, limit int) ([]entity.Module, int64, error)
	ListLessons(ctx context.Context, page, limit int) ([]entity.Lesson, int64, error)
	ListStudentLessons(ctx context.Context, page, limit int) ([]entity.StudentLesson, int64, error)
	SaveFull(ctx context.Context, levels []entity.Level, modules []entity.Module, lessons []entity.Lesson) error
	ClearAll(ctx context.Context) error
	LessonIDsByLevel(ctx context.Context, levelID string) ([]string, error)
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

// Urutan .order('sort_order') di curriculumService.ts; id menjadi pemecah seri supaya halaman stabil.
const sortOrder = "sort_order, id"

func (r *curriculumRepository) ListLevels(ctx context.Context, page, limit int) ([]entity.Level, int64, error) {
	return findPage[entity.Level](ctx, noFilter, r.db, sortOrder, page, limit)
}

func (r *curriculumRepository) ListModules(ctx context.Context, page, limit int) ([]entity.Module, int64, error) {
	return findPage[entity.Module](ctx, noFilter, r.db, sortOrder, page, limit)
}

func (r *curriculumRepository) ListLessons(ctx context.Context, page, limit int) ([]entity.Lesson, int64, error) {
	return findPage[entity.Lesson](ctx, noFilter, r.db, sortOrder, page, limit)
}

func (r *curriculumRepository) ListStudentLessons(ctx context.Context, page, limit int) ([]entity.StudentLesson, int64, error) {
	return findPage[entity.StudentLesson](ctx, noFilter, r.db, sortOrder, page, limit)
}

// Kolom upsert persis saveFullCurriculum; kolom lain (xp_reward, exercise_type, ...) tidak disentuh saat konflik.
var (
	levelUpsertColumns  = []string{"id", "title", "description", "access_mode", "locked", "sort_order"}
	moduleUpsertColumns = []string{"id", "level_id", "title", "sort_order"}
	lessonUpsertColumns = []string{"id", "module_id", "title", "explanation", "code_example", "initial_code", "solution", "hint", "quiz", "test_cases", "validation_rules", "sort_order"}
)

// Dulu sembilan request terpisah (curriculumService.ts:160-259) tanpa transaksi; kini gagal di tengah tidak meninggalkan kurikulum setengah jadi.
func (r *curriculumRepository) SaveFull(ctx context.Context, levels []entity.Level, modules []entity.Module, lessons []entity.Lesson) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := deleteMissing(tx, "lessons", idsOf(lessons, func(l entity.Lesson) string { return l.ID })); err != nil {
			return err
		}
		if err := deleteMissing(tx, "modules", idsOf(modules, func(m entity.Module) string { return m.ID })); err != nil {
			return err
		}
		if err := deleteMissing(tx, "levels", idsOf(levels, func(l entity.Level) string { return l.ID })); err != nil {
			return err
		}
		if err := upsertByID(tx, levels, levelUpsertColumns); err != nil {
			return err
		}
		if err := upsertByID(tx, modules, moduleUpsertColumns); err != nil {
			return err
		}
		return upsertByID(tx, lessons, lessonUpsertColumns)
	})
}

func idsOf[T any](rows []T, id func(T) string) []string {
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = id(row)
	}
	return ids
}

// Nama tabel berupa literal dari pemanggil. Daftar kosong berarti semua baris dihapus, sama dengan filter !includes lama.
func deleteMissing(tx *gorm.DB, table string, keep []string) error {
	if len(keep) == 0 {
		return tx.Exec("DELETE FROM " + table).Error
	}
	return tx.Exec("DELETE FROM "+table+" WHERE id NOT IN ?", keep).Error
}

func upsertByID[T any](tx *gorm.DB, rows []T, cols []string) error {
	if len(rows) == 0 {
		return nil
	}
	return tx.Select(cols).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns(cols[1:]),
	}).Create(&rows).Error
}

// Modul dan pelajaran ikut terhapus lewat ON DELETE CASCADE, seperti komentar clearCurriculum.
func (r *curriculumRepository) ClearAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM levels").Error
}

// Level tanpa baris berarti tidak ada (gorm.ErrRecordNotFound); level tanpa pelajaran mengembalikan daftar kosong.
func (r *curriculumRepository) LessonIDsByLevel(ctx context.Context, levelID string) ([]string, error) {
	var rows []struct{ LessonID *string }
	err := r.db.WithContext(ctx).Raw(`SELECT ls.id AS lesson_id
FROM levels l
LEFT JOIN modules m ON m.level_id = l.id
LEFT JOIN lessons ls ON ls.module_id = m.id
WHERE l.id = ?
ORDER BY m.sort_order, m.id, ls.sort_order, ls.id`, levelID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	ids := []string{}
	for _, row := range rows {
		if row.LessonID != nil {
			ids = append(ids, *row.LessonID)
		}
	}
	return ids, nil
}
