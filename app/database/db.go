package database

import (
	"log"

	"github.com/dariubs/altern/app/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	db, err := gorm.Open(postgres.Open(config.C.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database instance:", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	log.Println("Database connected")
}
