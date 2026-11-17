package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

// Dashboard.tsx:86 memakai limit(5).
const defaultRecentLimit = 5

type ProgressHandler struct {
	uc *usecase.ProgressUsecase
}

func NewProgressHandler(uc *usecase.ProgressUsecase) *ProgressHandler {
	return &ProgressHandler{uc: uc}
}

func (h *ProgressHandler) MyLessons(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Pelajaran selesai", func(ctx context.Context, p *dto.Principal) (*dto.Page[string], error) {
		return h.uc.MyLessons(ctx, p, page, limit)
	})
}

func (h *ProgressHandler) Recent(c *gin.Context) {
	limit := intQuery(c, "limit", defaultRecentLimit, maxPageLimit)
	serve(c, "Aktivitas terbaru", func(ctx context.Context, p *dto.Principal) ([]entity.StudentProgress, error) {
		return h.uc.Recent(ctx, p, limit)
	})
}

func (h *ProgressHandler) Complete(c *gin.Context) {
	lessonID := c.Param("lessonId")
	serveBody(c, "Pelajaran selesai", func(ctx context.Context, p *dto.Principal, req *dto.CompleteLessonRequest) (*dto.CompleteLessonResponse, error) {
		return h.uc.Complete(ctx, p, lessonID, req)
	})
}

func (h *ProgressHandler) GrantXP(c *gin.Context) {
	serveBody(c, "XP ditambahkan", h.uc.GrantXP)
}

func (h *ProgressHandler) StudyTime(c *gin.Context) {
	serveBody(c, "Waktu belajar disimpan", h.uc.StudyTime)
}

func (h *ProgressHandler) Streak(c *gin.Context) {
	serve(c, "Streak diperbarui", h.uc.Streak)
}
