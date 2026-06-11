package main

import (
	"log"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/handlers/root"
	"github.com/dariubs/altern/app/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal("config:", err)
	}
	database.InitDB()

	r := gin.Default()
	r.LoadHTMLGlob("views/root/*.html")

	store := cookie.NewStore([]byte(config.C.Session.Secret))
	store.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 60 * 60 * 24 * 7})
	r.Use(sessions.Sessions("altern_session", store))

	rootGroup := r.Group("/root")
	{
		rootGroup.GET("/login", root.LoginPage())
		rootGroup.GET("/auth/github", root.GitHubLogin())
		rootGroup.GET("/auth/github/callback", root.GitHubCallback(database.DB))
		rootGroup.GET("/logout", root.Logout())

		authed := rootGroup.Group("", middleware.RequireSuperuser(database.DB))
		{
			authed.GET("", root.DashboardHandler(database.DB))
			authed.GET("/", root.DashboardHandler(database.DB))
		}
	}

	log.Printf("listening on :%s", config.C.Server.Port)
	if err := r.Run(":" + config.C.Server.Port); err != nil {
		log.Fatal(err)
	}
}
