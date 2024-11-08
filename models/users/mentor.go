package users

import (
	"time"
)

// Модель наставника
type MentorProfile struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"uniqueIndex" json:"user_id"` // Связь с таблицей пользователей
	Name            string         `gorm:"size:255;not null" json:"name"`
	Specialization  string         `gorm:"size:255" json:"specialization"`
	Industry        string         `gorm:"size:255" json:"industry"`
	ExperienceYears int            `gorm:"not null" json:"experience_years"`
	HourlyRate      float64        `gorm:"not null" json:"hourly_rate"`
	Bio             string         `gorm:"type:text" json:"bio"`
	Availability    []Availability `gorm:"foreignKey:MentorID" json:"availability"`
	Reviews         []Review       `gorm:"foreignKey:MentorID" json:"reviews"`
	AverageRating   float64        `gorm:"-" json:"average_rating"` // Поле для среднего рейтинга (вычисляемое)
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Модель сеанса наставничества
type MentorshipSession struct {
	ID       uint `gorm:"primaryKey"`
	MentorID uint `gorm:"index"`
	UserID   uint `gorm:"index"`
	Date     time.Time
	Status   string `gorm:"size:50;default:'Pending'"`
	Review   string `gorm:"type:text"`
}

// Модель доступности
type Availability struct {
	ID        uint `gorm:"primaryKey"`
	MentorID  uint
	StartTime time.Time
	EndTime   time.Time
}

// Модель отзывов
type Review struct {
	ID        uint   `gorm:"primaryKey"`
	MentorID  uint   `gorm:"index"`
	UserID    uint   `gorm:"index"`    // пользователь, оставивший отзыв
	Rating    int    `gorm:"not null"` // Рейтинг от 1 до 5
	Comment   string `gorm:"type:text"`
	CreatedAt time.Time
}
