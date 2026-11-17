package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

type RPCHandler struct {
	uc *usecase.RPCUsecase
}

func NewRPCHandler(uc *usecase.RPCUsecase) *RPCHandler {
	return &RPCHandler{uc: uc}
}

func (h *RPCHandler) Call(c *gin.Context) {
	fn := c.Query("fn")
	if fn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing function name"})
		return
	}
	params, ok := rpcParams(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	out, f := h.uc.Execute(c.Request.Context(), fn, sessionToken(c), params)
	if f != nil {
		c.JSON(f.Status, f.Body)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", out)
}

// Body selain objek JSON ditolak supaya fungsi DB tidak dijalankan dengan parameter tebakan.
func rpcParams(c *gin.Context) (map[string]json.RawMessage, bool) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, false
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]json.RawMessage{}, true
	}
	var params map[string]json.RawMessage
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, false
	}
	return params, true
}
