package users

import (
	"gorm.io/gorm"
	"hired-valley-backend/models/story"
	"time"
)

type User struct {
	ID                 uint          `gorm:"primaryKey"`
	Name               string        `json:"name"`
	Email              string        `json:"email" gorm:"unique;not null"`
	Password           string        `json:"-" gorm:"not null"`
	Company            string        `json:"company"`
	Industry           string        `json:"industry"`
	Position           string        `json:"position"`
	City               string        `json:"city"`
	Income             int           `json:"income"`
	Role               string        `json:"role" gorm:"not null;default:user"`
	Skills             []Skill       `json:"skills" gorm:"many2many:user_skills"`
	Interests          []Interest    `json:"interests" gorm:"many2many:user_interests"`
	ContentPreferences string        `gorm:"type:text"`
	Visibility         string        `json:"visibility" gorm:"default:'public'"` // Контроль видимости профиля
	AccessToken        string        `json:"token" gorm:"column:token"`
	RefreshToken       string        `json:"refreshToken"`
	Provider           string        `json:"provider"`
	Stories            []story.Story `gorm:"foreignKey:UserID"` // Связь с историями
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
