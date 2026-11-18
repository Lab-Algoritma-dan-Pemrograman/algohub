package config

import "errors"

// MEDIUM-04: kunci VAPID private lama bocor lewat repo publik dan sudah dirotasi, jadi private
// wajib dari env. Kunci publik hasil rotasi aman dibundel karena memang dikirim ke browser.
const defaultVAPIDPublicKey = "BKAXOmOt7QXutFWQp9GEPf17gU0BkTAd_xzAeQtIWONnnnUo7XUI3m4hunhgqaY1bbuKMKw3GUq44jD5Xd2Dpc0"

type SiakadConfig struct {
	JWTSecret       string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	WAGatewayToken  string
	GeminiAPIKey    string
	BynaraAPIKey    string
}

func loadSiakad() SiakadConfig {
	return SiakadConfig{
		JWTSecret:       getEnv("SIAKAD_JWT_SECRET", ""),
		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", defaultVAPIDPublicKey),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		WAGatewayToken:  getEnv("WA_GATEWAY_TOKEN", ""),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		BynaraAPIKey:    getEnv("BYNARA_API_KEY", ""),
	}
}

// Siakad aktif hanya bila DSN-nya diisi, supaya env siakad yang belum siap tidak ikut mematikan
// praktikum dan elearning yang berbagi satu function.
func (s *SiakadConfig) Enabled(db DBConfig) bool {
	return db.Siakad != ""
}

// Dipanggil terpisah dari Load dan sebelum DB dibuka: DSN siakad yang diisi tanpa env pendampingnya
// adalah salah konfigurasi, jadi server gagal start tanpa mencoba konek.
func (s *SiakadConfig) Validate(db DBConfig) error {
	if !s.Enabled(db) {
		return nil
	}
	switch {
	case db.Elearning == "":
		return errors.New("config: DATABASE_URL_ELEARNING wajib diisi")
	case s.JWTSecret == "":
		return errors.New("config: SIAKAD_JWT_SECRET wajib diisi")
	}
	return nil
}
