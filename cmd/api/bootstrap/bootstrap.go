package bootstrap

import (
	"context"
	"log"
	"os"

	"github.com/rubenpad/stock-rating-system/internal/domain/service"
	"github.com/rubenpad/stock-rating-system/internal/infrastructure/server"
	"github.com/rubenpad/stock-rating-system/internal/infrastructure/storage/cockroach"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	connectionPoolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL") + "sslmode=require&pool_max_conns=40&pool_max_conn_lifetime=300s&pool_max_conn_lifetime_jitter=30s")
	if err != nil {
		log.Fatalf("Failed to parse connection pool config: %v", err)
	}

	connectionPool, err := pgxpool.NewWithConfig(context.Background(), connectionPoolConfig)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	defer connectionPool.Close()

	stockController := server.NewStockController(service.NewStockService(cockroach.NewStockRepository(connectionPool)))
	stockRatingController := server.NewStockRatingController(service.NewStockRatingService(cockroach.NewStockRatingRepository(connectionPool)))

	server := gin.Default()
	server.GET("/api/stocks", stockController.GetStocks)
	server.GET("/api/stocks-rating", stockRatingController.GetStockRatings)

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
