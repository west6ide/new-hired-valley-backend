package users

import (
	"gorm.io/gorm"
	"time"
)

type GoogleUser struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   // Foreign Key к User
	User        User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей User
	GoogleID    string `gorm:"unique_index"`
	Email       string `gorm:"not null"`
	FirstName   string
	LastName    string
	AccessToken string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
