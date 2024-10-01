package models

import "gorm.io/gorm"

type LinkedInUser struct {
	gorm.Model
	LinkedInID  string `json:"id" gorm:"not null"`
	FirstName   string `json:"localizedFirstName" gorm:"not null"`
	LastName    string `json:"localizedLastName" gorm:"not null"`
	Email       string `json:"email" gorm:"not null;unique"` // Email уникальный
	AccessToken string `json:"access_token"`                 // Поле для хранения токена
}
