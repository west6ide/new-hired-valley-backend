package models

import "gorm.io/gorm"

type LinkedInUser struct {
	gorm.Model
	LinkedInID  string `json:"id"`
	FirstName   string `json:"localizedFirstName"`
	LastName    string `json:"localizedLastName"`
	Email       string `json:"emailAddress"`
	AccessToken string `json:"access_token"` // Поле для хранения токена
}
