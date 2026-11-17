package dto

// TargetUserID bertipe any karena frontend lama mengirim angka maupun string.
type SendPushRequest struct {
	TargetUserID any    `json:"targetUserId"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	URL          string `json:"url"`
}

type SendWARequest struct {
	TargetPhone  string `json:"targetPhone"`
	TargetUserID any    `json:"targetUserId"`
	Message      string `json:"message"`
}
