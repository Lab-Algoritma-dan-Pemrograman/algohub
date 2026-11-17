package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Body 429 memakai {"error": message} mentah karena frontend siakad lama membaca field error.
// applies nil berarti semua request dihitung.
func SiakadRateLimit(limit int, win time.Duration, message string, applies func(*gin.Context) bool) gin.HandlerFunc {
	rl := &rateLimiter{
		hits:   make(map[string]*window),
		limit:  limit,
		window: win,
	}
	rl.startJanitor()

	return func(c *gin.Context) {
		if applies != nil && !applies(c) {
			c.Next()
			return
		}
		if !rl.allow(ClientIP(c)) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": message})
			return
		}
		c.Next()
	}
}
