package main

import (
	"log"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/handlers/root"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal("config:", err)
	}
	database.InitDB()

	r := gin.Default()
	r.LoadHTMLGlob("views/root/*.html")

	rootGroup := r.Group("/root")
	{
		rootGroup.GET("", root.DashboardHandler(database.DB))
		rootGroup.GET("/", root.DashboardHandler(database.DB))
	}

	log.Printf("listening on :%s", config.C.Server.Port)
	if err := r.Run(":" + config.C.Server.Port); err != nil {
		log.Fatal(err)
	}
}
