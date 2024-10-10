package coursesModels

import (
	"hired-valley-backend/models/authenticationUsers"
	"time"
)

type Course struct {
	ID           uint                     `gorm:"primaryKey"`
	Title        string                   `gorm:"size:255;not null"`
	Description  string                   `gorm:"type:text"`
	Instructor   authenticationUsers.User `gorm:"foreignKey:InstructorID"`
	InstructorID uint
	Modules      []Module
	Price        float64
	Rating       float64
	Reviews      []Review
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
