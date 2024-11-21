package career

import "time"

type PlanCareer struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `json:"user_id"`
	ShortTermGoals string `json:"short_term_goals"`
	LongTermGoals  string `json:"long_term_goals"`
	Steps          string `json:"steps" gorm:"type:text"` // JSON-строка с шагами
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
