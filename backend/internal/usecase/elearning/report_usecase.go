package usecase

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"time"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

type ReportUsecase struct {
	auth       *AuthUsecase
	curriculum repository.CurriculumRepository
	progress   repository.ProgressRepository
	reports    repository.ReportRepository
	now        func() time.Time
}

func NewReportUsecase(auth *AuthUsecase, curriculum repository.CurriculumRepository, progress repository.ProgressRepository, reports repository.ReportRepository, now func() time.Time) *ReportUsecase {
	return &ReportUsecase{auth: auth, curriculum: curriculum, progress: progress, reports: reports, now: now}
}

// api/report.ts mengembalikan error.status || 500 dengan error.message; error PostgREST tidak punya status.
func reportError(err error) error {
	return newError(http.StatusInternalServerError, dbMessage(err))
}

func (uc *ReportUsecase) Sync(ctx context.Context, token string, nim any) (*dto.ReportResponse, error) {
	if token == "" || !jsTruthy(nim) {
		return nil, BadRequest("Missing token or nim")
	}
	claims, err := uc.auth.verifyToken(token)
	if err != nil {
		return nil, Unauthorized("Unauthorized: Invalid token")
	}
	if !jsStrictEqual(claims["nim"], nim) {
		return nil, Forbidden("Identity mismatch")
	}
	nimStr := jsToString(nim)

	total, err := uc.curriculum.CountLessons(ctx)
	if err != nil {
		return nil, reportError(err)
	}
	completed, err := uc.progress.CountCompleted(ctx, nimStr)
	if err != nil {
		return nil, reportError(err)
	}
	levels, err := uc.curriculum.LevelTree(ctx)
	if err != nil {
		return nil, reportError(err)
	}
	ids, err := uc.progress.CompletedLessonIDs(ctx, nimStr)
	if err != nil {
		return nil, reportError(err)
	}
	done := make(map[string]bool, len(ids))
	for _, id := range ids {
		done[id] = true
	}

	completedLevels := []string{}
	currentLevel := ""
	for i := range levels {
		levelDone := 0
		for _, id := range levels[i].LessonIDs {
			if done[id] {
				levelDone++
			}
		}
		levelTotal := len(levels[i].LessonIDs)
		if levelTotal > 0 && levelDone >= levelTotal {
			completedLevels = append(completedLevels, levels[i].Title)
		} else if levelDone > 0 && currentLevel == "" {
			currentLevel = levels[i].Title
		}
	}

	percentage := 0.0
	if total > 0 {
		percentage = math.Round(float64(completed)/float64(total)*10000) / 100
	}
	levelsJSON, err := json.Marshal(completedLevels)
	if err != nil {
		return nil, err
	}
	studentName := jsToString(pick(claims, "Student", "nama", "full_name"))
	if currentLevel == "" {
		currentLevel = "Introduction"
	}
	now := uc.now().UTC()
	err = uc.reports.Upsert(ctx, &entity.ElearningProgress{
		NIM:                  nimStr,
		StudentName:          &studentName,
		LessonsCompleted:     int(completed),
		TotalLessons:         int(total),
		CompletionPercentage: percentage,
		IsCompleted:          total > 0 && completed >= total,
		CompletedLevels:      levelsJSON,
		CurrentLevel:         &currentLevel,
		LastAccessedAt:       &now,
	})
	if err != nil {
		return nil, reportError(err)
	}
	return &dto.ReportResponse{
		Success:      true,
		Recalculated: dto.ReportRecalculated{Completed: completed, Total: total, Percentage: percentage},
	}, nil
}

// Operator === JavaScript untuk nilai JSON primitif; objek dan array tidak pernah sama.
func jsStrictEqual(a, b any) bool {
	switch x := a.(type) {
	case string:
		y, ok := b.(string)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case json.Number, float64:
		switch b.(type) {
		case json.Number, float64:
			return jsNumber(a) == jsNumber(b)
		}
	case nil:
		return b == nil
	}
	return false
}
