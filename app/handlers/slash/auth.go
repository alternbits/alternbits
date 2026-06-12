package slash

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/handlers/totp"
	"github.com/dariubs/altern/app/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	googleoauth "golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func SignInPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "signin.tmpl", gin.H{
			"GitHubEnabled":  config.C.OAuthGitHubEnabled(),
			"GoogleEnabled":  config.C.OAuthGoogleEnabled(),
			"Error":          c.Query("error"),
		})
	}
}

func GitHubUserLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGitHubEnabled() {
			c.Redirect(http.StatusFound, "/signin?error=github_disabled")
			return
		}
		state := uuid.New().String()
		session := sessions.Default(c)
		session.Set("oauth_state", state)
		session.Set("oauth_flow", "public")
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/signin?error=session")
			return
		}
		cfg := &oauth2.Config{
			ClientID:     config.C.GitHubOAuth.ClientID,
			ClientSecret: config.C.GitHubOAuth.ClientSecret,
			RedirectURL:  config.C.GitHubOAuth.RedirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     githuboauth.Endpoint,
		}
		c.Redirect(http.StatusTemporaryRedirect, cfg.AuthCodeURL(state))
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
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/signin?error=session")
			return
		}
		cfg := &oauth2.Config{
			ClientID:     config.C.GoogleOAuth.ClientID,
			ClientSecret: config.C.GoogleOAuth.ClientSecret,
			RedirectURL:  config.C.GoogleOAuth.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		}
		c.Redirect(http.StatusTemporaryRedirect, cfg.AuthCodeURL(state))
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

		cfg := &oauth2.Config{
			ClientID:     config.C.GoogleOAuth.ClientID,
			ClientSecret: config.C.GoogleOAuth.ClientSecret,
			RedirectURL:  config.C.GoogleOAuth.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		}
		token, err := cfg.Exchange(context.Background(), code)
		if err != nil {
			c.Redirect(http.StatusFound, "/signin?error=exchange")
			return
		}

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		token.SetAuthHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.Redirect(http.StatusFound, "/signin?error=userinfo")
			return
		}
		defer resp.Body.Close()

		var gu googleUser
		if json.NewDecoder(resp.Body).Decode(&gu) != nil || gu.Email == "" {
			c.Redirect(http.StatusFound, "/signin?error=userinfo")
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
		session.Delete(sessionUserIDKey)
		session.Set(totp.PendingUserIDKey, user.ID)
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/signin?error=session")
			return
		}
		c.Redirect(http.StatusFound, "/2fa")
		return
	}
	session.Set(sessionUserIDKey, user.ID)
	if err := session.Save(); err != nil {
		c.Redirect(http.StatusFound, "/signin?error=session")
		return
	}
	c.Redirect(http.StatusFound, "/")
}

const sessionUserIDKey = "user_id"
