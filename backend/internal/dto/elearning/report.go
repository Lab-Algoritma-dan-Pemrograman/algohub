package dto

import "encoding/json"

// NIM any karena api/report.ts membandingkan dengan claim token memakai !== (tipe ikut dibandingkan).
type ReportRequest struct {
	Token string `json:"token"`
	NIM   any    `json:"nim"`
}

type ReportRecalculated struct {
	Completed  int64   `json:"completed"`
	Total      int64   `json:"total"`
	Percentage float64 `json:"percentage"`
}

type ReportResponse struct {
	Success      bool               `json:"success"`
	Recalculated ReportRecalculated `json:"recalculated"`
}

// RawMessage supaya selectedOption absen (400) bisa dibedakan dari null (dihitung Number(null) = 0).
type ValidateQuizRequest struct {
	LessonID       json.RawMessage `json:"lessonId"`
	SelectedOption json.RawMessage `json:"selectedOption"`
}

// CorrectAnswer nil bila Number(correctAnswer) NaN, sama dengan JSON.stringify(NaN) = null.
type ValidateQuizResponse struct {
	IsCorrect     bool     `json:"isCorrect"`
	CorrectAnswer *float64 `json:"correctAnswer"`
}
