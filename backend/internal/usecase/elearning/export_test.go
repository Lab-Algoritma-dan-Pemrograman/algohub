package usecase

import (
	"context"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

// Guard users dan trigger dibuka untuk test aturan RLS; T-check_user_updates-3 tidak bisa dicapai lewat endpoint
// karena column grant menolak assessment_access lebih dulu.
var (
	CheckUserUpdates  = checkUserUpdates
	CalculateStreak   = calculateStreak
	CalculateLevel    = calculateLevel
	ApplyPrestigeMult = applyPrestigeMult
)

func GuardUpdateUser(ctx context.Context, users repository.UserRepository, p *dto.Principal, nim string, cols map[string]any) (int64, error) {
	return userGuard{users: users}.update(ctx, p, nim, cols)
}

func GuardInsertUser(ctx context.Context, users repository.UserRepository, p *dto.Principal, cols map[string]any) error {
	return userGuard{users: users}.insert(ctx, p, cols)
}

func GuardDeleteUser(ctx context.Context, users repository.UserRepository, p *dto.Principal, nim string) (int64, error) {
	return userGuard{users: users}.delete(ctx, p, nim)
}
