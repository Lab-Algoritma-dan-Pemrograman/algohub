package repository

import (
	"context"
	"time"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TokenBlocklist: blocklist jti JWT praktikum (K-1), cermin AuthRepository
// siakad (IsTokenRevoked/RevokeToken) tapi memakai DB praktikum + GORM entity,
// bukan stored function RPC yang hanya ada di skema siakad.
type TokenBlocklist interface {
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
	RevokeToken(ctx context.Context, jti string, userID int, expiresAt time.Time) error
}

type revokedTokenRepository struct{ db *gorm.DB }

func NewTokenBlocklist(db *gorm.DB) TokenBlocklist { return &revokedTokenRepository{db: db} }

func (r *revokedTokenRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.RevokedToken{}).Where("jti = ?", jti).Count(&count).Error
	return count > 0, err
}

func (r *revokedTokenRepository) RevokeToken(ctx context.Context, jti string, userID int, expiresAt time.Time) error {
	row := entity.RevokedToken{JTI: jti, UserID: userID, ExpiresAt: expiresAt, RevokedAt: time.Now()}
	// OnConflict DoNothing: logout dua kali (dua tab) tak boleh error.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}
