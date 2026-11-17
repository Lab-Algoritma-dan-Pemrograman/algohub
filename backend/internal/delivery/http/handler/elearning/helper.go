package handler

import (
	"errors"
	"log"
	"net/http"

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

// Body kosong atau rusak diperlakukan sebagai field kosong supaya validasi dan pesannya sama dengan api/*.ts lama.
func bindLenient(c *gin.Context, dst any) {
	_ = c.ShouldBindJSON(dst)
}
