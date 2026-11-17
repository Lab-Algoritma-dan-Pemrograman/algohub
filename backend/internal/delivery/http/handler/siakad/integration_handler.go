package handler

import (
	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

type IntegrationHandler struct {
	schedule *usecase.ScheduleAIUsecase
	sso      *usecase.SSOUsecase
	sync     *usecase.ElearningSyncUsecase
}

func NewIntegrationHandler(schedule *usecase.ScheduleAIUsecase, sso *usecase.SSOUsecase, sync *usecase.ElearningSyncUsecase) *IntegrationHandler {
	return &IntegrationHandler{schedule: schedule, sso: sso, sync: sync}
}

func (h *IntegrationHandler) AnalyzeSchedule(c *gin.Context) {
	token, ok := bearerToken(c)
	req := bindBody[dto.AnalyzeScheduleRequest](c)
	body, f := h.schedule.Analyze(c.Request.Context(), token, ok, req)
	writeResult(c, body, f)
}

func (h *IntegrationHandler) GenerateJWT(c *gin.Context) {
	req := bindBody[dto.GenerateJWTRequest](c)
	body, f := h.sso.Generate(c.Request.Context(), req.HandshakeCode)
	writeResult(c, body, f)
}

func (h *IntegrationHandler) SyncElearning(c *gin.Context) {
	token, ok := bearerToken(c)
	body, f := h.sync.Sync(c.Request.Context(), token, ok)
	writeResult(c, body, f)
}
