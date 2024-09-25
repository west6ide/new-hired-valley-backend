package models

import (
	"gorm.io/gorm"
)

// LinkedInUser представляет пользователя LinkedIn
type LinkedInUser struct {
	gorm.Model
	LinkedInID string `json:"id"`
	FirstName  string `json:"localizedFirstName"`
	LastName   string `json:"localizedLastName"`
	Email      string `json:"emailAddress"`
}
