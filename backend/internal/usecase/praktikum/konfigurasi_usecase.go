package usecase

import (
	"errors"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"

	"gorm.io/gorm"
)

type KonfigurasiUsecase struct {
	repo repository.KonfigurasiRepository
}

func NewKonfigurasiUsecase(r repository.KonfigurasiRepository) *KonfigurasiUsecase {
	return &KonfigurasiUsecase{repo: r}
}

func (uc *KonfigurasiUsecase) Get(key string) (string, error) {
	k, err := uc.repo.Get(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return k.Value, nil
}

func (uc *KonfigurasiUsecase) Set(key, value string) error {
	return uc.repo.Set(key, value)
}

func (uc *KonfigurasiUsecase) All() ([]entity.Konfigurasi, error) {
	return uc.repo.All()
}
