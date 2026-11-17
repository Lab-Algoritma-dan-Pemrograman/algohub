package entity

import "encoding/json"

// Baris DB elearning dibaca lewat to_json. Id disimpan sebagai JSON mentah dan field lain bertipe
// longgar karena kode lama membandingkannya dengan aturan JavaScript tanpa mengenal tipe kolom.
type ElearningLevel struct {
	ID         json.RawMessage `json:"id"`
	Title      *string         `json:"title"`
	SortOrder  *float64        `json:"sort_order"`
	AccessMode any             `json:"access_mode"`
	Locked     any             `json:"locked"`
}

type ElearningModule struct {
	ID      json.RawMessage `json:"id"`
	LevelID json.RawMessage `json:"level_id"`
}

type ElearningLesson struct {
	ID       json.RawMessage `json:"id"`
	ModuleID json.RawMessage `json:"module_id"`
}

type ElearningStudentProgress struct {
	LessonID    json.RawMessage `json:"lesson_id"`
	Completed   any             `json:"completed"`
	CompletedAt any             `json:"completed_at"`
}

type ElearningSession struct {
	LastHeartbeat json.RawMessage `json:"last_heartbeat"`
}

// Urutan field = urutan key respons sync-elearning lama.
type ElearningProgress struct {
	NIM                  string          `json:"nim"`
	StudentName          string          `json:"student_name"`
	CompletedLessons     int             `json:"completed_lessons"`
	TotalLessons         int             `json:"total_lessons"`
	CompletionPercentage float64         `json:"completion_percentage"`
	IsCompleted          bool            `json:"is_completed"`
	CompletedLevels      []*string       `json:"completed_levels"`
	CurrentLevel         string          `json:"current_level"`
	LastAccessedAt       json.RawMessage `json:"last_accessed_at"`
}

// Baris tabel elearning_progress di DB siakad; lessons_completed diisi sama dengan completed_lessons
// karena kedua nama kolom dipakai di skema lama.
type ElearningProgressRow struct {
	NIM                  string          `json:"nim"`
	StudentName          string          `json:"student_name"`
	CompletedLessons     int             `json:"completed_lessons"`
	LessonsCompleted     int             `json:"lessons_completed"`
	TotalLessons         int             `json:"total_lessons"`
	CompletionPercentage float64         `json:"completion_percentage"`
	IsCompleted          bool            `json:"is_completed"`
	CompletedLevels      []*string       `json:"completed_levels"`
	CurrentLevel         string          `json:"current_level"`
	LastAccessedAt       json.RawMessage `json:"last_accessed_at"`
	UpdatedAt            string          `json:"updated_at"`
}
