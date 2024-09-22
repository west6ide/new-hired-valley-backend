package config

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
)

var (
	DB    *gorm.DB
	Store = sessions.NewCookieStore([]byte("something-very-secret"))
)

func InitDB() {
	var err error
	dsn := "host=dpg-cro07388fa8c738lmgng-a user=db password=8moWFtexXkL3DAmI2C3PFGqXjQv0G8jc dbname=hired_valley sslmode=disable"
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connection established")

}
