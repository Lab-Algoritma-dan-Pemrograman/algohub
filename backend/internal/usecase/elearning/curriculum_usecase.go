package usecase

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

type CurriculumUsecase struct {
	curriculum repository.CurriculumRepository
}

func NewCurriculumUsecase(curriculum repository.CurriculumRepository) *CurriculumUsecase {
	return &CurriculumUsecase{curriculum: curriculum}
}

func (uc *CurriculumUsecase) Levels(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[entity.Level], error) {
	return listPage(p, page, limit, func() ([]entity.Level, int64, error) { return uc.curriculum.ListLevels(ctx, page, limit) })
}

func (uc *CurriculumUsecase) Modules(ctx context.Context, p *dto.Principal, page, limit int) (*dto.Page[entity.Module], error) {
	return listPage(p, page, limit, func() ([]entity.Module, int64, error) { return uc.curriculum.ListModules(ctx, page, limit) })
}

// Pemilihan tabel sama dengan loadCurriculum di SupabaseProvider.tsx:401: staf membaca lessons, lainnya view student_lessons.
func (uc *CurriculumUsecase) Lessons(ctx context.Context, p *dto.Principal, page, limit int) (any, error) {
	if p != nil && IsStaff(p.Role) {
		return listPage(p, page, limit, func() ([]entity.Lesson, int64, error) { return uc.curriculum.ListLessons(ctx, page, limit) })
	}
	return listPage(p, page, limit, func() ([]entity.StudentLesson, int64, error) {
		rows, total, err := uc.curriculum.ListStudentLessons(ctx, page, limit)
		for i := range rows {
			applyStudentLessonView(&rows[i])
		}
		return rows, total, err
	})
}

// Dari koneksi Go auth_role() = admin, jadi kolom solution view selalu terisi; aturan view
// (migration-flowchart-solution-visible.sql:13) diterapkan ulang dengan peran token.
func applyStudentLessonView(l *entity.StudentLesson) {
	if l.ExerciseType == nil || *l.ExerciseType != "flowchart" {
		l.Solution = nil
	}
	l.Quiz = withoutKey(l.Quiz, "correctAnswer")
}

// Operator jsonb - 'kunci': hanya kunci tingkat atas objek yang dibuang, nilai selain objek dibiarkan.
func withoutKey(raw datatypes.JSON, key string) datatypes.JSON {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return raw
	}
	if _, ok := obj[key]; !ok {
		return raw
	}
	delete(obj, key)
	out, err := json.Marshal(obj)
	if err != nil {
		return raw
	}
	return out
}

// Upsert levels ditolak WITH CHECK lebih dulu, jadi pesan selalu memakai tabel levels; delete
// sebelumnya hanya menghapus 0 baris. Tanpa level sama sekali, kode lama tidak melakukan upsert dan sukses.
func (uc *CurriculumUsecase) Save(ctx context.Context, p *dto.Principal, req *dto.CurriculumRequest) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		if len(req.Levels) > 0 {
			return RLSInsert("levels")
		}
		return nil
	}
	var (
		levels  []entity.Level
		modules []entity.Module
		lessons []entity.Lesson
	)
	for li := range req.Levels {
		lv := &req.Levels[li]
		accessMode := lv.AccessMode
		if accessMode == "" {
			accessMode = "auto"
		}
		locked := lv.Locked
		levels = append(levels, entity.Level{ID: lv.ID, Title: lv.Title, Description: lv.Description, AccessMode: &accessMode, Locked: &locked, SortOrder: li})
		for mi := range lv.Modules {
			m := &lv.Modules[mi]
			modules = append(modules, entity.Module{ID: m.ID, LevelID: lv.ID, Title: m.Title, SortOrder: mi})
			for si := range m.Lessons {
				lessons = append(lessons, lessonRow(&m.Lessons[si], m.ID, si))
			}
		}
	}
	return MapDBError(uc.curriculum.SaveFull(ctx, levels, modules, lessons))
}

func lessonRow(l *dto.CurriculumLesson, moduleID string, sortOrder int) entity.Lesson {
	return entity.Lesson{
		ID:              l.ID,
		ModuleID:        moduleID,
		Title:           l.Title,
		Explanation:     l.Explanation,
		CodeExample:     l.CodeExample,
		InitialCode:     l.InitialCode,
		Solution:        l.Solution,
		Hint:            l.Hint,
		Quiz:            jsonOr(l.Quiz, "{}"),
		TestCases:       jsonOr(l.TestCases, "[]"),
		ValidationRules: jsonOr(l.ValidationRules, "[]"),
		SortOrder:       sortOrder,
	}
}

// Meniru `x || fallback` pada nilai JSON dari body.
func jsonOr(raw json.RawMessage, fallback string) datatypes.JSON {
	v, err := parseJS(raw)
	if len(raw) == 0 || err != nil || !jsTruthy(v) {
		return datatypes.JSON(fallback)
	}
	return datatypes.JSON(raw)
}

// clearCurriculum menghapus levels; untuk peran selain admin/kordas USING gagal sehingga 0 baris terhapus dan tetap sukses.
func (uc *CurriculumUsecase) Clear(ctx context.Context, p *dto.Principal) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	return MapDBError(uc.curriculum.ClearAll(ctx))
}
