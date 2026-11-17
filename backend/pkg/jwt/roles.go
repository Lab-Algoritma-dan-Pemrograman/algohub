package jwt

import "strings"

// DB menyimpan 'mahasiswa'/'kordas'/'admin', sedangkan whitelist siakad memakai role kanonik.
// 'admin' = asisten di siakad. 'superadmin' adalah username, bukan role, jadi sengaja
// tidak dipetakan supaya gagal tertutup.
var roleAliases = map[string]string{
	"mahasiswa":   "praktikan",
	"mhs":         "praktikan",
	"peminjam":    "penyewa",
	"kordas":      "koordinator",
	"korda":       "koordinator",
	"coordinator": "koordinator",
	"admin":       "asisten",
}

// NormalizeRole mengubah role mentah dari DB atau token menjadi role kanonik siakad.
func NormalizeRole(raw string) string {
	role := strings.ToLower(strings.TrimSpace(raw))
	if canon, ok := roleAliases[role]; ok {
		return canon
	}
	return role
}
