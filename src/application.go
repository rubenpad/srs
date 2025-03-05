package main

import (
	"context"
	"log"
	"os"

	"github.com/rubenpad/stock-rating-system/domain/service"
	persistence "github.com/rubenpad/stock-rating-system/infrastructure/persistence/cockroach/repository"
	"github.com/rubenpad/stock-rating-system/infrastructure/rest"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeApplication() {
	connectionPoolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL") + "sslmode=require&pool_max_conns=40&pool_max_conn_lifetime=300s&pool_max_conn_lifetime_jitter=30s")
	if err != nil {
		log.Fatalf("Failed to parse connection pool config: %v", err)
	}

	connectionPool, err := pgxpool.NewWithConfig(context.Background(), connectionPoolConfig)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	defer connectionPool.Close()

	stockController := rest.NewStockController(service.NewStockService(persistence.NewStockRepository(connectionPool)))
	stockRatingController := rest.NewStockRatingController(service.NewStockRatingService(persistence.NewStockRatingRepository(connectionPool)))

	server := gin.Default()
	server.GET("/api/stocks", stockController.GetStocks)
	server.GET("/api/stocks-rating", stockRatingController.GetStockRatings)

	if err := server.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
