package middleware

import (
	"net/http"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionUserIDKey = "user_id"

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
