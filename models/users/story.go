package users

import (
	"time"
)

type Story struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"` // Добавляем JSON-тег
	User       User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Content    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ExpiresAt  time.Time
	IsArchived bool `gorm:"default:false"`
}
