package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func mapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, usecase.ErrUnauthorized):
		response.Fail(c, http.StatusUnauthorized, "NIM atau password salah", nil)
	case errors.Is(err, usecase.ErrForbidden):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, usecase.ErrConflict):
		response.Fail(c, http.StatusConflict, "Data sudah ada / konflik", nil)
	case errors.Is(err, usecase.ErrHasReferences):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, usecase.ErrAlreadyDone):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, usecase.ErrTimeUp):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, usecase.ErrBadRequest):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	default:
		response.Fail(c, http.StatusInternalServerError, "Terjadi kesalahan server", err.Error())
	}
}

func idParam(c *gin.Context, name string) (int, bool) {
	v, err := strconv.Atoi(c.Param(name))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Parameter "+name+" tidak valid", nil)
		return 0, false
	}
	return v, true
}

func queryIntPtr(c *gin.Context, name string) *int {
	s := c.Query(name)
	if s == "" {
		return nil
	}
	if v, err := strconv.Atoi(s); err == nil {
		return &v
	}
	return nil
}

func bindJSON[T any](c *gin.Context) (*T, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return nil, false
	}
	return &req, true
}

// Audit log bersifat best-effort: gagal mencatat tidak boleh menggagalkan aksi.
func logAudit(c *gin.Context, a *usecase.AuditLogUsecase, action, desc string) {
	_ = a.LogAction(middleware.UserID(c), "", action, desc, middleware.ClientIP(c), c.Request.UserAgent())
}

// ok=false berarti response error sudah dikirim.
func bindCreate[T, R any](c *gin.Context, create func(*T) (R, error)) (req *T, res R, ok bool) {
	if req, ok = bindJSON[T](c); !ok {
		return req, res, false
	}
	res, err := create(req)
	if err != nil {
		mapError(c, err)
		return req, res, false
	}
	return req, res, true
}

func bindUpdate[T, R any](c *gin.Context, param string, update func(int, *T) (R, error)) (res R, ok bool) {
	id, ok := idParam(c, param)
	if !ok {
		return res, false
	}
	_, res, ok = bindCreate(c, func(req *T) (R, error) { return update(id, req) })
	return res, ok
}
