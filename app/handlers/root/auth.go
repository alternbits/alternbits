package root

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
	"gorm.io/gorm"
)

func githubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.C.GitHubOAuth.ClientID,
		ClientSecret: config.C.GitHubOAuth.ClientSecret,
		RedirectURL:  config.C.GitHubOAuth.RedirectURL,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     githuboauth.Endpoint,
	}
}

func LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"Title":         "Sign in",
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"Error":         c.Query("error"),
		})
	}
}

func GitHubLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGitHubEnabled() {
			c.Redirect(http.StatusFound, "/root/login?error=github_disabled")
			return
		}
		state := uuid.New().String()
		session := sessions.Default(c)
		session.Set("oauth_state", state)
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/root/login?error=session")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, githubOAuthConfig().AuthCodeURL(state))
	}
}

type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func GitHubCallback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGitHubEnabled() {
			c.Redirect(http.StatusFound, "/root/login?error=github_disabled")
			return
		}

		code := c.Query("code")
		state := c.Query("state")
		if code == "" {
			c.Redirect(http.StatusFound, "/root/login?error=code")
			return
		}

		session := sessions.Default(c)
		savedState, _ := session.Get("oauth_state").(string)
		if savedState == "" || savedState != state {
			c.Redirect(http.StatusFound, "/root/login?error=state")
			return
		}
		session.Delete("oauth_state")
		_ = session.Save()

		cfg := githubOAuthConfig()
		token, err := cfg.Exchange(context.Background(), code)
		if err != nil {
			c.Redirect(http.StatusFound, "/root/login?error=exchange")
			return
		}

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/user", nil)
		token.SetAuthHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.Redirect(http.StatusFound, "/root/login?error=userinfo")
			return
		}
		defer resp.Body.Close()

		var gu githubUser
		if json.NewDecoder(resp.Body).Decode(&gu) != nil || gu.Login == "" {
			c.Redirect(http.StatusFound, "/root/login?error=userinfo")
			return
		}

		githubID := fmt.Sprintf("%d", gu.ID)
		email := gu.Email
		if email == "" {
			email = gu.Login + "@github.user"
		}
		isSuperuser := config.C.IsSuperuserLogin(gu.Login)

		var user models.User
		err = db.Where(&models.User{GitHubID: githubID}).First(&user).Error
		if err != nil {
			if err := db.Where(&models.User{Email: email}).First(&user).Error; err != nil {
				user = models.User{
					Username:    gu.Login,
					Email:       email,
					Name:        gu.Name,
					AvatarURL:   gu.AvatarURL,
					GitHubID:    githubID,
					GitHubLogin: gu.Login,
					IsAdmin:     isSuperuser,
				}
				if err := db.Create(&user).Error; err != nil {
					c.Redirect(http.StatusFound, "/root/login?error=create")
					return
				}
			} else {
				user.GitHubID = githubID
				user.GitHubLogin = gu.Login
				user.AvatarURL = gu.AvatarURL
				user.Name = gu.Name
				if isSuperuser {
					user.IsAdmin = true
				}
				db.Save(&user)
			}
		} else {
			user.GitHubLogin = gu.Login
			user.AvatarURL = gu.AvatarURL
			user.Name = gu.Name
			if isSuperuser {
				user.IsAdmin = true
			}
			db.Save(&user)
		}

		if !isSuperuser {
			c.Redirect(http.StatusFound, "/root/login?error=forbidden")
			return
		}

		session.Delete(sessionUserIDKey)
		session.Set(totp.PendingUserIDKey, user.ID)
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/root/login?error=session")
			return
		}
		c.Redirect(http.StatusFound, "/root/2fa")
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/root/login")
	}
}

const sessionUserIDKey = "user_id"
