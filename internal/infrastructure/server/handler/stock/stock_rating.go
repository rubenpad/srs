package stock

import (
	"log/slog"
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

func (src *StockRatingController) GetStockRatings(ctx *gin.Context) {
	page := ctx.GetInt("page")
	pageSize := ctx.GetInt("pageSize")

	stockRatings, err := src.stockRatingService.GetStockRatings(ctx, page, pageSize)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": "error processing the request",
		})
		return
	}

	ctx.Header("Cache-Control", "private, max-age=900")
	ctx.JSON(http.StatusOK, stockRatings)
}

func (src *StockRatingController) GetStockRecommendations(ctx *gin.Context) {
	stockRecommendations, err := src.stockRatingService.GetStockRecommendations(ctx, ctx.GetInt("pageSize"))
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": "error processing the request",
		})
		return
	}

	ctx.Header("Cache-Control", "private, max-age=86400")
	ctx.JSON(http.StatusOK, stockRecommendations)
}

func (src *StockRatingController) LoadStockRatingData(ctx *gin.Context) {
	go src.stockRatingService.LoadStockRatingsData(ctx)

	ctx.JSON(http.StatusAccepted, gin.H{})
}
