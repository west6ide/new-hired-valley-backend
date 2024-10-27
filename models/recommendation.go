package models

import (
	"gorm.io/gorm"
)

type RecommendedContent struct {
	ID         uint    `gorm:"primaryKey"`
	UserID     uint    `gorm:"index"`      // ID пользователя
	Content    string  `json:"content"`    // Название контента
	Similarity float64 `json:"similarity"` // Сходство (от 0 до 1)
	gorm.Model         // Добавляет поля CreatedAt, UpdatedAt и DeletedAt
}
