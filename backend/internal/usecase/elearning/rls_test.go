package usecase_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Nama test = ID aturan di docs/updates/elearning-rls.md § Aturan usecase dengan '-' diganti '_'.

const (
	nimSelf  = "202711001"
	nimOther = "202711002"
)

var bg = context.Background()

func assertMissingToken(t *testing.T, err error) {
	t.Helper()
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Missing token")
}

func rlsMessage(table string) string {
	return `new row violates row-level security policy for table "` + table + `"`
}

func TestRLS_R_users_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.users.On("ListAdmin", mock.Anything, 1, 100, "").Return([]repository.AdminUser{{NIM: nimOther}}, int64(1), nil)
	page, err := f.adminUC.List(bg, who(usecase.RolePraktikan), 1, 100, "")
	require.NoError(t, err)
	assert.Equal(t, nimOther, page.Items[0].NIM)
}

func TestRLS_R_users_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).adminUC.List(bg, nil, 1, 100, "")
	assertMissingToken(t, err)
}

func TestRLS_R_users_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	own := map[string]any{"nim": nimSelf}
	other := map[string]any{"nim": nimOther}
	f.users.On("Insert", mock.Anything, own, true).Return(nil).Once()
	f.users.On("Insert", mock.Anything, other, false).Return(nil).Once()
	require.NoError(t, usecase.GuardInsertUser(bg, f.users, who(usecase.RolePraktikan), own))
	require.NoError(t, usecase.GuardInsertUser(bg, f.users, who(usecase.RoleKordas), other))
}

func TestRLS_R_users_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range []string{usecase.RolePraktikan, usecase.RoleAsisten} {
		err := usecase.GuardInsertUser(bg, f.users, who(role), map[string]any{"nim": nimOther})
		assertUCError(t, err, http.StatusForbidden, rlsMessage("users"))
	}
}

func TestRLS_R_users_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"streak": 2}
	f.users.On("UpdateUser", mock.Anything, nimSelf, cols, true).Return(int64(1), nil).Once()
	f.users.On("UpdateUser", mock.Anything, nimOther, cols, false).Return(int64(1), nil).Once()
	n, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RolePraktikan), nimSelf, cols)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	n, err = usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleAdmin), nimOther, cols)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestRLS_R_users_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range []string{usecase.RolePraktikan, usecase.RoleAsisten} {
		n, err := usecase.GuardUpdateUser(bg, f.users, who(role), nimOther, map[string]any{"xp": 0})
		require.NoError(t, err)
		assert.Zero(t, n)
	}
}

func TestRLS_R_users_UPDATE_2_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{
		"role": "asisten", "nama": "x", "xp": 1, "level": 1, "streak": 1, "study_time": 1, "last_active": learnNow,
		"kelas": "x", "jurusan": "x", "elearning_role": nil, "level_access_overrides": datatypes.JSON(`{}`),
	}
	f.users.On("UpdateUser", mock.Anything, nimOther, cols, false).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleKordas), nimOther, cols)
	require.NoError(t, err)
}

func TestRLS_R_users_UPDATE_2_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, p := range []*dto.Principal{who(usecase.RoleKordas), who(usecase.RolePraktikan)} {
		_, err := usecase.GuardUpdateUser(bg, f.users, p, nimSelf, map[string]any{"last_active": learnNow, "shift": "1"})
		assertUCError(t, err, http.StatusForbidden, "permission denied for table users")
	}
}

func TestRLS_R_users_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.users.On("Delete", mock.Anything, nimOther).Return(int64(1), nil)
	n, err := usecase.GuardDeleteUser(bg, f.users, who(usecase.RoleAdmin), nimOther)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestRLS_R_users_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	n, err := usecase.GuardDeleteUser(bg, f.users, who(usecase.RoleAsisten), nimOther)
	require.NoError(t, err)
	assert.Zero(t, n)
}

func TestRLS_R_levels_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ListLevels", mock.Anything, 1, 100).Return([]entity.Level{{ID: "level-1"}}, int64(1), nil)
	page, err := f.curriculumUC.Levels(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, &dto.Page[entity.Level]{Items: []entity.Level{{ID: "level-1"}}, Total: 1, Page: 1, Limit: 100}, page)
}

func TestRLS_R_levels_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).curriculumUC.Levels(bg, nil, 1, 100)
	assertMissingToken(t, err)
}

func TestRLS_R_modules_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ListModules", mock.Anything, 2, 10).Return([]entity.Module{{ID: "m1"}}, int64(11), nil)
	page, err := f.curriculumUC.Modules(bg, who(usecase.RolePraktikan), 2, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(11), page.Total)
}

func TestRLS_R_modules_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).curriculumUC.Modules(bg, nil, 1, 100)
	assertMissingToken(t, err)
}

func oneLevel() *dto.CurriculumRequest {
	return &dto.CurriculumRequest{Levels: []dto.CurriculumLevel{{
		ID: "level-1", Title: "Dasar",
		Modules: []dto.CurriculumModule{{ID: "m1", Title: "Modul", Lessons: []dto.CurriculumLesson{{ID: "l1", Title: "Pelajaran"}}}},
	}}}
}

// Upsert PUT /curriculum menulis ketiga tabel sekaligus; policy ALL admin/kordas berlaku sama untuk INSERT dan UPDATE.
func assertCurriculumSaved(t *testing.T, role string, check func(levels []entity.Level, modules []entity.Module, lessons []entity.Lesson) bool) {
	t.Helper()
	f := newLearn(t)
	f.curriculum.On("SaveFull", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		assert.True(t, check(args.Get(1).([]entity.Level), args.Get(2).([]entity.Module), args.Get(3).([]entity.Lesson)))
	})
	require.NoError(t, f.curriculumUC.Save(bg, who(role), oneLevel()))
}

// Upsert levels jalan pertama dan ditolak WITH CHECK, jadi pesan tabel selalu levels walaupun modules/lessons ikut dikirim.
func assertCurriculumRejected(t *testing.T, role string) {
	t.Helper()
	err := newLearn(t).curriculumUC.Save(bg, who(role), oneLevel())
	assertUCError(t, err, http.StatusForbidden, rlsMessage("levels"))
}

func TestRLS_R_levels_INSERT_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleKordas, func(levels []entity.Level, _ []entity.Module, _ []entity.Lesson) bool {
		return len(levels) == 1 && levels[0].ID == "level-1" && *levels[0].AccessMode == "auto" && !*levels[0].Locked && levels[0].SortOrder == 0
	})
}

func TestRLS_R_levels_INSERT_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RoleAsisten)
}

func TestRLS_R_levels_UPDATE_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleAdmin, func(levels []entity.Level, _ []entity.Module, _ []entity.Lesson) bool {
		return levels[0].Title == "Dasar"
	})
}

func TestRLS_R_levels_UPDATE_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RolePraktikan)
}

func TestRLS_R_modules_INSERT_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleAdmin, func(_ []entity.Level, modules []entity.Module, _ []entity.Lesson) bool {
		return len(modules) == 1 && modules[0].LevelID == "level-1" && modules[0].SortOrder == 0
	})
}

func TestRLS_R_modules_INSERT_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RolePraktikan)
}

func TestRLS_R_modules_UPDATE_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleKordas, func(_ []entity.Level, modules []entity.Module, _ []entity.Lesson) bool {
		return modules[0].Title == "Modul"
	})
}

func TestRLS_R_modules_UPDATE_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RoleAsisten)
}

func TestRLS_R_lessons_INSERT_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleKordas, func(_ []entity.Level, _ []entity.Module, lessons []entity.Lesson) bool {
		l := lessons[0]
		return l.ModuleID == "m1" && string(l.Quiz) == "{}" && string(l.TestCases) == "[]" && string(l.ValidationRules) == "[]"
	})
}

func TestRLS_R_lessons_INSERT_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RoleAsisten)
}

func TestRLS_R_lessons_UPDATE_1_Diizinkan(t *testing.T) {
	assertCurriculumSaved(t, usecase.RoleAdmin, func(_ []entity.Level, _ []entity.Module, lessons []entity.Lesson) bool {
		return lessons[0].Title == "Pelajaran" && lessons[0].SortOrder == 0
	})
}

func TestRLS_R_lessons_UPDATE_1_Ditolak(t *testing.T) {
	assertCurriculumRejected(t, usecase.RolePraktikan)
}

func TestRLS_R_levels_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ClearAll", mock.Anything).Return(nil)
	require.NoError(t, f.curriculumUC.Clear(bg, who(usecase.RoleKordas)))
}

func TestRLS_R_levels_DELETE_1_Ditolak(t *testing.T) {
	require.NoError(t, newLearn(t).curriculumUC.Clear(bg, who(usecase.RoleAsisten)))
}

// Kurikulum kosong dari admin/kordas berarti semua modul lama dihapus, sama dengan saveFullCurriculum([]).
func TestRLS_R_modules_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("SaveFull", mock.Anything, []entity.Level(nil), []entity.Module(nil), []entity.Lesson(nil)).Return(nil)
	require.NoError(t, f.curriculumUC.Save(bg, who(usecase.RoleAdmin), &dto.CurriculumRequest{Levels: []dto.CurriculumLevel{}}))
}

func TestRLS_R_modules_DELETE_1_Ditolak(t *testing.T) {
	err := newLearn(t).curriculumUC.Save(bg, who(usecase.RolePraktikan), &dto.CurriculumRequest{Levels: []dto.CurriculumLevel{}})
	require.NoError(t, err)
}

func TestRLS_R_lessons_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ClearAll", mock.Anything).Return(nil)
	require.NoError(t, f.curriculumUC.Clear(bg, who(usecase.RoleAdmin)))
}

func TestRLS_R_lessons_DELETE_1_Ditolak(t *testing.T) {
	require.NoError(t, newLearn(t).curriculumUC.Clear(bg, who(usecase.RolePraktikan)))
}

func TestRLS_R_lessons_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ListLessons", mock.Anything, 1, 100).Return([]entity.Lesson{{ID: "l1", Solution: str("print(1)")}}, int64(1), nil)
	res, err := f.curriculumUC.Lessons(bg, who(usecase.RoleAsisten), 1, 100)
	require.NoError(t, err)
	page := res.(*dto.Page[entity.Lesson])
	assert.Equal(t, "print(1)", *page.Items[0].Solution)
}

func TestRLS_R_lessons_SELECT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("ListStudentLessons", mock.Anything, 1, 100).Return([]entity.StudentLesson{}, int64(0), nil)
	res, err := f.curriculumUC.Lessons(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.IsType(t, &dto.Page[entity.StudentLesson]{}, res)
}

func studentLessons(t *testing.T, lessons ...entity.StudentLesson) []entity.StudentLesson {
	t.Helper()
	f := newLearn(t)
	f.curriculum.On("ListStudentLessons", mock.Anything, 1, 100).Return(lessons, int64(len(lessons)), nil)
	res, err := f.curriculumUC.Lessons(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	return res.(*dto.Page[entity.StudentLesson]).Items
}

func TestRLS_R_student_lessons_SELECT_1_Diizinkan(t *testing.T) {
	items := studentLessons(t, entity.StudentLesson{ID: "l1", Quiz: datatypes.JSON(`{"question":"q","options":["a","b"],"correctAnswer":1}`)})
	var quiz map[string]any
	require.NoError(t, json.Unmarshal(items[0].Quiz, &quiz))
	assert.Equal(t, map[string]any{"question": "q", "options": []any{"a", "b"}}, quiz)
}

func TestRLS_R_student_lessons_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).curriculumUC.Lessons(bg, nil, 1, 100)
	assertMissingToken(t, err)
}

func TestRLS_R_student_lessons_SELECT_2_Diizinkan(t *testing.T) {
	items := studentLessons(t, entity.StudentLesson{ID: "l1", ExerciseType: str("flowchart"), Solution: str("start")})
	assert.Equal(t, "start", *items[0].Solution)
}

func TestRLS_R_student_lessons_SELECT_2_Ditolak(t *testing.T) {
	items := studentLessons(t,
		entity.StudentLesson{ID: "l1", ExerciseType: str("code"), Solution: str("print(1)")},
		entity.StudentLesson{ID: "l2", Solution: str("print(2)")})
	assert.Nil(t, items[0].Solution)
	assert.Nil(t, items[1].Solution)
}

func TestRLS_R_student_progress_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.progress.On("ListByNIM", mock.Anything, nimOther, 1, 100).Return([]entity.StudentProgress{{LessonID: "l1"}}, int64(1), nil)
	f.progress.On("ListByNIM", mock.Anything, nimSelf, 1, 100).Return([]entity.StudentProgress{{LessonID: "l2"}}, int64(1), nil)
	page, err := f.adminUC.Progress(bg, who(usecase.RoleAsisten), nimOther, 1, 100)
	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
	mine, err := f.progressUC.MyLessons(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"l2"}, mine.Items)
}

func TestRLS_R_student_progress_SELECT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	page, err := f.adminUC.Progress(bg, who(usecase.RolePraktikan), nimOther, 1, 100)
	require.NoError(t, err)
	assert.Empty(t, page.Items)
	assert.Zero(t, page.Total)
}

func TestRLS_R_student_progress_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "mahasiswa", nil)
	f.progress.On("Exists", mock.Anything, nimSelf, "l1").Return(false, nil)
	f.progress.On("Insert", mock.Anything, nimSelf, "l1", learnNow).Return(nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, mock.Anything, true).Return(int64(1), nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nimSelf).Return([]string{"l1"}, nil)
	f.curriculum.On("LevelTree", mock.Anything).Return(nil, nil)
	f.achievements.On("List", mock.Anything, "").Return([]entity.Achievement{{ID: "x", RequirementType: str("xp"), RequirementValue: str("999999")}}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{}, nil)
	_, err := f.progressUC.Complete(bg, who(usecase.RolePraktikan), "l1", &dto.CompleteLessonRequest{XPReward: 60})
	require.NoError(t, err)
}

func TestRLS_R_student_progress_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	p := mismatched(usecase.RolePraktikan)
	f.profileRole(p.NIM, "mahasiswa", nil)
	_, err := f.progressUC.Complete(bg, p, "l1", &dto.CompleteLessonRequest{XPReward: 60})
	assertUCError(t, err, http.StatusForbidden, rlsMessage("student_progress"))
}

func TestRLS_R_student_progress_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "asisten", nil)
	f.progress.On("DeleteByNIM", mock.Anything, nimOther).Return(int64(3), nil)
	f.achievements.On("DeleteByNIM", mock.Anything, nimOther).Return(int64(1), nil)
	f.expectAudit("progress_reset", "Mereset seluruh progress belajar mahasiswa NIM: "+nimOther)
	require.NoError(t, f.adminUC.Reset(bg, who(usecase.RoleAsisten), nimOther))
}

// Praktikan ber-flag elearning_role admin lolos guard UI, tetapi DELETE student_progress hanya untuk staf token: 0 baris.
func TestRLS_R_student_progress_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "mahasiswa", str("admin"))
	f.expectAudit("progress_reset", "Mereset seluruh progress belajar mahasiswa NIM: "+nimOther)
	require.NoError(t, f.adminUC.Reset(bg, who(usecase.RolePraktikan), nimOther))
	f.progress.AssertNotCalled(t, "DeleteByNIM", mock.Anything, mock.Anything)
}

func TestRLS_R_achievements_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.achievements.On("List", mock.Anything, "xp").Return([]entity.Achievement{{ID: "xp-50", RequirementType: str("xp"), RequirementValue: str("50")}}, nil)
	list, err := f.achievementUC.List(bg, who(usecase.RolePraktikan), "xp")
	require.NoError(t, err)
	assert.InDelta(t, 50.0, list[0].RequirementValue, 0)
}

func TestRLS_R_achievements_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).achievementUC.List(bg, nil, "")
	assertMissingToken(t, err)
}

func TestRLS_R_unlocked_achievements_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{"xp-50"}, nil)
	ids, err := f.achievementUC.Mine(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, []string{"xp-50"}, ids)
}

func TestRLS_R_unlocked_achievements_SELECT_1_Ditolak(t *testing.T) {
	ids, err := newLearn(t).achievementUC.Mine(bg, mismatched(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func xpAchievement(f *learnFixture, nim string) {
	f.users.On("GetProfile", mock.Anything, nim).Return(&entity.User{NIM: nim, XP: 100}, nil)
	f.achievements.On("List", mock.Anything, "").Return([]entity.Achievement{{ID: "xp-50", RequirementType: str("xp"), RequirementValue: str("50")}}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nim).Return([]string{}, nil)
	f.progress.On("CountCompleted", mock.Anything, nim).Return(int64(0), nil)
}

func TestRLS_R_unlocked_achievements_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	xpAchievement(f, nimSelf)
	f.achievements.On("Insert", mock.Anything, nimSelf, "xp-50", learnNow).Return(nil)
	got, err := f.achievementUC.Check(bg, who(usecase.RoleKordas), &dto.AchievementCheckRequest{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "xp-50", got[0].ID)
}

// WITH CHECK nim = auth_nim() berlaku juga untuk admin/kordas.
func TestRLS_R_unlocked_achievements_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	p := mismatched(usecase.RoleKordas)
	xpAchievement(f, p.NIM)
	got, err := f.achievementUC.Check(bg, p, &dto.AchievementCheckRequest{})
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestRLS_R_unlocked_achievements_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.progress.On("DeleteByNIM", mock.Anything, nimOther).Return(int64(0), nil)
	f.achievements.On("DeleteByNIM", mock.Anything, nimOther).Return(int64(2), nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"xp": 0, "level": 1, "streak": 0}, false).Return(int64(1), nil)
	f.expectAudit("progress_reset", "Mereset seluruh progress belajar mahasiswa NIM: "+nimOther)
	require.NoError(t, f.adminUC.Reset(bg, who(usecase.RoleKordas), nimOther))
}

func TestRLS_R_unlocked_achievements_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "mahasiswa", str("admin"))
	f.expectAudit("progress_reset", "Mereset seluruh progress belajar mahasiswa NIM: "+nimOther)
	require.NoError(t, f.adminUC.Reset(bg, who(usecase.RolePraktikan), nimOther))
	f.achievements.AssertNotCalled(t, "DeleteByNIM", mock.Anything, mock.Anything)
}

func TestRLS_R_activity_logs_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.expectAudit("login", "Masuk ke sistem E-Learning")
	err := f.logUC.Insert(bg, who(usecase.RolePraktikan), &dto.ActivityLogRequest{EventType: "login", Details: "Masuk ke sistem E-Learning"})
	require.NoError(t, err)
}

// Log internal memakai nim tampilan (currentUser.nim); bila beda dari auth_nim() insert ditolak dan aksi tetap jalan.
func TestRLS_R_activity_logs_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	p := mismatched(usecase.RoleKordas)
	f.profileRole(p.NIM, "koordinator", nil)
	f.users.On("Delete", mock.Anything, nimOther).Return(int64(1), nil)
	require.NoError(t, f.adminUC.Delete(bg, p, nimOther))
	f.logs.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func oldUser(role string) *entity.User {
	return &entity.User{NIM: nimSelf, Role: role, LevelAccessOverrides: datatypes.JSON(`{"level-1":"auto"}`), AssessmentAccess: datatypes.JSON(`{"pre_test":false}`), PrestigeMult: 1}
}

func TestRLS_T_check_user_updates_1_Diizinkan(t *testing.T) {
	require.NoError(t, usecase.CheckUserUpdates(usecase.RoleKordas, oldUser("mahasiswa"), map[string]any{"role": "admin"}))

	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(oldUser("mahasiswa"), nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"role": "asisten"}, true).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleAsisten), nimSelf, map[string]any{"role": "asisten"})
	require.NoError(t, err)
}

func TestRLS_T_check_user_updates_1_Ditolak(t *testing.T) {
	cases := []struct {
		token, oldRole, newRole string
	}{
		{usecase.RolePraktikan, "asisten", "praktikan"},
		{usecase.RoleAsisten, "asisten", "kordas"},
		{usecase.RolePraktikan, "mahasiswa", "asisten"},
	}
	for _, tc := range cases {
		f := newLearn(t)
		f.users.On("GetProfile", mock.Anything, nimSelf).Return(oldUser(tc.oldRole), nil)
		_, err := usecase.GuardUpdateUser(bg, f.users, who(tc.token), nimSelf, map[string]any{"role": tc.newRole})
		assertUCError(t, err, http.StatusBadRequest, "Akses Ditolak: Anda tidak diperbolehkan mengubah peran (role) Anda sendiri!")
	}
}

func TestRLS_T_check_user_updates_2_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"level_access_overrides": datatypes.JSON(`{"level-1":"unlocked"}`)}
	f.users.On("UpdateUser", mock.Anything, nimOther, cols, false).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleAdmin), nimOther, cols)
	require.NoError(t, err)
	require.NoError(t, usecase.CheckUserUpdates(usecase.RolePraktikan, oldUser("mahasiswa"), map[string]any{"level_access_overrides": datatypes.JSON(`{ "level-1" : "auto" }`)}))
}

func TestRLS_T_check_user_updates_2_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(oldUser("asisten"), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleAsisten), nimSelf, map[string]any{"level_access_overrides": datatypes.JSON(`{"level-1":"unlocked"}`)})
	assertUCError(t, err, http.StatusBadRequest, "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah level_access_overrides!")
}

func TestRLS_T_check_user_updates_3_Diizinkan(t *testing.T) {
	require.NoError(t, usecase.CheckUserUpdates(usecase.RoleAdmin, oldUser("mahasiswa"), map[string]any{"assessment_access": datatypes.JSON(`{"pre_test":true}`)}))
	require.NoError(t, usecase.CheckUserUpdates(usecase.RolePraktikan, oldUser("mahasiswa"), map[string]any{"assessment_access": datatypes.JSON(`{"pre_test":false}`)}))
}

func TestRLS_T_check_user_updates_3_Ditolak(t *testing.T) {
	err := usecase.CheckUserUpdates(usecase.RolePraktikan, oldUser("mahasiswa"), map[string]any{"assessment_access": datatypes.JSON(`{"pre_test":true}`)})
	assertUCError(t, err, http.StatusBadRequest, "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah assessment_access!")
}

func TestRLS_T_check_user_updates_4_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"role": "koordinator", "elearning_role": "admin"}
	f.users.On("UpdateUser", mock.Anything, nimOther, cols, false).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleKordas), nimOther, cols)
	require.NoError(t, err)
}

func TestRLS_T_check_user_updates_4_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(oldUser("asisten"), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleAsisten), nimSelf, map[string]any{"role": "asisten", "elearning_role": "admin"})
	assertUCError(t, err, http.StatusBadRequest, "Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah elearning_role!")
}

func TestRLS_T_check_user_xp_level_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"xp": 500, "level": 4}
	f.users.On("UpdateUser", mock.Anything, nimOther, cols, false).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleKordas), nimOther, cols)
	require.NoError(t, err)
}

// recomputeXP=true: repository menimpa xp/level dari sumber dalam satu UPDATE dan meng-clamp prestige_mult.
func TestRLS_T_check_user_xp_level_1_Ditolak(t *testing.T) {
	for _, role := range []string{usecase.RoleAsisten, usecase.RolePraktikan} {
		f := newLearn(t)
		cols := map[string]any{"xp": 99999, "level": 50}
		f.users.On("UpdateUser", mock.Anything, nimSelf, cols, true).Return(int64(1), nil)
		_, err := usecase.GuardUpdateUser(bg, f.users, who(role), nimSelf, cols)
		require.NoError(t, err)
	}
}

func TestRLS_T_check_user_xp_level_2_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"nim": nimSelf, "xp": 0, "level": 1}
	f.users.On("Insert", mock.Anything, cols, false).Return(nil)
	require.NoError(t, usecase.GuardInsertUser(bg, f.users, who(usecase.RoleAdmin), cols))
}

func TestRLS_T_check_user_xp_level_2_Ditolak(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"nim": nimSelf, "xp": 0, "level": 1}
	f.users.On("Insert", mock.Anything, cols, true).Return(nil)
	require.NoError(t, usecase.GuardInsertUser(bg, f.users, who(usecase.RoleAsisten), cols))
}

func TestRLS_T_guard_prestige_columns_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"streak": 1}, false).Return(int64(1), nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"streak": 1}, true).Return(int64(1), nil)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RoleKordas), nimOther, map[string]any{"streak": 1})
	require.NoError(t, err)
	_, err = usecase.GuardUpdateUser(bg, f.users, who(usecase.RolePraktikan), nimSelf, map[string]any{"streak": 1})
	require.NoError(t, err)
}

func TestRLS_T_guard_prestige_columns_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"streak": 1}, true).Return(int64(0), repository.ErrPrestigeGuard)
	_, err := usecase.GuardUpdateUser(bg, f.users, who(usecase.RolePraktikan), nimSelf, map[string]any{"streak": 1})
	assertUCError(t, err, http.StatusBadRequest, "Akses Ditolak: kolom prestige hanya bisa diubah lewat Prestige resmi (tamat 6/6).")
}

func TestRLS_T_check_progress_velocity_1_Diizinkan(t *testing.T) {
	TestRLS_R_student_progress_INSERT_1_Diizinkan(t)
}

func TestRLS_T_check_progress_velocity_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "mahasiswa", nil)
	f.progress.On("Exists", mock.Anything, nimSelf, "l1").Return(false, nil)
	msg := "Aktivitas Mencurigakan: Anda menyelesaikan materi terlalu cepat! Harap baca materi dengan seksama."
	f.progress.On("Insert", mock.Anything, nimSelf, "l1", learnNow).Return(&pgconn.PgError{Code: "P0001", Message: msg})
	_, err := f.progressUC.Complete(bg, who(usecase.RolePraktikan), "l1", &dto.CompleteLessonRequest{XPReward: 60})
	assertUCError(t, err, http.StatusBadRequest, msg)
	f.users.AssertNotCalled(t, "UpdateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// Senin 16 November 2026 00:00 WIB, awal minggu learnNow.
var learnWeekStart = time.Date(2026, 11, 15, 17, 0, 0, 0, time.UTC)

func intp(v int) *int    { return &v }
func boolp(v bool) *bool { return &v }

func oneQuestion() *dto.GameQuestionsRequest {
	return &dto.GameQuestionsRequest{Questions: []dto.GameQuestionInput{{Language: str("c"), Title: str("Bug"), BugLine: intp(3)}}}
}

var oneQuestionFields = []repository.QuestionFields{{Language: str("c"), Title: str("Bug"), BugLine: intp(3)}}

var nonAdminRoles = []string{usecase.RoleAsisten, usecase.RolePraktikan}

func TestRLS_R_game_questions_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("QuestionsByLanguage", mock.Anything, "c").Return([]entity.GameQuestion{{ID: "q1", Language: "c"}}, nil)
	qs, err := f.gameUC.Questions(bg, who(usecase.RolePraktikan), "c", 5)
	require.NoError(t, err)
	assert.Equal(t, []dto.GameQuestion{{ID: "q1", Language: "c"}}, qs)
}

func TestRLS_R_game_questions_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).gameUC.Questions(bg, nil, "c", 5)
	assertMissingToken(t, err)
}

func TestRLS_R_game_questions_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("InsertQuestion", mock.Anything, &oneQuestionFields[0]).Return("q9", nil)
	f.games.On("CountQuestions", mock.Anything).Return(int64(0), nil)
	f.games.On("InsertQuestions", mock.Anything, oneQuestionFields).Return(nil)
	f.games.On("ReplaceAll", mock.Anything, oneQuestionFields).Return(nil)

	res, err := f.gameUC.Add(bg, who(usecase.RoleAdmin), &oneQuestion().Questions[0])
	require.NoError(t, err)
	assert.Equal(t, "q9", res.ID)
	seed, err := f.gameUC.Seed(bg, who(usecase.RoleKordas), oneQuestion())
	require.NoError(t, err)
	assert.True(t, seed.Seeded)
	require.NoError(t, f.gameUC.ForceReset(bg, who(usecase.RoleAdmin), oneQuestion()))
}

func TestRLS_R_game_questions_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.games.On("CountQuestions", mock.Anything).Return(int64(0), nil)
	for _, role := range nonAdminRoles {
		_, err := f.gameUC.Add(bg, who(role), &oneQuestion().Questions[0])
		assertUCError(t, err, http.StatusForbidden, rlsMessage("game_questions"))
		_, err = f.gameUC.Seed(bg, who(role), oneQuestion())
		assertUCError(t, err, http.StatusForbidden, rlsMessage("game_questions"))
		err = f.gameUC.ForceReset(bg, who(role), oneQuestion())
		assertUCError(t, err, http.StatusForbidden, rlsMessage("game_questions"))
	}
	f.games.AssertNotCalled(t, "ReplaceAll", mock.Anything, mock.Anything)
}

func TestRLS_R_game_questions_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("UpdateQuestion", mock.Anything, "q1", map[string]any{"title": "Baru", "bug_line": (*int)(nil)}).Return(int64(1), nil)
	req := &dto.GameQuestionUpdate{Title: "Baru", BugLine: dto.Optional[int]{Set: true}}
	require.NoError(t, f.gameUC.Update(bg, who(usecase.RoleKordas), "q1", req))
}

func TestRLS_R_game_questions_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		require.NoError(t, f.gameUC.Update(bg, who(role), "q1", &dto.GameQuestionUpdate{Title: "Baru"}))
	}
	f.games.AssertNotCalled(t, "UpdateQuestion", mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_game_questions_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("DeleteQuestion", mock.Anything, "q1").Return(int64(1), nil)
	require.NoError(t, f.gameUC.Delete(bg, who(usecase.RoleAdmin), "q1"))
}

// DELETE di forceResetGameQuestions juga 0 baris untuk selain admin/kordas; tanpa soal baru tidak ada INSERT yang ditolak.
func TestRLS_R_game_questions_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		require.NoError(t, f.gameUC.Delete(bg, who(role), "q1"))
		require.NoError(t, f.gameUC.ForceReset(bg, who(role), &dto.GameQuestionsRequest{Questions: []dto.GameQuestionInput{}}))
	}
	f.games.AssertNotCalled(t, "DeleteQuestion", mock.Anything, mock.Anything)
	f.games.AssertNotCalled(t, "ReplaceAll", mock.Anything, mock.Anything)
}

func TestRLS_R_game_settings_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("Settings", mock.Anything).Return(&entity.GameSettings{ID: "default", BugHuntActive: boolp(false), BugHuntWeeklyLimit: intp(0)}, nil)
	s, err := f.gameUC.Settings(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, &dto.GameSettings{BugHuntActive: boolp(false), BugHuntWeeklyLimit: intp(0), BugHuntQuestionCount: 5}, s)
}

func TestRLS_R_game_settings_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).gameUC.Settings(bg, nil)
	assertMissingToken(t, err)
}

func TestRLS_R_game_settings_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("UpsertSettings", mock.Anything, map[string]any{"bug_hunt_weekly_limit": intp(4)}).Return(nil)
	req := &dto.GameSettingsUpdate{BugHuntWeeklyLimit: dto.Optional[int]{Set: true, Value: intp(4)}}
	require.NoError(t, f.gameUC.UpdateSettings(bg, who(usecase.RoleAdmin), req))
}

func TestRLS_R_game_settings_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		err := f.gameUC.UpdateSettings(bg, who(role), &dto.GameSettingsUpdate{})
		assertUCError(t, err, http.StatusForbidden, rlsMessage("game_settings"))
	}
}

// updateGameSettings selalu upsert, jadi baris 'default' yang sudah ada ikut aturan R-game_settings-INSERT-1.
func TestRLS_R_game_settings_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.games.On("UpsertSettings", mock.Anything, map[string]any{"bug_hunt_active": boolp(false)}).Return(nil)
	req := &dto.GameSettingsUpdate{BugHuntActive: dto.Optional[bool]{Set: true, Value: boolp(false)}}
	require.NoError(t, f.gameUC.UpdateSettings(bg, who(usecase.RoleKordas), req))
}

func TestRLS_R_game_settings_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	req := &dto.GameSettingsUpdate{BugHuntActive: dto.Optional[bool]{Set: true, Value: boolp(false)}}
	assertUCError(t, f.gameUC.UpdateSettings(bg, who(usecase.RoleAsisten), req), http.StatusForbidden, rlsMessage("game_settings"))
	f.games.AssertNotCalled(t, "UpsertSettings", mock.Anything, mock.Anything)
}

func (f *learnFixture) bugHuntLimit(limit int) {
	f.games.On("Settings", mock.Anything).Return(&entity.GameSettings{ID: "default", BugHuntActive: boolp(true), BugHuntWeeklyLimit: intp(limit)}, nil)
}

func TestRLS_R_game_history_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.bugHuntLimit(3)
	f.games.On("CountPlaysSince", mock.Anything, nimSelf, "bug_hunt", learnWeekStart).Return(int64(2), nil)
	for _, p := range []*dto.Principal{who(usecase.RolePraktikan), mismatched(usecase.RoleAsisten)} {
		res, err := f.gameUC.PlaysMe(bg, p)
		require.NoError(t, err)
		assert.Equal(t, &dto.BugHuntAccess{Allowed: true, PlaysUsed: 2, Limit: 3}, res)
	}
}

func TestRLS_R_game_history_SELECT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.bugHuntLimit(3)
	res, err := f.gameUC.PlaysMe(bg, mismatched(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, &dto.BugHuntAccess{Allowed: true, PlaysUsed: 0, Limit: 3}, res)
	f.games.AssertNotCalled(t, "CountPlaysSince", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_game_history_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	yesterday := learnNow.Add(-24 * time.Hour)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 100, Level: 2, Streak: 1, LastActive: &yesterday, PrestigeMult: 1}, nil)
	cols := map[string]any{"xp": 140, "level": 2, "streak": 2, "last_active": learnNow}
	f.games.On("InsertHistory", mock.Anything, nimSelf, "quiz", 40, learnNow).Return(nil).Twice()
	f.users.On("UpdateUser", mock.Anything, nimSelf, cols, true).Return(int64(1), nil).Once()
	f.users.On("UpdateUser", mock.Anything, nimSelf, cols, false).Return(int64(1), nil).Once()

	for _, role := range []string{usecase.RolePraktikan, usecase.RoleAdmin} {
		res, err := f.gameUC.History(bg, who(role), &dto.GameHistoryRequest{GameType: "quiz", XPEarned: 40})
		require.NoError(t, err)
		assert.Equal(t, &dto.GameHistoryResponse{XP: 140, Level: 2, Streak: 2, LastActive: learnNow}, res)
	}
}

func TestRLS_R_game_history_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range []string{usecase.RoleAdmin, usecase.RolePraktikan} {
		_, err := f.gameUC.History(bg, mismatched(role), &dto.GameHistoryRequest{GameType: "quiz", XPEarned: 40})
		assertUCError(t, err, http.StatusForbidden, rlsMessage("game_history"))
	}
	f.games.AssertNotCalled(t, "InsertHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_playground_examples_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.playground.On("List", mock.Anything, "python", 1, 100).Return([]entity.PlaygroundExample{{ID: "e1", Title: "Halo", Language: "python", Code: "print()", SortOrder: 2}}, int64(1), nil)
	page, err := f.playgroundUC.List(bg, who(usecase.RolePraktikan), "python", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, []dto.PlaygroundExample{{ID: "e1", Title: "Halo", Language: "python", Code: "print()", SortOrder: 2}}, page.Items)
}

func TestRLS_R_playground_examples_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).playgroundUC.List(bg, nil, "", 1, 100)
	assertMissingToken(t, err)
}

func TestRLS_R_playground_examples_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	cols := map[string]any{"title": str("Halo"), "language": str("c"), "code": str("int main(){}"), "description": (*string)(nil), "sort_order": 1}
	f.playground.On("Insert", mock.Anything, cols).Return("e9", nil)
	res, err := f.playgroundUC.Add(bg, who(usecase.RoleAdmin), &dto.PlaygroundInput{Title: str("Halo"), Language: str("c"), Code: str("int main(){}"), SortOrder: intp(1)})
	require.NoError(t, err)
	assert.Equal(t, "e9", res.ID)
}

func TestRLS_R_playground_examples_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		_, err := f.playgroundUC.Add(bg, who(role), &dto.PlaygroundInput{Title: str("Halo")})
		assertUCError(t, err, http.StatusForbidden, rlsMessage("playground_examples"))
	}
}

func TestRLS_R_playground_examples_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.playground.On("Update", mock.Anything, "e1", map[string]any{"code": "x", "description": str("d")}).Return(int64(1), nil)
	req := &dto.PlaygroundUpdate{Code: "x", Description: dto.Optional[string]{Set: true, Value: str("d")}}
	require.NoError(t, f.playgroundUC.Update(bg, who(usecase.RoleKordas), "e1", req))
}

func TestRLS_R_playground_examples_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		require.NoError(t, f.playgroundUC.Update(bg, who(role), "e1", &dto.PlaygroundUpdate{Code: "x"}))
	}
	f.playground.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_playground_examples_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.playground.On("Delete", mock.Anything, "e1").Return(int64(1), nil)
	require.NoError(t, f.playgroundUC.Delete(bg, who(usecase.RoleAdmin), "e1"))
}

func TestRLS_R_playground_examples_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		require.NoError(t, f.playgroundUC.Delete(bg, who(role), "e1"))
	}
	f.playground.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestRLS_R_active_sessions_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.sessions.On("ListWithRole", mock.Anything, 1, 100).Return([]repository.SessionRole{
		{NIM: nimOther, LastHeartbeat: &learnNow, Role: str("mahasiswa"), HasUser: true},
		{NIM: "tanpa-user", LastHeartbeat: &learnNow},
	}, int64(2), nil)
	page, err := f.sessionUC.Sessions(bg, who(usecase.RoleAsisten), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, []dto.ActiveSessionRole{
		{NIM: nimOther, LastHeartbeat: &learnNow, Users: &dto.SessionUser{Role: str("mahasiswa")}},
		{NIM: "tanpa-user", LastHeartbeat: &learnNow},
	}, page.Items)
}

func TestRLS_R_active_sessions_SELECT_1_Ditolak(t *testing.T) {
	page, err := newLearn(t).sessionUC.Sessions(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.Empty(t, page.Items)
}

func TestRLS_R_active_sessions_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	s := &entity.ActiveSession{NIM: nimSelf, Nama: "Nama " + nimSelf, Kelas: "TI-1", LastHeartbeat: &learnNow, CurrentActivity: str("courses")}
	f.sessions.On("Upsert", mock.Anything, s, false).Return(nil)
	res, err := f.sessionUC.Heartbeat(bg, who(usecase.RoleAsisten), &dto.HeartbeatRequest{CurrentActivity: str("courses")})
	require.NoError(t, err)
	assert.True(t, res.SessionSaved)
}

// Layout.tsx menelan penolakan upsert dan tetap memperbarui users, jadi touchStreak tetap jalan.
func TestRLS_R_active_sessions_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Streak: 3, LastActive: &learnNow, PrestigeMult: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"last_active": learnNow}, true).Return(int64(1), nil)
	res, err := f.sessionUC.Heartbeat(bg, who(usecase.RolePraktikan), &dto.HeartbeatRequest{CurrentActivity: str("lesson"), TouchStreak: true})
	require.NoError(t, err)
	assert.Equal(t, &dto.HeartbeatResponse{Streak: &dto.StreakResponse{Streak: 3, LastActive: learnNow}}, res)
	f.sessions.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_active_sessions_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	s := &entity.ActiveSession{NIM: nimSelf, Nama: "Budi", Kelas: "TI-1", LastHeartbeat: &learnNow, AttemptID: str("att-1")}
	f.sessions.On("Upsert", mock.Anything, s, true).Return(nil)
	req := &dto.HeartbeatRequest{Nama: "Budi", AttemptID: dto.Optional[string]{Set: true, Value: str("att-1")}}
	res, err := f.sessionUC.Heartbeat(bg, who(usecase.RoleKordas), req)
	require.NoError(t, err)
	assert.True(t, res.SessionSaved)
}

func TestRLS_R_active_sessions_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	req := &dto.HeartbeatRequest{AttemptID: dto.Optional[string]{Set: true, Value: str("att-1")}}
	res, err := f.sessionUC.Heartbeat(bg, who(usecase.RolePraktikan), req)
	require.NoError(t, err)
	assert.False(t, res.SessionSaved)
	f.sessions.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
}

func TestRLS_R_active_sessions_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.sessions.On("DeleteByNIM", mock.Anything, nimSelf).Return(int64(1), nil)
	f.sessions.On("DeleteStale", mock.Anything, learnNow.Add(-5*time.Minute)).Return(int64(4), nil)
	require.NoError(t, f.sessionUC.End(bg, who(usecase.RoleAdmin)))
	require.NoError(t, f.sessionUC.CleanupStale(bg, who(usecase.RoleAsisten), 5))
}

func TestRLS_R_active_sessions_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	require.NoError(t, f.sessionUC.End(bg, who(usecase.RolePraktikan)))
	require.NoError(t, f.sessionUC.CleanupStale(bg, who(usecase.RolePraktikan), 2))
	f.sessions.AssertNotCalled(t, "DeleteByNIM", mock.Anything, mock.Anything)
	f.sessions.AssertNotCalled(t, "DeleteStale", mock.Anything, mock.Anything)
}

func TestRLS_R_activity_logs_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.logs.On("List", mock.Anything, 1, 100).Return([]entity.ActivityLog{{ID: "log-1", EventType: "login"}}, int64(1), nil)
	page, err := f.logUC.List(bg, who(usecase.RoleAsisten), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, "log-1", page.Items[0].ID)
}

func TestRLS_R_activity_logs_SELECT_1_Ditolak(t *testing.T) {
	page, err := newLearn(t).logUC.List(bg, who(usecase.RolePraktikan), 1, 100)
	require.NoError(t, err)
	assert.Empty(t, page.Items)
	assert.Zero(t, page.Total)
}

func TestRLS_R_activity_logs_DELETE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.logs.On("Delete", mock.Anything, "log-1").Return(int64(1), nil)
	require.NoError(t, f.logUC.Delete(bg, who(usecase.RoleKordas), "log-1"))
}

func TestRLS_R_activity_logs_DELETE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	for _, role := range nonAdminRoles {
		require.NoError(t, f.logUC.Delete(bg, who(role), "log-1"))
	}
	f.logs.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// Awal semester Ganjil 2026 (1 Agustus 2026 00:00 WIB) untuk learnNow.
var learnSemesterStart = time.Date(2026, 7, 31, 17, 0, 0, 0, time.UTC)

func TestRLS_R_prestige_history_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("LevelTree", mock.Anything).Return([]repository.LevelLessons{{ID: "lv1", LessonIDs: []string{"a", "b"}}}, nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nimOther).Return([]string{"a"}, nil)
	f.prestige.On("CountSince", mock.Anything, nimOther, learnSemesterStart).Return(int64(1), nil)
	f.users.On("PrestigeState", mock.Anything, nimOther).Return(&repository.PrestigeState{}, nil)
	e, err := f.prestigeUC.Eligibility(bg, who(usecase.RoleAsisten), nimOther)
	require.NoError(t, err)
	assert.Equal(t, int64(1), e.SemesterUsed)
	assert.Equal(t, 1, e.LessonsDone)
	assert.Equal(t, 2, e.TotalLessons)
}

// Praktikan juga tidak bisa membaca tabel lessons (R-lessons-SELECT-1), jadi total pelajaran 0.
func TestRLS_R_prestige_history_SELECT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("LevelTree", mock.Anything).Return([]repository.LevelLessons{{ID: "lv1", LessonIDs: []string{"a", "b"}}}, nil)
	f.users.On("PrestigeState", mock.Anything, nimOther).Return(nil, nil)
	e, err := f.prestigeUC.Eligibility(bg, who(usecase.RolePraktikan), nimOther)
	require.NoError(t, err)
	assert.Zero(t, e.SemesterUsed)
	assert.Zero(t, e.LessonsDone)
	assert.Zero(t, e.TotalLessons)
	f.prestige.AssertNotCalled(t, "CountSince", mock.Anything, mock.Anything, mock.Anything)
}

// prestige_history hanya ditulis do_prestige_reset (SECURITY DEFINER); tidak ada jalur Go yang menulisnya langsung.
func TestRLS_R_prestige_history_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.eligible(nimSelf)
	f.prestige.On("Reset", mock.Anything, nimSelf).Return(`{"tier":"Master","mult":1.5,"prestige_no":2}`, nil)
	f.expectAudit("progress_reset", "Prestige sukarela #2 NIM 202711001: tamat penuh 6/6 → tier Master x1.50 (lesson-only, RPC atomik). Progres diarsip, bukan dihapus.")
	res, err := f.prestigeUC.Reset(bg, who(usecase.RoleKordas), &dto.PrestigeResetRequest{})
	require.NoError(t, err)
	assert.Equal(t, &dto.PrestigeResetResponse{Tier: "Master", Mult: 1.5, PrestigeNo: 2}, res)
}

func TestRLS_R_prestige_history_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.eligible(nimSelf)
	_, err := f.prestigeUC.Reset(bg, who(usecase.RolePraktikan), &dto.PrestigeResetRequest{NIM: nimSelf})
	assertUCError(t, err, http.StatusBadRequest, "Syarat belum terpenuhi: tamat 0/6 level (6/0 lesson). Reset sukarela hanya untuk yang tamat penuh 6/6.")
	f.prestige.AssertNotCalled(t, "Reset", mock.Anything, mock.Anything)
}

func TestRLS_T_do_prestige_reset_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "asisten", nil)
	f.eligible(nimOther)
	f.prestige.On("Reset", mock.Anything, nimOther).Return(`{"tier":"Master","mult":1.5,"prestige_no":1}`, nil)
	f.expectAudit("progress_reset", "Prestige sukarela #1 NIM 202711002: tamat penuh 6/6 → tier Master x1.50 (lesson-only, RPC atomik). Progres diarsip, bukan dihapus.")
	_, err := f.prestigeUC.Reset(bg, who(usecase.RoleAsisten), &dto.PrestigeResetRequest{NIM: nimOther})
	require.NoError(t, err)
}

func TestRLS_T_do_prestige_reset_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "mahasiswa", nil)
	_, err := f.prestigeUC.Reset(bg, who(usecase.RolePraktikan), &dto.PrestigeResetRequest{NIM: nimOther})
	assertUCError(t, err, http.StatusForbidden, "Akses ditolak.")

	f = newLearn(t)
	f.eligible(nimSelf)
	f.prestige.On("Reset", mock.Anything, nimSelf).Return("", &pgconn.PgError{Code: "P0001", Message: "Cooldown 14 hari belum lewat."})
	_, err = f.prestigeUC.Reset(bg, who(usecase.RoleAdmin), &dto.PrestigeResetRequest{NIM: nimSelf})
	assertUCError(t, err, http.StatusBadRequest, "Cooldown 14 hari belum lewat.")
	f.logs.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestRLS_R_lessons_SELECT_2_Diizinkan(t *testing.T) {
	f := newLearn(t)
	f.board.On("Weekly", mock.Anything, repository.LeaderboardFilter{}, learnWeekStart, learnWeekStart.Add(7*24*time.Hour), (*string)(nil)).Return([]repository.WeeklyRow{}, nil)
	_, err := f.boardUC.Weekly(bg, who(usecase.RoleAsisten), repository.LeaderboardFilter{}, 1, 10)
	require.NoError(t, err)
}

// Penampil praktikan: agregat hanya baris miliknya dan reward lesson 60 (embed lessons null), lihat leaderboard_repository.go.
func TestRLS_R_lessons_SELECT_2_Ditolak(t *testing.T) {
	f := newLearn(t)
	f.board.On("Weekly", mock.Anything, repository.LeaderboardFilter{Kelas: "TI-1"}, learnWeekStart, learnWeekStart.Add(7*24*time.Hour), str(nimSelf)).Return([]repository.WeeklyRow{}, nil)
	_, err := f.boardUC.Weekly(bg, who(usecase.RolePraktikan), repository.LeaderboardFilter{Kelas: "TI-1"}, 1, 10)
	require.NoError(t, err)
}

func pngFile() *dto.ImageFile {
	return &dto.ImageFile{Name: "diagram.png", ContentType: "image/png", Content: []byte("png")}
}

func TestRLS_R_assessment_grading_rules_SELECT_1_Diizinkan(t *testing.T) {
	q := easyQuestion()
	p := gradePrompt(t, "post_test", `{"difficulties":{"easy":{"criteria":{"benar":20}}}}`, &q)
	assert.Contains(t, p, "KRITERIA SCORING KHUSUS UNTUK SOAL INI:\n- EASY: benar: 20 poin\n")
}

func TestRLS_R_assessment_grading_rules_SELECT_1_Ditolak(t *testing.T) {
	_, err := newGrade(t, usecase.AIClients{}).uc.Grade(bg, "", gradeReq(`["a1"]`))
	assertMissingToken(t, err)
}

func TestRLS_R_storage_objects_SELECT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	res, err := f.storageUC.UploadLessonImage(who(usecase.RoleAdmin), pngFile())
	require.NoError(t, err)
	assert.Equal(t, "https://storage.test/"+f.uploader.path, res.PublicURL)
}

func TestRLS_R_storage_objects_SELECT_1_Ditolak(t *testing.T) {
	_, err := newLearn(t).storageUC.UploadLessonImage(nil, pngFile())
	assertMissingToken(t, err)
}

func TestRLS_R_storage_objects_INSERT_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	_, err := f.storageUC.UploadLessonImage(who(usecase.RoleAsisten), pngFile())
	require.NoError(t, err)
	assert.Equal(t, "image/png", f.uploader.contentType)
}

// Validasi tipe file jalan lebih dulu, baru syarat staf.
func TestRLS_R_storage_objects_INSERT_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	_, err := f.storageUC.UploadLessonImage(who(usecase.RolePraktikan), pngFile())
	assertUCError(t, err, http.StatusForbidden, "new row violates row-level security policy")
	_, err = f.storageUC.UploadLessonImage(who(usecase.RolePraktikan), &dto.ImageFile{Name: "a.txt", ContentType: "text/plain"})
	assertUCError(t, err, http.StatusBadRequest, "Format file tidak didukung. Harap upload gambar (JPG, PNG, GIF, WEBP).")
	assert.False(t, f.uploader.called)
}

// upsert:false: staf tidak pernah menimpa objek yang sudah ada.
func TestRLS_R_storage_objects_UPDATE_1_Diizinkan(t *testing.T) {
	f := newLearn(t)
	_, err := f.storageUC.UploadLessonImage(who(usecase.RoleKordas), pngFile())
	require.NoError(t, err)
	assert.False(t, f.uploader.upsert)
}

func TestRLS_R_storage_objects_UPDATE_1_Ditolak(t *testing.T) {
	f := newLearn(t)
	_, err := f.storageUC.UploadLessonImage(who(usecase.RolePraktikan), pngFile())
	assertUCError(t, err, http.StatusForbidden, "new row violates row-level security policy")
	assert.False(t, f.uploader.called)
}
