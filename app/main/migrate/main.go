package main

import (
	"log"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/models"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal("config:", err)
	}
	database.InitDB()

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Genus{},
		&models.Tool{},
		&models.List{},
		&models.ListTool{},
	); err != nil {
		log.Fatal("migrate:", err)
	}

	log.Println("migrations applied")
}
