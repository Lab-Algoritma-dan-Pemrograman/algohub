package dto

type TokenRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// Nilai any karena api/*.ts lama meneruskan claim apa adanya (string, angka, atau null).
type AuthPayload struct {
	NIM     any     `json:"nim"`
	Nama    any     `json:"nama"`
	Kelas   any     `json:"kelas"`
	Shift   *string `json:"shift"`
	Jurusan any     `json:"jurusan"`
	Email   any     `json:"email"`
	Role    string  `json:"role"`
}

type VerifyResponse struct {
	Payload AuthPayload `json:"payload"`
	Token   string      `json:"token"`
}

type ReceiveTokenResponse struct {
	Success     bool        `json:"success"`
	Token       string      `json:"token"`
	RedirectURL string      `json:"redirectUrl"`
	Payload     AuthPayload `json:"payload"`
}

type LoginUser struct {
	NIM       string  `json:"nim"`
	Nama      string  `json:"nama"`
	Kelas     string  `json:"kelas"`
	Role      string  `json:"role"`
	NamaKelas *string `json:"nama_kelas"`
	Shift     *int    `json:"shift"`
}

type LoginResponse struct {
	Payload      AuthPayload `json:"payload"`
	Token        string      `json:"token"`
	BackendToken string      `json:"backendToken"`
	User         LoginUser   `json:"user"`
}

type SupabaseTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn,omitempty"`
}
