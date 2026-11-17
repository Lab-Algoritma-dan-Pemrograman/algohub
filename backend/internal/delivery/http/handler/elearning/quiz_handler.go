package handler

import (
	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type QuizHandler struct {
	quiz *usecase.QuizUsecase
}

func NewQuizHandler(q *usecase.QuizUsecase) *QuizHandler { return &QuizHandler{quiz: q} }

// Hanya header Bearer, tanpa fallback body, sama dengan api/validate-quiz.ts.
func (h *QuizHandler) Validate(c *gin.Context) {
	var req dto.ValidateQuizRequest
	bindLenient(c, &req)
	res, err := h.quiz.Validate(c.Request.Context(), c.GetHeader("Authorization"), &req)
	reply(c, "Jawaban diperiksa", res, err)
}
