package users

import (
	"gorm.io/gorm"
	"time"
)

type LinkedInUser struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   // Foreign Key к User
	User        User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей User
	FirstName   string `json:"localizedFirstName" gorm:"not null"`
	LastName    string `json:"localizedLastName" gorm:"not null"`
	Email       string `json:"email" gorm:"not null;unique"` // Email уникальный
	Sub         string `gorm:"unique"`                       // LinkedIn OpenID идентификатор
	AccessToken string `json:"access_token"`                 // Поле для хранения токена
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
