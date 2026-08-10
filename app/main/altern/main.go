package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/handlers/auth"
	"github.com/dariubs/altern/app/handlers/maker"
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
		"faviconURL": func(website string) string {
			if website == "" {
				return ""
			}
			u := website
			if !strings.Contains(u, "://") {
				u = "https://" + u
			}
			return "https://www.google.com/s2/favicons?domain=" + url.QueryEscape(u) + "&sz=64"
		},
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				if m := int(d.Minutes()); m == 1 {
					return "1 minute ago"
				} else {
					return fmt.Sprintf("%d minutes ago", m)
				}
			case d < 24*time.Hour:
				if h := int(d.Hours()); h == 1 {
					return "1 hour ago"
				} else {
					return fmt.Sprintf("%d hours ago", h)
				}
			case d < 30*24*time.Hour:
				if days := int(d.Hours() / 24); days == 1 {
					return "1 day ago"
				} else {
					return fmt.Sprintf("%d days ago", days)
				}
			case d < 365*24*time.Hour:
				if months := int(d.Hours() / 24 / 30); months == 1 {
					return "1 month ago"
				} else {
					return fmt.Sprintf("%d months ago", months)
				}
			default:
				if years := int(d.Hours() / 24 / 365); years == 1 {
					return "1 year ago"
				} else {
					return fmt.Sprintf("%d years ago", years)
				}
			}
		},
		"dict": func(pairs ...any) (map[string]any, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("dict expects an even number of arguments")
			}
			m := make(map[string]any, len(pairs)/2)
			for i := 0; i < len(pairs); i += 2 {
				key, ok := pairs[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				m[key] = pairs[i+1]
			}
			return m, nil
		},
	})
	r.LoadHTMLGlob("views/*/*.tmpl")
	r.Static("/static", "./static")
	r.Static("/assets", "./assets")

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

	var r2svc *utils.R2Service
	if config.C.R2Enabled() {
		svc, err := utils.NewR2Service()
		if err != nil {
			log.Printf("r2: %v (uploads disabled)", err)
		} else {
			r2svc = svc
		}
	}

	var tgsvc *utils.TelegramService
	if config.C.TelegramEnabled() {
		tgsvc = utils.NewTelegramService()
	}

	r.GET("/", slash.Handler(database.DB))
	r.GET("/maker", maker.Handler(database.DB))
	r.GET("/ai/:slug", slash.AIHandler(database.DB))
	r.GET("/lists", slash.ListsHandler(database.DB))
	r.GET("/lists/:slug", slash.ListHandler(database.DB))
	r.GET("/categories", slash.CategoriesHandler(database.DB))
	r.GET("/categories/:slug", slash.CategoryHandler(database.DB))
	r.GET("/page/:slug", slash.PageHandler(database.DB))
	r.GET("/alternatives/:slug", slash.AlternativesHandler(database.DB))
	r.GET("/search", slash.SearchPage(database.DB))
	r.GET("/search/suggest", slash.SearchAPI(database.DB))

	userRoutes := r.Group("", middleware.RequireAuth(database.DB))
	{
		userRoutes.GET("/maker/submit", maker.SubmitForm(database.DB))
		userRoutes.POST("/maker/submit", maker.SubmitCreate(database.DB, r2svc, tgsvc))
		userRoutes.GET("/settings", slash.SettingsPage(database.DB))
		userRoutes.GET("/settings/2fa/setup", slash.Settings2FASetupGet(database.DB))
		userRoutes.POST("/settings/2fa/setup", slash.Settings2FASetupPost(database.DB, tgsvc))
		userRoutes.POST("/settings/2fa/done", slash.Settings2FASetupDone())
		userRoutes.POST("/settings/2fa/disable", slash.Settings2FADisable(database.DB, tgsvc))
		userRoutes.GET("/saved", slash.SavedPage(database.DB))
		userRoutes.POST("/saved/toggle", slash.SavedToggle(database.DB))
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
	r.GET("/signup", auth.SignUpPage())
	r.GET("/auth/github", auth.GitHubUserLogin())
	r.GET("/auth/github/callback", auth.GitHubCallback(database.DB, tgsvc))
	r.GET("/auth/google", auth.GoogleLogin())
	r.GET("/auth/google/callback", auth.GoogleCallback(database.DB, tgsvc))
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

		authed := rootGroup.Group("", middleware.RequireSuperuser(database.DB))
		{
			authed.GET("", root.DashboardHandler(database.DB))
			authed.GET("/", root.DashboardHandler(database.DB))
			authed.GET("/users", root.UsersListHandler(database.DB))
			authed.GET("/users/search", root.UsersSearchAPI(database.DB))
			authed.GET("/users/:id/edit", root.UserEditForm(database.DB))
			authed.POST("/users/:id", root.UserUpdate(database.DB, tgsvc))
			authed.GET("/ais", root.AIsListHandler(database.DB))
			authed.GET("/ais/review", root.AIReviewHandler(database.DB))
			authed.POST("/ais/:id/approve", root.AIApprove(database.DB, tgsvc))
			authed.POST("/ais/:id/reject", root.AIReject(database.DB, tgsvc))
			authed.GET("/categories", root.CategoriesListHandler(database.DB))
			authed.GET("/lists", root.ListsListHandler(database.DB))
			authed.GET("/lists/new", root.ListNewForm(database.DB))
			authed.POST("/lists", root.ListCreate(database.DB, tgsvc))
			authed.GET("/lists/:id/edit", root.ListEditForm(database.DB))
			authed.POST("/lists/:id", root.ListUpdate(database.DB, tgsvc))
			authed.GET("/categories/new", root.CategoryNewForm(database.DB))
			authed.POST("/categories", root.CategoryCreate(database.DB, tgsvc))
			authed.GET("/categories/:id/edit", root.CategoryEditForm(database.DB))
			authed.POST("/categories/:id", root.CategoryUpdate(database.DB, tgsvc))
			authed.GET("/artifacts", root.ArtifactsListHandler(database.DB))
			authed.GET("/artifacts/new", root.ArtifactNewForm())
			authed.POST("/artifacts", root.ArtifactCreate(database.DB, tgsvc))
			authed.GET("/artifacts/:id/edit", root.ArtifactEditForm(database.DB))
			authed.POST("/artifacts/:id", root.ArtifactUpdate(database.DB, tgsvc))
			authed.GET("/genera", root.GeneraListHandler(database.DB))
			authed.GET("/genera/new", root.GenusNewForm())
			authed.POST("/genera", root.GenusCreate(database.DB, tgsvc))
			authed.GET("/pages", root.PagesListHandler(database.DB))
			authed.GET("/pages/new", root.PageNewForm())
			authed.POST("/pages", root.PageCreate(database.DB, tgsvc))
			authed.GET("/pages/:id/edit", root.PageEditForm(database.DB))
			authed.POST("/pages/:id", root.PageUpdate(database.DB, tgsvc))
			authed.POST("/upload/screenshot", root.UploadScreenshot(r2svc))
			authed.GET("/ais/new", root.AINewForm(database.DB))
			authed.POST("/ais", root.AICreate(database.DB, r2svc, tgsvc))
			authed.GET("/ais/:id/edit", root.AIEditForm(database.DB))
			authed.POST("/ais/:id", root.AIUpdate(database.DB, r2svc, tgsvc))
			authed.GET("/ais/search", root.AIsSearchAPI(database.DB))
			authed.GET("/ais/:id/alternatives", root.AIAlternativesAPI(database.DB))
			authed.GET("/alternatives", root.AlternativesListHandler(database.DB))
			authed.GET("/alternatives/ai/:id", root.AIAlternativesEditHandler(database.DB))
			authed.GET("/alternatives/new", root.AlternativeNewForm(database.DB))
			authed.POST("/alternatives", root.AlternativeCreate(database.DB, tgsvc))
			authed.POST("/alternatives/:id/approve", root.AlternativeApprove(database.DB, tgsvc))
			authed.POST("/alternatives/:id/reject", root.AlternativeReject(database.DB, tgsvc))
			authed.POST("/alternatives/:id/delete", root.AlternativeDelete(database.DB, tgsvc))
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
	auth.RedirectNext(c, "/")
}
