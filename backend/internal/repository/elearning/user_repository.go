package repository

import (
	"context"

	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type UserRepository interface {
	RoleByNIM(ctx context.Context, nim string) (string, error)
	RoleByEmail(ctx context.Context, email string) (string, error)
	NamaByNIM(ctx context.Context, nim string) (string, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) RoleByNIM(ctx context.Context, nim string) (string, error) {
	return r.single(ctx, "COALESCE(role, '')", "nim = ?", nim)
}

func (r *userRepository) RoleByEmail(ctx context.Context, email string) (string, error) {
	return r.single(ctx, "COALESCE(role, '')", "email = ?", email)
}

func (r *userRepository) NamaByNIM(ctx context.Context, nim string) (string, error) {
	return r.single(ctx, "COALESCE(nama, '')", "nim = ?", nim)
}

// Meniru .single()/.maybeSingle() PostgREST: nol atau lebih dari satu baris sama-sama dianggap tidak ditemukan.
func (r *userRepository) single(ctx context.Context, column, where string, arg any) (string, error) {
	var vals []string
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where(where, arg).Limit(2).Pluck(column, &vals).Error; err != nil {
		return "", err
	}
	if len(vals) != 1 {
		return "", gorm.ErrRecordNotFound
	}
	return vals[0], nil
}
