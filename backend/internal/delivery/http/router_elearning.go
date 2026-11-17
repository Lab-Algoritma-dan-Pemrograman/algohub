package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	elhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/elearning"
	eldto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/response"
)

type ElearningAuthenticator interface {
	Authenticate(token string) (*eldto.Principal, error)
}

type ElearningHandlers struct {
	Auth *elhandler.AuthHandler
}

type Elearning struct {
	Auth     ElearningAuthenticator
	Handlers ElearningHandlers
}

func RegisterElearning(r *gin.Engine, e *Elearning) {
	h := e.Handlers
	// Per IP klien; dulu semua login elearning berbagi IP server Vercel saat memproksi ke backend praktikum.
	loginLimit := middleware.RateLimit(10, time.Minute)
	authmw := middleware.ElearningAuth(e.Auth)

	g := r.Group("/api/elearning")
	g.POST("/sso", h.Auth.SSO)
	g.POST("/receive-token", h.Auth.ReceiveToken)
	g.POST("/login", loginLimit, h.Auth.Login)
	g.POST("/verify", h.Auth.Verify)

	authed := g.Group("", authmw)
	authed.POST("/auth/supabase-token", h.Auth.SupabaseToken)
}

// Tanpa DATABASE_URL_ELEARNING semua path elearning dijawab 503 supaya praktikum tetap bisa jalan.
func RegisterElearningUnavailable(r *gin.Engine) {
	r.Any("/api/elearning/*path", func(c *gin.Context) {
		response.Fail(c, http.StatusServiceUnavailable, "Backend elearning belum dikonfigurasi (DATABASE_URL_ELEARNING).", nil)
	})
}
