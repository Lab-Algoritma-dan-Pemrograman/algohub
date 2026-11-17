package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
)

// Secret kosong selalu ditolak supaya salah konfigurasi tidak membuka endpoint.
// Perbandingan constant-time agar secret tidak bisa ditebak lewat waktu respons.
func SecretMatches(got, secret string) bool {
	return secret != "" && subtle.ConstantTimeCompare([]byte(got), []byte(secret)) == 1
}

// Vercel Cron mengirim Authorization: Bearer <CRON_SECRET>; scheduler eksternal memakai X-Cron-Secret.
func CronAuthorized(c *gin.Context, secret string) bool {
	if bearer, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok && SecretMatches(bearer, secret) {
		return true
	}
	return SecretMatches(c.GetHeader("X-Cron-Secret"), secret)
}
