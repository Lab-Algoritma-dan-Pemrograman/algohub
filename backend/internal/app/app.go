package app

import (
	"net/http"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/database"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
)

func New() (http.Handler, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Siakad.Validate(cfg.DB); err != nil {
		return nil, err
	}
	dbs, err := database.Open(cfg.DB, cfg.AppEnv == "development")
	if err != nil {
		return nil, err
	}
	r, err := deliveryhttp.NewRouter(cfg)
	if err != nil {
		return nil, err
	}
	deliveryhttp.RegisterPraktikum(r, newPraktikum(cfg, dbs.Praktikum))
	registerSiakad(r, cfg, dbs)
	registerElearning(r, cfg, dbs)
	return r, nil
}
