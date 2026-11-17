package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning/mocks"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func newQuiz(t *testing.T) (*usecase.QuizUsecase, *mocks.CurriculumRepository) {
	t.Helper()
	repo := mocks.NewCurriculumRepository(t)
	return usecase.NewQuizUsecase(newAuth(testConfig(), nil), repo), repo
}

func quizBearer(t *testing.T) string {
	return "Bearer " + supabaseToken(t, jwt.MapClaims{"nim": "202711001"})
}

func quizReq(lessonID, selected string) *dto.ValidateQuizRequest {
	return &dto.ValidateQuizRequest{LessonID: json.RawMessage(lessonID), SelectedOption: json.RawMessage(selected)}
}

func TestValidateQuizAuth(t *testing.T) {
	ctx := context.Background()
	uc, _ := newQuiz(t)
	_, err := uc.Validate(ctx, "", quizReq(`"l1"`, `1`))
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Missing token")
	_, err = uc.Validate(ctx, "Token abc", quizReq(`"l1"`, `1`))
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Missing token")
	_, err = uc.Validate(ctx, "Bearer rusak", quizReq(`"l1"`, `1`))
	assertUCError(t, err, http.StatusInternalServerError, "Internal server error")
	_, err = uc.Validate(ctx, "Bearer "+supabaseToken(t, jwt.MapClaims{"sub": "202711001"}), quizReq(`"l1"`, `1`))
	assertUCError(t, err, http.StatusUnauthorized, "Unauthorized: Invalid token")
}

func TestValidateQuizRequiredFields(t *testing.T) {
	uc, _ := newQuiz(t)
	for _, req := range []*dto.ValidateQuizRequest{quizReq("", `1`), quizReq(`""`, `1`), quizReq(`0`, `1`), quizReq(`"l1"`, "")} {
		_, err := uc.Validate(context.Background(), quizBearer(t), req)
		assertUCError(t, err, http.StatusBadRequest, "Missing required fields (lessonId, selectedOption)")
	}
}

func TestValidateQuizLessonLookup(t *testing.T) {
	ctx := context.Background()
	t.Run("tidak ada -> 404", func(t *testing.T) {
		uc, repo := newQuiz(t)
		repo.On("LessonQuiz", mock.Anything, "l404").Return(nil, gorm.ErrRecordNotFound)
		_, err := uc.Validate(ctx, quizBearer(t), quizReq(`"l404"`, `1`))
		assertUCError(t, err, http.StatusNotFound, "Lesson not found")
	})
	t.Run("error postgres -> 500 dengan SQLSTATE", func(t *testing.T) {
		uc, repo := newQuiz(t)
		repo.On("LessonQuiz", mock.Anything, "l1").Return(nil, &pgconn.PgError{Code: "42P01", Message: `relation "lessons" does not exist`})
		_, err := uc.Validate(ctx, quizBearer(t), quizReq(`"l1"`, `1`))
		detail := assertUCError(t, err, http.StatusInternalServerError, "Server tidak dapat membaca bank soal. Periksa konfigurasi Supabase di server.")
		assert.Equal(t, map[string]any{"code": "42P01"}, detail)
	})
	t.Run("error lain -> 500 tanpa teks koneksi", func(t *testing.T) {
		uc, repo := newQuiz(t)
		repo.On("LessonQuiz", mock.Anything, "7").Return(nil, errors.New("failed to connect to `user=postgres.ref database=postgres`: dial tcp: timeout"))
		_, err := uc.Validate(ctx, quizBearer(t), quizReq(`7`, `1`))
		detail := assertUCError(t, err, http.StatusInternalServerError, "Server tidak dapat membaca bank soal. Periksa konfigurasi Supabase di server.")
		assert.Equal(t, map[string]any{"code": ""}, detail)
	})
}

func TestValidateQuizContent(t *testing.T) {
	ctx := context.Background()
	quiz := func(t *testing.T, stored *string, selected string) (*dto.ValidateQuizResponse, error) {
		uc, repo := newQuiz(t)
		repo.On("LessonQuiz", mock.Anything, "l1").Return(stored, nil)
		return uc.Validate(ctx, quizBearer(t), quizReq(`"l1"`, selected))
	}

	_, err := quiz(t, strp(`{"correctAnswer":`), `1`)
	assertUCError(t, err, http.StatusInternalServerError, "Data kuis rusak (bukan JSON yang valid)")

	for _, stored := range []*string{nil, strp(`null`), strp(`{"question":"x"}`), strp(`[1]`), strp(`"teks"`)} {
		_, err = quiz(t, stored, `1`)
		assertUCError(t, err, http.StatusBadRequest, "Lesson does not have a quiz configured")
	}

	res, err := quiz(t, strp(`{"correctAnswer":"2"}`), `2`)
	require.NoError(t, err)
	assert.True(t, res.IsCorrect)
	assert.InDelta(t, 2, *res.CorrectAnswer, 0)

	res, err = quiz(t, strp(`{"correctAnswer":1}`), `"2"`)
	require.NoError(t, err)
	assert.False(t, res.IsCorrect)

	var nullReq dto.ValidateQuizRequest
	require.NoError(t, json.Unmarshal([]byte(`{"lessonId":"l1","selectedOption":null}`), &nullReq))
	uc, repo := newQuiz(t)
	repo.On("LessonQuiz", mock.Anything, "l1").Return(strp(`{"correctAnswer":0}`), nil)
	res, err = uc.Validate(ctx, quizBearer(t), &nullReq)
	require.NoError(t, err)
	assert.True(t, res.IsCorrect, "selectedOption null bukan absen, dan Number(null) === 0")

	res, err = quiz(t, strp(`{"correctAnswer":null}`), `0`)
	require.NoError(t, err)
	assert.True(t, res.IsCorrect, "correctAnswer null tetap terkonfigurasi")

	res, err = quiz(t, strp(`{"correctAnswer":"abc"}`), `"abc"`)
	require.NoError(t, err)
	assert.False(t, res.IsCorrect, "NaN !== NaN")
	assert.Nil(t, res.CorrectAnswer)

	body, _ := json.Marshal(res)
	assert.JSONEq(t, `{"isCorrect":false,"correctAnswer":null}`, string(body))
}
