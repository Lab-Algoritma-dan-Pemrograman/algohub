package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Tabel assessment_attempts dan assessment_questions tidak punya CREATE TABLE di repo lama; kolom diambil
// dari pemakaian di api/grade.ts dan migration-*.sql.
type AssessmentAttempt struct {
	ID                string         `gorm:"column:id;primaryKey" json:"id"`
	NIM               string         `gorm:"column:nim" json:"nim"`
	MenuType          string         `gorm:"column:menu_type" json:"menu_type"`
	Status            string         `gorm:"column:status" json:"status"`
	SelectedQuestions datatypes.JSON `gorm:"column:selected_questions" json:"selected_questions"`
	Answers           datatypes.JSON `gorm:"column:answers" json:"answers"`
	AIGrades          datatypes.JSON `gorm:"column:ai_grades" json:"ai_grades"`
	FinalScore        *float64       `gorm:"column:final_score" json:"final_score"`
	GradedAt          *time.Time     `gorm:"column:graded_at" json:"graded_at"`
	TabSwitchCount    *int           `gorm:"column:tab_switch_count" json:"tab_switch_count"`
}

func (AssessmentAttempt) TableName() string { return "assessment_attempts" }

type AssessmentQuestion struct {
	ID                string  `gorm:"column:id;primaryKey" json:"id"`
	MenuType          *string `gorm:"column:menu_type" json:"menu_type"`
	Title             *string `gorm:"column:title" json:"title"`
	Instruction       *string `gorm:"column:instruction" json:"instruction"`
	Type              *string `gorm:"column:type" json:"type"`
	Difficulty        *string `gorm:"column:difficulty" json:"difficulty"`
	ModuleAssociation *string `gorm:"column:module_association" json:"module_association"`
	InitialCode       *string `gorm:"column:initial_code" json:"initial_code"`
	ReferenceSolution *string `gorm:"column:reference_solution" json:"reference_solution"`
}

func (AssessmentQuestion) TableName() string { return "assessment_questions" }

type AssessmentGradingRule struct {
	ID        string         `gorm:"column:id;primaryKey" json:"id"`
	Rules     datatypes.JSON `gorm:"column:rules" json:"rules"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updated_at"`
}

func (AssessmentGradingRule) TableName() string { return "assessment_grading_rules" }
