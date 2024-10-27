package users

import (
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"time"
)

//type User struct {
//	ID           uint   `gorm:"primaryKey"`
//	Name         string `json:"name"`                              // Имя пользователя
//	Email        string `json:"email" gorm:"unique;not null"`      // Email пользователя, уникальный в базе данных
//	Password     string `json:"-" gorm:"not null"`                 // Хэш пароля (не передается в JSON)
//	Position     string `json:"position"`                          // Должность или позиция пользователя
//	City         string `json:"city"`                              // Город пользователя
//	Income       int    `json:"income"`                            // Уровень дохода пользователя
//	Role         string `json:"role" gorm:"not null;default:user"` // Роль пользователя (например, user или admin)
//	AccessToken  string `json:"token"`                             // Access токен
//	RefreshToken string `json:"refreshToken"`                      // Refresh токен
//	Provider     string `json:"provider"`                          // Обычная авторизация, Google или LinkedIn
//	CreatedAt    time.Time
//	UpdatedAt    time.Time
//	DeletedAt    gorm.DeletedAt `gorm:"index"`
//}

// user.go

type User struct {
	ID                 uint       `gorm:"primaryKey"`
	Name               string     `json:"name"`
	Email              string     `json:"email" gorm:"unique;not null"`
	Password           string     `json:"-" gorm:"not null"`
	Position           string     `json:"position"`
	City               string     `json:"city"`
	Income             int        `json:"income"`
	Role               string     `json:"role" gorm:"not null;default:user"`
	Skills             []Skill    `json:"skills" gorm:"many2many:user_skills"`
	Interests          []Interest `json:"interests" gorm:"many2many:user_interests"`
	ContentPreferences string     `gorm:"type:text"`
	Visibility         string     `json:"visibility" gorm:"default:'public'"` // Контроль видимости профиля
	AccessToken        string     `json:"token"`
	RefreshToken       string     `json:"refreshToken"`
	Provider           string     `json:"provider"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

type Skill struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type Interest struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

func GetUserByID(userID interface{}) (*User, error) {
	var user User
	result := config.DB.First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
