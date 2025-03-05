package server

import (
	"github.com/rubenpad/stock-rating-system/internal/domain/ports"

	"github.com/gin-gonic/gin"
)

type StockRatingController struct {
	stockRatingService ports.IStockRatingService
}

func NewStockRatingController(stockRatingService ports.IStockRatingService) *StockRatingController {
	return &StockRatingController{stockRatingService}
}

func (src *StockRatingController) GetStockRatings(context *gin.Context) {
	context.JSON(200, gin.H{})
}
