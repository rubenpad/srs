package entity

import (
	"context"
	"time"

	"github.com/Finnhub-Stock-API/finnhub-go/v2"
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
	Score             float32   `json:"score"`
}
type StockRatingAggregate struct {
	Ticker            string    `json:"ticker"`
	Time              time.Time `json:"time"`
	StrongBuyRatings  int       `json:"strong_buy_ratings"`
	BuyRatings        int       `json:"buy_ratings"`
	HoldRatings       int       `json:"hold_ratings"`
	SellRatings       int       `json:"sell_ratings"`
	Rating            string    `json:"rating"`
	TargetPriceChange float64   `json:"target_price_change"`
	Score             float32   `json:"score"`
}

type StockDetails struct {
	KeyFacts        string                         `json:"keyFacts"`
	Quote           *finnhub.Quote                 `json:"quote"`
	Recommendations *[]finnhub.RecommendationTrend `json:"recommendations"`
}

type IStockRatingApi interface {
	GetStockDetails(ctx context.Context, ticker string) *StockDetails
	GetStockRatings(ctx context.Context, nextPage string, useCustomFormat bool) ([]StockRating, string, error)
}

type IStockRatingRepository interface {
	Save(ctx context.Context, stock StockRating)
	GetStockRatings(ctx context.Context, nextPage string, pageSize int, search string) ([]StockRating, error)
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
