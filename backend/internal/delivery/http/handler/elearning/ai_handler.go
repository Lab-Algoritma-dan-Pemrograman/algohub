package handler

import (
	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type AIHandler struct {
	ai *usecase.AIUsecase
}

func NewAIHandler(ai *usecase.AIUsecase) *AIHandler { return &AIHandler{ai: ai} }

func (h *AIHandler) Generate(c *gin.Context) {
	var req dto.AIRequest
	bindLenient(c, &req)
	res, err := h.ai.Generate(c.Request.Context(), bearerOr(c, req.Token), &req)
	reply(c, "Respons AI", res, err)
}
