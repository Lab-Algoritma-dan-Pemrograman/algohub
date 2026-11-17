package dto

type KonfigurasiRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}
