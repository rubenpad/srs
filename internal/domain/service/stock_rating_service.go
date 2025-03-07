package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

var bullishRatings = []string{
	"Strong-Buy",
	"Buy",
	"Top Pick",
	"Positive",
	"Outperform",
	"Outperformer",
	"Market Outperform",
	"Sector Outperform",
	"Market Outperform",
}

var intermediateBullishRatings = []string{
	"Overweight",
	"Equal Weight",
	"Sector Weight",
	"Peer Perform",
	"In-Line",
	"Inline",
}

var neutralRatings = []string{
	"Neutral",
	"Market Perform",
	"Sector Perform",
	"Hold",
}

var intermediateBearishRatings = []string{
	"Reduce",
	"Negative",
	"Underweight",
	"Underperform",
	"Sector Underperform",
}

var bearishRatings = []string{
	"Sell",
}

type StockRatingService struct {
	stockRepository       entity.IStockRepository
	stockRatingRepository entity.IStockRatingRepository
	stockRatingApi        entity.IStockRatingApi
}

func NewStockRatingService(stockRatingRepository entity.IStockRatingRepository, stockRepository entity.IStockRepository, stockRatingApi entity.IStockRatingApi) StockRatingService {
	return StockRatingService{
		stockRatingApi:        stockRatingApi,
		stockRepository:       stockRepository,
		stockRatingRepository: stockRatingRepository,
	}
}

func (s StockRatingService) GetStockRatings(ctx context.Context, limit int, offset int) ([]entity.StockRating, error) {
	stockRatings, err := s.stockRatingRepository.GetStockRatings(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return stockRatings, nil
}

func (s StockRatingService) AnalyzeStockRatings(ctx context.Context) {
	slog.Info("process to load stock ratings started")
	nextPage := ""

	for {
		stockRatings, nNextPage, err := s.stockRatingApi.GetStockRatings(ctx, nextPage)
		if err != nil {
			errorMessage := "failed to get stock ratings from API"
			slog.Error(errorMessage, "error", err)
			break
		}

		for _, rating := range stockRatings {

			priceTargetChange, ratingScore := calculateRatingScore(rating)
			stockRating := entity.NewStockRating(rating.Brokerage, rating.Action, rating.Company, rating.Ticker, rating.RatingFrom, rating.RatingTo, rating.TargetFrom, rating.TargetTo, rating.Time, priceTargetChange)

			if err := s.stockRatingRepository.Save(ctx, stockRating); err != nil {
				slog.Error("error saving stock rating", "error", err)
				continue
			}

			s.saveStock(ctx, stockRating, ratingScore)

		}

		nextPage = nNextPage

		if nNextPage == "" {
			break
		}
	}

	slog.Info("process to load stock ratings finished")
}

func (s StockRatingService) saveStock(ctx context.Context, stockRating entity.StockRating, ratingScore float64) {
	maybeExistingStock, err := s.stockRepository.GetStock(ctx, stockRating.Ticker)
	if err != nil {
		slog.Error("error getting stock from database", "ticker", stockRating.Ticker)
	}

	newScore := ratingScore
	if maybeExistingStock != nil {
		newScore += maybeExistingStock.Score
	}

	newStock := entity.NewStock(stockRating.Ticker, stockRating.Company, newScore)
	if err := s.stockRepository.Save(ctx, newStock); err != nil {
		slog.Error("error saving stock", "error", err)
	}
}

func buildRatingsScaleMap() map[string]int {
	ratingScale := make(map[string]int)

	for _, rating := range bullishRatings {
		ratingScale[rating] = 1
	}

	for _, rating := range intermediateBullishRatings {
		ratingScale[rating] = 2
	}

	for _, rating := range neutralRatings {
		ratingScale[rating] = 3
	}

	for _, rating := range intermediateBearishRatings {
		ratingScale[rating] = 4
	}

	for _, rating := range bearishRatings {
		ratingScale[rating] = 5
	}

	return ratingScale
}

func calculateRatingScore(stockRating entity.StockRating) (float64, float64) {
	var score float64 = 0
	ratingScale := buildRatingsScaleMap()

	ratingFrom := ratingScale[stockRating.RatingFrom]
	ratingTo := ratingScale[stockRating.RatingTo]

	if stockRating.RatingFrom != stockRating.RatingTo {
		// upgrade
		if ratingFrom > ratingTo {
			score += 3
		} else {
			score -= 3
		}
	}

	dateScore := calculateDateScore(stockRating.Time)
	currentRatingScore := calculateCurrentRatingScore(ratingTo)
	priceTargetChange, priceTargetChangeScore := calculatePriceTargetChangeScore(stockRating.TargetFrom, stockRating.TargetTo)
	scoreResult := (score + currentRatingScore + priceTargetChangeScore) * dateScore

	return priceTargetChange, scoreResult
}

func calculateCurrentRatingScore(ratingToScaleValue int) float64 {
	switch ratingToScaleValue {
	case 1, 2:
		return 2
	case 3:
		return 0
	case 4, 5:
		return -2
	default:
		return 0
	}
}

func calculatePriceTargetChangeScore(rawTargetFrom, rawTargetTo string) (float64, float64) {
	targetFrom, err := strconv.ParseFloat(strings.TrimPrefix(rawTargetFrom, "$"), 64)
	if err != nil {
		return 0, 0
	}

	targetTo, err := strconv.ParseFloat(strings.TrimPrefix(rawTargetTo, "$"), 64)
	if err != nil {
		return 0, 0
	}

	variation := (targetFrom - targetTo) / targetFrom * 100

	if variation < 0 {
		return variation, -3
	}

	if variation >= 0.3 {
		return variation, 3
	}

	if variation >= 0.1 {
		return variation, 2
	}

	return variation, 1
}

func calculateDateScore(reportDate time.Time) float64 {
	daysSinceReport := time.Since(reportDate).Hours() / 24

	switch {
	case daysSinceReport <= 7:
		return 1.0
	case daysSinceReport <= 30:
		return 0.75
	case daysSinceReport <= 90:
		return 0.5
	default:
		return 0.25
	}
}
