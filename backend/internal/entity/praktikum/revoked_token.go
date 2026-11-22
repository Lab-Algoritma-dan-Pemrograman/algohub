package entity

import "time"

// RevokedToken: blocklist jti JWT praktikum yang sudah logout (K-1).
// Baris dibersihkan otomatis saat exp lewat (lihat migration 001), jadi tabel
// tetap kecil tanpa job terpisah. Pola mengikuti blocklist siakad.
type RevokedToken struct {
	JTI       string    `gorm:"primaryKey;column:jti;type:varchar(64)" json:"jti"`
	UserID    int       `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	RevokedAt time.Time `gorm:"not null" json:"revoked_at"`
}

func (RevokedToken) TableName() string { return "revoked_tokens" }
