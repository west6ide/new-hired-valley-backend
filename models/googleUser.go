package models

import (
	"gorm.io/gorm"
	"time"
)

type GoogleUser struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null"`                                      // Foreign Key к User
	User        User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей User
	GoogleID    string `gorm:"unique_index;not null"`
	Email       string `gorm:"not null"`
	FirstName   string
	LastName    string
	AccessToken string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
