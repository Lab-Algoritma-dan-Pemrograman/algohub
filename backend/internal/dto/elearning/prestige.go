package dto

import "time"

// Bentuk PrestigeEligibility di prestigeService.ts.
type PrestigeEligibility struct {
	Eligible       bool       `json:"eligible"`
	Reason         string     `json:"reason"`
	LevelsDone     int        `json:"levelsDone"`
	TotalLevels    int        `json:"totalLevels"`
	LessonsDone    int        `json:"lessonsDone"`
	TotalLessons   int        `json:"totalLessons"`
	Tier           string     `json:"tier"`
	Mult           float64    `json:"mult"`
	PrestigeCount  int        `json:"prestigeCount"`
	LastPrestigeAt *time.Time `json:"lastPrestigeAt"`
	SemesterUsed   int64      `json:"semesterUsed"`
}

// NIM kosong berarti diri sendiri.
type PrestigeResetRequest struct {
	NIM string `json:"nim"`
}

type PrestigeResetResponse struct {
	Tier       string  `json:"tier"`
	Mult       float64 `json:"mult"`
	PrestigeNo float64 `json:"prestigeNo"`
}
