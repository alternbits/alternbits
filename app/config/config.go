package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database struct {
		DSN string
	}
	Server struct {
		Port string
	}
}

var C *Config

func Load() error {
	_ = godotenv.Load()

	C = &Config{}

	C.Database.DSN = os.Getenv("DB_DSN")
	if C.Database.DSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}

	C.Server.Port = os.Getenv("PORT")
	if C.Server.Port == "" {
		C.Server.Port = "1337"
	}

	return nil
}
