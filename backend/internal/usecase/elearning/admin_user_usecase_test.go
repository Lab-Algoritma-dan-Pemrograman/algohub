package usecase_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Guard service lama memakai peran efektif UI dari DB, bukan peran token.
func TestAdminServiceGuardMessages(t *testing.T) {
	cases := []struct {
		name, dbRole, token, message string
		call                         func(f *learnFixture, p *dto.Principal) error
	}{
		{"reset", "mahasiswa", usecase.RolePraktikan, "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres.", func(f *learnFixture, p *dto.Principal) error {
			return f.adminUC.Reset(bg, p, nimOther)
		}},
		{"reset-level", "mahasiswa", usecase.RolePraktikan, "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres level.", func(f *learnFixture, p *dto.Principal) error {
			return f.adminUC.ResetLevel(bg, p, nimOther, &dto.ResetLevelRequest{LevelID: "level-1"})
		}},
		{"reset-lesson", "mahasiswa", usecase.RolePraktikan, "Akses Ditolak: Hanya staf yang diperbolehkan mereset progres pelajaran.", func(f *learnFixture, p *dto.Principal) error {
			return f.adminUC.ResetLesson(bg, p, nimOther, &dto.ResetLessonRequest{LessonID: "l1"})
		}},
		{"xp", "asisten", usecase.RoleAsisten, "Akses Ditolak: Hanya Admin atau Kordas yang dapat menyesuaikan XP.", func(f *learnFixture, p *dto.Principal) error {
			_, err := f.adminUC.AdjustXP(bg, p, nimOther, &dto.AdjustXPRequest{XP: f64(10)})
			return err
		}},
		{"delete", "asisten", usecase.RoleAsisten, "Akses Ditolak: Hanya Admin atau Kordas yang dapat menghapus pengguna.", func(f *learnFixture, p *dto.Principal) error {
			return f.adminUC.Delete(bg, p, nimOther)
		}},
		{"level-access", "asisten", usecase.RoleAsisten, "Maaf, area ini hanya dapat diakses oleh Administrator atau Editor.", func(f *learnFixture, p *dto.Principal) error {
			return f.adminUC.LevelAccess(bg, p, nimOther, &dto.LevelAccessRequest{LevelAccessOverrides: map[string]any{}})
		}},
	}
	for _, tc := range cases {
		f := newLearn(t)
		f.profileRole(nimSelf, tc.dbRole, nil)
		err := tc.call(f, who(tc.token))
		assertUCError(t, err, http.StatusForbidden, tc.message)
	}
}

func TestAdminRoleRules(t *testing.T) {
	f := newLearn(t)
	err := f.adminUC.Role(bg, who(usecase.RoleKordas), nimSelf, &dto.RoleRequest{Role: "praktikan"})
	assertUCError(t, err, http.StatusForbidden, "Anda tidak dapat mengubah peran akun Anda sendiri.")

	f = newLearn(t)
	f.profileRole(nimSelf, "asisten", nil)
	err = f.adminUC.Role(bg, who(usecase.RoleAsisten), nimOther, &dto.RoleRequest{Role: "admin"})
	assertUCError(t, err, http.StatusForbidden, "Hanya Kordas (Super Admin) yang dapat mengelola akun Admin dan Kordas.")

	f = newLearn(t)
	f.profileRole(nimSelf, "asisten", nil)
	f.users.On("RoleByNIM", mock.Anything, nimOther).Return("koordinator", nil)
	err = f.adminUC.Role(bg, who(usecase.RoleAsisten), nimOther, &dto.RoleRequest{Role: "asisten"})
	assertUCError(t, err, http.StatusForbidden, "Hanya Kordas (Super Admin) yang dapat mengelola akun Admin dan Kordas.")

	f = newLearn(t)
	f.profileRole(nimSelf, "asisten", nil)
	f.users.On("RoleByNIM", mock.Anything, nimOther).Return("", gorm.ErrRecordNotFound)
	err = f.adminUC.Role(bg, who(usecase.RoleAsisten), nimOther, &dto.RoleRequest{Role: "asisten"})
	assertUCError(t, err, http.StatusForbidden, "Anda tidak memiliki otoritas untuk mengubah peran.")

	f = newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"role": "koordinator", "elearning_role": "admin"}, false).Return(int64(1), nil)
	require.NoError(t, f.adminUC.Role(bg, who(usecase.RoleKordas), nimOther, &dto.RoleRequest{Role: "admin"}))

	f = newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"role": "mahasiswa", "elearning_role": nil}, false).Return(int64(1), nil)
	require.NoError(t, f.adminUC.Role(bg, who(usecase.RoleKordas), nimOther, &dto.RoleRequest{Role: "praktikan"}))
}

func TestAdminResetLevelDeductsPerLesson(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.curriculum.On("LessonIDsByLevel", mock.Anything, "level-1").Return([]string{"l1", "l2", "l3"}, nil)
	f.progress.On("LessonIDsIn", mock.Anything, nimOther, []string{"l1", "l2", "l3"}).Return([]string{"l1", "l3"}, nil)
	f.progress.On("DeleteByNIMLessons", mock.Anything, nimOther, []string{"l1", "l3"}).Return(int64(2), nil)
	f.users.On("XPLevel", mock.Anything, nimOther).Return(&repository.XPLevel{XP: 250, Level: 3}, nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"xp": 130, "level": 2}, false).Return(int64(1), nil)
	f.expectAudit("progress_reset", `Mereset progres level "level-1" untuk mahasiswa NIM: `+nimOther)
	require.NoError(t, f.adminUC.ResetLevel(bg, who(usecase.RoleKordas), nimOther, &dto.ResetLevelRequest{LevelID: "level-1"}))
}

func TestAdminResetLevelNotFoundAndNothingDone(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.curriculum.On("LessonIDsByLevel", mock.Anything, "x").Return(nil, gorm.ErrRecordNotFound)
	err := f.adminUC.ResetLevel(bg, who(usecase.RoleKordas), nimOther, &dto.ResetLevelRequest{LevelID: "x"})
	assertUCError(t, err, http.StatusNotFound, "Level x not found")

	f = newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.curriculum.On("LessonIDsByLevel", mock.Anything, "level-2").Return([]string{"l9"}, nil)
	f.progress.On("LessonIDsIn", mock.Anything, nimOther, []string{"l9"}).Return([]string{}, nil)
	require.NoError(t, f.adminUC.ResetLevel(bg, who(usecase.RoleKordas), nimOther, &dto.ResetLevelRequest{LevelID: "level-2"}))
}

func TestAdminResetLessonAlwaysDeducts60(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.progress.On("DeleteByNIMLessons", mock.Anything, nimOther, []string{"l1"}).Return(int64(0), nil)
	f.users.On("XPLevel", mock.Anything, nimOther).Return(&repository.XPLevel{XP: 40, Level: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"xp": 0, "level": 1}, false).Return(int64(1), nil)
	f.expectAudit("progress_reset", `Mereset progres pelajaran "Perulangan" untuk mahasiswa NIM: `+nimOther)
	require.NoError(t, f.adminUC.ResetLesson(bg, who(usecase.RoleKordas), nimOther, &dto.ResetLessonRequest{LessonID: "l1", LessonTitle: str("Perulangan")}))
}

func TestAdminAdjustXP(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.users.On("XPLevel", mock.Anything, nimOther).Return(&repository.XPLevel{XP: 0, Level: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"xp": 201, "level": 3}, false).Return(int64(1), nil)
	f.expectAudit("xp_adjusted", "Mengubah XP mahasiswa NIM: "+nimOther+" menjadi 201 (Level: 3)")
	f.achievements.On("List", mock.Anything, "xp").Return([]entity.Achievement{{ID: "xp-50", RequirementValue: str("50")}}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimOther).Return([]string{}, nil)
	res, err := f.adminUC.AdjustXP(bg, who(usecase.RoleKordas), nimOther, &dto.AdjustXPRequest{XP: f64(200.5)})
	require.NoError(t, err)
	assert.Equal(t, 201, res.XP)
	assert.True(t, res.LeveledUp)
	assert.Nil(t, res.Achievement, "insert unlocked untuk NIM lain ditolak WITH CHECK")
}

func TestAdminLevelAccessAndDelete(t *testing.T) {
	f := newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.users.On("UpdateUser", mock.Anything, nimOther, map[string]any{"level_access_overrides": datatypes.JSON(`{"level-1":"unlocked"}`)}, false).Return(int64(1), nil)
	require.NoError(t, f.adminUC.LevelAccess(bg, who(usecase.RoleKordas), nimOther, &dto.LevelAccessRequest{LevelAccessOverrides: map[string]any{"level-1": "unlocked"}}))

	f = newLearn(t)
	f.profileRole(nimSelf, "koordinator", nil)
	f.expectAudit("user_deleted", "Menghapus akun mahasiswa NIM: "+nimOther)
	f.users.On("Delete", mock.Anything, nimOther).Return(int64(1), nil)
	require.NoError(t, f.adminUC.Delete(bg, who(usecase.RoleKordas), nimOther))
}
