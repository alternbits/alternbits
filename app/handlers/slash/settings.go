package slash

import (
	"encoding/json"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func currentUser(c *gin.Context) *models.User {
	if u, ok := c.MustGet("user").(*models.User); ok {
		return u
	}
	return nil
}

func headerData(db *gorm.DB, user *models.User) gin.H {
	var categories []models.Category
	db.Where("parent_id IS NULL").Find(&categories)
	var toolCount int64
	db.Model(&models.Tool{}).Count(&toolCount)
	return gin.H{
		"CurrentUser": user,
		"Categories":  categories,
		"ToolCount":   toolCount,
	}
}

func SettingsPage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		data := headerData(db, user)
		data["User"] = user
		data["Flash"] = c.Query("2fa")
		data["Connected"] = c.Query("connected")
		data["FlashError"] = c.Query("error")
		data["GitHubEnabled"] = config.C.OAuthGitHubEnabled()
		data["GoogleEnabled"] = config.C.OAuthGoogleEnabled()

		var recoveryCount int
		if user.RecoveryCodes != "" {
			var codes []string
			if json.Unmarshal([]byte(user.RecoveryCodes), &codes) == nil {
				recoveryCount = len(codes)
			}
		}
		data["RecoveryCount"] = recoveryCount

		c.HTML(http.StatusOK, "settings.tmpl", data)
	}
}

func Settings2FASetupGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if user.TOTPEnabled {
			c.Redirect(http.StatusFound, "/settings")
			return
		}
		if user.TOTPSecret == "" {
			if err := totp.NewSecret(db, user); err != nil {
				c.HTML(http.StatusInternalServerError, "settings_2fa_setup.tmpl", gin.H{
					"Error": "Failed to generate secret.",
				})
				return
			}
		}
		qr, err := totp.QRBase64(user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "settings_2fa_setup.tmpl", gin.H{
				"Error": "Failed to render QR code.",
			})
			return
		}
		data := headerData(db, user)
		data["Secret"] = user.TOTPSecret
		data["QR"] = qr
		c.HTML(http.StatusOK, "settings_2fa_setup.tmpl", data)
	}
}

func Settings2FASetupPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if user.TOTPEnabled {
			c.Redirect(http.StatusFound, "/settings")
			return
		}
		code := c.PostForm("code")
		if !totp.ValidateCode(code, user.TOTPSecret) {
			qr, _ := totp.QRBase64(user)
			data := headerData(db, user)
			data["Secret"] = user.TOTPSecret
			data["QR"] = qr
			data["Error"] = "Invalid code — try again."
			c.HTML(http.StatusOK, "settings_2fa_setup.tmpl", data)
			return
		}

		plain, hashed, err := totp.GenerateCodes(8)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "settings_2fa_setup.tmpl", gin.H{
				"Error": "Failed to generate recovery codes.",
			})
			return
		}
		encoded, _ := json.Marshal(hashed)
		user.TOTPEnabled = true
		user.RecoveryCodes = string(encoded)
		if err := db.Save(user).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "settings_2fa_setup.tmpl", gin.H{
				"Error": "Failed to enable 2FA.",
			})
			return
		}

		data := headerData(db, user)
		data["Codes"] = plain
		c.HTML(http.StatusOK, "settings_2fa_recovery.tmpl", data)
	}
}

func Settings2FASetupDone() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/settings?2fa=enabled")
	}
}

func Settings2FADisable(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := currentUser(c)
		if !user.TOTPEnabled {
			c.Redirect(http.StatusFound, "/settings")
			return
		}
		code := c.PostForm("code")
		valid := totp.ValidateCode(code, user.TOTPSecret)
		if !valid {
			_, valid = totp.ConsumeCode(user, code)
		}
		if !valid {
			c.Redirect(http.StatusFound, "/settings?error=invalid_code")
			return
		}
		user.TOTPEnabled = false
		user.TOTPSecret = ""
		user.RecoveryCodes = ""
		if err := db.Save(user).Error; err != nil {
			c.Redirect(http.StatusFound, "/settings?error=save")
			return
		}
		c.Redirect(http.StatusFound, "/settings?2fa=disabled")
	}
}
