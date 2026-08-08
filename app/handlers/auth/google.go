package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.C.GoogleOAuth.ClientID,
		ClientSecret: config.C.GoogleOAuth.ClientSecret,
		RedirectURL:  config.C.GoogleOAuth.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     googleoauth.Endpoint,
	}
}

func GoogleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGoogleEnabled() {
			c.Redirect(http.StatusFound, "/signin?error=google_disabled")
			return
		}
		state := uuid.New().String()
		session := sessions.Default(c)
		session.Set("google_oauth_state", state)
		setAuthNext(session, c.Query("next"))
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/signin?error=session")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, googleOAuthConfig().AuthCodeURL(state))
	}
}

// ConnectGoogle initiates a Google connect flow for an already-authenticated user.
func ConnectGoogle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGoogleEnabled() {
			c.Redirect(http.StatusFound, "/settings?error=google_disabled")
			return
		}
		user := c.MustGet("user").(*models.User)
		state := uuid.New().String()
		session := sessions.Default(c)
		session.Set("google_oauth_state", state)
		session.Set(connectUserIDKey, user.ID)
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/settings?error=session")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, googleOAuthConfig().AuthCodeURL(state))
	}
}

type googleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func GoogleCallback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGoogleEnabled() {
			c.Redirect(http.StatusFound, "/signin?error=google_disabled")
			return
		}

		code := c.Query("code")
		state := c.Query("state")
		if code == "" {
			c.Redirect(http.StatusFound, "/signin?error=code")
			return
		}

		session := sessions.Default(c)
		savedState, _ := session.Get("google_oauth_state").(string)
		if savedState == "" || savedState != state {
			c.Redirect(http.StatusFound, "/signin?error=state")
			return
		}
		session.Delete("google_oauth_state")
		_ = session.Save()

		// Check if this is a connect flow (connectUserIDKey set by ConnectGoogle).
		rawUID := session.Get(connectUserIDKey)
		isConnect := rawUID != nil
		session.Delete(connectUserIDKey)
		_ = session.Save()

		token, err := googleOAuthConfig().Exchange(context.Background(), code)
		if err != nil {
			if isConnect {
				c.Redirect(http.StatusFound, "/settings?error=exchange")
			} else {
				c.Redirect(http.StatusFound, "/signin?error=exchange")
			}
			return
		}

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		token.SetAuthHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if isConnect {
				c.Redirect(http.StatusFound, "/settings?error=userinfo")
			} else {
				c.Redirect(http.StatusFound, "/signin?error=userinfo")
			}
			return
		}
		defer resp.Body.Close()

		var gu googleUser
		if json.NewDecoder(resp.Body).Decode(&gu) != nil || gu.Email == "" {
			if isConnect {
				c.Redirect(http.StatusFound, "/settings?error=userinfo")
			} else {
				c.Redirect(http.StatusFound, "/signin?error=userinfo")
			}
			return
		}

		// Connect flow: link Google to an already-authenticated user.
		if isConnect {
			uid, ok := rawUID.(uint)
			if !ok {
				c.Redirect(http.StatusFound, "/settings?error=connect")
				return
			}
			var user models.User
			if db.First(&user, uid).Error != nil {
				c.Redirect(http.StatusFound, "/settings?error=connect")
				return
			}
			// Reject if this Google account is already used by a different user.
			var existing models.User
			if db.Where("google_id = ?", gu.ID).First(&existing).Error == nil && existing.ID != uid {
				c.Redirect(http.StatusFound, "/settings?error=google_taken")
				return
			}
			gid := gu.ID
			user.GoogleID = &gid
			if user.AvatarURL == "" {
				user.AvatarURL = gu.Picture
			}
			if user.Name == "" {
				user.Name = gu.Name
			}
			db.Save(&user)
			c.Redirect(http.StatusFound, "/settings?connected=google")
			return
		}

		var user models.User
		err = db.Where("google_id = ?", gu.ID).First(&user).Error
		if err != nil {
			if db.Where(&models.User{Email: gu.Email}).First(&user).Error != nil {
				gid := gu.ID
				user = models.User{
					Username:  fmt.Sprintf("g_%s", gu.ID),
					Email:     gu.Email,
					Name:      gu.Name,
					AvatarURL: gu.Picture,
					GoogleID:  &gid,
				}
				if err := db.Create(&user).Error; err != nil {
					c.Redirect(http.StatusFound, "/signin?error=create")
					return
				}
			} else {
				gid := gu.ID
				user.GoogleID = &gid
				if user.AvatarURL == "" {
					user.AvatarURL = gu.Picture
				}
				if user.Name == "" {
					user.Name = gu.Name
				}
				db.Save(&user)
			}
		} else {
			if user.AvatarURL == "" {
				user.AvatarURL = gu.Picture
			}
			if user.Name == "" {
				user.Name = gu.Name
			}
			db.Save(&user)
		}

		loginPublicUser(c, &user)
	}
}
