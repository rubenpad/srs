package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockRatingService struct {
	stockRatingRepository entity.IStockRatingRepository
	stockRatingApi        entity.IStockRatingApi
}

func NewStockRatingService(stockRatingRepository entity.IStockRatingRepository, stockRatingApi entity.IStockRatingApi) StockRatingService {
	return StockRatingService{
		stockRatingApi:        stockRatingApi,
		stockRatingRepository: stockRatingRepository,
	}
}

func (s StockRatingService) GetStockRatings(ctx context.Context, page int, size int) ([]entity.StockRating, error) {
	return s.stockRatingRepository.GetStockRatings(ctx, size, (page-1)*size)
}

func (s StockRatingService) GetStockRecommendations(ctx context.Context, limit int) ([]entity.StockRatingAggregate, error) {
	return s.stockRatingRepository.GetStockRecommendations(ctx, limit)
}

func (s StockRatingService) LoadStockRatingsData(ctx context.Context) {
	slog.Info("process to load stock ratings started")
	nextPage := ""

	for {
		stockRatings, nNextPage, err := s.stockRatingApi.GetStockRatings(ctx, nextPage)
		if err != nil {
			errorMessage := "failed to get stock ratings from API"
			slog.Error(errorMessage, "error", err)
			break
		}

		formattedStockRatings := make([]entity.StockRating, 0, len(stockRatings))

		for _, rating := range stockRatings {
			stockRating := entity.NewStockRating(
				rating.Brokerage,
				rating.Action,
				rating.Company,
				rating.Ticker,
				rating.RatingFrom,
				rating.RatingTo,
				rating.TargetFrom,
				rating.TargetTo,
				rating.Time,
				calculatePriceTargetChange(rating.TargetFrom, rating.TargetTo))

			formattedStockRatings = append(formattedStockRatings, stockRating)
		}

		s.stockRatingRepository.BatchSave(ctx, formattedStockRatings)

		nextPage = nNextPage

		if nNextPage == "" {
			break
		}
	}

	slog.Info("process to load stock ratings finished")
}

func calculatePriceTargetChange(rawTargetFrom, rawTargetTo string) float64 {
	targetFrom, fromErr := strconv.ParseFloat(strings.TrimPrefix(rawTargetFrom, "$"), 64)
	targetTo, toErr := strconv.ParseFloat(strings.TrimPrefix(rawTargetTo, "$"), 64)

	if fromErr != nil || toErr != nil {
		return 0
	}

	return (targetFrom - targetTo) / targetFrom * 100
}
