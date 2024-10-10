package coursesModels

import "time"

type Review struct {
	ID        uint `gorm:"primaryKey"`
	CourseID  uint
	UserID    uint
	Rating    float64 `gorm:"not null"`
	Comment   string  `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
