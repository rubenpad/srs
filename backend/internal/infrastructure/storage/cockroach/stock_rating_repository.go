package cockroach

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

const insertQuery = `INSERT INTO stock_rating (
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
			  VALUES (
			  	@brokerage,
				@action,
				@company,
				@ticker,
				@rating_from,
				@rating_to,
				@target_from,
				@target_to,
				@time,
				@target_price_change)`

type StockRatingRepository struct {
	pool *pgxpool.Pool
}

func NewStockRatingRepository(pool *pgxpool.Pool) *StockRatingRepository {
	return &StockRatingRepository{pool}
}

func (srr *StockRatingRepository) GetStockRatings(context context.Context, nextPage string, pageSize int, search string) ([]entity.StockRating, error) {
	query := `
        SELECT 
            brokerage,
            action,
            company,
            ticker,
            rating_from,
            rating_to,
            target_from,
            target_to,
            time,
            target_price_change
        FROM stock_rating
        WHERE (@nextPage = '' OR ticker > @nextPage)
		AND (@search = '' OR UPPER(ticker) BETWEEN UPPER(@search) AND CONCAT(UPPER(@search), 'ÿ'))
        ORDER BY ticker ASC
        LIMIT @pageSize
	`

	args := pgx.NamedArgs{"nextPage": nextPage, "pageSize": pageSize, "search": search}
	rows, err := srr.pool.Query(context, query, args)

	if err != nil {
		errorMessage := "error getting stock ratings"
		slog.Error(errorMessage, "error", err)
		return nil, errors.New(errorMessage)
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.StockRating])
}

func (srr *StockRatingRepository) Save(ctx context.Context, stockRating entity.StockRating) {
	args := pgx.NamedArgs{
		"brokerage":           stockRating.Brokerage,
		"action":              stockRating.Action,
		"company":             stockRating.Company,
		"ticker":              stockRating.Ticker,
		"rating_from":         stockRating.RatingFrom,
		"rating_to":           stockRating.RatingTo,
		"target_from":         stockRating.TargetFrom,
		"target_to":           stockRating.TargetTo,
		"time":                stockRating.Time,
		"target_price_change": stockRating.TargetPriceChange,
	}
	_, err := srr.pool.Exec(ctx, insertQuery, args)

	if err != nil {
		slog.Error("error saving stock rating", "error", err)
	}
}

func (srr *StockRatingRepository) BatchSave(ctx context.Context, stockRatings []entity.StockRating) {
	batch := &pgx.Batch{}
	for _, stockRating := range stockRatings {
		args := pgx.NamedArgs{
			"brokerage":           stockRating.Brokerage,
			"action":              stockRating.Action,
			"company":             stockRating.Company,
			"ticker":              stockRating.Ticker,
			"rating_from":         stockRating.RatingFrom,
			"rating_to":           stockRating.RatingTo,
			"target_from":         stockRating.TargetFrom,
			"target_to":           stockRating.TargetTo,
			"time":                stockRating.Time,
			"target_price_change": stockRating.TargetPriceChange,
		}

		batch.Queue(insertQuery, args)
	}

	results := srr.pool.SendBatch(ctx, batch)
	defer results.Close()

	for _, stockRating := range stockRatings {
		_, err := results.Exec()

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				slog.Info(
					"skipping duplicate stock rating - probably running the load data again",
					"constraint", pgErr.ConstraintName,
					"data", stockRating,
				)
				continue
			}
			slog.Error("error saving stock rating", "error", err)
		}
	}
}

func (ssr *StockRatingRepository) GetStockRecommendations(ctx context.Context, pageSize int) ([]entity.StockRatingAggregate, error) {
	query := `
		WITH latest_stock_ratings AS
  		(SELECT
			ticker,
          	MAX(time) AS "time",
          	COUNT(CASE WHEN rating_to IN ('Strong-Buy', 'Buy', 'Top Pick', 'Positive', 'Outperform', 'Outperformer', 'Market Outperform', 'Sector Outperform') THEN 1 ELSE NULL END) AS strong_buy_ratings,
          	COUNT(CASE WHEN rating_to IN ('Overweight', 'Equal Weight', 'Sector Weight', 'Peer Perform', 'In-Line', 'Inline') THEN 1 ELSE NULL END) AS buy_ratings,
          	COUNT(CASE WHEN rating_to IN ('Neutral', 'Market Perform', 'Sector Perform', 'Hold') THEN 1 ELSE NULL END) AS hold_ratings,
          	COUNT(CASE WHEN rating_to IN ('Sell', 'Reduce', 'Negative', 'Underweight', 'Underperform', 'Sector Underperform') THEN 1 ELSE NULL END) AS sell_ratings
   		FROM (SELECT
				ticker,
             	brokerage,
             	rating_to,
             	time,
             	ROW_NUMBER() OVER (PARTITION BY ticker, brokerage ORDER BY time DESC) AS rn
      		FROM stock_rating) AS ranked_stock_ratings
   		WHERE rn <= 5
   		GROUP BY ticker)
		SELECT
			ticker,
			time,
			strong_buy_ratings,
			buy_ratings,
			hold_ratings,
			sell_ratings
		FROM latest_stock_ratings
		ORDER BY strong_buy_ratings DESC, buy_ratings DESC, time DESC
		LIMIT @pageSize;	
	`

	args := pgx.NamedArgs{
		"pageSize": pageSize,
	}

	rows, err := ssr.pool.Query(ctx, query, args)

	if err != nil {
		errorMessage := "error getting stock ratings"
		slog.Error(errorMessage, "error", err)
		return nil, errors.New(errorMessage)
	}

	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.StockRatingAggregate])
}
