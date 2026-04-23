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

	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3Bucket         string
	S3UseSSL         bool
	S3PublicEndpoint string

	RabbitURL      string
	TranscodeQueue string
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

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		return nil, errors.New("S3_ENDPOINT is required")
	}
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	if s3AccessKey == "" {
		return nil, errors.New("S3_ACCESS_KEY is required")
	}
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	if s3SecretKey == "" {
		return nil, errors.New("S3_SECRET_KEY is required")
	}
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "onetube"
	}
	s3PublicEndpoint := os.Getenv("S3_PUBLIC_ENDPOINT")
	if s3PublicEndpoint == "" {
		s3PublicEndpoint = s3Endpoint
	}

	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		return nil, errors.New("RABBIT_URL is required")
	}
	transcodeQueue := os.Getenv("TRANSCODE_QUEUE")
	if transcodeQueue == "" {
		transcodeQueue = "transcode.jobs"
	}

	return &Config{
		DatabaseURL:      databaseURL,
		Port:             port,
		JWTSecret:        []byte(secret),
		CookieName:       "onetube_session",
		Secure:           false,
		S3Endpoint:       s3Endpoint,
		S3AccessKey:      s3AccessKey,
		S3SecretKey:      s3SecretKey,
		S3Bucket:         s3Bucket,
		S3UseSSL:         os.Getenv("S3_USE_SSL") == "true",
		S3PublicEndpoint: s3PublicEndpoint,
		RabbitURL:        rabbitURL,
		TranscodeQueue:   transcodeQueue,
	}, nil
}
