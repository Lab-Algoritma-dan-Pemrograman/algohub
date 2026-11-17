package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

const sessionMaxAge = 3600

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// Token juga dikirim di body untuk klien lama yang masih memakai localStorage.
func (h *AuthHandler) Login(c *gin.Context) {
	req := bindBody[dto.LoginRequest](c)
	body, token, f := h.uc.Login(c.Request.Context(), req.Username, req.Password)
	if f != nil {
		c.JSON(f.Status, f.Body)
		return
	}
	setSessionCookie(c, token, sessionMaxAge)
	c.JSON(http.StatusOK, body)
}

// LOW-01: 204 tanpa body karena endpoint ini hanya menghapus cookie dan mencabut token,
// jadi tidak ada informasi yang perlu dikembalikan ke pemanggil tanpa auth.
func (h *AuthHandler) Logout(c *gin.Context) {
	clearSessionCookie(c)
	h.uc.Logout(c.Request.Context(), sessionToken(c))
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Verify(c *gin.Context) {
	body, clearCookie, f := h.uc.Verify(c.Request.Context(), sessionToken(c))
	if clearCookie {
		clearSessionCookie(c)
	}
	if f != nil {
		c.JSON(f.Status, f.Body)
		return
	}
	c.JSON(http.StatusOK, body)
}
