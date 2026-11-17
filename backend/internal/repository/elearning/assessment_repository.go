package repository

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
)

type AssessmentRepository interface {
	Attempt(ctx context.Context, id string) (*entity.AssessmentAttempt, error)
	Questions(ctx context.Context, ids []string) ([]entity.AssessmentQuestion, error)
	GradingRules(ctx context.Context, menuType string) (json.RawMessage, error)
	SaveGrade(ctx context.Context, id string, aiGrades json.RawMessage, finalScore any, gradedAt time.Time) error
}

type assessmentRepository struct{ db *gorm.DB }

func NewAssessmentRepository(db *gorm.DB) AssessmentRepository {
	return &assessmentRepository{db: db}
}

func (r *assessmentRepository) Attempt(ctx context.Context, id string) (*entity.AssessmentAttempt, error) {
	var a entity.AssessmentAttempt
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Tanpa ORDER BY, sama dengan .in('id', ...) di api/grade.ts yang memakai urutan fisik tabel.
func (r *assessmentRepository) Questions(ctx context.Context, ids []string) ([]entity.AssessmentQuestion, error) {
	var qs []entity.AssessmentQuestion
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&qs).Error
	return qs, err
}

func (r *assessmentRepository) GradingRules(ctx context.Context, menuType string) (json.RawMessage, error) {
	var g entity.AssessmentGradingRule
	if err := r.db.WithContext(ctx).Select("rules").Where("id = ?", menuType).Take(&g).Error; err != nil {
		return nil, err
	}
	return json.RawMessage(g.Rules), nil
}

func (r *assessmentRepository) SaveGrade(ctx context.Context, id string, aiGrades json.RawMessage, finalScore any, gradedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.AssessmentAttempt{}).Where("id = ?", id).Updates(map[string]any{
		"ai_grades":   datatypes.JSON(aiGrades),
		"final_score": finalScore,
		"status":      "graded",
		"graded_at":   gradedAt,
	}).Error
}
