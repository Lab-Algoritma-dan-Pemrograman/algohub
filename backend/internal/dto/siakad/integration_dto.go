package dto

type AnalyzeScheduleRequest struct {
	FileBase64 string `json:"fileBase64"`
	MimeType   string `json:"mimeType"`
	Model      string `json:"model"`
}

type GenerateJWTRequest struct {
	HandshakeCode string `json:"handshake_code"`
}
