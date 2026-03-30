package config

import (
	"os"
	"strings"
	"strconv"
)

type Config struct {
	DatabaseURL        string
	RedisAddr          string
	HTTPAddr           string // host:port for Listen, e.g. ":8080"
	CORSAllowedOrigins []string
	CouponBaseURLs     []string
	CouponDiscountPct  float64
	CouponCachePath    string
}

func Load() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Matches docker-compose postgres service when bound to the host.
		dbURL = "postgres://foodapp:foodapp@127.0.0.1:5433/foodapp?sslmode=disable"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	httpAddr := ":" + port

	corsOrigins := parseCSVEnv("CORS_ALLOWED_ORIGINS")
	if len(corsOrigins) == 0 {
		corsOrigins = []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://food-ordering-app-two-pearl.vercel.app",
		}
	}

	couponURLs := parseCSVEnv("COUPON_BASE_URLS")
	if len(couponURLs) == 0 {
		// These are the three .gz files you provided.
		couponURLs = []string{
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
		}
	}

	discountPct := 15.0
	if raw := strings.TrimSpace(os.Getenv("COUPON_DISCOUNT_PCT")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			discountPct = v
		}
	}

	cachePath := os.Getenv("COUPON_CACHE_PATH")
	if cachePath == "" {
		cachePath = "./data/coupon_cache.bin"
	}

	return Config{
		DatabaseURL:        dbURL,
		RedisAddr:          redisAddr,
		HTTPAddr:           httpAddr,
		CORSAllowedOrigins: corsOrigins,
		CouponBaseURLs:     couponURLs,
		CouponDiscountPct:  discountPct,
		CouponCachePath:    cachePath,
	}
}

func parseCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
