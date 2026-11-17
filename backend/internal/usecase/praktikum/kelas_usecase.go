package usecase

import (
	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
)

type KelasUsecase struct {
	repo repository.KelasRepository
}

func NewKelasUsecase(r repository.KelasRepository) *KelasUsecase { return &KelasUsecase{repo: r} }

func (uc *KelasUsecase) List() ([]entity.Kelas, error) { return uc.repo.List() }

func (uc *KelasUsecase) Create(req dto.KelasRequest) (*entity.Kelas, error) {
	k := &entity.Kelas{NamaKelas: req.NamaKelas}
	if err := uc.repo.Create(k); err != nil {
		return nil, err
	}
	return k, nil
}

func (uc *KelasUsecase) Update(id int, req dto.KelasRequest) (*entity.Kelas, error) {
	k, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	k.NamaKelas = req.NamaKelas
	if err := uc.repo.Update(k); err != nil {
		return nil, err
	}
	return k, nil
}

func (uc *KelasUsecase) Delete(id int) error { return mapDeleteErr(uc.repo.Delete(id)) }
