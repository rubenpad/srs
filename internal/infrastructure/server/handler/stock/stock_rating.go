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
			"message": "error processing the request",
		})
		return
	}

	context.Header("Cache-Control", "private, max-age=86400")
	context.JSON(http.StatusOK, stockRatings)
}

func (src *StockRatingController) GetStockRecommendations(context *gin.Context) {
	stockRecommendations, err := src.stockRatingService.GetStockRecommendations(context, 1)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": "error processing the request",
		})
		return
	}

	context.Header("Cache-Control", "private, max-age=900")
	context.JSON(http.StatusOK, stockRecommendations)
}

func (src *StockRatingController) LoadStockRatingData(ctx *gin.Context) {
	go src.stockRatingService.LoadStockRatingsData(ctx)

	ctx.JSON(http.StatusAccepted, gin.H{})
}
