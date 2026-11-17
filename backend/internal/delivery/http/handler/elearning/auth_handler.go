package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
)

type AuthHandler struct {
	auth   *usecase.AuthUsecase
	appURL string
}

func NewAuthHandler(a *usecase.AuthUsecase, appURL string) *AuthHandler {
	return &AuthHandler{auth: a, appURL: appURL}
}

// Pengganti POST ke root elearning (api/index.ts). Token panjang tidak lagi ditaruh di URL; SPA menukar code via /receive-token.
func (h *AuthHandler) SSO(c *gin.Context) {
	token := ssoToken(c)
	if token == "" {
		c.Redirect(http.StatusFound, h.appURL+"/")
		return
	}
	code, err := h.auth.SSO(token)
	switch {
	case err == nil:
		c.Redirect(http.StatusSeeOther, h.appURL+"/sso?code="+url.QueryEscape(code))
	case errors.Is(err, usecase.ErrSSOInvalidToken):
		c.Redirect(http.StatusFound, h.appURL+"/?error=invalid_token")
	case errors.Is(err, usecase.ErrSSOMissingFields):
		c.Redirect(http.StatusFound, h.appURL+"/?error=missing_fields")
	default:
		log.Printf("elearning sso: %v", err)
		c.Redirect(http.StatusFound, h.appURL+"/?error=server_error")
	}
}

// Siakad mengirim x-www-form-urlencoded; JSON juga diterima seperti api/index.ts. Jenis lain dianggap tanpa token.
func ssoToken(c *gin.Context) string {
	ct := c.GetHeader("Content-Type")
	switch {
	case strings.Contains(ct, "application/json"):
		var req dto.TokenRequest
		bindLenient(c, &req)
		return req.Token
	case strings.Contains(ct, "application/x-www-form-urlencoded"):
		return c.PostForm("token")
	}
	return ""
}

func (h *AuthHandler) ReceiveToken(c *gin.Context) {
	var req dto.TokenRequest
	bindLenient(c, &req)
	var (
		res *dto.ReceiveTokenResponse
		err error
	)
	if req.Token == "" && req.Code != "" {
		res, err = h.auth.ReceiveCode(req.Code)
	} else {
		res, err = h.auth.ReceiveToken(req.Token)
	}
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Token diterima", res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	bindLenient(c, &req)
	res, err := h.auth.Login(c.Request.Context(), req, middleware.ClientIP(c), c.Request.UserAgent())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Login berhasil", res)
}

func (h *AuthHandler) Verify(c *gin.Context) {
	var req dto.TokenRequest
	bindLenient(c, &req)
	res, err := h.auth.Verify(req.Token)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Token valid", res)
}

func (h *AuthHandler) SupabaseToken(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	res, err := h.auth.SupabaseToken(token, middleware.ElearningPrincipal(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, http.StatusOK, "Token Supabase", res)
}
