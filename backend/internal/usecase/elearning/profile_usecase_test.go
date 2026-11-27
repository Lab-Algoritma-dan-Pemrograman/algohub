package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestProfileSyncCreatesMissingProfile(t *testing.T) {
	f := newLearn(t)
	p := who(usecase.RolePraktikan)
	p.Shift = str("2")
	p.Jurusan = str("")
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(nil, nil)
	f.users.On("Insert", mock.Anything, mock.MatchedBy(func(c map[string]any) bool {
		return c["nim"] == nimSelf && c["role"] == "praktikan" && c["jurusan"] == (*string)(nil) && *(c["shift"].(*string)) == "2" &&
			c["streak"] == 1 && c["xp"] == 0 && string(c["assessment_access"].(datatypes.JSON)) == `{"pre_test":false,"post_test":false,"program_keterampilan":false,"ujian_praktik":false}`
	}), true).Return(errors.New("duplicate key"))

	res, err := f.profileUC.Sync(bg, p)
	require.NoError(t, err, "error insert ditelan seperti SupabaseProvider")
	assert.Equal(t, "praktikan", res.Profile.Role)
	assert.Equal(t, 1, res.Profile.Streak)
	assert.InDelta(t, 1.0, res.Profile.PrestigeMult, 0)
	assert.JSONEq(t, `{}`, string(res.Profile.LevelAccessOverrides))
	assert.Empty(t, res.NewlyUnlocked)
}

// Shift berbeda membuat update ikut menulis shift; column grant menolak seluruh statement dan error ditelan (T-26).
func TestProfileSyncShiftChangeSwallowedByGrant(t *testing.T) {
	f := newLearn(t)
	p := who(usecase.RolePraktikan)
	p.Shift = str("2")
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Nama: "Lama", Kelas: "TI-1", Role: "mahasiswa", Shift: str("1"), Streak: 1, LastActive: wibTime(17, 8, 0), PrestigeMult: 1}, nil)

	res, err := f.profileUC.Sync(bg, p)
	require.NoError(t, err)
	assert.Equal(t, "2", *res.Profile.Shift)
	assert.Equal(t, p.Nama, res.Profile.Nama)
	assert.True(t, res.Profile.LastActive.Equal(learnNow))
	f.users.AssertNotCalled(t, "UpdateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestProfileSyncUpdatesAndRoleChange(t *testing.T) {
	f := newLearn(t)
	p := who(usecase.RoleAsisten)
	p.Kelas = ""
	p.Jurusan = str("Informatika")
	old := &entity.User{NIM: nimSelf, Nama: "Nama " + nimSelf, Kelas: "TI-1", Jurusan: str("Sistem Informasi"), Role: "mahasiswa", Streak: 2, LastActive: wibTime(16, 8, 0), PrestigeMult: 1, XP: 10}
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(old, nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{
		"last_active": learnNow, "streak": 3, "role": "asisten", "jurusan": "Informatika",
	}, true).Return(int64(1), nil)
	f.achievements.On("List", mock.Anything, "").Return([]entity.Achievement{{ID: "streak-3", RequirementType: str("streak"), RequirementValue: str("3")}}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{}, nil)
	f.progress.On("CountCompleted", mock.Anything, nimSelf).Return(int64(0), nil)
	f.achievements.On("Insert", mock.Anything, nimSelf, "streak-3", learnNow).Return(nil)

	res, err := f.profileUC.Sync(bg, p)
	require.NoError(t, err)
	assert.Equal(t, "asisten", res.Profile.Role)
	assert.Equal(t, "Informatika", *res.Profile.Jurusan)
	assert.Empty(t, res.Profile.Kelas)
	require.Len(t, res.NewlyUnlocked, 1)
}

// elearning_role admin mengalahkan role token dan role tidak ditulis balik.
func TestProfileSyncElearningRoleWins(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Nama: "Nama " + nimSelf, Kelas: "TI-1", Role: "asisten", ElearningRole: str("admin"), Streak: 1, LastActive: wibTime(17, 1, 0), PrestigeMult: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"last_active": learnNow, "streak": 1}, true).Return(int64(1), nil)
	res, err := f.profileUC.Sync(bg, who(usecase.RoleAsisten))
	require.NoError(t, err)
	assert.Equal(t, "admin", res.Profile.Role)
}

// T-30: token tanpa kelas tidak boleh menulis "kelas" kosong, karena trigger lalu mengosongkan class_code dan kelas_id.
func TestProfileSyncIgnoresEmptyKelas(t *testing.T) {
	for _, tokenKelas := range []string{"", "   "} {
		f := newLearn(t)
		p := who(usecase.RolePraktikan)
		p.Kelas = tokenKelas
		f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Nama: "Nama " + nimSelf, Kelas: "TI-1", Role: "mahasiswa", Streak: 1, LastActive: wibTime(17, 8, 0), PrestigeMult: 1}, nil)
		f.users.On("UpdateUser", mock.Anything, nimSelf, mock.MatchedBy(func(cols map[string]any) bool {
			_, ada := cols["kelas"]
			return !ada
		}), true).Return(int64(1), nil)

		_, err := f.profileUC.Sync(bg, p)
		require.NoError(t, err)
		f.users.AssertExpectations(t)
	}
}

func TestProfileSyncRequiresPrincipal(t *testing.T) {
	_, err := newLearn(t).profileUC.Sync(bg, nil)
	assertMissingToken(t, err)
}
