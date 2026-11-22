package middleware

import (
	"github.com/gin-gonic/gin"
)

// ClientIP mengambil IP asli klien lewat mekanisme Gin.
//
// Router memakai Engine.TrustedPlatform = "X-Vercel-Forwarded-For" (lihat
// router.go), header yang DISET OLEH Vercel dan tidak bisa di-spoof client.
// Gin membacanya lebih dulu di c.ClientIP(); di luar Vercel ia jatuh ke
// RemoteIP, jadi dev lokal tetap dapat IP nyata.
//
// CATATAN: jangan pernah baca X-Forwarded-For/X-Real-IP mentah di sini.
// SetTrustedProxies(nil) membuat semua request dianggap datang dari proxy tak
// terpercaya, jadi XFF mentah diabaikan Gin — mem-bypass-nya berarti membuka
// spoofing IP dan mem-bypass rate limit per-IP.
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}
