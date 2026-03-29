package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := http.StatusOK
	body := gin.H{"status": "ok"}

	if err := h.db.Ping(ctx); err != nil {
		status = http.StatusServiceUnavailable
		body["postgres"] = "error"
		body["postgres_error"] = err.Error()
	} else {
		body["postgres"] = "ok"
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		status = http.StatusServiceUnavailable
		body["redis"] = "error"
		body["redis_error"] = err.Error()
	} else {
		body["redis"] = "ok"
	}

	if status != http.StatusOK {
		body["status"] = "degraded"
	}

	c.JSON(status, body)
}
