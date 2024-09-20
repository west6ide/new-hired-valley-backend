package config

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"
)

var (
	DB    *gorm.DB
	Store = sessions.NewCookieStore([]byte("something-very-secret"))
)

func InitDB() {
	var _ error

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	DB, _ = gorm.Open("postgres", dsn)

}
