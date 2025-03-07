package cockroach

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
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
		errorMessage := "error getting stock ratings"
		slog.Error(errorMessage, "error", err)
		return nil, errors.New(errorMessage)
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.StockRating])
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
										time,
										target_price_change)
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *`,
		stockRating.Brokerage,
		stockRating.Action,
		stockRating.Company,
		stockRating.Ticker,
		stockRating.RatingFrom,
		stockRating.RatingTo,
		stockRating.TargetFrom,
		stockRating.TargetTo,
		stockRating.Time,
		stockRating.TargetPriceChange)

	if err != nil {
		slog.Error("error saving stock rating", "error", err)
		return fmt.Errorf("error saving stock rating")
	}

	return nil
}
