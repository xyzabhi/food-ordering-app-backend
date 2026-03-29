package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/xyzabhi/food-ordering-app-backend/internal/config"
	"github.com/xyzabhi/food-ordering-app-backend/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("migrations applied (nothing new if already up to date)")
}
