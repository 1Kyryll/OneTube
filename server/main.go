package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	config "github.com/1kyryll/onetube/server/internal/config"
	server "github.com/1kyryll/onetube/server/internal/http"
	service "github.com/1kyryll/onetube/server/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env", ".env")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	queries := gen.New(pool)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	svc := service.NewUserService(queries)

	srv := server.NewServer(cfg, svc)
	log.Printf("running on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, srv.SetupRoutes()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
