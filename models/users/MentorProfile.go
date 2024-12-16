package users

import "time"

type MentorProfile struct {
	ID             uint `gorm:"primaryKey"`
	UserID         uint `gorm:"index;unique"`
	User           User `gorm:"constraint:OnDelete:CASCADE;"`
	Bio            string
	Skills         string
	PricePerHour   float64
	AvailableSlots []Slot `gorm:"foreignKey:MentorID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Slot struct {
	ID        uint `gorm:"primaryKey"`
	MentorID  uint `gorm:"index"`
	UserID    uint `gorm:"index;default:null"` // ID пользователя, который забронировал слот
	StartTime time.Time
	EndTime   time.Time
	IsBooked  bool
	CreatedAt time.Time
	User      User `gorm:"foreignKey:UserID"` // Связь с пользователем
}

type NotificationMentor struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	Message   string
	IsRead    bool
	CreatedAt time.Time
}
