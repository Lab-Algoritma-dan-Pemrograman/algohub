package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

type QuizUsecase struct {
	auth       *AuthUsecase
	curriculum repository.CurriculumRepository
}

func NewQuizUsecase(auth *AuthUsecase, curriculum repository.CurriculumRepository) *QuizUsecase {
	return &QuizUsecase{auth: auth, curriculum: curriculum}
}

func (uc *QuizUsecase) Validate(ctx context.Context, authHeader string, req *dto.ValidateQuizRequest) (*dto.ValidateQuizResponse, error) {
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return nil, Unauthorized("Unauthorized: Missing token")
	}
	claims, err := uc.auth.verifyToken(token)
	if err != nil {
		// Di validate-quiz.ts verifyToken yang gagal jatuh ke catch umum, jadi token rusak dijawab 500 (T-12).
		log.Printf("elearning validate-quiz: %v", err)
		return nil, newError(http.StatusInternalServerError, "Internal server error")
	}
	if !truthy(claims["nim"]) {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}

	lessonID := optionalJS(req.LessonID)
	selected := optionalJS(req.SelectedOption)
	if _, absent := selected.(jsUndefined); !jsTruthy(lessonID) || absent {
		return nil, BadRequest("Missing required fields (lessonId, selectedOption)")
	}

	quiz, err := uc.curriculum.LessonQuiz(ctx, jsToString(lessonID))
	// Bedakan "soal tidak ada" dari "backend salah konfigurasi". Dulu keduanya 404 'Lesson not found',
	// sehingga env Supabase yang hilang di Vercel menyamar sebagai soal yang tidak ditemukan (PGRST116 = tidak ada baris).
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("Lesson not found")
	}
	if err != nil {
		log.Printf("elearning validate-quiz: query lesson gagal: %v", err)
		e := newError(http.StatusInternalServerError, "Server tidak dapat membaca bank soal. Periksa konfigurasi Supabase di server.")
		e.Detail = map[string]any{"code": dbErrorCode(err)}
		return nil, e
	}

	// Kolom lessons.quiz bertipe TEXT berisi JSON (bukan jsonb). Tanpa parse, correctAnswer selalu
	// undefined dan setiap kuis dianggap tidak terkonfigurasi.
	var parsed any
	if quiz != nil {
		if parsed, err = parseJS([]byte(*quiz)); err != nil {
			return nil, newError(http.StatusInternalServerError, "Data kuis rusak (bukan JSON yang valid)")
		}
	}
	answer := prop(parsed, "correctAnswer")
	if _, missing := answer.(jsUndefined); missing {
		return nil, BadRequest("Lesson does not have a quiz configured")
	}

	correct := jsNumber(answer)
	res := &dto.ValidateQuizResponse{IsCorrect: jsNumber(selected) == correct}
	// JSON.stringify mengubah NaN/Infinity menjadi null dan -0 menjadi 0.
	if !math.IsNaN(correct) && !math.IsInf(correct, 0) {
		if correct == 0 {
			correct = 0
		}
		res.CorrectAnswer = &correct
	}
	return res, nil
}

func optionalJS(raw json.RawMessage) any {
	if len(raw) == 0 {
		return undefined
	}
	v, err := parseJS(raw)
	if err != nil {
		return undefined
	}
	return v
}

func dbErrorCode(err error) string {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code
	}
	return err.Error()
}
