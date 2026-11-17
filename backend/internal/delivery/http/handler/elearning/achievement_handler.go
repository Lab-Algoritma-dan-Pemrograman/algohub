package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type AchievementHandler struct {
	uc *usecase.AchievementUsecase
}

func NewAchievementHandler(uc *usecase.AchievementUsecase) *AchievementHandler {
	return &AchievementHandler{uc: uc}
}

func (h *AchievementHandler) List(c *gin.Context) {
	requirementType := c.Query("requirementType")
	serve(c, "Daftar pencapaian", func(ctx context.Context, p *dto.Principal) ([]dto.Achievement, error) {
		return h.uc.List(ctx, p, requirementType)
	})
}

func (h *AchievementHandler) Mine(c *gin.Context) {
	serve(c, "Pencapaian saya", h.uc.Mine)
}

func (h *AchievementHandler) Check(c *gin.Context) {
	serveBody(c, "Pencapaian diperiksa", h.uc.Check)
}

func (h *AchievementHandler) CheckXP(c *gin.Context) {
	serveBody(c, "Pencapaian XP diperiksa", h.uc.CheckXP)
}
