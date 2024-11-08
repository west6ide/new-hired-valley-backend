package users

import (
	"time"
)

type Story struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"` // ID пользователя, связанного с историей
	User       User      `gorm:"foreignKey:UserID"`       // Связь с моделью User
	Content    string    `gorm:"type:text"`               // Текст (опционально)
	MediaURL   string    `json:"media_url"`               // URL медиафайла (фото или видео)
	MediaType  string    `json:"media_type"`              // Тип медиа (image или video)
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ExpiresAt  time.Time // Дата истечения истории
	IsArchived bool      `gorm:"default:false"` // Архивирована ли история
	IsPublic   bool      `gorm:"default:true"`  // Видимость истории (публичная или нет)
}
