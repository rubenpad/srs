package cockroach

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool}
}

func (shr *StockRepository) GetStocks(ctx context.Context, limit int, offset int) ([]entity.Stock, error) {
	rows, err := shr.pool.Query(ctx, `SELECT * FROM stock_rating_history LIMIT $1 OFFSET $2`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("error getting stocks")
	}

	defer rows.Close()

	var stocks []entity.Stock
	for rows.Next() {
		var stock entity.Stock
		if err := rows.Scan(&stock.Ticker, &stock.Company, &stock.Score); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}

func (shr *StockRepository) Save(ctx context.Context, stock entity.Stock) error {
	if err := shr.pool.QueryRow(ctx, `INSERT INTO stock (ticker, company, score) VALUES ($1, $2, $3) RETURNING *`, stock.Ticker, stock.Company, stock.Score).Scan(&stock); err != nil {
		return err
	}

	return nil
}
