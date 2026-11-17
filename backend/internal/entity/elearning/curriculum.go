package entity

import "gorm.io/datatypes"

type Level struct {
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	Title       string  `gorm:"column:title" json:"title"`
	Description *string `gorm:"column:description" json:"description"`
	AccessMode  *string `gorm:"column:access_mode" json:"access_mode"`
	Locked      *bool   `gorm:"column:locked" json:"locked"`
	SortOrder   int     `gorm:"column:sort_order" json:"sort_order"`
}

func (Level) TableName() string { return "levels" }

type Module struct {
	ID        string `gorm:"column:id;primaryKey" json:"id"`
	LevelID   string `gorm:"column:level_id" json:"level_id"`
	Title     string `gorm:"column:title" json:"title"`
	SortOrder int    `gorm:"column:sort_order" json:"sort_order"`
}

func (Module) TableName() string { return "modules" }

type Lesson struct {
	ID                   string         `gorm:"column:id;primaryKey" json:"id"`
	ModuleID             string         `gorm:"column:module_id" json:"module_id"`
	Title                string         `gorm:"column:title" json:"title"`
	Explanation          *string        `gorm:"column:explanation" json:"explanation"`
	CodeExample          *string        `gorm:"column:code_example" json:"code_example"`
	InitialCode          *string        `gorm:"column:initial_code" json:"initial_code"`
	Solution             *string        `gorm:"column:solution" json:"solution"`
	Hint                 *string        `gorm:"column:hint" json:"hint"`
	Quiz                 datatypes.JSON `gorm:"column:quiz" json:"quiz"`
	TestCases            datatypes.JSON `gorm:"column:test_cases" json:"test_cases"`
	ValidationRules      datatypes.JSON `gorm:"column:validation_rules" json:"validation_rules"`
	XPReward             *int           `gorm:"column:xp_reward" json:"xp_reward"`
	SortOrder            int            `gorm:"column:sort_order" json:"sort_order"`
	ExerciseType         *string        `gorm:"column:exercise_type" json:"exercise_type"`
	FlowchartDistractors *string        `gorm:"column:flowchart_distractors" json:"flowchart_distractors"`
}

func (Lesson) TableName() string { return "lessons" }

// View read-only: aturan kolom solution dan quiz tanpa correctAnswer ditiru di usecase (keputusan 3).
type StudentLesson struct {
	ID                   string         `gorm:"column:id" json:"id"`
	ModuleID             string         `gorm:"column:module_id" json:"module_id"`
	Title                string         `gorm:"column:title" json:"title"`
	Explanation          *string        `gorm:"column:explanation" json:"explanation"`
	CodeExample          *string        `gorm:"column:code_example" json:"code_example"`
	InitialCode          *string        `gorm:"column:initial_code" json:"initial_code"`
	Hint                 *string        `gorm:"column:hint" json:"hint"`
	Quiz                 datatypes.JSON `gorm:"column:quiz" json:"quiz"`
	TestCases            datatypes.JSON `gorm:"column:test_cases" json:"test_cases"`
	ValidationRules      datatypes.JSON `gorm:"column:validation_rules" json:"validation_rules"`
	SortOrder            int            `gorm:"column:sort_order" json:"sort_order"`
	XPReward             *int           `gorm:"column:xp_reward" json:"xp_reward"`
	ExerciseType         *string        `gorm:"column:exercise_type" json:"exercise_type"`
	FlowchartDistractors *string        `gorm:"column:flowchart_distractors" json:"flowchart_distractors"`
	Solution             *string        `gorm:"column:solution" json:"solution"`
}

func (StudentLesson) TableName() string { return "student_lessons" }
