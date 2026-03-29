package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/xyzabhi/food-ordering-app-backend/internal/handler"
	"github.com/xyzabhi/food-ordering-app-backend/internal/repository"
	"github.com/xyzabhi/food-ordering-app-backend/internal/service"
)

func SetUp(db *pgxpool.Pool, rdb *redis.Client, corsAllowedOrigins []string) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	healthHandler := handler.NewHealthHandler(db, rdb)
	r.GET("/health", healthHandler.HealthCheck)

	productRepo := repository.NewProductRepository(db)
	productHandler := handler.NewProductHandler(productRepo)
	r.GET("/products", productHandler.ListProducts)
	r.GET("/products/:id", productHandler.GetProductByID)

	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, productRepo)
	orderHandler := handler.NewOrderHandler(orderSvc)
	r.POST("/orders", orderHandler.CreateOrder)

	return r
}
