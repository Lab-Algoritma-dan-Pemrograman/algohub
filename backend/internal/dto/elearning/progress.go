package dto

import (
	"encoding/json"
	"time"
)

type CompleteLessonRequest struct {
	XPReward float64 `json:"xpReward"`
}

// XP dan level adalah nilai optimistis completeLesson; untuk peran selain admin/kordas DB menimpanya dari sumber XP.
type CompleteLessonResponse struct {
	XP               int           `json:"xp"`
	Level            int           `json:"level"`
	Streak           int           `json:"streak"`
	LastActive       *time.Time    `json:"lastActive"`
	LeveledUp        bool          `json:"leveledUp"`
	AlreadyCompleted bool          `json:"alreadyCompleted"`
	NewlyUnlocked    []Achievement `json:"newlyUnlocked"`
}

type GrantXPRequest struct {
	Amount int `json:"amount"`
}

type XPResponse struct {
	XP          int          `json:"xp"`
	Level       int          `json:"level"`
	LeveledUp   bool         `json:"leveledUp"`
	Achievement *Achievement `json:"achievement"`
}

type StudyTimeRequest struct {
	Seconds int64 `json:"seconds" binding:"gte=0"`
}

type StudyTimeResponse struct {
	StudyTime *int64 `json:"studyTime"`
}

type StreakResponse struct {
	Streak     int       `json:"streak"`
	LastActive time.Time `json:"lastActive"`
}

// Bentuk UserProfile di src/store/useStore.ts yang diisi SupabaseProvider.
type Profile struct {
	NIM                  string          `json:"nim"`
	Nama                 string          `json:"nama"`
	Kelas                string          `json:"kelas"`
	Jurusan              *string         `json:"jurusan"`
	Shift                *string         `json:"shift"`
	Email                *string         `json:"email"`
	PrestigeTier         *string         `json:"prestigeTier"`
	PrestigeMult         float64         `json:"prestigeMult"`
	PrestigeCount        int             `json:"prestigeCount"`
	XP                   int             `json:"xp"`
	Level                int             `json:"level"`
	Streak               int             `json:"streak"`
	LastActive           *time.Time      `json:"lastActive"`
	CreatedAt            *time.Time      `json:"createdAt"`
	Role                 string          `json:"role"`
	AssessmentAccess     json.RawMessage `json:"assessmentAccess"`
	LevelAccessOverrides json.RawMessage `json:"levelAccessOverrides"`
	StudyTime            int             `json:"studyTime"`
}

type ProfileSyncResponse struct {
	Profile       Profile       `json:"profile"`
	NewlyUnlocked []Achievement `json:"newlyUnlocked"`
}
