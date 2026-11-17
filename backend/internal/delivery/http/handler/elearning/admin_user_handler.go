package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type AdminUserHandler struct {
	uc *usecase.AdminUserUsecase
}

func NewAdminUserHandler(uc *usecase.AdminUserUsecase) *AdminUserHandler {
	return &AdminUserHandler{uc: uc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	page, limit := pageParams(c)
	q := c.Query("q")
	serve(c, "Daftar pengguna", func(ctx context.Context, p *dto.Principal) (*dto.Page[repository.AdminUser], error) {
		return h.uc.List(ctx, p, page, limit, q)
	})
}

func (h *AdminUserHandler) Progress(c *gin.Context) {
	page, limit := pageParams(c)
	nim := c.Param("nim")
	serve(c, "Progres pengguna", func(ctx context.Context, p *dto.Principal) (*dto.Page[entity.StudentProgress], error) {
		return h.uc.Progress(ctx, p, nim, page, limit)
	})
}

// Satu pola untuk aksi admin yang hanya mengembalikan error: parameter :nim + body JSON.
func adminAction[B any](c *gin.Context, message string, fn func(ctx context.Context, p *dto.Principal, nim string, req *B) error) {
	nim := c.Param("nim")
	serveBody(c, message, func(ctx context.Context, p *dto.Principal, req *B) (any, error) {
		return done(fn(ctx, p, nim, req))
	})
}

func (h *AdminUserHandler) LevelAccess(c *gin.Context) {
	adminAction(c, "Akses level diperbarui", h.uc.LevelAccess)
}

func (h *AdminUserHandler) Role(c *gin.Context) {
	adminAction(c, "Peran diperbarui", h.uc.Role)
}

func (h *AdminUserHandler) ResetLevel(c *gin.Context) {
	adminAction(c, "Progres level direset", h.uc.ResetLevel)
}

func (h *AdminUserHandler) ResetLesson(c *gin.Context) {
	adminAction(c, "Progres pelajaran direset", h.uc.ResetLesson)
}

func (h *AdminUserHandler) Reset(c *gin.Context) {
	nim := c.Param("nim")
	serve(c, "Progres direset", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.uc.Reset(ctx, p, nim))
	})
}

func (h *AdminUserHandler) AdjustXP(c *gin.Context) {
	nim := c.Param("nim")
	serveBody(c, "XP diperbarui", func(ctx context.Context, p *dto.Principal, req *dto.AdjustXPRequest) (*dto.XPResponse, error) {
		return h.uc.AdjustXP(ctx, p, nim, req)
	})
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	nim := c.Param("nim")
	serve(c, "Pengguna dihapus", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.uc.Delete(ctx, p, nim))
	})
}
