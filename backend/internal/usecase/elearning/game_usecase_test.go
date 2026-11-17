package usecase_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestGameQuestionsShuffledThenLimited(t *testing.T) {
	f := newLearn(t)
	f.games.On("QuestionsByLanguage", mock.Anything, "python").Return([]entity.GameQuestion{{ID: "q1"}, {ID: "q2"}, {ID: "q3"}}, nil)
	qs, err := f.gameUC.Questions(bg, who(usecase.RolePraktikan), "python", 2)
	require.NoError(t, err)
	assert.Equal(t, []dto.GameQuestion{{ID: "q3"}, {ID: "q2"}}, qs)
}

func TestGameSettingsDefaultOnMissingOrError(t *testing.T) {
	want := &dto.GameSettings{BugHuntActive: boolp(true), BugHuntCActive: boolp(true), BugHuntPythonActive: boolp(true), BugHuntWeeklyLimit: intp(3), BugHuntQuestionCount: 5}
	for _, err := range []error{nil, errors.New("db mati")} {
		f := newLearn(t)
		f.games.On("Settings", mock.Anything).Return(nil, err)
		s, gotErr := f.gameUC.Settings(bg, who(usecase.RolePraktikan))
		require.NoError(t, gotErr)
		assert.Equal(t, want, s)
	}
}

func TestGamePlaysMe(t *testing.T) {
	f := newLearn(t)
	f.games.On("Settings", mock.Anything).Return(&entity.GameSettings{BugHuntActive: boolp(false), BugHuntWeeklyLimit: intp(3)}, nil)
	res, err := f.gameUC.PlaysMe(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, &dto.BugHuntAccess{}, res)

	f = newLearn(t)
	f.games.On("Settings", mock.Anything).Return(&entity.GameSettings{BugHuntActive: boolp(true)}, nil)
	res, err = f.gameUC.PlaysMe(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, &dto.BugHuntAccess{Allowed: true}, res)

	f = newLearn(t)
	f.bugHuntLimit(3)
	f.games.On("CountPlaysSince", mock.Anything, nimSelf, "bug_hunt", learnWeekStart).Return(int64(3), nil)
	res, err = f.gameUC.PlaysMe(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, &dto.BugHuntAccess{PlaysUsed: 3, Limit: 3}, res)
}

func TestGameHistoryWeeklyLimitReached(t *testing.T) {
	f := newLearn(t)
	f.bugHuntLimit(3)
	f.games.On("CountPlaysSince", mock.Anything, nimSelf, "bug_hunt", learnWeekStart).Return(int64(3), nil)
	_, err := f.gameUC.History(bg, who(usecase.RolePraktikan), &dto.GameHistoryRequest{GameType: "bug_hunt", XPEarned: 50})
	assertUCError(t, err, http.StatusBadRequest, "WEEKLY_LIMIT_REACHED")
	f.games.AssertNotCalled(t, "InsertHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGameHistoryUnlimitedLevelUp(t *testing.T) {
	f := newLearn(t)
	f.bugHuntLimit(0)
	f.games.On("InsertHistory", mock.Anything, nimSelf, "bug_hunt", 20, learnNow).Return(nil)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 190, Level: 2, PrestigeMult: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"xp": 210, "level": 3, "streak": 1, "last_active": learnNow}, false).Return(int64(1), nil)
	res, err := f.gameUC.History(bg, who(usecase.RoleAdmin), &dto.GameHistoryRequest{GameType: "bug_hunt", XPEarned: 20})
	require.NoError(t, err)
	assert.Equal(t, &dto.GameHistoryResponse{XP: 210, Level: 3, Streak: 1, LastActive: learnNow, LeveledUp: true}, res)
}

func TestGameSeedSkipsWhenNotEmpty(t *testing.T) {
	f := newLearn(t)
	f.games.On("CountQuestions", mock.Anything).Return(int64(12), nil)
	res, err := f.gameUC.Seed(bg, who(usecase.RolePraktikan), oneQuestion())
	require.NoError(t, err)
	assert.False(t, res.Seeded)
}

func TestGameUpdateSkipsEmptyFields(t *testing.T) {
	f := newLearn(t)
	f.games.On("UpdateQuestion", mock.Anything, "q1", map[string]any{"code": "x = 1"}).Return(int64(1), nil)
	require.NoError(t, f.gameUC.Update(bg, who(usecase.RoleAdmin), "q1", &dto.GameQuestionUpdate{Code: "x = 1"}))
}

// Batas minggu Bug Hunt: Minggu 23:59 WIB masih minggu lama, Senin 00:00 WIB minggu baru.
func TestGamePlaysWeekBoundaryWIB(t *testing.T) {
	cases := []struct{ now, start time.Time }{
		{time.Date(2026, 11, 22, 16, 59, 0, 0, time.UTC), learnWeekStart},
		{time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC), time.Date(2026, 11, 22, 17, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		f := newLearn(t)
		f.bugHuntLimit(3)
		f.games.On("CountPlaysSince", mock.Anything, nimSelf, "bug_hunt", tc.start).Return(int64(0), nil)
		uc := usecase.NewGameUsecase(f.games, f.users, func() time.Time { return tc.now }, nil)
		_, err := uc.PlaysMe(bg, who(usecase.RolePraktikan))
		require.NoError(t, err)
	}
}
