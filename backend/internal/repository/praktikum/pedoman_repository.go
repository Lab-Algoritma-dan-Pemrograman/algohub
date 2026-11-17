package repository

import (
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"

	"gorm.io/gorm"
)

type PedomanRepository interface {
	crudRepo[entity.PedomanLaporan]
}

type pedomanRepository struct {
	crudRepository[entity.PedomanLaporan]
}

func NewPedomanRepository(db *gorm.DB) PedomanRepository {
	return &pedomanRepository{crudRepository[entity.PedomanLaporan]{db: db, order: "diunggah_pada desc"}}
}
