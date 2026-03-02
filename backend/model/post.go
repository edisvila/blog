package model

import "time"

type Post struct {
	ID        int    `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Slug      string `gorm:"uniqueIndex;not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
