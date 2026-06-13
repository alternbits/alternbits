package root

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserEditForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/users")
			return
		}
		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/users")
			return
		}
		c.HTML(http.StatusOK, "root_user_edit.tmpl", gin.H{
			"ActiveNav": "users",
			"User":      user,
		})
	}
}

func UserUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/users")
			return
		}
		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.Redirect(http.StatusFound, "/root/users")
			return
		}

		name := strings.TrimSpace(c.PostForm("name"))
		username := strings.TrimSpace(c.PostForm("username"))
		email := strings.TrimSpace(c.PostForm("email"))
		bio := strings.TrimSpace(c.PostForm("bio"))
		avatarURL := strings.TrimSpace(c.PostForm("avatar_url"))
		isAdmin := c.PostForm("is_admin") == "on"
		disableTOTP := c.PostForm("disable_totp") == "on"

		renderErr := func(msg string) {
			c.HTML(http.StatusBadRequest, "root_user_edit.tmpl", gin.H{
				"ActiveNav": "users",
				"User":      user,
				"Error":     msg,
			})
		}

		if email == "" {
			renderErr("Email is required.")
			return
		}

		var conflict models.User
		if db.Where("email = ? AND id != ?", email, user.ID).First(&conflict).Error == nil {
			renderErr("Email already in use by another account.")
			return
		}
		if username != "" {
			if db.Where("username = ? AND id != ?", username, user.ID).First(&conflict).Error == nil {
				renderErr("Username already in use by another account.")
				return
			}
		}

		user.Name = name
		user.Username = username
		user.Email = email
		user.Bio = bio
		user.AvatarURL = avatarURL
		user.IsAdmin = isAdmin
		if disableTOTP {
			user.TOTPEnabled = false
			user.TOTPSecret = ""
			user.RecoveryCodes = ""
		}

		if err := db.Save(&user).Error; err != nil {
			renderErr("Failed to save user: " + err.Error())
			return
		}

		c.Redirect(http.StatusFound, "/root/users")
	}
}
