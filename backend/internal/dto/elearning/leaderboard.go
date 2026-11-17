package dto

// Bentuk UserProfile hasil getLeaderboard; email, lastActive, createdAt, role, assessmentAccess berisi nilai tetap kode lama.
type LeaderboardEntry struct {
	NIM              string         `json:"nim"`
	Nama             *string        `json:"nama"`
	Kelas            *string        `json:"kelas"`
	Jurusan          *string        `json:"jurusan"`
	Shift            *string        `json:"shift"`
	Email            *string        `json:"email"`
	XP               int            `json:"xp"`
	Level            int            `json:"level"`
	Streak           int            `json:"streak"`
	LastActive       string         `json:"lastActive"`
	CreatedAt        string         `json:"createdAt"`
	Role             string         `json:"role"`
	AssessmentAccess map[string]any `json:"assessmentAccess"`
}

type WeeklyEntry struct {
	LeaderboardEntry
	WeeklyXP         int  `json:"weeklyXp"`
	LessonXP         int  `json:"lessonXp"`
	BughuntXP        int  `json:"bughuntXp"`
	LessonsCompleted int  `json:"lessonsCompleted"`
	IsCarry          bool `json:"isCarry"`
	CarryXP          int  `json:"carryXp"`
}

type RankResponse struct {
	Rank *int64 `json:"rank"`
}

type LeaderboardFilters struct {
	Kelas   []string `json:"kelas"`
	Jurusan []string `json:"jurusan"`
	Shift   []string `json:"shift"`
}
