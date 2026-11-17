package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type PlaygroundHandler struct {
	uc *usecase.PlaygroundUsecase
}

func NewPlaygroundHandler(uc *usecase.PlaygroundUsecase) *PlaygroundHandler {
	return &PlaygroundHandler{uc: uc}
}

func (h *PlaygroundHandler) List(c *gin.Context) {
	language := c.Query("language")
	page, limit := pageParams(c)
	serve(c, "Contoh playground", func(ctx context.Context, p *dto.Principal) (*dto.Page[dto.PlaygroundExample], error) {
		return h.uc.List(ctx, p, language, page, limit)
	})
}

func (h *PlaygroundHandler) Add(c *gin.Context) {
	serveBody(c, "Contoh playground ditambahkan", h.uc.Add)
}

func (h *PlaygroundHandler) Update(c *gin.Context) {
	id := c.Param("id")
	serveBody(c, "Contoh playground diperbarui", func(ctx context.Context, p *dto.Principal, req *dto.PlaygroundUpdate) (any, error) {
		return done(h.uc.Update(ctx, p, id, req))
	})
}

func (h *PlaygroundHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	serve(c, "Contoh playground dihapus", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.uc.Delete(ctx, p, id))
	})
}
