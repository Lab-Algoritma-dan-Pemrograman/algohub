package dto

// Bentuk Achievement di achievementService.ts; RequirementValue angka bila Number() valid, selain itu string.
type Achievement struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Icon             *string `json:"icon"`
	RequirementType  *string `json:"requirementType"`
	RequirementValue any     `json:"requirementValue"`
}

// Field absen dibedakan dari nol karena checkAndUnlockAchievements memakai stats.x !== undefined.
type AchievementCheckRequest struct {
	XP                *float64 `json:"xp"`
	GamesPlayed       *float64 `json:"gamesPlayed"`
	PerfectGames      *float64 `json:"perfectGames"`
	LessonCount       *float64 `json:"lessonCount"`
	CompletedLevelIDs []string `json:"completedLevelIds"`
}

type CheckXPRequest struct {
	XP float64 `json:"xp"`
}

type CheckXPResponse struct {
	Achievement *Achievement `json:"achievement"`
}
