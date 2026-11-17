package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Kolom id, kelas_id, shift, divisi, kelompok, gelombang, tersembunyi, dan is_aktif hanya diketahui dari
// select SupabaseProvider.tsx tanpa DDL di repo lama, jadi tipenya dibuat pointer string/bool yang aman di-scan.
type User struct {
	NIM                  string         `gorm:"column:nim;primaryKey" json:"nim"`
	ID                   *string        `gorm:"column:id" json:"id"`
	Nama                 string         `gorm:"column:nama" json:"nama"`
	Kelas                string         `gorm:"column:kelas" json:"kelas"`
	Jurusan              *string        `gorm:"column:jurusan" json:"jurusan"`
	Email                *string        `gorm:"column:email" json:"email"`
	Role                 string         `gorm:"column:role" json:"role"`
	ElearningRole        *string        `gorm:"column:elearning_role" json:"elearning_role"`
	XP                   int            `gorm:"column:xp" json:"xp"`
	Level                int            `gorm:"column:level" json:"level"`
	Streak               int            `gorm:"column:streak" json:"streak"`
	StudyTime            *int           `gorm:"column:study_time" json:"study_time"`
	LastActive           *time.Time     `gorm:"column:last_active" json:"last_active"`
	CreatedAt            *time.Time     `gorm:"column:created_at" json:"created_at"`
	AssessmentAccess     datatypes.JSON `gorm:"column:assessment_access" json:"assessment_access"`
	LevelAccessOverrides datatypes.JSON `gorm:"column:level_access_overrides" json:"level_access_overrides"`
	Tersembunyi          *bool          `gorm:"column:tersembunyi" json:"tersembunyi"`
	IsAktif              *bool          `gorm:"column:is_aktif" json:"is_aktif"`
	KelasID              *string        `gorm:"column:kelas_id" json:"kelas_id"`
	Shift                *string        `gorm:"column:shift" json:"shift"`
	Divisi               *string        `gorm:"column:divisi" json:"divisi"`
	Kelompok             *string        `gorm:"column:kelompok" json:"kelompok"`
	Gelombang            *string        `gorm:"column:gelombang" json:"gelombang"`
	PrestigeTier         *string        `gorm:"column:prestige_tier" json:"prestige_tier"`
	PrestigeMult         float64        `gorm:"column:prestige_mult" json:"prestige_mult"`
	PrestigeCount        int            `gorm:"column:prestige_count" json:"prestige_count"`
	LastPrestigeAt       *time.Time     `gorm:"column:last_prestige_at" json:"last_prestige_at"`
}

func (User) TableName() string { return "users" }
