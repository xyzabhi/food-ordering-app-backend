package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/xyzabhi/food-ordering-app-backend/internal/coupons"
	"github.com/xyzabhi/food-ordering-app-backend/internal/config"
	"github.com/xyzabhi/food-ordering-app-backend/internal/database"
	"github.com/xyzabhi/food-ordering-app-backend/internal/redisclient"
	"github.com/xyzabhi/food-ordering-app-backend/internal/router"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v\n(set DATABASE_URL in .env — see .env.example; or run: docker compose up -d postgres redis)", err)
	}
	defer pool.Close()

	rdb, err := redisclient.New(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	log.Printf("Loading promo codes from %d gzip files...", len(cfg.CouponBaseURLs))
	couponStore, err := coupons.LoadValidCouponsWithCache(ctx, cfg.CouponBaseURLs, cfg.CouponCachePath)
	if err != nil {
		log.Fatalf("coupons: %v", err)
	}

	r := router.SetUp(pool, rdb, cfg.CORSAllowedOrigins, couponStore, cfg.CouponDiscountPct)

	log.Printf("Server is listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
