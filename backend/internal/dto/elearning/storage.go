package dto

type ImageFile struct {
	Name        string
	ContentType string
	Content     []byte
}

type UploadResponse struct {
	PublicURL string `json:"publicUrl"`
}
