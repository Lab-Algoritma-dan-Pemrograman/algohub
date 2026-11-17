package entity

import "time"

type WeeklyStanding struct {
	ID               string     `gorm:"column:id;primaryKey" json:"id"`
	WeekStart        time.Time  `gorm:"column:week_start" json:"week_start"`
	NIM              string     `gorm:"column:nim" json:"nim"`
	Nama             string     `gorm:"column:nama" json:"nama"`
	Kelas            *string    `gorm:"column:kelas" json:"kelas"`
	Shift            *string    `gorm:"column:shift" json:"shift"`
	Jurusan          *string    `gorm:"column:jurusan" json:"jurusan"`
	WeeklyXP         int        `gorm:"column:weekly_xp" json:"weekly_xp"`
	LessonXP         int        `gorm:"column:lesson_xp" json:"lesson_xp"`
	BughuntXP        int        `gorm:"column:bughunt_xp" json:"bughunt_xp"`
	LessonsCompleted int        `gorm:"column:lessons_completed" json:"lessons_completed"`
	PrestigeTier     *string    `gorm:"column:prestige_tier" json:"prestige_tier"`
	PrestigeCount    int        `gorm:"column:prestige_count" json:"prestige_count"`
	Rank             *int       `gorm:"column:rank" json:"rank"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
}

func (WeeklyStanding) TableName() string { return "weekly_standings" }
