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

	// Add slug column without unique constraint first so existing rows can be backfilled.
	database.DB.Exec(`ALTER TABLE lists ADD COLUMN IF NOT EXISTS slug TEXT NOT NULL DEFAULT ''`)
	database.DB.Exec(`UPDATE lists SET slug = 'list-' || id::text WHERE slug = ''`)

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Genus{},
		&models.AI{},
		&models.List{},
		&models.ListAI{},
		&models.Artifact{},
		&models.ArtifactField{},
		&models.Page{},
		&models.Alternative{},
		&models.SavedAI{},
	); err != nil {
		log.Fatal("migrate:", err)
	}

	log.Println("migrations applied")
}
