package models

import (
	"time"
)

// Модель для хранения OAuth токена
type OAuthToken struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      // Внешний ключ к User
	AccessToken string    // OAuth токен
	TokenType   string    // Тип токена
	Expiry      time.Time // Время истечения токена
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
