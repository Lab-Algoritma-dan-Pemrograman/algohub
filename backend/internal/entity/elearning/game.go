package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Kolom question, buggy_code, options, correct_option hanya ada di setup-new-database.sql (dianggap tidak
// dijalankan, keputusan 5); gameService.ts memakai difficulty, title, code, bug_line. Keduanya dibawa sebagai
// pointer supaya scan SELECT * aman; tulis wajib memakai daftar kolom eksplisit.
type GameQuestion struct {
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
	Language      string         `gorm:"column:language" json:"language"`
	Difficulty    *string        `gorm:"column:difficulty" json:"difficulty"`
	Title         *string        `gorm:"column:title" json:"title"`
	Code          *string        `gorm:"column:code" json:"code"`
	BugLine       *int           `gorm:"column:bug_line" json:"bug_line"`
	Explanation   *string        `gorm:"column:explanation" json:"explanation"`
	Question      *string        `gorm:"column:question" json:"question"`
	BuggyCode     *string        `gorm:"column:buggy_code" json:"buggy_code"`
	Options       datatypes.JSON `gorm:"column:options" json:"options"`
	CorrectOption *int           `gorm:"column:correct_option" json:"correct_option"`
	CreatedAt     *time.Time     `gorm:"column:created_at" json:"created_at"`
}

func (GameQuestion) TableName() string { return "game_questions" }

type GameSettings struct {
	ID                  string `gorm:"column:id;primaryKey" json:"id"`
	BugHuntActive       *bool  `gorm:"column:bug_hunt_active" json:"bug_hunt_active"`
	BugHuntCActive      *bool  `gorm:"column:bug_hunt_c_active" json:"bug_hunt_c_active"`
	BugHuntPythonActive *bool  `gorm:"column:bug_hunt_python_active" json:"bug_hunt_python_active"`
	BugHuntWeeklyLimit  *int   `gorm:"column:bug_hunt_weekly_limit" json:"bug_hunt_weekly_limit"`
}

func (GameSettings) TableName() string { return "game_settings" }

// score NOT NULL di setup-new-database.sql, tetapi saveGameHistory tidak mengirim score; jadi pointer.
type GameHistory struct {
	ID       string     `gorm:"column:id;primaryKey" json:"id"`
	NIM      string     `gorm:"column:nim" json:"nim"`
	GameType string     `gorm:"column:game_type" json:"game_type"`
	Score    *int       `gorm:"column:score" json:"score"`
	XPEarned int        `gorm:"column:xp_earned" json:"xp_earned"`
	PlayedAt *time.Time `gorm:"column:played_at" json:"played_at"`
}

func (GameHistory) TableName() string { return "game_history" }
