package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database struct {
		DSN string
	}
	Server struct {
		Port string
	}
	Session struct {
		Secret string
	}
	GitHubOAuth struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Superuser struct {
		GitHubLogins []string
	}
}

var C *Config

func Load() error {
	_ = godotenv.Load()

	C = &Config{}

	C.Database.DSN = os.Getenv("DB_DSN")

	C.Server.Port = os.Getenv("PORT")
	if C.Server.Port == "" {
		C.Server.Port = "1337"
	}

	C.Session.Secret = os.Getenv("SESSION_SECRET")
	if C.Session.Secret == "" {
		C.Session.Secret = "altern-dev-insecure-session-secret"
		log.Println("config: SESSION_SECRET not set, using insecure dev default")
	}

	C.GitHubOAuth.ClientID = os.Getenv("GITHUB_CLIENT_ID")
	C.GitHubOAuth.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	C.GitHubOAuth.RedirectURL = os.Getenv("GITHUB_REDIRECT_URL")

	for login := range strings.SplitSeq(os.Getenv("SUPERUSER_GITHUB_LOGINS"), ",") {
		login = strings.TrimSpace(login)
		if login != "" {
			C.Superuser.GitHubLogins = append(C.Superuser.GitHubLogins, login)
		}
	}

	return nil
}

func (c *Config) OAuthGitHubEnabled() bool {
	return c.GitHubOAuth.ClientID != "" && c.GitHubOAuth.ClientSecret != "" && c.GitHubOAuth.RedirectURL != ""
}

func (c *Config) IsSuperuserLogin(login string) bool {
	for _, allowed := range c.Superuser.GitHubLogins {
		if strings.EqualFold(allowed, login) {
			return true
		}
	}
	return false
}
