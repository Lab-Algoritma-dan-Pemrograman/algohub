package entity

import "time"

type PrestigeHistory struct {
	ID               string     `gorm:"column:id;primaryKey" json:"id"`
	NIM              string     `gorm:"column:nim" json:"nim"`
	PrestigeNo       int        `gorm:"column:prestige_no" json:"prestige_no"`
	ResetAt          *time.Time `gorm:"column:reset_at" json:"reset_at"`
	LevelsCompleted  int        `gorm:"column:levels_completed" json:"levels_completed"`
	LessonsCompleted int        `gorm:"column:lessons_completed" json:"lessons_completed"`
	BaseXPArchived   int        `gorm:"column:base_xp_archived" json:"base_xp_archived"`
	Tier             *string    `gorm:"column:tier" json:"tier"`
	Mult             float64    `gorm:"column:mult" json:"mult"`
}

func (PrestigeHistory) TableName() string { return "prestige_history" }

type ProgressArchive struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	NIM         string     `gorm:"column:nim" json:"nim"`
	LessonID    string     `gorm:"column:lesson_id" json:"lesson_id"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
	ArchivedAt  *time.Time `gorm:"column:archived_at" json:"archived_at"`
	PrestigeNo  int        `gorm:"column:prestige_no" json:"prestige_no"`
}

func (ProgressArchive) TableName() string { return "progress_archive" }
