package models

import (
	"hired-valley-backend/models/users"
	"time"
)

type Story struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"not null"`
	User       users.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Ассоциация с пользователем
	Content    string     `gorm:"type:text;not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	ExpiresAt  time.Time
	IsArchived bool `gorm:"default:false"`
}
