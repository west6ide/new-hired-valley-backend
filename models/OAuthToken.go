package models

import (
	"time"
)

// OAuthToken — структура для хранения OAuth-токена
type OAuthToken struct {
	ID          uint      `gorm:"primaryKey"`     // Уникальный идентификатор токена
	UserID      uint      `gorm:"not null"`       // Связь с пользователем
	AccessToken string    `gorm:"not null"`       // Токен доступа
	TokenType   string    `gorm:"not null"`       // Тип токена (например, "Bearer")
	Expiry      time.Time `gorm:"not null"`       // Время истечения токена
	CreatedAt   time.Time `gorm:"autoCreateTime"` // Время создания записи
	UpdatedAt   time.Time `gorm:"autoUpdateTime"` // Время обновления записи
}
