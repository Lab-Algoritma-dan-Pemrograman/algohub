package entity

import "time"

// attempt_id hanya diketahui dari monitoringService.ts (tanpa DDL di repo lama).
type ActiveSession struct {
	NIM             string     `gorm:"column:nim;primaryKey" json:"nim"`
	Nama            string     `gorm:"column:nama" json:"nama"`
	Kelas           string     `gorm:"column:kelas" json:"kelas"`
	LastHeartbeat   *time.Time `gorm:"column:last_heartbeat" json:"last_heartbeat"`
	CurrentActivity *string    `gorm:"column:current_activity" json:"current_activity"`
	AttemptID       *string    `gorm:"column:attempt_id" json:"attempt_id"`
}

func (ActiveSession) TableName() string { return "active_sessions" }
