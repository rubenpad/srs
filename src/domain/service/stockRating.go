package service

import (
	"github.com/rubenpad/stock-rating-system/domain/ports"
)

type StockRatinService struct {
	repository ports.IStockRatingRepository
}

func NewStockRatingService(stockRatingRepository ports.IStockRatingRepository) *StockRatinService {
	return &StockRatinService{repository: stockRatingRepository}
}
