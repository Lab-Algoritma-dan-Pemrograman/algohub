package entity

type PushSubscription struct {
	UserID   *int64
	Endpoint string
	P256dh   string
	Auth     string
}

// Hanya kolom yang dibaca cron; nilai NULL menjadi string kosong karena kode lama memakai `x || default`.
type ShiftAssignment struct {
	UserID       *int64
	StartTime    string
	UserRole     string
	TaskRole     string
	ActivityName string
}

// AutoNotifyEnabled nil berarti kolom NULL atau baris tidak ada, yang di kode lama dianggap aktif.
type WASettings struct {
	Provider          string
	Token             string
	AutoNotifyEnabled *bool
}
