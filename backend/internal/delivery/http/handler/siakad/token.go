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
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.Split(h, " ")[1]
	}
	return ""
}
