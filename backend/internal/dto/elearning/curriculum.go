package dto

import "encoding/json"

// Bentuk Level di src/data/curriculum.ts yang dikirim saveFullCurriculum.
type CurriculumRequest struct {
	Levels []CurriculumLevel `json:"levels" binding:"required"`
}

type CurriculumLevel struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	AccessMode  string             `json:"accessMode"`
	Locked      bool               `json:"locked"`
	Modules     []CurriculumModule `json:"modules"`
}

type CurriculumModule struct {
	ID      string             `json:"id"`
	Title   string             `json:"title"`
	Lessons []CurriculumLesson `json:"lessons"`
}

type CurriculumLesson struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Explanation     *string         `json:"explanation"`
	CodeExample     *string         `json:"codeExample"`
	InitialCode     *string         `json:"initialCode"`
	Solution        *string         `json:"solution"`
	Hint            *string         `json:"hint"`
	Quiz            json.RawMessage `json:"quiz"`
	TestCases       json.RawMessage `json:"testCases"`
	ValidationRules json.RawMessage `json:"validationRules"`
}
