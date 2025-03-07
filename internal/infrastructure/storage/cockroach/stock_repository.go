package cockroach

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{pool}
}

func (sr *StockRepository) GetStock(ctx context.Context, ticker string) (*entity.Stock, error) {
	row := sr.pool.QueryRow(ctx, `SELECT * FROM stock WHERE ticker = $1`, ticker)

	var stock *entity.Stock
	if err := row.Scan(&stock); err != nil {
		slog.Error(fmt.Sprintf("error getting stock: %s", ticker), "error", err)
		return nil, fmt.Errorf("error getting stock")
	}

	return stock, nil
}

func (sr *StockRepository) GetStocks(ctx context.Context, limit int, offset int) ([]entity.Stock, error) {
	rows, err := sr.pool.Query(ctx, `SELECT * FROM stock LIMIT $1 OFFSET $2`, limit, offset)

	if err != nil {
		slog.Error("error getting stocks", "error", err)
		return nil, fmt.Errorf("error getting stocks")
	}

	defer rows.Close()

	stocks := make([]entity.Stock, 0, limit)
	for rows.Next() {
		var stock entity.Stock
		if err := rows.Scan(&stock); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}

func (sr *StockRepository) Save(ctx context.Context, stock entity.Stock) error {
	err := sr.pool.QueryRow(
		ctx, `INSERT INTO stock (ticker, company, score)
			  VALUES ($1, $2, $3) RETURNING *`,
		stock.Ticker,
		stock.Company,
		stock.Score).Scan(&stock)

	if err != nil {
		slog.Error("error saving stock", "error", err)
		return fmt.Errorf("error saving stock")
	}

	return nil
}
