package videos

import "time"

type Video struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	VideoLink   string    `json:"video_link"`
	YouTubeID   string    `json:"youtube_id"`
	UploadedBy  uint      `json:"uploaded_by"` // ID пользователя, загрузившего видео
	LessonID    uint      `json:"lesson_id"`   // Связь с уроком
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
