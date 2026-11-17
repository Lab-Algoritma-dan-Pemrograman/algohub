package app

import (
	"github.com/gin-gonic/gin"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/database"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
	handler "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http/handler/siakad"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/siakad"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/siakad"
)

func registerSiakad(r *gin.Engine, cfg *config.Config, dbs *database.DBs) {
	tokens := usecase.NewTokens(cfg.Siakad.JWTSecret)
	authRepo := repository.NewAuthRepository(dbs.Siakad)
	rpcRepo := repository.NewRPCRepository(dbs.Siakad)

	deliveryhttp.RegisterSiakad(r, &deliveryhttp.Siakad{
		Handlers: deliveryhttp.SiakadHandlers{
			Auth: handler.NewAuthHandler(usecase.NewAuthUsecase(authRepo, tokens)),
			RPC:  handler.NewRPCHandler(usecase.NewRPCUsecase(rpcRepo, authRepo, tokens)),
		},
	})
}
