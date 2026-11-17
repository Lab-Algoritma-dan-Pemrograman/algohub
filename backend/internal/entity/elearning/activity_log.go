package entity

import "time"

type ActivityLog struct {
	ID        string     `gorm:"column:id;primaryKey" json:"id"`
	NIM       *string    `gorm:"column:nim" json:"nim"`
	Nama      *string    `gorm:"column:nama" json:"nama"`
	EventType string     `gorm:"column:event_type" json:"event_type"`
	Timestamp *time.Time `gorm:"column:timestamp" json:"timestamp"`
	Details   *string    `gorm:"column:details" json:"details"`
	IPAddress *string    `gorm:"column:ip_address" json:"ip_address"`
}

func (ActivityLog) TableName() string { return "activity_logs" }
