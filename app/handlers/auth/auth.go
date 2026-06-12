package auth

import (
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SessionUserIDKey = "user_id"

// LoginPage renders the /root/login page.
func LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "root_login.tmpl", gin.H{
			"Title":         "Sign in",
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"Error":         c.Query("error"),
		})
	}
}

// Logout clears the root session and redirects to /root/login.
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/root/login")
	}
}

// SignInPage renders the public /signin page.
func SignInPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uid, ok := session.Get(SessionUserIDKey).(uint); ok && uid > 0 {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "signin.tmpl", gin.H{
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"GoogleEnabled": config.C.OAuthGoogleEnabled(),
			"Error":         c.Query("error"),
		})
	}
}

// SignOut clears the public session and redirects to /.
func SignOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/")
	}
}

// loginPublicUser sets the session for a public user, gating through /2fa if TOTP is enabled.
func loginPublicUser(c *gin.Context, user *models.User) {
	session := sessions.Default(c)
	if user.TOTPEnabled {
		session.Delete(SessionUserIDKey)
		session.Set(totp.PendingUserIDKey, user.ID)
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/signin?error=session")
			return
		}
		c.Redirect(http.StatusFound, "/2fa")
		return
	}
	session.Set(SessionUserIDKey, user.ID)
	if err := session.Save(); err != nil {
		c.Redirect(http.StatusFound, "/signin?error=session")
		return
	}
	c.Redirect(http.StatusFound, "/")
}
