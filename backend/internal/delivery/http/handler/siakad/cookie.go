package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Ditulis manual karena http.SetCookie membuang titik di depan Domain, padahal cookie harus
// terbaca di semua subdomain algohub.web.id.
func setSessionCookie(c *gin.Context, value string, maxAge int) {
	c.Header("Set-Cookie", sessionCookie+"="+value+"; Domain=.algohub.web.id; Path=/; Max-Age="+strconv.Itoa(maxAge)+"; HttpOnly; Secure; SameSite=Lax")
}

func clearSessionCookie(c *gin.Context) {
	setSessionCookie(c, "", 0)
}
