package entity

import "context"

type StockRating struct {
	Brokerage         string
	Action            string
	Company           string
	Ticker            string
	RatingFrom        string
	RatingTo          string
	TargetFrom        float64
	TargetTo          float64
	Time              string
	TargetPriceChange float64
}

type IStockRatingRepository interface {
	Save(ctx context.Context, stock StockRating) error
	GetStockRatings(ctx context.Context, limit int, offset int) ([]StockRating, error)
}

func NewStockRating(brokerage, action, company, ticker, ratingFrom, ratingTo string, targetFrom, targetTo float64, time string, targetPriceChange float64) StockRating {
	return StockRating{
		Brokerage:         brokerage,
		Action:            action,
		Company:           company,
		Ticker:            ticker,
		RatingFrom:        ratingFrom,
		RatingTo:          ratingTo,
		TargetFrom:        targetFrom,
		TargetTo:          targetTo,
		Time:              time,
		TargetPriceChange: targetPriceChange,
	}
}
