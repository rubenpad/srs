package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rubenpad/srs/internal/domain/entity"
)

const (
	strongBuyRatingScore    = 5
	buyRatingScore          = 4
	holdRatingScore         = 3
	sellRatingScore         = 2
	strongSellRatingScore   = 1
	reportDateWeight        = 5
	ratingChangeWeight      = 50
	currentTargetWeight     = 15
	brokerageActionWeight   = 25
	targetPriceChangeWeight = 5

	workers           = 4
	itemsBatchSize    = 10
	channelBufferSize = itemsBatchSize * (workers / 2)
)

var actionsScaleMap = map[string]int{
	"upgraded by":       5,
	"target raised by":  5,
	"initiated by":      3,
	"target set by":     2,
	"reiterated by":     2,
	"target lowered by": 1,
	"downgraded by":     1,
}

var ratingScaleMap = map[string]int{
	"Strong-Buy":        strongBuyRatingScore,
	"Buy":               strongBuyRatingScore,
	"Top Pick":          strongBuyRatingScore,
	"Positive":          strongBuyRatingScore,
	"Outperform":        strongBuyRatingScore,
	"Outperformer":      strongBuyRatingScore,
	"Sector Outperform": strongBuyRatingScore,
	"Market Outperform": strongBuyRatingScore,

	"Overweight":    buyRatingScore,
	"Equal Weight":  buyRatingScore,
	"Sector Weight": buyRatingScore,
	"Peer Perform":  buyRatingScore,
	"In-Line":       buyRatingScore,
	"Inline":        buyRatingScore,

	"Neutral":        holdRatingScore,
	"Market Perform": holdRatingScore,
	"Sector Perform": holdRatingScore,
	"Hold":           holdRatingScore,

	"Reduce":              sellRatingScore,
	"Negative":            sellRatingScore,
	"Underweight":         sellRatingScore,
	"Underperform":        sellRatingScore,
	"Sector Underperform": sellRatingScore,

	"Sell": strongSellRatingScore,
}

type serviceResponse[T any] struct {
	Data     []T    `json:"data"`
	NextPage string `json:"nextPage"`
}

type StockRatingService struct {
	isLoading             atomic.Bool
	stockRatingApi        entity.IStockRatingApi
	stockRatingRepository entity.IStockRatingRepository
}

func NewStockRatingService(stockRatingRepository entity.IStockRatingRepository, stockRatingApi entity.IStockRatingApi) *StockRatingService {
	return &StockRatingService{
		stockRatingApi:        stockRatingApi,
		stockRatingRepository: stockRatingRepository,
	}
}

func (s *StockRatingService) GetStockDetails(ctx context.Context, ticker string) *entity.StockDetails {
	return s.stockRatingApi.GetStockDetails(ctx, ticker)
}

func (s *StockRatingService) GetStockRatings(ctx context.Context, nextPage string, pageSize int, search string) (*serviceResponse[entity.StockRating], error) {
	pageSizePlusOne := pageSize + 1
	stockRatings, err := s.stockRatingRepository.GetStockRatings(ctx, nextPage, pageSizePlusOne, search)

	if err != nil {
		return nil, err
	}

	nNextPage := ""
	responseSize := len(stockRatings)
	existsMoreItems := responseSize == pageSizePlusOne

	if responseSize > 0 && existsMoreItems {
		lastItemCurrentPage := stockRatings[responseSize-2]
		nNextPage = lastItemCurrentPage.Ticker
		stockRatings = stockRatings[:responseSize-1]
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
		Data: recommendations,
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

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	ratingsChannel := make(chan entity.StockRating, channelBufferSize)

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rating := range ratingsChannel {
				select {
				case <-timeoutCtx.Done():
					return
				default:
					s.stockRatingRepository.Save(ctx, s.formatStockRating(rating))
				}
			}
		}()
	}

	nextPage := ""
	for {
		stockRatings, nNextPage, err := s.stockRatingApi.GetStockRatings(ctx, nextPage)
		if err != nil {
			errorMessage := "failed to get stock ratings from API"
			slog.Error(errorMessage, "error", err)
			break
		}

		for _, rating := range stockRatings {
			select {
			case <-ctx.Done():
				goto Cleanup
			case ratingsChannel <- rating:
			}
		}

		nextPage = nNextPage
		if nNextPage == "" {
			break
		}
	}

Cleanup:
	close(ratingsChannel)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Minute):
		slog.Warn("worker timeout exceeded during cleanup")
	}

	elapsed := time.Since(start)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	milliseconds := int(elapsed.Milliseconds()) % 1000
	duration := fmt.Sprintf("%dm %ds %dms", minutes, seconds, milliseconds)
	slog.Info("process to load stock ratings finished", "duration", duration)
}

func (s *StockRatingService) formatStockRating(rating entity.StockRating) entity.StockRating {
	reportDateScore := calculateDateScore(rating.Time)
	currentRatingScore := ratingScaleMap[rating.RatingTo]
	ratingChangeScore := calculateRatingChangeScore(rating)
	brokerageActionScore := calculateBrokerageActionScore(rating)
	targetPriceChange := calculateTargetPriceChange(rating)
	targetPriceChangeScore := calculateTargetPriceChangeScore(targetPriceChange)
	score := calculateScore(ratingChangeScore, currentRatingScore, brokerageActionScore, reportDateScore, targetPriceChangeScore)

	return entity.StockRating{
		Brokerage:         rating.Brokerage,
		Action:            rating.Action,
		Company:           rating.Company,
		Ticker:            rating.Ticker,
		RatingFrom:        rating.RatingFrom,
		RatingTo:          rating.RatingTo,
		TargetFrom:        rating.TargetFrom,
		TargetTo:          rating.TargetTo,
		Time:              rating.Time.Truncate(24 * time.Hour),
		TargetPriceChange: targetPriceChange,
		Score:             score,
	}
}

func calculateScore(ratingChangeScore, currentRatingScore, brokerageActionScore, reportDateScore, targetPriceChangeScore int) float32 {
	ratingValue := ratingChangeScore * ratingChangeWeight
	currentRatingValue := currentRatingScore * currentTargetWeight
	actionValue := brokerageActionScore * brokerageActionWeight
	reportDateValue := reportDateScore * reportDateWeight
	targetPriceValue := targetPriceChangeScore * targetPriceChangeWeight

	score := ratingValue + currentRatingValue + actionValue + reportDateValue + targetPriceValue
	return float32(score) / 100
}

func calculateTargetPriceChange(rating entity.StockRating) float64 {
	targetFrom, fromErr := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(rating.TargetFrom, ",", ""), "$"), 64)
	targetTo, toErr := strconv.ParseFloat(strings.TrimPrefix(strings.ReplaceAll(rating.TargetTo, ",", ""), "$"), 64)

	if fromErr != nil || toErr != nil || targetFrom == 0 {
		slog.Warn(
			"target price change calculation error",
			"fromError",
			fromErr,
			"toError",
			toErr,
			"targetFrom",
			targetFrom,
			"targetTo",
			targetTo,
			"ticker", rating.Ticker,
			"brokerage", rating.Brokerage,
			"time", rating.Time)
		return 0
	}

	return (targetTo - targetFrom) / targetFrom
}

func calculateTargetPriceChangeScore(priceTargetChange float64) int {
	switch {
	case priceTargetChange < 0:
		return 0
	case priceTargetChange >= 0.5:
		return 5
	case priceTargetChange >= 0.25:
		return 3
	default:
		return 1
	}
}

func calculateBrokerageActionScore(stockRating entity.StockRating) int {
	actionScore := actionsScaleMap[stockRating.Action]

	switch {
	case actionScore == 2:
		return calculateRatingChangeScore(stockRating)
	default:
		return actionScore
	}
}

func calculateRatingChangeScore(stockRating entity.StockRating) int {
	ratingFrom := ratingScaleMap[stockRating.RatingFrom]
	ratingTo := ratingScaleMap[stockRating.RatingTo]

	switch {
	case ratingFrom == ratingTo:
		return ratingTo
	case ratingFrom < ratingTo:
		return 5
	default:
		return 1
	}
}

func calculateDateScore(reportTime time.Time) int {
	daysSinceReport := time.Since(reportTime).Hours() / 24

	switch {
	case daysSinceReport <= 3:
		return 5
	case daysSinceReport <= 7:
		return 4
	case daysSinceReport <= 15:
		return 3
	case daysSinceReport <= 30:
		return 2
	default:
		return 1
	}
}
