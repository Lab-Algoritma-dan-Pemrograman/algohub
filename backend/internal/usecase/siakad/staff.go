package usecase

import (
	"slices"
	"strings"
)

// K-11: daftar mentah lama tanpa trim dan tanpa alias. NormalizeRole akan ikut meloloskan
// 'korda'/'coordinator', yang di prod lama ditolak.
var notifyStaffRoles = []string{"asisten", "koordinator", "admin", "kordas"}

func isNotifyStaff(role string) bool {
	return slices.Contains(notifyStaffRoles, strings.ToLower(role))
}
