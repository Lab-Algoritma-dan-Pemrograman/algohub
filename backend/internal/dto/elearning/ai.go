package dto

import "encoding/json"

// FileMimeType pointer karena default "application/pdf" hanya berlaku bila field tidak dikirim (api/ai.ts).
type AIRequest struct {
	Prompt           string          `json:"prompt"`
	Model            string          `json:"model"`
	Token            string          `json:"token"`
	ResponseMimeType string          `json:"responseMimeType"`
	ResponseSchema   json.RawMessage `json:"responseSchema"`
	FileData         string          `json:"fileData"`
	FileMimeType     *string         `json:"fileMimeType"`
}

type AIResponse struct {
	Text string `json:"text"`
}

// RawMessage supaya field absen, null, dan tipe lain bisa dibedakan seperti destructuring default di api/grade.ts.
type GradeRequest struct {
	Token          string          `json:"token"`
	AttemptIDs     json.RawMessage `json:"attemptIds"`
	RequestedModel json.RawMessage `json:"requestedModel"`
}

type GradeResult struct {
	AttemptID   any    `json:"attemptId"`
	Success     bool   `json:"success"`
	FinalScore  any    `json:"finalScore,omitempty"`
	StudentName string `json:"studentName,omitempty"`
	Error       string `json:"error,omitempty"`
}

type GradeResponse struct {
	Success     bool          `json:"success"`
	GradedCount int           `json:"gradedCount"`
	Results     []GradeResult `json:"results"`
}
