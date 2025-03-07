package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

const defaultItemsSize = 10

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
		stockRatingRepository: stockRatingRepository,
		stockRatingApi:        stockRatingApi,
	}
}

func (s StockRatingService) GetStockRatings(ctx context.Context, limit int, offset int) ([]entity.StockRating, error) {
	stockRatings, err := s.stockRatingRepository.GetStockRatings(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return stockRatings, nil
}

func (s StockRatingService) AnalyzeStockRatings(ctx context.Context) error {
	const workerCount = 4

	results := make(chan error, defaultItemsSize)
	jobs := make(chan entity.StockRating, defaultItemsSize)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rating := range jobs {
				score := calculateRatingScore(rating)
				stockRating := entity.NewStockRating(rating.Brokerage, rating.Action, rating.Company, rating.Ticker, rating.RatingFrom, rating.RatingTo, rating.TargetFrom, rating.TargetTo, rating.Time)

				if err := s.stockRatingRepository.Save(ctx, stockRating); err != nil {
					results <- fmt.Errorf("failed to save stock rating %s: %w", rating.Ticker, err)
					continue
				}

				maybeExistingStock, err := s.stockRepository.GetStock(ctx, rating.Ticker)
				if err != nil {
					slog.Warn("error getting stock from database", "ticker", stockRating.Ticker)
				}

				newStock := entity.NewStock(rating.Ticker, rating.Company, score+maybeExistingStock.Score)
				if err := s.stockRepository.Save(ctx, newStock); err != nil {
					results <- fmt.Errorf("failed to save stock %s: %w", rating.Ticker, err)
					continue
				}

				results <- nil
			}
		}()
	}

	nextPage := ""
	for {
		stockRatings, nNextPage, err := s.stockRatingApi.GetStockRatings(ctx, nextPage)
		if err != nil {
			return fmt.Errorf("failed to get stock ratings from API")
		}

		for _, rating := range stockRatings {
			jobs <- rating
		}

		for range stockRatings {
			if err := <-results; err != nil {
				slog.Error("Error processing stock rating: %v\n", "error", err)
			}
		}

		if nNextPage == "" {
			break
		}

		nextPage = nNextPage
	}

	close(jobs)
	wg.Wait()
	close(results)

	return nil
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

func calculateRatingScore(stockRating entity.StockRating) int {
	score := 0
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

	score += calculateCurrentRatingScore(ratingTo) + calculatePriceTargetChangeScore(stockRating.TargetFrom, stockRating.TargetTo)
	return score * int(calculateDateScore(stockRating.Time))
}

func calculateCurrentRatingScore(ratingToScaleValue int) int {
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

func calculatePriceTargetChangeScore(rawTargetFrom, rawTargetTo string) int {
	targetFrom, err := strconv.ParseFloat(strings.TrimPrefix(rawTargetFrom, "$"), 64)
	if err != nil {
		return 0
	}

	targetTo, err := strconv.ParseFloat(strings.TrimPrefix(rawTargetTo, "$"), 64)
	if err != nil {
		return 0
	}

	variation := (targetFrom - targetTo) / targetFrom * 100

	if variation < 0 {
		return -3
	} else if variation >= 0.3 {
		return 3
	} else if variation >= 0.1 {
		return 2
	} else {
		return 1
	}
}

func calculateDateScore(dateStr string) float32 {
	reportDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return 0
	}

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
