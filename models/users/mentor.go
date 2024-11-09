package users

import (
	"time"
)

type MentorProfile struct {
	ID          uint            `gorm:"primaryKey"`
	UserID      uint            `gorm:"uniqueIndex"` // Связь с моделью пользователя
	Name        string          `gorm:"size:255;not null;default:'Unknown'"`
	PhotoURL    string          `gorm:"size:255"`
	City        string          `gorm:"size:100"`
	Position    string          `gorm:"size:100"`
	Company     string          `gorm:"size:255"`
	Skills      []MentorSkill   `gorm:"many2many:mentor_skills"`
	Experience  string          `gorm:"type:text"`
	Education   string          `gorm:"type:text"`
	Services    string          `gorm:"type:text"`
	Rating      float32         `gorm:"default:0"`
	ReviewCount int             `gorm:"default:0"`
	Schedule    []AvailableTime `gorm:"foreignKey:MentorID"`
	SocialLinks SocialLinks     `gorm:"embedded"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MentorSkill struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100;unique;not null;default:'Unknown'"`
}

type AvailableTime struct {
	ID        uint      `gorm:"primaryKey"`
	MentorID  uint      `gorm:"index"` // ForeignKey для связи с MentorProfile
	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
	Mentor    MentorProfile
}

type SocialLinks struct {
	LinkedIn string `gorm:"size:255"`
	Website  string `gorm:"size:255"`
}
