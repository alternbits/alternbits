package slash

import (
	"net/http"

	"github.com/dariubs/altern/app/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

func SignInPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uid, ok := session.Get(sessionUserIDKey).(uint); ok && uid > 0 {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "signin.tmpl", gin.H{
			"GitHubEnabled": config.C.OAuthGitHubEnabled(),
			"Error":         c.Query("error"),
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

func SignOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/")
	}
}

const sessionUserIDKey = "user_id"
