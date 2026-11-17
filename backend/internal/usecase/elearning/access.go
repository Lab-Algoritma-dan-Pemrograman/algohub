package usecase

import (
	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
)

// Semua policy elearning ditujukan TO authenticated; principal kosong setara request anon (elearning-rls.md § Aturan usecase).
func requirePrincipal(p *dto.Principal) error {
	if p == nil {
		return Unauthorized("Unauthorized: Missing token")
	}
	return nil
}

// Policy SELECT/INSERT "milik sendiri atau staf" pada student_progress dan unlocked_achievements (nim = auth_nim()).
func ownsOrStaff(p *dto.Principal, nim string) bool {
	return nim == p.AuthNIM || IsStaff(p.Role)
}

func listPage[T any](p *dto.Principal, page, limit int, list func() ([]T, int64, error)) (*dto.Page[T], error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	rows, total, err := list()
	if err != nil {
		return nil, err
	}
	return pageOf(rows, total, page, limit), nil
}

func pageOf[T any](items []T, total int64, page, limit int) *dto.Page[T] {
	if items == nil {
		items = []T{}
	}
	return &dto.Page[T]{Items: items, Total: total, Page: page, Limit: limit}
}
