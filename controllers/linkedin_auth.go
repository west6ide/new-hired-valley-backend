package controllers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/models"
	"net/http"
)

func decodeIDToken(idToken string, user *models.LinkedInUser) error {
	// Парсинг JWT
	tkn, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("недействительный метод подписи")
		}

		// Получение открытого ключа LinkedIn
		keySet, err := getLinkedInPublicKeys()
		if err != nil {
			return nil, err
		}

		// Получение kid из токена
		if kid, ok := token.Header["kid"].(string); ok {
			if key, ok := keySet[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("ключ не найден")
	})

	if err != nil {
		return err
	}

	if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
		user.FirstName = claims["given_name"].(string)
		user.LastName = claims["family_name"].(string)
		user.Email = claims["email"].(string)
		return nil
	}

	return fmt.Errorf("недействительный токен")
}

// Функция для получения публичных ключей LinkedIn
func getLinkedInPublicKeys() (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get("https://api.linkedin.com/v2/publicKeys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при получении публичных ключей: %s", resp.Status)
	}

	var keySet map[string]*rsa.PublicKey
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return nil, err
	}

	return keySet, nil
}
