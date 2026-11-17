package handler

import (
	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type ReportHandler struct {
	report *usecase.ReportUsecase
}

func NewReportHandler(r *usecase.ReportUsecase) *ReportHandler { return &ReportHandler{report: r} }

func (h *ReportHandler) Sync(c *gin.Context) {
	var req dto.ReportRequest
	bindLenient(c, &req)
	res, err := h.report.Sync(c.Request.Context(), bearerOr(c, req.Token), req.NIM)
	reply(c, "Progres disinkronkan", res, err)
}
