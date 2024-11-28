package courses

import (
	"time"
)

type Lesson struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `json:"title"`         // Название урока
	Content        string    `json:"content"`       // Описание или текст урока
	VimeoVideoLink string    `json:"video_url"`     // Ссылка на видео
	CourseID       uint      `json:"course_id"`     // ID курса, к которому принадлежит урок
	InstructorID   uint      `json:"instructor_id"` // ID инструктора, создавшего урок
	CreatedAt      time.Time `json:"created_at"`    // Время создания
	UpdatedAt      time.Time `json:"updated_at"`    // Время обновления
}
