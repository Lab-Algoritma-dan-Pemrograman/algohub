package handler

import (
	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

type GradeHandler struct {
	grade *usecase.GradeUsecase
}

func NewGradeHandler(g *usecase.GradeUsecase) *GradeHandler { return &GradeHandler{grade: g} }

func (h *GradeHandler) Grade(c *gin.Context) {
	var req dto.GradeRequest
	bindLenient(c, &req)
	res, err := h.grade.Grade(c.Request.Context(), bearerOr(c, req.Token), &req)
	reply(c, "Penilaian selesai", res, err)
}
