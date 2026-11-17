package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const sessionCookie = "siakad_token"

// Cookie httpOnly diutamakan; header Authorization tetap diterima untuk klien lama yang
// menyimpan token di localStorage.
func sessionToken(c *gin.Context) string {
	if ck, err := c.Request.Cookie(sessionCookie); err == nil && ck.Value != "" {
		return ck.Value
	}
	token, _ := bearerToken(c)
	return token
}

// present membedakan header tanpa skema Bearer (401 "Missing") dari token kosong/rusak (401 "Invalid")
// pada endpoint lama yang hanya membaca header.
func bearerToken(c *gin.Context) (token string, present bool) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	return strings.Split(h, " ")[1], true
}
