package dto

import "time"

// touchStreak dan studySeconds menggabungkan heartbeat, streak, dan sinkron waktu belajar 30 detik Layout.tsx ke satu request.
type HeartbeatRequest struct {
	CurrentActivity *string          `json:"currentActivity"`
	Nama            string           `json:"nama"`
	Kelas           string           `json:"kelas"`
	AttemptID       Optional[string] `json:"attemptId"`
	TouchStreak     bool             `json:"touchStreak"`
	StudySeconds    *int64           `json:"studySeconds" binding:"omitempty,gte=0"`
}

// SessionSaved false berarti upsert active_sessions ditolak RLS dan ditelan seperti Layout.tsx.
type HeartbeatResponse struct {
	SessionSaved bool               `json:"sessionSaved"`
	Streak       *StreakResponse    `json:"streak,omitempty"`
	StudyTime    *StudyTimeResponse `json:"studyTime,omitempty"`
}

// Bentuk select('nim, last_heartbeat, users(role)') di StudentMonitoring.tsx.
type ActiveSessionRole struct {
	NIM           string       `json:"nim"`
	LastHeartbeat *time.Time   `json:"last_heartbeat"`
	Users         *SessionUser `json:"users"`
}

type SessionUser struct {
	Role *string `json:"role"`
}
