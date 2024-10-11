package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `json:"name"`                                // Имя пользователя
	Email        string `json:"email" gorm:"unique" gorm:"not null"` // Email пользователя, уникальный в базе данных
	Password     string `json:"-" gorm:"not null"`                   // Хэш пароля (не передается в JSON)
	Position     string `json:"position"`                            // Должность или позиция пользователя
	City         string `json:"city"`                                // Город пользователя
	Income       int    `json:"income"`                              // Уровень дохода пользователя
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refreshToken"` // Access токен
	Provider     string `json:"provider"`     // Обычная авторизация, Google или LinkedIn
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
