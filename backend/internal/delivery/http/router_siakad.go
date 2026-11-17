package http

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	cronhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/cron"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/middleware"
)

type SiakadHandlers struct {
	Auth   *handler.AuthHandler
	RPC    *handler.RPCHandler
	Notify *handler.NotifyHandler
	Cron   *cronhandler.CronHandler
}

type Siakad struct {
	Handlers SiakadHandlers
}

var siakadMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodHead, http.MethodOptions,
}

// Prod lama memakai function terpisah untuk login dan rpc, jadi kuota rate limit-nya juga terpisah.
func RegisterSiakad(r *gin.Engine, s *Siakad) {
	h := s.Handlers
	loginLimit := middleware.SiakadRateLimit(5, time.Minute,
		"Terlalu banyak percobaan login. Silakan coba lagi dalam satu menit.", nil)
	rpcLimit := middleware.SiakadRateLimit(5, time.Minute,
		"Terlalu banyak percobaan. Silakan coba lagi dalam satu menit.", func(c *gin.Context) bool {
			fn := c.Query("fn")
			return fn == "login_user" || fn == "register_user"
		})

	g := r.Group("/api/siakad")
	allowOnly(g, "/auth/login", []string{http.MethodPost}, loginLimit, h.Auth.Login)
	allowOnly(g, "/auth/logout", []string{http.MethodPost}, h.Auth.Logout)
	allowOnly(g, "/auth/verify", []string{http.MethodGet, http.MethodPost}, h.Auth.Verify)
	allowOnly(g, "/rpc", []string{http.MethodPost}, rpcLimit, h.RPC.Call)
	allowOnly(g, "/send-push", []string{http.MethodPost, http.MethodOptions}, okOnOptions(h.Notify.SendPush))
	allowOnly(g, "/send-wa", []string{http.MethodPost, http.MethodOptions}, okOnOptions(h.Notify.SendWA))

	cron := r.Group("/api/cron")
	allowOnly(cron, "/siakad/check-shifts", []string{http.MethodGet}, h.Cron.CheckShifts)
	allowOnly(cron, "/keep-alive", []string{http.MethodGet}, h.Cron.KeepAlive)
}

// Endpoint notifikasi lama membalas OPTIONS non-preflight dengan 200 kosong; preflight sudah
// dijawab middleware CORS sebelum sampai ke sini.
func okOnOptions(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusOK)
			return
		}
		next(c)
	}
}

// Gin membalas 404 untuk method yang tidak terdaftar, sedangkan endpoint lama membalas 405.
func allowOnly(g *gin.RouterGroup, path string, methods []string, handlers ...gin.HandlerFunc) {
	for _, m := range siakadMethods {
		if slices.Contains(methods, m) {
			g.Handle(m, path, handlers...)
			continue
		}
		g.Handle(m, path, methodNotAllowed)
	}
}

func methodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
}
