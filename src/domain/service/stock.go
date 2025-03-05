package service

import "github.com/rubenpad/stock-rating-system/domain/ports"

type Stock struct {
	repository ports.IStockRepository
}

func NewStockService(stockRepository ports.IStockRepository) *Stock {
	return &Stock{repository: stockRepository}
}
