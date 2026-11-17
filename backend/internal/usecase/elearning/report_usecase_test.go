package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type reportFixture struct {
	uc         *usecase.ReportUsecase
	curriculum *mocks.CurriculumRepository
	progress   *mocks.ProgressRepository
	reports    *mocks.ReportRepository
}

func newReport(t *testing.T) *reportFixture {
	t.Helper()
	f := &reportFixture{curriculum: mocks.NewCurriculumRepository(t), progress: mocks.NewProgressRepository(t), reports: mocks.NewReportRepository(t)}
	f.uc = usecase.NewReportUsecase(newAuth(testConfig(), nil), f.curriculum, f.progress, f.reports, func() time.Time { return gradeNow })
	return f
}

func TestReportValidation(t *testing.T) {
	ctx := context.Background()
	tok := supabaseToken(t, jwt.MapClaims{"nim": "202711001"})

	_, err := newReport(t).uc.Sync(ctx, "", "202711001")
	assertUCError(t, err, http.StatusBadRequest, "Missing token or nim")
	_, err = newReport(t).uc.Sync(ctx, tok, nil)
	assertUCError(t, err, http.StatusBadRequest, "Missing token or nim")
	_, err = newReport(t).uc.Sync(ctx, "rusak", "202711001")
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")
	_, err = newReport(t).uc.Sync(ctx, tok, "202711002")
	assertUCError(t, err, http.StatusForbidden, "Identity mismatch")
	_, err = newReport(t).uc.Sync(ctx, tok, float64(202711001))
	assertUCError(t, err, http.StatusForbidden, "Identity mismatch")
	_, err = newReport(t).uc.Sync(ctx, supabaseToken(t, jwt.MapClaims{"username": "202711001"}), "202711001")
	assertUCError(t, err, http.StatusForbidden, "Identity mismatch")
}

func TestReportCalculation(t *testing.T) {
	f := newReport(t)
	nim := "202711001"
	f.curriculum.On("CountLessons", mock.Anything).Return(int64(6), nil)
	f.progress.On("CountCompleted", mock.Anything, nim).Return(int64(4), nil)
	f.curriculum.On("LevelTree", mock.Anything).Return([]repository.LevelLessons{
		{ID: "l0", Title: "Kosong"},
		{ID: "l1", Title: "Dasar", LessonIDs: []string{"x1", "x2"}},
		{ID: "l2", Title: "Menengah", LessonIDs: []string{"x3", "x4"}},
		{ID: "l3", Title: "Lanjut", LessonIDs: []string{"x5"}},
		{ID: "l4", Title: "Ahli", LessonIDs: []string{"x6"}},
	}, nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, nim).Return([]string{"x1", "x2", "x3", "x6"}, nil)
	f.reports.On("Upsert", mock.Anything, mock.MatchedBy(func(p *entity.ElearningProgress) bool {
		return p.NIM == nim && *p.StudentName == "Siti" && p.LessonsCompleted == 4 && p.TotalLessons == 6 &&
			p.CompletionPercentage == 66.67 && !p.IsCompleted && string(p.CompletedLevels) == `["Dasar","Ahli"]` &&
			*p.CurrentLevel == "Menengah" && p.LastAccessedAt.Equal(gradeNow)
	})).Return(nil)

	res, err := f.uc.Sync(context.Background(), supabaseToken(t, jwt.MapClaims{"nim": nim, "full_name": "Siti"}), nim)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, int64(4), res.Recalculated.Completed)
	assert.Equal(t, int64(6), res.Recalculated.Total)
	assert.InDelta(t, 66.67, res.Recalculated.Percentage, 0)
}

func TestReportDefaultsAndEmptyCurriculum(t *testing.T) {
	f := newReport(t)
	f.curriculum.On("CountLessons", mock.Anything).Return(int64(0), nil)
	f.progress.On("CountCompleted", mock.Anything, "1").Return(int64(0), nil)
	f.curriculum.On("LevelTree", mock.Anything).Return(nil, nil)
	f.progress.On("CompletedLessonIDs", mock.Anything, "1").Return(nil, nil)
	f.reports.On("Upsert", mock.Anything, mock.MatchedBy(func(p *entity.ElearningProgress) bool {
		return *p.StudentName == "Student" && p.CompletionPercentage == 0 && !p.IsCompleted &&
			string(p.CompletedLevels) == `[]` && *p.CurrentLevel == "Introduction"
	})).Return(nil)

	res, err := f.uc.Sync(context.Background(), supabaseToken(t, jwt.MapClaims{"nim": "1"}), "1")
	require.NoError(t, err)
	assert.Zero(t, res.Recalculated.Percentage)
}

func TestReportDBError(t *testing.T) {
	f := newReport(t)
	f.curriculum.On("CountLessons", mock.Anything).Return(int64(0), errors.New("koneksi putus"))
	_, err := f.uc.Sync(context.Background(), supabaseToken(t, jwt.MapClaims{"nim": "1"}), "1")
	assertUCError(t, err, http.StatusInternalServerError, "koneksi putus")
}
