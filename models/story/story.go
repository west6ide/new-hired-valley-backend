package story

import (
	"time"
)

type Story struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"index;constraint:OnDelete:CASCADE;not null"`
	ContentURL  string `gorm:"type:text;not null"` // Ссылка на медиафайл (фото или видео)
	DriveFileID string
	CreatedAt   time.Time  `gorm:"autoCreateTime"` // Время создания истории
	ExpireAt    time.Time  // Время истечения истории
	IsArchived  bool       `gorm:"default:false"`      // Флаг, сохранена ли история в архиве
	Views       uint       `gorm:"default:0"`          // Счетчик просмотров
	Privacy     string     `gorm:"default:'public'"`   // Приватность (public, friends, private)
	Reactions   []Reaction `gorm:"foreignKey:StoryID"` // Реакции пользователей
	Comments    []Comment  `gorm:"foreignKey:StoryID"`
}

type Reaction struct {
	ID        uint      `gorm:"primaryKey"`
	StoryID   uint      `gorm:"index;not null"`
	UserID    uint      `gorm:"index;constraint:OnDelete:CASCADE;not null"`
	Emoji     string    `gorm:"type:varchar(20);not null"` // Реакция (например, 😊, ❤️)
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey"`
	StoryID   uint      `gorm:"index;not null"`
	UserID    uint      `gorm:"index;constraint:OnDelete:CASCADE;not null"`
	Content   string    `gorm:"type:text;not null"` // Текст комментария
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type ViewStory struct {
	ID       uint      `gorm:"primaryKey"`
	StoryID  uint      `gorm:"index;not null"`
	UserID   uint      `gorm:"index;constraint:OnDelete:CASCADE;not null"`
	ViewedAt time.Time `gorm:"autoCreateTime"`
}
