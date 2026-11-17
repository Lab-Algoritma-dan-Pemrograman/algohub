package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

// Cron siakad dan keep-alive dipicu Vercel Cron (Bearer CRON_SECRET), bukan scheduler praktikum.
type CronHandler struct {
	secret    string
	shifts    *usecase.ShiftCronUsecase
	keepAlive *usecase.KeepAliveUsecase
}

func NewCronHandler(secret string, shifts *usecase.ShiftCronUsecase, keepAlive *usecase.KeepAliveUsecase) *CronHandler {
	return &CronHandler{secret: secret, shifts: shifts, keepAlive: keepAlive}
}

func (h *CronHandler) CheckShifts(c *gin.Context) {
	if !h.authorized(c) {
		return
	}
	body, f := h.shifts.Run(c.Request.Context())
	write(c, body, f)
}

func (h *CronHandler) KeepAlive(c *gin.Context) {
	if !h.authorized(c) {
		return
	}
	body, f := h.keepAlive.Run(c.Request.Context())
	write(c, body, f)
}

func (h *CronHandler) authorized(c *gin.Context) bool {
	if middleware.CronAuthorized(c, h.secret) {
		return true
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Access restricted to Vercel Cron only"})
	return false
}

func write(c *gin.Context, body map[string]any, f *usecase.Failure) {
	if f != nil {
		c.JSON(f.Status, f.Body)
		return
	}
	c.JSON(http.StatusOK, body)
}
