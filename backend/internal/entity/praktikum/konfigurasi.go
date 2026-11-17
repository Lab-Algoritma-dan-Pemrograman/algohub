package entity

import "time"

type Konfigurasi struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"column:key;type:varchar(100);uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Konfigurasi) TableName() string { return "konfigurasi" }

const (
	KeyGDriveJadwalURL = "gdrive_jadwal_url"
	KeyJadwalMode      = "jadwal_mode" // "gdrive" | "internal"
	KeyModulFileURL    = "modul_file_url"
	KeyAIModel         = "ai_model" // override model AI grading; kosong = default env OLLAMA_MODEL
)
