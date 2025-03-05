package server

import (
	"github.com/rubenpad/stock-rating-system/internal/domain/ports"

	"github.com/gin-gonic/gin"
)

type StockController struct {
	stockService ports.IStockService
}

func NewStockController(stockService ports.IStockService) *StockController {
	return &StockController{}
}

func (sc *StockController) GetStocks(context *gin.Context) {
	context.JSON(200, gin.H{})
}
