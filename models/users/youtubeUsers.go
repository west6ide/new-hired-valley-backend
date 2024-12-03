package users

import "time"

type YoutubeUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`                   // Уникальный идентификатор GoogleUser
	UserID       uint      `gorm:"not null" json:"user_id"`                // Связь с таблицей User
	GoogleID     string    `gorm:"unique;not null" json:"google_id"`       // Уникальный ID Google-аккаунта
	Email        string    `gorm:"unique;not null" json:"email"`           // Email Google-аккаунта
	FirstName    string    `gorm:"not null" json:"first_name"`             // Имя пользователя
	LastName     string    `gorm:"not null" json:"last_name"`              // Фамилия пользователя
	AccessToken  string    `gorm:"type:text;not null" json:"access_token"` // Токен доступа
	RefreshToken string    `gorm:"type:text" json:"refresh_token"`         // Токен обновления
	Expiry       time.Time `json:"expiry"`                                 // Срок действия токена
	CreatedAt    time.Time `json:"created_at"`                             // Время создания записи
	UpdatedAt    time.Time `json:"updated_at"`                             // Время последнего обновления записи
}
