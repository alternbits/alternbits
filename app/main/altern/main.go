package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/handlers/root"
	"github.com/dariubs/altern/app/handlers/slash"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/middleware"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
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
	r.SetFuncMap(template.FuncMap{
		"appName":   func() string { return config.C.App.Name },
		"appDomain": func() string { return config.C.App.Domain },
	})
	r.LoadHTMLGlob("views/*/*.tmpl")

	store := cookie.NewStore([]byte(config.C.Session.Secret))
	store.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 60 * 60 * 24 * 7})
	r.Use(sessions.Sessions("altern_session", store))

	rootTOTPOpts := totp.Options{
		BaseURL:         "/root/2fa",
		LoginRedirect:   "/root/login",
		FinalizeSession: finalizeRootSession,
	}

	publicTOTPOpts := totp.Options{
		BaseURL:         "/2fa",
		LoginRedirect:   "/signin",
		FinalizeSession: finalizePublicSession,
	}

	r.GET("/", slash.Handler(database.DB))
	r.GET("/2fa", totp.Dispatch(database.DB, publicTOTPOpts))
	r.POST("/2fa/setup", totp.Setup(database.DB, publicTOTPOpts))
	r.POST("/2fa/setup/done", totp.SetupDone(database.DB, publicTOTPOpts))
	r.POST("/2fa/verify", totp.Verify(database.DB, publicTOTPOpts))

	r.GET("/signin", slash.SignInPage())
	r.GET("/auth/github", slash.GitHubUserLogin())
	r.GET("/auth/github/callback", root.GitHubCallback(database.DB))
	r.POST("/signout", slash.SignOut())

	rootGroup := r.Group("/root")
	{
		rootGroup.GET("/login", root.LoginPage())
		rootGroup.GET("/auth/github", root.GitHubLogin())
		rootGroup.POST("/logout", root.Logout())

		rootGroup.GET("/2fa", totp.Dispatch(database.DB, rootTOTPOpts))
		rootGroup.POST("/2fa/setup", totp.Setup(database.DB, rootTOTPOpts))
		rootGroup.POST("/2fa/setup/done", totp.SetupDone(database.DB, rootTOTPOpts))
		rootGroup.POST("/2fa/verify", totp.Verify(database.DB, rootTOTPOpts))

		var r2svc *utils.R2Service
		if config.C.R2Enabled() {
			svc, err := utils.NewR2Service()
			if err != nil {
				log.Printf("r2: %v (uploads disabled)", err)
			} else {
				r2svc = svc
			}
		}

		authed := rootGroup.Group("", middleware.RequireSuperuser(database.DB))
		{
			authed.GET("", root.DashboardHandler(database.DB))
			authed.GET("/", root.DashboardHandler(database.DB))
			authed.GET("/users", root.UsersListHandler(database.DB))
			authed.GET("/tools", root.ToolsListHandler(database.DB))
			authed.GET("/categories", root.CategoriesListHandler(database.DB))
			authed.GET("/lists", root.ListsListHandler(database.DB))
			authed.GET("/lists/new", root.ListNewForm(database.DB))
			authed.POST("/lists", root.ListCreate(database.DB))
			authed.GET("/categories/new", root.CategoryNewForm(database.DB))
			authed.POST("/categories", root.CategoryCreate(database.DB))
			authed.GET("/tools/new", root.ToolNewForm())
			authed.POST("/tools", root.ToolCreate(database.DB, r2svc))
		}
	}

	log.Printf("listening on :%s", config.C.Server.Port)
	if err := r.Run(":" + config.C.Server.Port); err != nil {
		log.Fatal(err)
	}
}

// finalizeRootSession promotes the TOTP-pending user into the final
// /root session and sends them to the dashboard.
func finalizeRootSession(c *gin.Context, user *models.User) {
	session := sessions.Default(c)
	session.Delete(totp.PendingUserIDKey)
	session.Set("user_id", user.ID)
	if err := session.Save(); err != nil {
		c.Redirect(http.StatusFound, "/root/login?error=session")
		return
	}
	c.Redirect(http.StatusFound, "/root")
}

// finalizePublicSession promotes the TOTP-pending user into the public
// session and sends them to the home page.
func finalizePublicSession(c *gin.Context, user *models.User) {
	session := sessions.Default(c)
	session.Delete(totp.PendingUserIDKey)
	session.Set("user_id", user.ID)
	if err := session.Save(); err != nil {
		c.Redirect(http.StatusFound, "/signin?error=session")
		return
	}
	c.Redirect(http.StatusFound, "/")
}
