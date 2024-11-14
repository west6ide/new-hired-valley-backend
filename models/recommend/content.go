package recommend

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Content представляет структуру одной рекомендации
type Content struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ContentURL  string      `json:"content_url"`
	Category    string      `json:"category"`
	SkillLevel  string      `json:"skill_level"`
	Tags        StringArray `json:"tags" gorm:"type:json"` // Используем кастомный тип для массива строк
}

// StringArray представляет собой кастомный тип для []string, чтобы сохранять его как JSON в базе данных
type StringArray []string

// Value сериализует массив в JSON при сохранении в базе данных
func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan десериализует JSON из базы данных в массив
func (a *StringArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported data type: %T", value)
	}
	return json.Unmarshal(b, &a)
}
