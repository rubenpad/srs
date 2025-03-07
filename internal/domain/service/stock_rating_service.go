package service

import (
	"context"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockRatingService struct {
	repository entity.IStockRatingRepository
}

func NewStockRatingService(stockRatingRepository entity.IStockRatingRepository) StockRatingService {
	return StockRatingService{repository: stockRatingRepository}
}

func (s StockRatingService) GetStockRatings(ctx context.Context, limit int, offset int) ([]entity.StockRating, error) {
	stockRatings, err := s.repository.GetStockRatings(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return stockRatings, nil
}
