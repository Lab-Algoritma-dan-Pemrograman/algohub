package handler

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
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

// Endpoint pengganti query browser tidak punya padanan pesan lama untuk body rusak; body rusak tidak boleh
// dibaca sebagai field kosong karena PUT /curriculum kosong berarti menghapus seluruh kurikulum.
func bindJSON[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return nil, false
	}
	return &req, true
}

const maxPageLimit = 100

// Keputusan 14: default dan maksimum 100; nilai tidak valid jatuh ke default.
func intQuery(c *gin.Context, name string, def, maxValue int) int {
	v, err := strconv.Atoi(c.Query(name))
	if err != nil || v <= 0 {
		return def
	}
	return min(v, maxValue)
}

func pageParams(c *gin.Context) (page, limit int) {
	return intQuery(c, "page", 1, math.MaxInt32), intQuery(c, "limit", maxPageLimit, maxPageLimit)
}

func serve[T any](c *gin.Context, message string, fn func(ctx context.Context, p *dto.Principal) (T, error)) {
	res, err := fn(c.Request.Context(), middleware.ElearningPrincipal(c))
	reply(c, message, res, err)
}

func serveBody[B, T any](c *gin.Context, message string, fn func(ctx context.Context, p *dto.Principal, req *B) (T, error)) {
	req, ok := bindJSON[B](c)
	if !ok {
		return
	}
	serve(c, message, func(ctx context.Context, p *dto.Principal) (T, error) { return fn(ctx, p, req) })
}

func done(err error) (any, error) { return nil, err }
