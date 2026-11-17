package usecase

import "strings"

const (
	RoleAdmin     = "admin"
	RoleKordas    = "kordas"
	RoleAsisten   = "asisten"
	RolePraktikan = "praktikan"

	CanonicalMahasiswa   = "mahasiswa"
	CanonicalAsisten     = "asisten"
	CanonicalKoordinator = "koordinator"
)

// Kosakata server (api/index.ts, receive-token.ts, verify.ts, login.ts): superadmin dipetakan ke kordas.
func NormalizeTokenRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "admin":
		return RoleAdmin
	case "kordas", "koordinator", "korda", "superadmin", "super_admin", "administrator":
		return RoleKordas
	case "asisten", "assistant", "laboran":
		return RoleAsisten
	}
	return RolePraktikan
}

// Disalin dari normalizeRole di src/services/tokenService.ts: berbeda dari NormalizeTokenRole, superadmin dipetakan ke admin.
func NormalizeUIRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "superadmin", "super_admin", "administrator", "super admin":
		return RoleAdmin
	case "kordas", "koordinator", "korda", "coordinator":
		return RoleKordas
	case "asisten", "assistant", "laboran", "ast":
		return RoleAsisten
	}
	return RolePraktikan
}

// Kolom users.role dibaca juga oleh Praktikum yang hanya mengenal mahasiswa/asisten/koordinator.
func ToCanonicalRole(uiRole string) string {
	switch NormalizeUIRole(uiRole) {
	case RoleAdmin, RoleKordas:
		return CanonicalKoordinator
	case RoleAsisten:
		return CanonicalAsisten
	}
	return CanonicalMahasiswa
}

// Flag admin E-Learning disimpan terpisah di users.elearning_role karena users.role harus tetap kanonik.
func ResolveElearningRole(role, elearningRole string) string {
	if strings.ToLower(strings.TrimSpace(elearningRole)) == RoleAdmin {
		return RoleAdmin
	}
	return NormalizeUIRole(role)
}

func IsStaff(role string) bool {
	return role == RoleAdmin || role == RoleKordas || role == RoleAsisten
}

func IsAdminOrKordas(role string) bool {
	return role == RoleAdmin || role == RoleKordas
}
