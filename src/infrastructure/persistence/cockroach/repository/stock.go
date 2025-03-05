package persistence

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool}
}

func (shr *StockRepository) GetStocks() {
	shr.pool.Query(context.Background(), "SELECT * FROM stock_rating_history")
}
