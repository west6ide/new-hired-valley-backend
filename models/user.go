package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model        // Включает поля ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `json:"name"`                // Имя пользователя
	Email      string `json:"email" gorm:"unique"` // Email пользователя, уникальный в базе данных
	Password   string `json:"-"`                   // Хэш пароля (не передается в JSON)
	Position   string `json:"position"`            // Должность или позиция пользователя
	City       int    `json:"city"`                // Город пользователя
	Income     int    `json:"income"`              // Уровень дохода пользователя
	Token      string `json:"token"`
}
