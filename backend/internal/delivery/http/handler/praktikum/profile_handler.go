package handler

import (
	"net/http"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	_ "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	uc *usecase.ProfileUsecase
}

func NewProfileHandler(uc *usecase.ProfileUsecase) *ProfileHandler { return &ProfileHandler{uc: uc} }

// @Summary		Lihat Profil
// @Description	Mengambil data profil user yang sedang login
// @Tags			General
// @Security		bearerAuth
// @Produce		json
// @Success		200	{object}	response.Envelope{data=entity.User}
// @Router			/praktikum/profile [get]
func (h *ProfileHandler) Get(c *gin.Context) {
	res, err := h.uc.Get(middleware.UserID(c))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Profil", res)
}

// @Summary		Perbarui Profil
// @Description	Mengubah data profil user yang sedang login (termasuk password)
// @Tags			General
// @Security		bearerAuth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.UpdateProfileRequest	true	"Payload Update Profil"
// @Success		200		{object}	response.Envelope{data=entity.User}
// @Router			/praktikum/profile [put]
func (h *ProfileHandler) Update(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	res, err := h.uc.Update(middleware.UserID(c), req)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Profil diperbarui", res)
}
