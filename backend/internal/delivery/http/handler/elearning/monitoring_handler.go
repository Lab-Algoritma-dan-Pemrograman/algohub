package handler

import (
	"context"
	"math"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type MonitoringHandler struct {
	sessions *usecase.SessionUsecase
	logs     *usecase.ActivityLogUsecase
}

func NewMonitoringHandler(sessions *usecase.SessionUsecase, logs *usecase.ActivityLogUsecase) *MonitoringHandler {
	return &MonitoringHandler{sessions: sessions, logs: logs}
}

func (h *MonitoringHandler) Heartbeat(c *gin.Context) {
	serveBody(c, "Sesi diperbarui", h.sessions.Heartbeat)
}

func (h *MonitoringHandler) EndSession(c *gin.Context) {
	serve(c, "Sesi dihapus", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.sessions.End(ctx, p))
	})
}

func (h *MonitoringHandler) CleanupStale(c *gin.Context) {
	minutes := intQuery(c, "thresholdMinutes", usecase.DefaultStaleMinutes, math.MaxInt32)
	serve(c, "Sesi usang dibersihkan", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.sessions.CleanupStale(ctx, p, minutes))
	})
}

func (h *MonitoringHandler) Sessions(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Sesi aktif", func(ctx context.Context, p *dto.Principal) (*dto.Page[dto.ActiveSessionRole], error) {
		return h.sessions.Sessions(ctx, p, page, limit)
	})
}

func (h *MonitoringHandler) Students(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Daftar mahasiswa", func(ctx context.Context, p *dto.Principal) (*dto.Page[repository.MonitoringStudent], error) {
		return h.sessions.Students(ctx, p, page, limit)
	})
}

func (h *MonitoringHandler) Progress(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Progres mahasiswa", func(ctx context.Context, p *dto.Principal) (*dto.Page[repository.NIMLessons], error) {
		return h.sessions.Progress(ctx, p, page, limit)
	})
}

func (h *MonitoringHandler) Logs(c *gin.Context) {
	page, limit := pageParams(c)
	serve(c, "Log aktivitas", func(ctx context.Context, p *dto.Principal) (*dto.Page[entity.ActivityLog], error) {
		return h.logs.List(ctx, p, page, limit)
	})
}

func (h *MonitoringHandler) DeleteLog(c *gin.Context) {
	id := c.Param("id")
	serve(c, "Log dihapus", func(ctx context.Context, p *dto.Principal) (any, error) {
		return done(h.logs.Delete(ctx, p, id))
	})
}
