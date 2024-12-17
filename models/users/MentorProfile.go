package users

import "time"

type MentorProfile struct {
	ID             uint `gorm:"primaryKey;autoIncrement:false"` // Отключаем автоинкремент
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
	ID        uint      `gorm:"primaryKey" json:"id"`
	MentorID  uint      `gorm:"not null" json:"mentor_id"`
	UserID    *uint     `gorm:"default:null" json:"user_id"`
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`
	IsBooked  bool      `gorm:"default:false" json:"is_booked"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"` // Связь с пользователем
}

type NotificationMentor struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	Message   string
	IsRead    bool
	CreatedAt time.Time
}
