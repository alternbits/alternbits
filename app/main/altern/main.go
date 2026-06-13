package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/handlers/auth"
	"github.com/dariubs/altern/app/handlers/root"
	"github.com/dariubs/altern/app/handlers/slash"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/middleware"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
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
		"markdownHTML": func(s string) template.HTML {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(s), &buf); err != nil {
				return template.HTML(template.HTMLEscapeString(s))
			}
			return template.HTML(buf.String())
		},
	})
	r.LoadHTMLGlob("views/*/*.tmpl")
	r.Static("/static", "./static")

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
	r.GET("/ai/:slug", slash.AIHandler(database.DB))
	r.GET("/lists", slash.ListsHandler(database.DB))
	r.GET("/lists/:slug", slash.ListHandler(database.DB))
	r.GET("/categories", slash.CategoriesHandler(database.DB))
	r.GET("/categories/:slug", slash.CategoryHandler(database.DB))
	r.GET("/page/:slug", slash.PageHandler(database.DB))

	userRoutes := r.Group("", middleware.RequireAuth(database.DB))
	{
		userRoutes.GET("/settings", slash.SettingsPage(database.DB))
		userRoutes.GET("/settings/2fa/setup", slash.Settings2FASetupGet(database.DB))
		userRoutes.POST("/settings/2fa/setup", slash.Settings2FASetupPost(database.DB))
		userRoutes.POST("/settings/2fa/done", slash.Settings2FASetupDone())
		userRoutes.POST("/settings/2fa/disable", slash.Settings2FADisable(database.DB))
		if config.C.OAuthGitHubEnabled() {
			userRoutes.GET("/settings/connect/github", auth.ConnectGitHub())
		}
		if config.C.OAuthGoogleEnabled() {
			userRoutes.GET("/settings/connect/google", auth.ConnectGoogle())
		}
	}
	r.GET("/2fa", totp.Dispatch(database.DB, publicTOTPOpts))
	r.POST("/2fa/setup", totp.Setup(database.DB, publicTOTPOpts))
	r.POST("/2fa/setup/done", totp.SetupDone(database.DB, publicTOTPOpts))
	r.POST("/2fa/verify", totp.Verify(database.DB, publicTOTPOpts))

	r.GET("/signin", auth.SignInPage())
	r.GET("/auth/github", auth.GitHubUserLogin())
	r.GET("/auth/github/callback", auth.GitHubCallback(database.DB))
	r.GET("/auth/google", auth.GoogleLogin())
	r.GET("/auth/google/callback", auth.GoogleCallback(database.DB))
	r.POST("/signout", auth.SignOut())

	rootGroup := r.Group("/root")
	{
		rootGroup.GET("/login", auth.LoginPage())
		rootGroup.GET("/auth/github", auth.GitHubLogin())
		rootGroup.POST("/logout", auth.Logout())

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
			authed.GET("/users/search", root.UsersSearchAPI(database.DB))
			authed.GET("/users/:id/edit", root.UserEditForm(database.DB))
			authed.POST("/users/:id", root.UserUpdate(database.DB))
			authed.GET("/ais", root.AIsListHandler(database.DB))
			authed.GET("/categories", root.CategoriesListHandler(database.DB))
			authed.GET("/lists", root.ListsListHandler(database.DB))
			authed.GET("/lists/new", root.ListNewForm(database.DB))
			authed.POST("/lists", root.ListCreate(database.DB))
			authed.GET("/lists/:id/edit", root.ListEditForm(database.DB))
			authed.POST("/lists/:id", root.ListUpdate(database.DB))
			authed.GET("/categories/new", root.CategoryNewForm(database.DB))
			authed.POST("/categories", root.CategoryCreate(database.DB))
			authed.GET("/categories/:id/edit", root.CategoryEditForm(database.DB))
			authed.POST("/categories/:id", root.CategoryUpdate(database.DB))
			authed.GET("/artifacts", root.ArtifactsListHandler(database.DB))
			authed.GET("/artifacts/new", root.ArtifactNewForm())
			authed.POST("/artifacts", root.ArtifactCreate(database.DB))
			authed.GET("/artifacts/:id/edit", root.ArtifactEditForm(database.DB))
			authed.POST("/artifacts/:id", root.ArtifactUpdate(database.DB))
			authed.GET("/genera", root.GeneraListHandler(database.DB))
			authed.GET("/genera/new", root.GenusNewForm())
			authed.POST("/genera", root.GenusCreate(database.DB))
			authed.GET("/pages", root.PagesListHandler(database.DB))
			authed.GET("/pages/new", root.PageNewForm())
			authed.POST("/pages", root.PageCreate(database.DB))
			authed.GET("/pages/:id/edit", root.PageEditForm(database.DB))
			authed.POST("/pages/:id", root.PageUpdate(database.DB))
			authed.GET("/ais/new", root.AINewForm(database.DB))
			authed.POST("/ais", root.AICreate(database.DB, r2svc))
			authed.GET("/ais/:id/edit", root.AIEditForm(database.DB))
			authed.POST("/ais/:id", root.AIUpdate(database.DB, r2svc))
		}
	}

	log.Printf("listening on :%s", config.C.Server.Port)
	if err := r.Run(":" + config.C.Server.Port); err != nil {
		log.Fatal(err)
	}
}

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
