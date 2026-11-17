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
	Auth        *elhandler.AuthHandler
	AI          *elhandler.AIHandler
	Grade       *elhandler.GradeHandler
	Report      *elhandler.ReportHandler
	Quiz        *elhandler.QuizHandler
	Curriculum  *elhandler.CurriculumHandler
	Progress    *elhandler.ProgressHandler
	Profile     *elhandler.ProfileHandler
	Achievement *elhandler.AchievementHandler
	AdminUser   *elhandler.AdminUserHandler
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
	// Auth di dalam usecase supaya urutan cek dan pesan sama dengan api/*.ts (misal ai cek kunci Gemini sebelum token).
	g.POST("/ai", h.AI.Generate)
	g.POST("/grade", h.Grade.Grade)
	g.POST("/report", h.Report.Sync)
	g.POST("/validate-quiz", h.Quiz.Validate)

	authed := g.Group("", noStore, authmw)
	authed.POST("/auth/supabase-token", h.Auth.SupabaseToken)

	authed.GET("/curriculum/levels", h.Curriculum.Levels)
	authed.GET("/curriculum/modules", h.Curriculum.Modules)
	authed.GET("/curriculum/lessons", h.Curriculum.Lessons)
	authed.PUT("/curriculum", h.Curriculum.Save)
	authed.DELETE("/curriculum", h.Curriculum.Clear)

	authed.GET("/progress/me/lessons", h.Progress.MyLessons)
	authed.GET("/progress/me/recent", h.Progress.Recent)
	authed.POST("/progress/lessons/:lessonId/complete", h.Progress.Complete)
	authed.POST("/progress/xp", h.Progress.GrantXP)
	authed.POST("/progress/study-time", h.Progress.StudyTime)
	authed.POST("/streak", h.Progress.Streak)
	authed.POST("/profile/sync", h.Profile.Sync)
	authed.POST("/activity-logs", h.Profile.LogActivity)

	authed.GET("/achievements", h.Achievement.List)
	authed.GET("/achievements/me", h.Achievement.Mine)
	authed.POST("/achievements/check", h.Achievement.Check)
	authed.POST("/achievements/check-xp", h.Achievement.CheckXP)

	authed.GET("/admin/users", h.AdminUser.List)
	authed.GET("/admin/users/:nim/progress", h.AdminUser.Progress)
	authed.PATCH("/admin/users/:nim/level-access", h.AdminUser.LevelAccess)
	authed.PATCH("/admin/users/:nim/role", h.AdminUser.Role)
	authed.POST("/admin/users/:nim/reset", h.AdminUser.Reset)
	authed.POST("/admin/users/:nim/reset-level", h.AdminUser.ResetLevel)
	authed.POST("/admin/users/:nim/reset-lesson", h.AdminUser.ResetLesson)
	authed.PUT("/admin/users/:nim/xp", h.AdminUser.AdjustXP)
	authed.DELETE("/admin/users/:nim", h.AdminUser.Delete)
}

// Respons elearning ber-token bergantung pada pengguna dan perannya, jadi tidak boleh disimpan CDN Vercel atau cache bersama.
func noStore(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	c.Next()
}

// Tanpa DATABASE_URL_ELEARNING semua path elearning dijawab 503 supaya praktikum tetap bisa jalan.
func RegisterElearningUnavailable(r *gin.Engine) {
	r.Any("/api/elearning/*path", func(c *gin.Context) {
		response.Fail(c, http.StatusServiceUnavailable, "Backend elearning belum dikonfigurasi (DATABASE_URL_ELEARNING).", nil)
	})
}
