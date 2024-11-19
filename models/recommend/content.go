package recommend

import "time"

type Recommendation struct {
	ID         uint    `gorm:"primaryKey"`
	Title      string  `gorm:"not null"`
	Content    string  `gorm:"type:text"`
	SkillID    *uint   `gorm:"index"`     // Nullable для рекомендаций по интересам
	InterestID *uint   `gorm:"index"`     // Nullable для рекомендаций по навыкам
	Popularity float64 `gorm:"default:0"` // Метрика популярности
	Relevance  float64 `gorm:"default:0"` // Релевантность для пользователя
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
