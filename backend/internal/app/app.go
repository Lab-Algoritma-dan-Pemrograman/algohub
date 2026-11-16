package app

import (
	"net/http"

	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/config"
	deliveryhttp "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/delivery/http"
)

func New() (http.Handler, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return deliveryhttp.NewRouter(cfg)
}
