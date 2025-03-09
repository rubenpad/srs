package entity

import (
	"context"
	"time"
)

type StockRating struct {
	Brokerage         string    `json:"brokerage"`
	Action            string    `json:"action"`
	Company           string    `json:"company"`
	Ticker            string    `json:"ticker"`
	RatingFrom        string    `json:"rating_from"`
	RatingTo          string    `json:"rating_to"`
	TargetFrom        string    `json:"target_from"`
	TargetTo          string    `json:"target_to"`
	Time              time.Time `json:"time"`
	TargetPriceChange float64   `json:"target_price_change"`
}

type StockRatingAggregate struct {
	Ticker           string    `json:"ticker"`
	Time             time.Time `json:"time"`
	StrongBuyRatings int       `json:"strong_buy_ratings"`
	BuyRatings       int       `json:"buy_ratings"`
	HoldRatings      int       `json:"hold_ratings"`
	SellRatings      int       `json:"sell_ratings"`
}

type IStockRatingApi interface {
	GetStockRatings(ctx context.Context, nextPage string) ([]StockRating, string, error)
}

type IStockRatingRepository interface {
	Save(ctx context.Context, stock StockRating)
	BatchSave(ctx context.Context, stockRatings []StockRating)
	GetStockRatings(ctx context.Context, nextPage string, pageSize int) ([]StockRating, error)
	GetStockRecommendations(ctx context.Context, pageSize int) ([]StockRatingAggregate, error)
}

func NewStockRating(brokerage, action, company, ticker, ratingFrom, ratingTo, targetFrom, targetTo string, time time.Time, targetPriceChange float64) StockRating {
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
