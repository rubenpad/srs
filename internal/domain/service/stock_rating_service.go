package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type serviceResponse[T any] struct {
	Data     []T    `json:"data"`
	NextPage string `json:"nextPage"`
}

type StockRatingService struct {
	stockRatingRepository entity.IStockRatingRepository
	stockRatingApi        entity.IStockRatingApi
	isLoading             atomic.Bool
}

func NewStockRatingService(stockRatingRepository entity.IStockRatingRepository, stockRatingApi entity.IStockRatingApi) *StockRatingService {
	return &StockRatingService{
		stockRatingApi:        stockRatingApi,
		stockRatingRepository: stockRatingRepository,
	}
}

func (s *StockRatingService) GetStockRatings(ctx context.Context, nextPage string, pageSize int) (*serviceResponse[entity.StockRating], error) {
	stockRatings, err := s.stockRatingRepository.GetStockRatings(ctx, nextPage, pageSize)

	if err != nil {
		return nil, err
	}

	nNextPage := ""
	responseSize := len(stockRatings)
	if responseSize > 0 {
		lastItem := stockRatings[responseSize-1]
		nNextPage = lastItem.Ticker
	}

	return &serviceResponse[entity.StockRating]{
		Data:     stockRatings,
		NextPage: nNextPage,
	}, nil
}

func (s *StockRatingService) GetStockRecommendations(ctx context.Context, pageSize int) (*serviceResponse[entity.StockRatingAggregate], error) {
	recommendations, err := s.stockRatingRepository.GetStockRecommendations(ctx, pageSize)

	if err != nil {
		return nil, err
	}

	return &serviceResponse[entity.StockRatingAggregate]{
		Data:     recommendations,
		NextPage: "",
	}, nil
}

func (s *StockRatingService) LoadStockRatingsData(ctx context.Context) {
	if !s.isLoading.CompareAndSwap(false, true) {
		slog.Info("load stock ratings process already running")
		return
	}

	defer s.isLoading.Store(false)

	slog.Info("process to load stock ratings started")
	start := time.Now()
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
	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds()) % 1000
	duration := fmt.Sprintf("%dm %ds %dms", minutes, seconds, milliseconds)
	slog.Info("process to load stock ratings finished", "duration", duration)
}

func calculatePriceTargetChange(rawTargetFrom, rawTargetTo string) float64 {
	targetFrom, fromErr := strconv.ParseFloat(strings.TrimPrefix(rawTargetFrom, "$"), 64)
	targetTo, toErr := strconv.ParseFloat(strings.TrimPrefix(rawTargetTo, "$"), 64)

	if fromErr != nil || toErr != nil {
		return 0
	}

	return (targetTo - targetFrom) / targetTo
}
