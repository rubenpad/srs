package persistence

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StockRatingRepository struct {
	pool *pgxpool.Pool
}

func NewStockRatingRepository(pool *pgxpool.Pool) *StockRatingRepository {
	return &StockRatingRepository{pool}
}

func (shr *StockRatingRepository) GetStockRatings() {
	shr.pool.Query(context.Background(), "SELECT * FROM stock_rating")
}
