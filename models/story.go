package models

import "time"

type Story struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null"`
	Content    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ExpiresAt  time.Time
	IsArchived bool `gorm:"default:false"`
}
