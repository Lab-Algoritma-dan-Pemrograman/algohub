package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type ProfileHandler struct {
	profile *usecase.ProfileUsecase
	logs    *usecase.ActivityLogUsecase
}

func NewProfileHandler(profile *usecase.ProfileUsecase, logs *usecase.ActivityLogUsecase) *ProfileHandler {
	return &ProfileHandler{profile: profile, logs: logs}
}

func (h *ProfileHandler) Sync(c *gin.Context) {
	serve(c, "Profil disinkronkan", h.profile.Sync)
}

func (h *ProfileHandler) LogActivity(c *gin.Context) {
	serveBody(c, "Aktivitas dicatat", func(ctx context.Context, p *dto.Principal, req *dto.ActivityLogRequest) (any, error) {
		return done(h.logs.Insert(ctx, p, req))
	})
}
