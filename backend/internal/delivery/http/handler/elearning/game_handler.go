package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Bawaan limit getGameQuestions(language, limit = 5).
const defaultQuestionLimit = 5

type GameHandler struct {
	uc *usecase.GameUsecase
}

func NewGameHandler(uc *usecase.GameUsecase) *GameHandler {
	return &GameHandler{uc: uc}
}

func (h *GameHandler) Questions(c *gin.Context) {
	language := c.Query("language")
	limit := intQuery(c, "limit", defaultQuestionLimit, maxPageLimit)
	serve(c, "Soal game", func(ctx context.Context, p *dto.Principal) ([]dto.GameQuestion, error) {
		return h.uc.Questions(ctx, p, language, limit)
	})
}

func (h *GameHandler) Seed(c *gin.Context) {
	serveBody(c, "Seed soal game", h.uc.Seed)
}

func (h *GameHandler) Reset(c *gin.Context) {
	serveBody(c, "Soal game direset", func(ctx context.Context, p *dto.Principal, req *dto.GameQuestionsRequest) (any, error) {
		return done(h.uc.ForceReset(ctx, p, req))
	})
}

func (h *GameHandler) Add(c *gin.Context) {
	serveBody(c, "Soal game ditambahkan", h.uc.Add)
}

func (h *GameHandler) Update(c *gin.Context) {
	id := c.Param("id")
	serveBody(c, "Soal game diperbarui", func(ctx context.Context, p *dto.Principal, req *dto.GameQuestionUpdate) (any, error) {
		return done(h.uc.Update(ctx, p, id, req))
	})
}

func (h *GameHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	serve(c, "Soal game dihapus", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.uc.Delete(ctx, p, id))
	})
}

func (h *GameHandler) Settings(c *gin.Context) {
	serve(c, "Pengaturan game", h.uc.Settings)
}

func (h *GameHandler) UpdateSettings(c *gin.Context) {
	serveBody(c, "Pengaturan game disimpan", func(ctx context.Context, p *dto.Principal, req *dto.GameSettingsUpdate) (any, error) {
		return done(h.uc.UpdateSettings(ctx, p, req))
	})
}

func (h *GameHandler) PlaysMe(c *gin.Context) {
	serve(c, "Kuota Bug Hunt", h.uc.PlaysMe)
}

func (h *GameHandler) History(c *gin.Context) {
	serveBody(c, "Hasil game disimpan", h.uc.History)
}
