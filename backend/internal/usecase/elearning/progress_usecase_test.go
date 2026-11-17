package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func wibTime(day, hour, minute int) *time.Time {
	t := time.Date(2026, 11, day, hour, minute, 0, 0, time.FixedZone("WIB", 7*3600))
	return &t
}

func TestCalculateStreak(t *testing.T) {
	now := *wibTime(17, 0, 10)
	cases := []struct {
		name    string
		last    *time.Time
		current int
		want    int
		changed bool
	}{
		{"tanpa last_active", nil, 0, 1, true},
		{"tanpa last_active streak 1", nil, 1, 1, false},
		{"hari yang sama", wibTime(17, 0, 1), 4, 4, false},
		{"hari yang sama streak 0", wibTime(17, 0, 1), 0, 1, true},
		{"kemarin menurut WIB walau UTC sama hari", wibTime(16, 23, 30), 2, 3, true},
		{"lewat dua hari", wibTime(14, 12, 0), 9, 1, true},
		{"jam masa depan", wibTime(19, 8, 0), 3, 3, false},
		{"streak negatif", wibTime(16, 8, 0), -5, 1, true},
	}
	for _, tc := range cases {
		got, changed := usecase.CalculateStreak(tc.last, tc.current, now)
		assert.Equal(t, tc.want, got, tc.name)
		assert.Equal(t, tc.changed, changed, tc.name)
	}
}

func TestCalculateLevelAndPrestigeMult(t *testing.T) {
	for xp, want := range map[int]int{-5: 1, 0: 1, 49: 1, 50: 2, 199: 2, 200: 3, 450: 4} {
		assert.Equal(t, want, usecase.CalculateLevel(xp), "xp %d", xp)
	}
	assert.Equal(t, 60, usecase.ApplyPrestigeMult(60, 0))
	assert.Equal(t, 75, usecase.ApplyPrestigeMult(60, 1.25))
	assert.Equal(t, 90, usecase.ApplyPrestigeMult(60, 3))
	assert.Equal(t, 60, usecase.ApplyPrestigeMult(60, 0.5))
	assert.Equal(t, 60, usecase.ApplyPrestigeMult(60.9, 1))
}

func TestCompleteLessonFlow(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 100, Level: 2, Streak: 2, LastActive: wibTime(16, 20, 0), PrestigeMult: 1.25}, nil)
	f.progress.On("Exists", mock.Anything, nimSelf, "l2").Return(false, nil)
	f.progress.On("Insert", mock.Anything, nimSelf, "l2", learnNow).Return(nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"xp": 175, "level": 2, "streak": 3, "last_active": learnNow}, false).Return(int64(1), nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nimSelf).Return([]string{"l1"}, nil)
	f.curriculum.On("LevelTree", mock.Anything).Return([]repository.LevelLessons{
		{ID: "level-1", Title: "Dasar", LessonIDs: []string{"l1", "l2"}},
		{ID: "level-2", Title: "Lanjut", LessonIDs: []string{"l3"}},
	}, nil)
	f.achievements.On("List", mock.Anything, "").Return([]entity.Achievement{
		{ID: "level-1-done", RequirementType: str("level_completed"), RequirementValue: str("level-1")},
		{ID: "lesson-2", RequirementType: str("lesson_count"), RequirementValue: str("2")},
		{ID: "lesson-3", RequirementType: str("lesson_count"), RequirementValue: str("3")},
		{ID: "streak-3", RequirementType: str("streak"), RequirementValue: str("3")},
		{ID: "xp-150", RequirementType: str("xp"), RequirementValue: str("150")},
	}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{"xp-150"}, nil)
	for _, id := range []string{"level-1-done", "lesson-2", "streak-3"} {
		f.achievements.On("Insert", mock.Anything, nimSelf, id, learnNow).Return(nil).Once()
	}

	res, err := f.progressUC.Complete(bg, who(usecase.RoleKordas), "l2", &dto.CompleteLessonRequest{XPReward: 60})
	require.NoError(t, err)
	assert.Equal(t, 175, res.XP)
	assert.Equal(t, 2, res.Level)
	assert.Equal(t, 3, res.Streak)
	assert.False(t, res.LeveledUp)
	ids := make([]string, len(res.NewlyUnlocked))
	for i, a := range res.NewlyUnlocked {
		ids[i] = a.ID
	}
	assert.Equal(t, []string{"level-1-done", "lesson-2", "streak-3"}, ids)
}

func TestCompleteLessonAlreadyDone(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 100, Level: 2}, nil)
	f.progress.On("Exists", mock.Anything, nimSelf, "l1").Return(true, nil)
	res, err := f.progressUC.Complete(bg, who(usecase.RolePraktikan), "l1", &dto.CompleteLessonRequest{XPReward: 60})
	require.NoError(t, err)
	assert.True(t, res.AlreadyCompleted)
	assert.Equal(t, 100, res.XP)
	assert.Empty(t, res.NewlyUnlocked)
}

func TestGrantXP(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 40, Level: 1}, nil)
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"xp": 60, "level": 2, "last_active": learnNow}, true).Return(int64(1), nil)
	f.achievements.On("List", mock.Anything, "xp").Return([]entity.Achievement{
		{ID: "xp-100", RequirementValue: str("100")},
		{ID: "xp-50", RequirementValue: str("50")},
	}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{}, nil)
	f.achievements.On("Insert", mock.Anything, nimSelf, "xp-50", learnNow).Return(nil)
	res, err := f.progressUC.GrantXP(bg, who(usecase.RolePraktikan), &dto.GrantXPRequest{Amount: 20})
	require.NoError(t, err)
	assert.True(t, res.LeveledUp)
	require.NotNil(t, res.Achievement)
	assert.Equal(t, "xp-50", res.Achievement.ID)
}

func TestStudyTimeIncrement(t *testing.T) {
	f := newLearn(t)
	total := int64(130)
	f.users.On("IncrementStudyTime", mock.Anything, nimSelf, int64(30), true).Return(&total, nil)
	res, err := f.progressUC.StudyTime(bg, who(usecase.RolePraktikan), &dto.StudyTimeRequest{Seconds: 30})
	require.NoError(t, err)
	assert.Equal(t, int64(130), *res.StudyTime)

	res, err = f.progressUC.StudyTime(bg, mismatched(usecase.RolePraktikan), &dto.StudyTimeRequest{Seconds: 30})
	require.NoError(t, err)
	assert.Nil(t, res.StudyTime)
}

func TestStreakHeartbeat(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Streak: 4, LastActive: wibTime(17, 8, 0)}, nil).Once()
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"last_active": learnNow}, true).Return(int64(1), nil).Once()
	res, err := f.progressUC.Streak(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, 4, res.Streak)

	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, Streak: 4, LastActive: wibTime(16, 8, 0)}, nil).Once()
	f.users.On("UpdateUser", mock.Anything, nimSelf, map[string]any{"last_active": learnNow, "streak": 5}, true).Return(int64(1), nil).Once()
	res, err = f.progressUC.Streak(bg, who(usecase.RolePraktikan))
	require.NoError(t, err)
	assert.Equal(t, 5, res.Streak)
}

func TestRecentAndMyLessonsFiltered(t *testing.T) {
	f := newLearn(t)
	f.progress.On("Recent", mock.Anything, nimSelf, 5).Return([]entity.StudentProgress{{LessonID: "l9"}}, nil)
	rows, err := f.progressUC.Recent(bg, who(usecase.RolePraktikan), 5)
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	rows, err = f.progressUC.Recent(bg, mismatched(usecase.RolePraktikan), 5)
	require.NoError(t, err)
	assert.Empty(t, rows)
}
