package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   []byte
	CookieName  string
	Secure      bool
}

func LoadConfig() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: databaseURL,
		Port:        port,
		JWTSecret:   []byte(secret),
		CookieName:  "onetube_session",
		Secure:      false,
	}, nil
}
