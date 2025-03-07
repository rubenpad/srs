package cockroach

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
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
	var stock entity.Stock
	err := sr.pool.QueryRow(ctx, `
		SELECT ticker, company, score
		FROM stock
		WHERE ticker = $1`, ticker).Scan(&stock.Ticker, &stock.Company, &stock.Score)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		slog.Error("error getting stock", "ticker", ticker, "error", err)
		return nil, fmt.Errorf("error getting stock")
	}

	return &stock, nil
}

func (sr *StockRepository) GetStocks(ctx context.Context, limit int, offset int) ([]entity.Stock, error) {
	query := `
		SELECT ticker, company, score
		FROM stock
		LIMIT @limit
		OFFSET @offset
	`

	args := pgx.NamedArgs{
		"limit":  limit,
		"offset": offset,
	}

	rows, err := sr.pool.Query(ctx, query, args)

	if err != nil {
		slog.Error("error getting stocks", "error", err)
		return nil, fmt.Errorf("error getting stocks")
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.Stock])
}

func (sr *StockRepository) Save(ctx context.Context, stock entity.Stock) error {
	query := `
		INSERT INTO stock (ticker, company, score)
		VALUES (@ticker, @company, @score) RETURNING *`

	args := pgx.NamedArgs{
		"ticker":  stock.Ticker,
		"company": stock.Company,
		"score":   stock.Score,
	}

	_, err := sr.pool.Exec(ctx, query, args)

	if err != nil {
		slog.Error("error saving stock", "error", err)
		return fmt.Errorf("error saving stock")
	}

	return nil
}
