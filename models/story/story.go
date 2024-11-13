package story

import (
	"time"
)

type Story struct {
	ID         uint       `gorm:"primaryKey"`
	UserID     uint       `gorm:"index"`
	ContentURL string     `gorm:"type:text"`      // Ссылка на медиафайл (фото или видео)
	CreatedAt  time.Time  `gorm:"autoCreateTime"` // Время создания истории
	ExpireAt   time.Time  // Время истечения истории
	IsArchived bool       `gorm:"default:false"`    // Флаг, сохранена ли история в архиве
	Views      uint       `gorm:"default:0"`        // Счетчик просмотров
	Privacy    string     `gorm:"default:'public'"` // Приватность (public, friends, private)
	Reactions  []Reaction // Реакции пользователей
}

type Reaction struct {
	ID        uint      `gorm:"primaryKey"`
	StoryID   uint      `gorm:"index"`
	UserID    uint      `gorm:"index"`
	Emoji     string    `gorm:"type:varchar(10)"` // Реакция (например, 😊, ❤️)
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type ViewStory struct {
	ID       uint      `gorm:"primaryKey"`
	StoryID  uint      `gorm:"index"`
	UserID   uint      `gorm:"index"`
	ViewedAt time.Time `gorm:"autoCreateTime"`
}
