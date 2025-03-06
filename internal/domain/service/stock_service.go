package service

import (
	"context"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

type StockService struct {
	repository entity.IStockRepository
}

func NewStockService(stockRepository entity.IStockRepository) StockService {
	return StockService{repository: stockRepository}
}

func (s StockService) CreateStock(ctx context.Context, ticker string, company string, score int) error {
	if err := s.repository.Save(ctx, entity.NewStock(ticker, company, score)); err != nil {
		return err
	}

	return nil
}

func (s StockService) GetStocks(ctx context.Context, limit int, offset int) ([]entity.Stock, error) {
	stocks, err := s.repository.GetStocks(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return stocks, nil
}
