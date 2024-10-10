package coursesModels

import "time"

type Module struct {
	ID        uint `gorm:"primaryKey"`
	CourseID  uint
	Course    Course `gorm:"foreignKey:CourseID"`
	Title     string `gorm:"size:255;not null"`
	VideoURL  string `gorm:"size:500;not null"`
	Content   string `gorm:"type:text"`
	Order     int    `gorm:"not null"`
	Duration  int    `gorm:"not null"` // длительность урока в минутах
	CreatedAt time.Time
	UpdatedAt time.Time
}
