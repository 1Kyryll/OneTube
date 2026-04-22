package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/1kyryll/onetube/server/internal/config"
	userhandlers "github.com/1kyryll/onetube/server/internal/user/handlers"
	userservices "github.com/1kyryll/onetube/server/internal/user/services"
)

func main() {
	_ = godotenv.Load("../../.env", "../.env", ".env")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	queries := gen.New(pool)

	userSvc := userservices.NewUserService(queries)
	userH := userhandlers.NewUserHTTPHandler(cfg, userSvc)

	router := newRouter(cfg, userH)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
