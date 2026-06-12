package auth

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

// GitHubLogin initiates the root GitHub OAuth flow.
func GitHubLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGitHubEnabled() {
			c.Redirect(http.StatusFound, "/root/login?error=github_disabled")
			return
		}
		state := uuid.New().String()
		session := sessions.Default(c)
		session.Set("oauth_state", state)
		session.Set("oauth_flow", "root")
		if err := session.Save(); err != nil {
			c.Redirect(http.StatusFound, "/root/login?error=session")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, githubOAuthConfig().AuthCodeURL(state))
	}
}

// GitHubUserLogin initiates the public GitHub OAuth flow.
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

// GitHubCallback handles the shared OAuth callback for both root and public flows.
// The flow is determined by the "oauth_flow" session key set before the redirect.
func GitHubCallback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.C.OAuthGitHubEnabled() {
			c.Redirect(http.StatusFound, "/signin?error=github_disabled")
			return
		}

		session := sessions.Default(c)
		flow, _ := session.Get("oauth_flow").(string)

		errTo := func(code string) {
			if flow == "root" {
				c.Redirect(http.StatusFound, "/root/login?error="+code)
			} else {
				c.Redirect(http.StatusFound, "/signin?error="+code)
			}
		}

		code := c.Query("code")
		state := c.Query("state")
		if code == "" {
			errTo("code")
			return
		}

		savedState, _ := session.Get("oauth_state").(string)
		if savedState == "" || savedState != state {
			errTo("state")
			return
		}
		session.Delete("oauth_state")
		session.Delete("oauth_flow")
		_ = session.Save()

		token, err := githubOAuthConfig().Exchange(context.Background(), code)
		if err != nil {
			errTo("exchange")
			return
		}

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/user", nil)
		token.SetAuthHeader(req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			errTo("userinfo")
			return
		}
		defer resp.Body.Close()

		var gu githubUser
		if json.NewDecoder(resp.Body).Decode(&gu) != nil || gu.Login == "" {
			errTo("userinfo")
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
			if db.Where(&models.User{Email: email}).First(&user).Error != nil {
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
					errTo("create")
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

		if flow == "root" {
			if !isSuperuser {
				c.Redirect(http.StatusFound, "/root/login?error=forbidden")
				return
			}
			session.Delete(SessionUserIDKey)
			session.Set(totp.PendingUserIDKey, user.ID)
			if err := session.Save(); err != nil {
				c.Redirect(http.StatusFound, "/root/login?error=session")
				return
			}
			c.Redirect(http.StatusFound, "/root/2fa")
			return
		}

		loginPublicUser(c, &user)
	}
}
