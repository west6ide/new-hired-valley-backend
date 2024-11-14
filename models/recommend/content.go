package recommend

// Content представляет структуру одной рекомендации
type Content struct {
	Title       string   `json:"title"`       // Название контента
	Description string   `json:"description"` // Описание или краткое содержание
	ContentURL  string   `json:"content_url"` // Ссылка на материал
	Category    string   `json:"category"`    // Категория (например, "Программирование", "Финансы")
	SkillLevel  string   `json:"skill_level"` // Уровень сложности (например, "Новичок", "Средний", "Эксперт")
	Tags        []string `json:"tags"`        // Теги, описывающие содержание
}
