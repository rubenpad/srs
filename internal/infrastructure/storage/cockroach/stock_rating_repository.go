package cockroach

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockRatingRepository struct {
	pool *pgxpool.Pool
}

func NewStockRatingRepository(pool *pgxpool.Pool) *StockRatingRepository {
	return &StockRatingRepository{pool}
}

func (srr *StockRatingRepository) GetStockRatings(context context.Context, limit int, offset int) ([]entity.StockRating, error) {
	rows, err := srr.pool.Query(context, "SELECT * FROM stock_rating")

	if err != nil {
		slog.Error("error getting stockRatings", "error", err)
		return nil, fmt.Errorf("error getting stock ratings")
	}

	defer rows.Close()

	stockRatings := make([]entity.StockRating, 0, limit)
	for rows.Next() {
		var stockRating entity.StockRating
		if err := rows.Scan(&stockRating); err != nil {
			return nil, err
		}

		stockRatings = append(stockRatings, stockRating)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stockRatings, nil
}

func (srr *StockRatingRepository) Save(ctx context.Context, stockRating entity.StockRating) error {
	_, err := srr.pool.Exec(ctx, `INSERT INTO stock_rating (
										brokerage,
										action,
										company,
										ticker,
										rating_from,
										rating_to,
										target_from,
										target_to,
										time)
									VALUES ($1, $2, $3) RETURNING *`,
		stockRating.Brokerage,
		stockRating.Action,
		stockRating.Company,
		stockRating.Ticker,
		stockRating.RatingFrom,
		stockRating.RatingTo,
		stockRating.TargetFrom,
		stockRating.TargetTo,
		stockRating.Time)

	if err != nil {
		slog.Error("error saving stock rating", "error", err)
		return fmt.Errorf("error saving stock rating")
	}

	return nil
}
