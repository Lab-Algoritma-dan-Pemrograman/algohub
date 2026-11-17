package handler

import (
	"net/http"

	_ "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type PraktikumHandler struct {
	uc *usecase.PraktikumUsecase
}

func NewPraktikumHandler(uc *usecase.PraktikumUsecase) *PraktikumHandler {
	return &PraktikumHandler{uc: uc}
}

// @Summary		Dashboard Mahasiswa
// @Description	Mengambil data statistik dashboard khusus mahasiswa
// @Tags			Praktikum
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope{data=dto.DashboardUserResponse}
// @Router			/praktikum/dashboard [get]
func (h *PraktikumHandler) Dashboard(c *gin.Context) {
	res, err := h.uc.Dashboard(middleware.UserID(c))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Dashboard mahasiswa", res)
}

// @Summary		Daftar Sesi Mahasiswa
// @Description	Mengambil daftar sesi praktikum yang tersedia untuk mahasiswa terkait
// @Tags			Praktikum
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope{data=[]dto.SesiUserItem}
// @Router			/praktikum/sesi [get]
func (h *PraktikumHandler) ListSesi(c *gin.Context) {
	res, err := h.uc.ListSesi(middleware.UserID(c))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Daftar sesi", res)
}
