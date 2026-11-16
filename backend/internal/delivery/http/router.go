package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
)

func NewRouter(cfg *config.Config) (*gin.Engine, error) {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.Use(
		middleware.CORS(middleware.NewOriginMatcher(cfg.CORSOrigins, cfg.CORSOriginPatterns)),
		gin.Logger(),
		gin.Recovery(),
	)

	r.GET("/api/health", health)
	return r, nil
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "algohub-api"})
}
