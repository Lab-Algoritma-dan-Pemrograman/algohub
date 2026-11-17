package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type CurriculumHandler struct {
	uc *usecase.CurriculumUsecase
}

func NewCurriculumHandler(uc *usecase.CurriculumUsecase) *CurriculumHandler {
	return &CurriculumHandler{uc: uc}
}

func (h *CurriculumHandler) Levels(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Daftar level", func(ctx context.Context, p *dto.Principal) (*dto.Page[entity.Level], error) {
		return h.uc.Levels(ctx, p, page, limit)
	})
}

func (h *CurriculumHandler) Modules(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Daftar modul", func(ctx context.Context, p *dto.Principal) (*dto.Page[entity.Module], error) {
		return h.uc.Modules(ctx, p, page, limit)
	})
}

func (h *CurriculumHandler) Lessons(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Daftar pelajaran", func(ctx context.Context, p *dto.Principal) (any, error) {
		return h.uc.Lessons(ctx, p, page, limit)
	})
}

func (h *CurriculumHandler) Save(c *gin.Context) {
	serveBody(c, "Kurikulum disimpan", func(ctx context.Context, p *dto.Principal, req *dto.CurriculumRequest) (any, error) {
		return done(h.uc.Save(ctx, p, req))
	})
}

func (h *CurriculumHandler) Clear(c *gin.Context) {
	serve(c, "Kurikulum dikosongkan", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.uc.Clear(ctx, p))
	})
}
