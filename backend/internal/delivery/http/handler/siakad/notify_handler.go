package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

type NotifyHandler struct {
	push       *usecase.PushUsecase
	wa         *usecase.WAUsecase
	cronSecret string
}

func NewNotifyHandler(push *usecase.PushUsecase, wa *usecase.WAUsecase, cronSecret string) *NotifyHandler {
	return &NotifyHandler{push: push, wa: wa, cronSecret: cronSecret}
}

func (h *NotifyHandler) SendPush(c *gin.Context) {
	req := bindBody[dto.SendPushRequest](c)
	body, f := h.push.Send(c.Request.Context(), h.fromCron(c), sessionToken(c), req)
	writeResult(c, body, f)
}

func (h *NotifyHandler) SendWA(c *gin.Context) {
	req := bindBody[dto.SendWARequest](c)
	body, f := h.wa.Send(c.Request.Context(), h.fromCron(c), sessionToken(c), req)
	writeResult(c, body, f)
}

// Hanya X-Cron-Secret yang diterima di sini, sama dengan kode lama; Bearer dipakai untuk token sesi.
func (h *NotifyHandler) fromCron(c *gin.Context) bool {
	return middleware.SecretMatches(c.GetHeader("X-Cron-Secret"), h.cronSecret)
}

func writeResult(c *gin.Context, body any, f *usecase.Failure) {
	if f != nil {
		c.JSON(f.Status, f.Body)
		return
	}
	c.JSON(http.StatusOK, body)
}
