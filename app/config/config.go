package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Name   string
		Domain string
	}
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
	GoogleOAuth struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Superuser struct {
		GitHubLogins []string
	}
	CloudflareR2 struct {
		AccountID       string
		AccessKeyID     string
		SecretAccessKey string
		Bucket          string
		Region          string
		PublicURL       string
	}
}

var C *Config

func Load() error {
	_ = godotenv.Load()

	C = &Config{}

	C.App.Name = os.Getenv("APP_NAME")
	if C.App.Name == "" {
		C.App.Name = "Altern"
	}
	C.App.Domain = os.Getenv("APP_DOMAIN")
	if C.App.Domain == "" {
		C.App.Domain = "altern.ai"
	}

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

	C.GoogleOAuth.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	C.GoogleOAuth.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	C.GoogleOAuth.RedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	if C.GoogleOAuth.RedirectURL == "" {
		C.GoogleOAuth.RedirectURL = "http://localhost:1337/auth/google/callback"
	}

	C.CloudflareR2.AccountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	C.CloudflareR2.AccessKeyID = os.Getenv("CLOUDFLARE_ACCESS_KEY_ID")
	C.CloudflareR2.SecretAccessKey = os.Getenv("CLOUDFLARE_SECRET_ACCESS_KEY")
	C.CloudflareR2.Bucket = os.Getenv("CLOUDFLARE_R2_BUCKET")
	C.CloudflareR2.Region = os.Getenv("CLOUDFLARE_R2_REGION")
	if C.CloudflareR2.Region == "" {
		C.CloudflareR2.Region = "auto"
	}
	C.CloudflareR2.PublicURL = strings.TrimRight(os.Getenv("CLOUDFLARE_R2_PUBLIC_URL"), "/")

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

func (c *Config) OAuthGoogleEnabled() bool {
	return c.GoogleOAuth.ClientID != "" && c.GoogleOAuth.ClientSecret != ""
}


func (c *Config) R2Enabled() bool {
	return c.CloudflareR2.AccountID != "" &&
		c.CloudflareR2.AccessKeyID != "" &&
		c.CloudflareR2.SecretAccessKey != "" &&
		c.CloudflareR2.Bucket != ""
}

func (c *Config) IsSuperuserLogin(login string) bool {
	for _, allowed := range c.Superuser.GitHubLogins {
		if strings.EqualFold(allowed, login) {
			return true
		}
	}
	return false
}
