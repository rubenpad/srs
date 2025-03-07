package entity

import "context"

type Stock struct {
	Ticker  string
	Company string
	Score   int
}

type IStockRepository interface {
	Save(ctx context.Context, stock Stock) error
	GetStock(ctx context.Context, ticker string) (*Stock, error)
	GetStocks(ctx context.Context, limit int, offset int) ([]Stock, error)
}

func NewStock(ticker string, company string, score int) Stock {
	return Stock{ticker, company, score}
}
