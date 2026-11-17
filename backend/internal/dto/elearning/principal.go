package dto

// NIM mengikuti normalisasi api/*.ts (nim || username || sub), sedangkan AuthNIM mengikuti
// public.auth_nim() (COALESCE nim, sub, username) yang dipakai aturan RLS. Untuk token re-sign keduanya sama.
type Principal struct {
	NIM       string
	AuthNIM   string
	UID       string
	Nama      string
	Kelas     string
	Shift     *string
	Jurusan   *string
	Email     *string
	Role      string
	RawClaims map[string]any
}
