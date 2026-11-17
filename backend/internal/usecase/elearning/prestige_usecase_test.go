package usecase_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Syarat 6/6 level selesai sudah terpenuhi dengan satu pelajaran per level.
func (f *learnFixture) eligibleWith(nim string, state *repository.PrestigeState, used int64, semesterStart time.Time) {
	tree := make([]repository.LevelLessons, 6)
	ids := make([]string, 6)
	for i := range tree {
		ids[i] = fmt.Sprintf("l%d", i)
		tree[i] = repository.LevelLessons{ID: fmt.Sprintf("lv%d", i), LessonIDs: []string{ids[i]}}
	}
	f.curriculum.On("LevelTree", mock.Anything).Return(tree, nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nim).Return(ids, nil)
	f.prestige.On("CountSince", mock.Anything, nim, semesterStart).Return(used, nil)
	f.users.On("PrestigeState", mock.Anything, nim).Return(state, nil)
}

func (f *learnFixture) eligible(nim string) {
	f.eligibleWith(nim, &repository.PrestigeState{Count: 1}, 0, learnSemesterStart)
}

func TestPrestigeEligibilityReasons(t *testing.T) {
	f := newLearn(t)
	f.eligible(nimSelf)
	e, err := f.prestigeUC.Eligibility(bg, who(usecase.RoleAdmin), "")
	require.NoError(t, err)
	assert.Equal(t, &dto.PrestigeEligibility{
		Eligible: true, Reason: "Layak prestige: tamat penuh 6/6 → Master x1.50 lesson-only. Badge: Prestige-2. Papan mingguan tetap base.",
		LevelsDone: 6, TotalLevels: 6, LessonsDone: 6, TotalLessons: 6, Tier: "Master", Mult: 1.5, PrestigeCount: 1,
	}, e)

	f = newLearn(t)
	last := learnNow.Add(-10*24*time.Hour + time.Hour)
	f.eligibleWith(nimSelf, &repository.PrestigeState{Count: 1, LastPrestigeAt: &last}, 0, learnSemesterStart)
	e, err = f.prestigeUC.Eligibility(bg, who(usecase.RoleAdmin), nimSelf)
	require.NoError(t, err)
	assert.False(t, e.Eligible)
	assert.Equal(t, "Cooldown: reset berikutnya bisa dilakukan 5 hari lagi (jeda 14 hari antar reset).", e.Reason)

	f = newLearn(t)
	f.eligibleWith(nimSelf, &repository.PrestigeState{}, 2, learnSemesterStart)
	e, err = f.prestigeUC.Eligibility(bg, who(usecase.RoleAdmin), nimSelf)
	require.NoError(t, err)
	assert.Equal(t, "Kuota semester habis (2/2). Reset lagi semester depan (Ganjil Agu 2026–Jan 2027).", e.Reason)
	assert.Equal(t, "Master", e.Tier)
}

func TestPrestigeEligibilityIncomplete(t *testing.T) {
	f := newLearn(t)
	f.curriculum.On("LevelTree", mock.Anything).Return([]repository.LevelLessons{{ID: "lv1", LessonIDs: []string{"a", "b"}}, {ID: "lv2"}}, nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nimSelf).Return([]string{"a", "b", "b", "lama"}, nil)
	f.prestige.On("CountSince", mock.Anything, nimSelf, learnSemesterStart).Return(int64(0), nil)
	f.users.On("PrestigeState", mock.Anything, nimSelf).Return(nil, nil)
	e, err := f.prestigeUC.Eligibility(bg, who(usecase.RoleAsisten), nimSelf)
	require.NoError(t, err)
	assert.Equal(t, "Syarat belum terpenuhi: tamat 1/2 level (3/2 lesson). Reset sukarela hanya untuk yang tamat penuh 6/6.", e.Reason)
	assert.Equal(t, "Base", e.Tier)
	assert.Equal(t, 1.0, e.Mult)
}

// Batas semester mengikuti kalender WIB: 1 Februari 00:00 WIB sudah masuk Genap.
func TestPrestigeSemesterBoundaryWIB(t *testing.T) {
	cases := []struct {
		now   time.Time
		start time.Time
		label string
	}{
		{time.Date(2027, 1, 31, 16, 59, 0, 0, time.UTC), learnSemesterStart, "Ganjil Agu 2026–Jan 2027"},
		{time.Date(2027, 1, 31, 17, 0, 0, 0, time.UTC), time.Date(2027, 1, 31, 17, 0, 0, 0, time.UTC), "Genap Feb–Jul 2027"},
		{time.Date(2027, 7, 31, 17, 0, 0, 0, time.UTC), time.Date(2027, 7, 31, 17, 0, 0, 0, time.UTC), "Ganjil Agu 2027–Jan 2028"},
	}
	for _, tc := range cases {
		f := newLearn(t)
		f.eligibleWith(nimSelf, &repository.PrestigeState{}, 2, tc.start)
		uc := usecase.NewPrestigeUsecase(f.users, f.curriculum, f.progress, f.prestige, f.logUC, func() time.Time { return tc.now })
		e, err := uc.Eligibility(bg, who(usecase.RoleAdmin), nimSelf)
		require.NoError(t, err)
		assert.Equal(t, "Kuota semester habis (2/2). Reset lagi semester depan ("+tc.label+").", e.Reason)
	}
}

func TestPrestigeResetIgnoresAuditFailureAndDefaults(t *testing.T) {
	f := newLearn(t)
	f.eligible(nimSelf)
	f.prestige.On("Reset", mock.Anything, nimSelf).Return(`{"prestige_no":"3"}`, nil)
	f.logs.On("Insert", mock.Anything, mock.Anything).Return(errors.New("db mati"))
	res, err := f.prestigeUC.Reset(bg, who(usecase.RoleAdmin), &dto.PrestigeResetRequest{NIM: nimSelf})
	require.NoError(t, err)
	assert.Equal(t, &dto.PrestigeResetResponse{Tier: "Master", Mult: 1.5, PrestigeNo: 3}, res)
}

func TestPrestigeResetNotEligible(t *testing.T) {
	f := newLearn(t)
	f.eligibleWith(nimSelf, &repository.PrestigeState{}, 2, learnSemesterStart)
	_, err := f.prestigeUC.Reset(bg, who(usecase.RoleKordas), &dto.PrestigeResetRequest{})
	assertUCError(t, err, http.StatusBadRequest, "Kuota semester habis (2/2). Reset lagi semester depan (Ganjil Agu 2026–Jan 2027).")
	f.prestige.AssertNotCalled(t, "Reset", mock.Anything, mock.Anything)
}
