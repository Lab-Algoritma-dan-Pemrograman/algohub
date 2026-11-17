package dto

type LevelAccessRequest struct {
	LevelAccessOverrides map[string]any `json:"levelAccessOverrides" binding:"required"`
}

type RoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type ResetLevelRequest struct {
	LevelID string `json:"levelId" binding:"required"`
}

type ResetLessonRequest struct {
	LessonID    string  `json:"lessonId" binding:"required"`
	LessonTitle *string `json:"lessonTitle"`
}

type AdjustXPRequest struct {
	XP *float64 `json:"xp" binding:"required"`
}

type ActivityLogRequest struct {
	EventType string  `json:"eventType" binding:"required"`
	Details   string  `json:"details"`
	Nama      *string `json:"nama"`
}
