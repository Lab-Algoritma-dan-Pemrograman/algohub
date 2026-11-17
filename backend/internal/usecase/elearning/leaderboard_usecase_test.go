package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestLeaderboardListMapsLikeOldProfile(t *testing.T) {
	f := newLearn(t)
	filter := repository.LeaderboardFilter{Kelas: "TI-1", Shift: "2"}
	f.board.On("Leaderboard", mock.Anything, filter, 1, 20).Return([]repository.LeaderboardUser{{NIM: nimOther, Nama: str("Budi"), Shift: str("2"), XP: intp(300)}}, int64(1), nil)
	page, err := f.boardUC.List(bg, who(usecase.RolePraktikan), filter, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []dto.LeaderboardEntry{{
		NIM: nimOther, Nama: str("Budi"), Shift: str("2"), XP: 300, Level: 1, Role: "praktikan", AssessmentAccess: map[string]any{},
	}}, page.Items)
}

func TestLeaderboardRank(t *testing.T) {
	f := newLearn(t)
	res, err := f.boardUC.Rank(bg, who(usecase.RolePraktikan), 0, repository.LeaderboardFilter{})
	require.NoError(t, err)
	assert.Nil(t, res.Rank)

	f.board.On("CountAbove", mock.Anything, 120.0, repository.LeaderboardFilter{Jurusan: "TI"}).Return(int64(4), nil)
	res, err = f.boardUC.Rank(bg, who(usecase.RolePraktikan), 120, repository.LeaderboardFilter{Jurusan: "TI"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), *res.Rank)
}

func TestLeaderboardFiltersDistinctSorted(t *testing.T) {
	f := newLearn(t)
	f.board.On("Filters", mock.Anything).Return([]repository.FilterValue{
		{Kind: "kelas", Value: "TI-2"}, {Kind: "kelas", Value: "TI-1"}, {Kind: "kelas", Value: ""},
		{Kind: "shift", Value: "2"}, {Kind: "shift", Value: "1"}, {Kind: "jurusan", Value: "SI"},
	}, nil)
	res, err := f.boardUC.Filters(bg, who(usecase.RoleAsisten))
	require.NoError(t, err)
	assert.Equal(t, &dto.LeaderboardFilters{Kelas: []string{"TI-1", "TI-2"}, Jurusan: []string{"SI"}, Shift: []string{"1", "2"}}, res)
}

func TestWeeklySortedByXPThenNameAndPaged(t *testing.T) {
	f := newLearn(t)
	end := learnWeekStart.Add(7 * 24 * time.Hour)
	f.board.On("Weekly", mock.Anything, repository.LeaderboardFilter{}, learnWeekStart, end, (*string)(nil)).Return([]repository.WeeklyRow{
		{NIM: "1", Nama: str("Citra"), LessonXP: 60},
		{NIM: "2", Nama: str("budi"), LessonXP: 60},
		{NIM: "3", Nama: str("Andi"), LessonXP: 60, BughuntXP: 40, LessonsCompleted: 1, Level: intp(4), Streak: intp(2)},
	}, nil)
	page, err := f.boardUC.Weekly(bg, who(usecase.RoleKordas), repository.LeaderboardFilter{}, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), page.Total)
	require.Len(t, page.Items, 2)
	assert.Equal(t, "3", page.Items[0].NIM)
	assert.Equal(t, 100, page.Items[0].WeeklyXP)
	assert.Equal(t, 100, page.Items[0].XP)
	assert.Equal(t, 4, page.Items[0].Level)
	assert.Equal(t, "2", page.Items[1].NIM, "budi sebelum Citra seperti localeCompare, bukan urutan byte")

	page, err = f.boardUC.Weekly(bg, who(usecase.RoleKordas), repository.LeaderboardFilter{}, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, "1", page.Items[0].NIM)
}

// Minggu 19-25 Okt 2026: semua entry mendapat floor(50% XP minggu lalu) di atas XP minggu ini.
func TestWeeklyCarryBonusWeek(t *testing.T) {
	f := newLearn(t)
	start := time.Date(2026, 10, 18, 17, 0, 0, 0, time.UTC)
	prevStart := start.Add(-7 * 24 * time.Hour)
	viewer := str(nimSelf)
	f.board.On("Weekly", mock.Anything, repository.LeaderboardFilter{}, start, start.Add(7*24*time.Hour), viewer).Return([]repository.WeeklyRow{
		{NIM: "a", Nama: str("A"), LessonXP: 10}, {NIM: "b", Nama: str("B")}, {NIM: "c", Nama: str("C")},
	}, nil)
	f.board.On("Weekly", mock.Anything, repository.LeaderboardFilter{}, prevStart, start, viewer).Return([]repository.WeeklyRow{
		{NIM: "a", LessonXP: 30}, {NIM: "b", BughuntXP: 7},
	}, nil)
	uc := usecase.NewLeaderboardUsecase(f.board, func() time.Time { return time.Date(2026, 10, 21, 3, 0, 0, 0, time.UTC) })
	page, err := uc.Weekly(bg, who(usecase.RolePraktikan), repository.LeaderboardFilter{}, 1, 10)
	require.NoError(t, err)
	got := map[string]dto.WeeklyEntry{}
	for _, e := range page.Items {
		got[e.NIM] = e
	}
	assert.Equal(t, 25, got["a"].WeeklyXP)
	assert.Equal(t, 25, got["a"].XP)
	assert.Equal(t, 15, got["a"].CarryXP)
	assert.True(t, got["a"].IsCarry)
	assert.Equal(t, 3, got["b"].WeeklyXP)
	assert.False(t, got["c"].IsCarry)
	assert.Equal(t, []string{"a", "b", "c"}, []string{page.Items[0].NIM, page.Items[1].NIM, page.Items[2].NIM})
}
