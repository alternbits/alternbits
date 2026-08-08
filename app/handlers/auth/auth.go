package auth

import (
	"net/http"
	"net/url"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SessionUserIDKey = "user_id"

// authNextKey stashes a validated post-login redirect target in the session
// so it survives the round trip through an OAuth provider (and, if enabled,
// the TOTP gate) until FinalizeSession runs.
const authNextKey = "auth_next"

// setAuthNext validates and stores (or clears) the post-login redirect
// target. Call at the start of every public login-initiating handler so
// stale values from an abandoned previous attempt don't leak through.
func setAuthNext(session sessions.Session, next string) {
	if safe := utils.SafeNext(next); safe != "" {
		session.Set(authNextKey, safe)
	} else {
		session.Delete(authNextKey)
	}
}

// RedirectNext redirects to the stored post-login target if one was set,
// otherwise to fallback. Clears the stored value either way.
func RedirectNext(c *gin.Context, fallback string) {
	session := sessions.Default(c)
	next, _ := session.Get(authNextKey).(string)
	session.Delete(authNextKey)
	_ = session.Save()
	if next == "" {
		next = fallback
	}
	c.Redirect(http.StatusFound, next)
}

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
		next := utils.SafeNext(c.Query("next"))
		session := sessions.Default(c)
		if uid, ok := session.Get(SessionUserIDKey).(uint); ok && uid > 0 {
			dest := "/"
			if next != "" {
				dest = next
			}
			c.Redirect(http.StatusFound, dest)
			return
		}
		c.HTML(http.StatusOK, "signin.tmpl", gin.H{
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"GoogleEnabled": config.C.OAuthGoogleEnabled(),
			"Error":         c.Query("error"),
			"Next":          next,
			"NextParam":     nextParam(next),
		})
	}
}

// SignUpPage renders the public /signup page.
func SignUpPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		next := utils.SafeNext(c.Query("next"))
		session := sessions.Default(c)
		if uid, ok := session.Get(SessionUserIDKey).(uint); ok && uid > 0 {
			dest := "/"
			if next != "" {
				dest = next
			}
			c.Redirect(http.StatusFound, dest)
			return
		}
		c.HTML(http.StatusOK, "signup.tmpl", gin.H{
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"GoogleEnabled": config.C.OAuthGoogleEnabled(),
			"Error":         c.Query("error"),
			"Next":          next,
			"NextParam":     nextParam(next),
		})
	}
}

// nextParam renders a "?next=..." query-string fragment (URL-escaped) for
// use in template hrefs, or "" when there's nothing to carry forward.
func nextParam(next string) string {
	if next == "" {
		return ""
	}
	return "?next=" + url.QueryEscape(next)
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
	RedirectNext(c, "/")
}
