package stock

import (
	"net/http"

	"github.com/rubenpad/stock-rating-system/internal/domain/service"

	"github.com/gin-gonic/gin"
)

type StockRatingController struct {
	stockRatingService service.StockRatingService
}

func NewStockRatingController(stockRatingService service.StockRatingService) *StockRatingController {
	return &StockRatingController{stockRatingService}
}

func (src *StockRatingController) GetStockRatings(context *gin.Context) {
	stockRatings, err := src.stockRatingService.GetStockRatings(context, 1, 0)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, stockRatings)
}
