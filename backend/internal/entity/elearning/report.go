package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Kolom mengikuti upsert api/report.ts. Skema siakad (20260410060000_create_elearning_progress.sql) memakai
// completed_lessons dan tidak punya completed_levels/current_level, jadi tipe completed_levels belum pasti.
type ElearningProgress struct {
	ID                   string         `gorm:"column:id;primaryKey" json:"id"`
	NIM                  string         `gorm:"column:nim" json:"nim"`
	StudentName          *string        `gorm:"column:student_name" json:"student_name"`
	LessonsCompleted     int            `gorm:"column:lessons_completed" json:"lessons_completed"`
	TotalLessons         int            `gorm:"column:total_lessons" json:"total_lessons"`
	CompletionPercentage float64        `gorm:"column:completion_percentage" json:"completion_percentage"`
	IsCompleted          bool           `gorm:"column:is_completed" json:"is_completed"`
	CompletedLevels      datatypes.JSON `gorm:"column:completed_levels" json:"completed_levels"`
	CurrentLevel         *string        `gorm:"column:current_level" json:"current_level"`
	LastAccessedAt       *time.Time     `gorm:"column:last_accessed_at" json:"last_accessed_at"`
	CreatedAt            *time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (ElearningProgress) TableName() string { return "elearning_progress" }
