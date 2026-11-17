package entity

import "time"

type StudentProgress struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	NIM         string     `gorm:"column:nim" json:"nim"`
	LessonID    string     `gorm:"column:lesson_id" json:"lesson_id"`
	Completed   *bool      `gorm:"column:completed" json:"completed"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
	Score       *int       `gorm:"column:score" json:"score"`
}

func (StudentProgress) TableName() string { return "student_progress" }
