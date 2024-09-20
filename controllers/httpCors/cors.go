package httpCors

import (
	"github.com/rs/cors"
)

func CorsSettings() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowedOrigins:     []string{"*"}, // Установите конкретные домены, если нужно ограничить доступ
		AllowCredentials:   true,
		AllowedHeaders:     []string{"Content-Type", "Authorization"},
		OptionsPassthrough: true,
		ExposedHeaders:     []string{"Authorization"},
		Debug:              true,
	})
	return c
}
