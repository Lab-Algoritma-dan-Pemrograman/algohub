package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type PrestigeHandler struct {
	uc *usecase.PrestigeUsecase
}

func NewPrestigeHandler(uc *usecase.PrestigeUsecase) *PrestigeHandler {
	return &PrestigeHandler{uc: uc}
}

func (h *PrestigeHandler) Eligibility(c *gin.Context) {
	nim := c.Query("nim")
	serve(c, "Kelayakan prestige", func(ctx context.Context, p *dto.Principal) (*dto.PrestigeEligibility, error) {
		return h.uc.Eligibility(ctx, p, nim)
	})
}

func (h *PrestigeHandler) Reset(c *gin.Context) {
	serveBody(c, "Prestige berhasil", h.uc.Reset)
}
