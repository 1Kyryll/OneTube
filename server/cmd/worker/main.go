package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/1kyryll/onetube/server/internal/common/queue"
	s3client "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/1kyryll/onetube/server/internal/config"
	transcodesvc "github.com/1kyryll/onetube/server/internal/transcode/services"
)

func main() {
	_ = godotenv.Load("../../.env", "../.env", ".env")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	consumer, err := queue.NewConsumer(cfg.RabbitURL, cfg.TranscodeQueue)
	if err != nil {
		log.Fatalf("Failed to create Rabbit consumer: %v", err)
	}
	defer consumer.Close()

	transcoder := transcodesvc.NewTranscoder(queries, s3)

	log.Printf("worker: consuming queue %q", cfg.TranscodeQueue)
	if err := consumer.Run(ctx, transcoder.Process); err != nil && err != context.Canceled {
		log.Fatalf("consumer: %v", err)
	}
}
