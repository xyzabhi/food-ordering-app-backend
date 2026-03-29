package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	RedisAddr   string
	HTTPAddr    string // host:port for Listen, e.g. ":8080"
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
	return Config{
		DatabaseURL: dbURL,
		RedisAddr:   redisAddr,
		HTTPAddr:    httpAddr,
	}
}
