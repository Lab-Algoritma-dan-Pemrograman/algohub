package entity

import "time"

type PlaygroundExample struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	Title       string     `gorm:"column:title" json:"title"`
	Description *string    `gorm:"column:description" json:"description"`
	Language    string     `gorm:"column:language" json:"language"`
	Code        string     `gorm:"column:code" json:"code"`
	Category    *string    `gorm:"column:category" json:"category"`
	Difficulty  *string    `gorm:"column:difficulty" json:"difficulty"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	SortOrder   int        `gorm:"column:sort_order" json:"sort_order"`
}

func (PlaygroundExample) TableName() string { return "playground_examples" }
