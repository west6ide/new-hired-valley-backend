package models

import "time"

type Progress struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	CourseID  uint
	ModuleID  uint
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
