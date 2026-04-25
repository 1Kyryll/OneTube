package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/1kyryll/onetube/server/internal/common/queue"
	s3client "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/1kyryll/onetube/server/internal/config"
	userhandlers "github.com/1kyryll/onetube/server/internal/user/handlers"
	userservices "github.com/1kyryll/onetube/server/internal/user/services"
	videohandlers "github.com/1kyryll/onetube/server/internal/video/handlers"
	videoservices "github.com/1kyryll/onetube/server/internal/video/services"
)

func main() {
	_ = godotenv.Load("../../.env", "../.env", ".env")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fail to connect to DB: %v", err)
	}
	defer pool.Close()
	queries := gen.New(pool)

	s3, err := s3client.NewClient(s3client.Options{
		Endpoint:       cfg.S3Endpoint,
		PublicEndpoint: cfg.S3PublicEndpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		Bucket:         cfg.S3Bucket,
		UseSSL:         cfg.S3UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create s3 client: %v", err)
	}

	publisher, err := queue.NewPublisher(cfg.RabbitURL, cfg.TranscodeQueue)
	if err != nil {
		log.Fatalf("Failed to create Rabbit publisher: %v", err)
	}
	defer publisher.Close()

	userSvc := userservices.NewUserService(queries)
	userH := userhandlers.NewUserHTTPHandler(cfg, userSvc)

	videoSvc := videoservices.NewVideoService(queries, s3, publisher, cfg.APIPublicBaseURL)
	videoH := videohandlers.NewVideoHTTPHandler(cfg, videoSvc)

	router := newRouter(cfg, userH, videoH)

	log.Printf("api listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
