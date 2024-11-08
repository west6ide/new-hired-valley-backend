package users

import (
	"time"
)

// Модель наставника
type MentorProfile struct {
	ID              uint           `gorm:"primaryKey"`
	UserID          uint           `gorm:"uniqueIndex"`       // Связь с таблицей пользователей
	User            User           `gorm:"foreignKey:UserID"` // Связь с моделью User
	Specialization  string         `gorm:"size:255"`
	Industry        string         `gorm:"size:255"`
	ExperienceYears int            `gorm:"not null"`
	HourlyRate      float64        `gorm:"not null"`
	Bio             string         `gorm:"type:text"`
	Availability    []Availability `gorm:"foreignKey:MentorID"`
	Reviews         []Review       `gorm:"foreignKey:MentorID"`
	AverageRating   float64        `gorm:"-"` // Поле для среднего рейтинга (вычисляемое)
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
