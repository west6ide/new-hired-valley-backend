package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `json:"name"`                         // Имя пользователя
	Email        string `json:"email" gorm:"unique;not null"` // Email пользователя, уникальный в базе данных
	Password     string `json:"-" gorm:"not null"`            // Хэш пароля (не передается в JSON)
	Position     string `json:"position"`                     // Должность или позиция пользователя
	City         string `json:"city"`                         // Город пользователя
	Income       int    `json:"income"`                       // Уровень дохода пользователя
	AccessToken  string `json:"token"`                        // Access токен
	RefreshToken string `json:"refreshToken"`                 // Refresh токен
	Provider     string `json:"provider"`                     // Обычная авторизация, Google или LinkedIn
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
