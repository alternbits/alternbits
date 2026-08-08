package middleware

import (
	"net/http"
	"net/url"

	"github.com/dariubs/altern/app/models"
	"github.com/dariubs/altern/app/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionUserIDKey = "user_id"

// signInWithNext builds a /signin redirect that carries the originally
// requested URL along as ?next=, so the user lands back where they started
// after completing login.
func signInWithNext(c *gin.Context) string {
	next := utils.SafeNext(c.Request.URL.RequestURI())
	if next == "" {
		return "/signin"
	}
	return "/signin?next=" + url.QueryEscape(next)
}

func RequireAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		raw := session.Get(sessionUserIDKey)
		if raw == nil {
			c.Redirect(http.StatusFound, signInWithNext(c))
			c.Abort()
			return
		}
		userID, ok := raw.(uint)
		if !ok {
			next := signInWithNext(c)
			session.Clear()
			_ = session.Save()
			c.Redirect(http.StatusFound, next)
			c.Abort()
			return
		}
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			next := signInWithNext(c)
			session.Clear()
			_ = session.Save()
			c.Redirect(http.StatusFound, next)
			c.Abort()
			return
		}
		c.Set("user", &user)
		c.Next()
	}
}

func RequireSuperuser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		raw := session.Get(sessionUserIDKey)
		if raw == nil {
			c.Redirect(http.StatusFound, "/root/login")
			c.Abort()
			return
		}
		userID, ok := raw.(uint)
		if !ok {
			session.Clear()
			_ = session.Save()
			c.Redirect(http.StatusFound, "/root/login")
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil || !user.IsAdmin {
			session.Clear()
			_ = session.Save()
			c.Redirect(http.StatusFound, "/root/login?error=forbidden")
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}
