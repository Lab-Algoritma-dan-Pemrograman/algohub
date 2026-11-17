package entity

import "time"

type Achievement struct {
	ID               string  `gorm:"column:id;primaryKey" json:"id"`
	Title            string  `gorm:"column:title" json:"title"`
	Description      string  `gorm:"column:description" json:"description"`
	Icon             *string `gorm:"column:icon" json:"icon"`
	RequirementType  *string `gorm:"column:requirement_type" json:"requirement_type"`
	RequirementValue *string `gorm:"column:requirement_value" json:"requirement_value"`
}

func (Achievement) TableName() string { return "achievements" }

type UnlockedAchievement struct {
	ID            string     `gorm:"column:id;primaryKey" json:"id"`
	NIM           string     `gorm:"column:nim" json:"nim"`
	AchievementID string     `gorm:"column:achievement_id" json:"achievement_id"`
	UnlockedAt    *time.Time `gorm:"column:unlocked_at" json:"unlocked_at"`
}

func (UnlockedAchievement) TableName() string { return "unlocked_achievements" }
