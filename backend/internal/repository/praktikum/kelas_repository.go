package repository

import (
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"

	"gorm.io/gorm"
)

type KelasRepository interface{ crudRepo[entity.Kelas] }

type kelasRepository struct{ crudRepository[entity.Kelas] }

func NewKelasRepository(db *gorm.DB) KelasRepository {
	return &kelasRepository{crudRepository[entity.Kelas]{db: db, order: "nama_kelas asc"}}
}
