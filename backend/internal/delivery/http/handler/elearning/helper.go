package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
)

func fail(c *gin.Context, err error) {
	var e *usecase.Error
	if errors.As(err, &e) {
		response.Fail(c, e.Status, e.Message, e.Detail)
		return
	}
	log.Printf("elearning: %v", err)
	response.Fail(c, http.StatusInternalServerError, "Terjadi kesalahan server", nil)
}

func reply(c *gin.Context, message string, data any, err error) {
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, message, data)
}

// Urutan api/grade.ts: header Bearer dulu, lalu field token di body. Header "Bearer " kosong tidak jatuh ke body.
func bearerOr(c *gin.Context, bodyToken string) string {
	if token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok {
		return token
	}
	return bodyToken
}

// Body kosong atau rusak diperlakukan sebagai field kosong supaya validasi dan pesannya sama dengan api/*.ts lama.
func bindLenient(c *gin.Context, dst any) {
	_ = c.ShouldBindJSON(dst)
}
