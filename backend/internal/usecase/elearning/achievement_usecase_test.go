package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func f64(v float64) *float64 { return &v }

func TestAchievementListFallsBackToEmbeddedJSON(t *testing.T) {
	f := newLearn(t)
	f.achievements.On("List", mock.Anything, "").Return(nil, errors.New("tabel tidak ada")).Once()
	f.achievements.On("List", mock.Anything, "streak").Return([]entity.Achievement{}, nil).Once()

	all, err := f.achievementUC.List(bg, who(usecase.RolePraktikan), "")
	require.NoError(t, err)
	require.NotEmpty(t, all)
	assert.Equal(t, "xp-50", all[0].ID)
	assert.InDelta(t, 50.0, all[0].RequirementValue, 0)

	streaks, err := f.achievementUC.List(bg, who(usecase.RolePraktikan), "streak")
	require.NoError(t, err)
	require.NotEmpty(t, streaks)
	for _, a := range streaks {
		assert.Equal(t, "streak", *a.RequirementType)
	}
}

func TestAchievementRequirementRules(t *testing.T) {
	f := newLearn(t)
	f.users.On("GetProfile", mock.Anything, nimSelf).Return(&entity.User{NIM: nimSelf, XP: 10, Streak: 7}, nil)
	f.achievements.On("List", mock.Anything, "").Return([]entity.Achievement{
		{ID: "perfect-hunt", RequirementType: str("game_score"), RequirementValue: str("1")},
		{ID: "hunter-3", RequirementType: str("game_score"), RequirementValue: str("3")},
		{ID: "streak-7", RequirementType: str("streak"), RequirementValue: str("7")},
		{ID: "level-1", RequirementType: str("level_completed"), RequirementValue: str("level-1")},
		{ID: "level-angka", RequirementType: str("level_completed"), RequirementValue: str("1")},
		{ID: "xp-teks", RequirementType: str("xp"), RequirementValue: str("banyak")},
		{ID: "lesson-5", RequirementType: str("lesson_count"), RequirementValue: str("5")},
	}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{"streak-7"}, nil)
	f.achievements.On("Insert", mock.Anything, nimSelf, "hunter-3", learnNow).Return(nil).Once()
	f.achievements.On("Insert", mock.Anything, nimSelf, "level-1", learnNow).Return(errors.New("duplikat")).Once()
	f.achievements.On("Insert", mock.Anything, nimSelf, "lesson-5", learnNow).Return(nil).Once()

	got, err := f.achievementUC.Check(bg, who(usecase.RolePraktikan), &dto.AchievementCheckRequest{
		GamesPlayed: f64(3), PerfectGames: f64(0), LessonCount: f64(5), CompletedLevelIDs: []string{"level-1", "1"},
	})
	require.NoError(t, err)
	ids := []string{}
	for _, a := range got {
		ids = append(ids, a.ID)
	}
	assert.Equal(t, []string{"hunter-3", "lesson-5"}, ids)
}

func TestCheckXPFirstUnlockedByThreshold(t *testing.T) {
	f := newLearn(t)
	f.achievements.On("List", mock.Anything, "xp").Return([]entity.Achievement{
		{ID: "xp-500", RequirementValue: str("500")},
		{ID: "xp-100", RequirementValue: str("100")},
		{ID: "xp-50", RequirementValue: str("50")},
		{ID: "xp-rusak", RequirementValue: str("x")},
	}, nil)
	f.achievements.On("UnlockedIDs", mock.Anything, nimSelf).Return([]string{"xp-50"}, nil)
	f.achievements.On("Insert", mock.Anything, nimSelf, "xp-100", learnNow).Return(nil)
	res, err := f.achievementUC.CheckXP(bg, who(usecase.RolePraktikan), &dto.CheckXPRequest{XP: 120})
	require.NoError(t, err)
	require.NotNil(t, res.Achievement)
	assert.Equal(t, "xp-100", res.Achievement.ID)
}

func TestCheckXPNoAchievementWhenListFails(t *testing.T) {
	f := newLearn(t)
	f.achievements.On("List", mock.Anything, "xp").Return(nil, errors.New("putus"))
	res, err := f.achievementUC.CheckXP(bg, who(usecase.RolePraktikan), &dto.CheckXPRequest{XP: 120})
	require.NoError(t, err)
	assert.Nil(t, res.Achievement)
}
