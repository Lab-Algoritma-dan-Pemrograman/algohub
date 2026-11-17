package app

import (
	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/database"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
	cronhandler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/cron"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/bynara"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/gemini"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/webpush"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/pkg/whatsapp"
)

const vapidSubject = "mailto:admin@lab-ap.web.id"

func registerSiakad(r *gin.Engine, cfg *config.Config, dbs *database.DBs) {
	tokens := usecase.NewTokens(cfg.Siakad.JWTSecret)
	authRepo := repository.NewAuthRepository(dbs.Siakad)
	rpcRepo := repository.NewRPCRepository(dbs.Siakad)
	pushRepo := repository.NewPushRepository(dbs.Siakad)
	pusher := webpush.NewSender(cfg.Siakad.VAPIDPublicKey, cfg.Siakad.VAPIDPrivateKey, vapidSubject, nil)

	pushUC := usecase.NewPushUsecase(pushRepo, pusher, tokens)
	waUC := usecase.NewWAUsecase(repository.NewNotifyRepository(dbs.Siakad), whatsapp.NewClient(), tokens, cfg.Siakad.WAGatewayToken)
	shiftUC := usecase.NewShiftCronUsecase(repository.NewShiftRepository(dbs.Siakad), pushRepo, pusher)
	keepAliveUC := usecase.NewKeepAliveUsecase(repository.NewSiakadPinger(dbs.Siakad), repository.NewElearningPinger(dbs.Elearning))
	integration := handler.NewIntegrationHandler(
		usecase.NewScheduleAIUsecase(tokens, aiProviders(&cfg.Siakad)),
		usecase.NewSSOUsecase(repository.NewSSORepository(dbs.Siakad), tokens),
		usecase.NewElearningSyncUsecase(tokens, repository.NewElearningRepository(dbs.Elearning), repository.NewProgressRepository(dbs.Siakad)),
	)

	deliveryhttp.RegisterSiakad(r, &deliveryhttp.Siakad{
		Handlers: deliveryhttp.SiakadHandlers{
			Auth:        handler.NewAuthHandler(usecase.NewAuthUsecase(authRepo, tokens)),
			RPC:         handler.NewRPCHandler(usecase.NewRPCUsecase(rpcRepo, authRepo, tokens)),
			Notify:      handler.NewNotifyHandler(pushUC, waUC, cfg.CronSecret),
			Integration: integration,
			Cron:        cronhandler.NewCronHandler(cfg.CronSecret, shiftUC, keepAliveUC),
		},
	})
}

// Klien hanya dibuat kalau key-nya ada, supaya usecase bisa membalas pesan "belum dikonfigurasi" lama.
func aiProviders(cfg *config.SiakadConfig) usecase.AIProviders {
	var ai usecase.AIProviders
	if cfg.GeminiAPIKey != "" {
		ai.Gemini = gemini.NewClient(cfg.GeminiAPIKey)
	}
	if cfg.BynaraAPIKey != "" {
		ai.Bynara = bynara.NewClient(cfg.BynaraAPIKey)
	}
	return ai
}
