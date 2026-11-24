package app

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/database"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
	elhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/elearning"
	prakdto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	elrepo "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
	prakrepo "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
	elusecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
	prakusecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/jwt"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/objectstore"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/ollama"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/supabase"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/vercelblob"
)

func registerElearning(r *gin.Engine, cfg *config.Config, dbs *database.DBs) {
	if dbs.Elearning == nil {
		deliveryhttp.RegisterElearningUnavailable(r)
		return
	}
	deliveryhttp.RegisterElearning(r, newElearning(cfg, dbs))
}

func newElearning(cfg *config.Config, dbs *database.DBs) *deliveryhttp.Elearning {
	ecfg := config.LoadElearning()
	authUC := elusecase.NewAuthUsecase(&ecfg, newPraktikumLogin(cfg, dbs.Praktikum), time.Now)

	db := dbs.Elearning
	users := elrepo.NewUserRepository(db)
	curriculum := elrepo.NewCurriculumRepository(db)
	progress := elrepo.NewProgressRepository(db)
	unlocked := elrepo.NewAchievementRepository(db)
	logs := elusecase.NewActivityLogUsecase(elrepo.NewActivityLogRepository(db), time.Now)
	achievements := elusecase.NewAchievementUsecase(unlocked, users, progress, time.Now)
	learning := elusecase.NewProgressUsecase(users, progress, curriculum, achievements, time.Now)
	images := objectstore.NewFailover(
		vercelblob.New(cfg.PublicBlobReadWriteToken, vercelblob.Public),
		supabase.New(ecfg.SupabaseURL, ecfg.SupabaseServiceRoleKey, "lesson-images"),
	)
	ai := elusecase.AIClients{Gemini: gemini.NewClient(""), GeminiKeys: ecfg.GeminiAPIKeys}
	if ecfg.OllamaURL != "" {
		ai.Ollama = ollama.NewClient(ollama.Config{URL: ecfg.OllamaURL, Model: ecfg.OllamaModel, APIKey: ecfg.OllamaAPIKey})
	}

	return &deliveryhttp.Elearning{
		Auth: authUC,
		Handlers: deliveryhttp.ElearningHandlers{
			Auth:        elhandler.NewAuthHandler(authUC, ecfg.AppURL),
			AI:          elhandler.NewAIHandler(elusecase.NewAIUsecase(authUC, users, ai)),
			Grade:       elhandler.NewGradeHandler(elusecase.NewGradeUsecase(authUC, users, elrepo.NewAssessmentRepository(db), ai, time.Now)),
			Report:      elhandler.NewReportHandler(elusecase.NewReportUsecase(authUC, curriculum, progress, elrepo.NewReportRepository(db), time.Now)),
			Quiz:        elhandler.NewQuizHandler(elusecase.NewQuizUsecase(authUC, curriculum)),
			Curriculum:  elhandler.NewCurriculumHandler(elusecase.NewCurriculumUsecase(curriculum)),
			Progress:    elhandler.NewProgressHandler(learning),
			Profile:     elhandler.NewProfileHandler(elusecase.NewProfileUsecase(users, achievements, time.Now), logs),
			Achievement: elhandler.NewAchievementHandler(achievements),
			AdminUser:   elhandler.NewAdminUserHandler(elusecase.NewAdminUserUsecase(users, progress, curriculum, unlocked, achievements, logs)),
			Game:        elhandler.NewGameHandler(elusecase.NewGameUsecase(elrepo.NewGameRepository(db), users, time.Now, nil)),
			Leaderboard: elhandler.NewLeaderboardHandler(elusecase.NewLeaderboardUsecase(elrepo.NewLeaderboardRepository(db), time.Now)),
			Prestige:    elhandler.NewPrestigeHandler(elusecase.NewPrestigeUsecase(users, curriculum, progress, elrepo.NewPrestigeRepository(db), logs, time.Now)),
			Playground:  elhandler.NewPlaygroundHandler(elusecase.NewPlaygroundUsecase(elrepo.NewPlaygroundRepository(db))),
			Monitoring:  elhandler.NewMonitoringHandler(elusecase.NewSessionUsecase(elrepo.NewSessionRepository(db), users, progress, learning, time.Now), logs),
			Storage:     elhandler.NewStorageHandler(elusecase.NewStorageUsecase(images, time.Now)),
		},
	}
}

// Efek audit LOGIN/LOGIN_FAILED sama dengan proksi lama ke POST /api/auth/login praktikum.
type praktikumLogin struct {
	auth  *prakusecase.AuthUsecase
	audit *prakusecase.AuditLogUsecase
}

func newPraktikumLogin(cfg *config.Config, db *gorm.DB) praktikumLogin {
	users := prakrepo.NewUserRepository(db)
	auth := prakusecase.NewAuthUsecase(users, prakrepo.NewKelasRepository(db), jwt.NewManager(cfg.JWTSecret, cfg.JWTExpireHours), cfg, fbScryptConfig(cfg)).
		WithBlocklist(prakrepo.NewTokenBlocklist(db))
	return praktikumLogin{auth: auth, audit: prakusecase.NewAuditLogUsecase(prakrepo.NewAuditLogRepository(db), users)}
}

func (p praktikumLogin) Login(_ context.Context, identifier, password, ip, userAgent string) (*elusecase.PraktikumLoginResult, error) {
	res, err := p.auth.Login(prakdto.LoginRequest{Identifier: identifier, Password: password})
	if err != nil {
		_ = p.audit.LogAction(0, identifier, "LOGIN_FAILED", "Gagal login", ip, userAgent)
		return nil, err
	}
	_ = p.audit.LogAction(0, identifier, "LOGIN", "Login berhasil", ip, userAgent)
	return &elusecase.PraktikumLoginResult{
		Token:     res.Token,
		NIM:       res.User.NIM,
		Nama:      res.User.Nama,
		NamaKelas: res.User.NamaKelas,
		Shift:     res.User.Shift,
		Role:      res.User.Role,
	}, nil
}
