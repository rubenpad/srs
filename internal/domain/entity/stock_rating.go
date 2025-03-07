package entity

import "context"

type StockRating struct {
	Brokerage  string `json:"brokerage"`
	Action     string `json:"action"`
	Company    string `json:"company"`
	Ticker     string `json:"ticker"`
	RatingFrom string `json:"rating_from"`
	RatingTo   string `json:"rating_to"`
	TargetFrom string `json:"target_from"`
	TargetTo   string `json:"target_to"`
	Time       string `json:"time"`
}

type IStockRatingApi interface {
	GetStockRatings(ctx context.Context, nextPage string) ([]StockRating, string, error)
}

type IStockRatingRepository interface {
	Save(ctx context.Context, stock StockRating) error
	GetStockRatings(ctx context.Context, limit int, offset int) ([]StockRating, error)
}

func NewStockRating(brokerage, action, company, ticker, ratingFrom, ratingTo, targetFrom, targetTo, time string) StockRating {
	return StockRating{
		Brokerage:  brokerage,
		Action:     action,
		Company:    company,
		Ticker:     ticker,
		RatingFrom: ratingFrom,
		RatingTo:   ratingTo,
		TargetFrom: targetFrom,
		TargetTo:   targetTo,
		Time:       time,
	}
}
