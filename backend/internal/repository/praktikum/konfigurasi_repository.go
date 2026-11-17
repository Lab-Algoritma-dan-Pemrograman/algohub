package repository

import (
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KonfigurasiRepository interface {
	Get(key string) (*entity.Konfigurasi, error)
	Set(key, value string) error
	All() ([]entity.Konfigurasi, error)
}

type konfigurasiRepository struct {
	db *gorm.DB
}

func NewKonfigurasiRepository(db *gorm.DB) KonfigurasiRepository {
	return &konfigurasiRepository{db: db}
}

func (r *konfigurasiRepository) Get(key string) (*entity.Konfigurasi, error) {
	var k entity.Konfigurasi
	if err := r.db.Where("\"key\" = ?", key).First(&k).Error; err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *konfigurasiRepository) Set(key, value string) error {
	k := entity.Konfigurasi{Key: key, Value: value, UpdatedAt: time.Now()}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&k).Error
}

func (r *konfigurasiRepository) All() ([]entity.Konfigurasi, error) {
	var ks []entity.Konfigurasi
	return ks, r.db.Find(&ks).Error
}
