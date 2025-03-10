package stock

import (
	"log/slog"
	"net/http"

	"github.com/rubenpad/stock-rating-system/internal/domain/service"
	"github.com/rubenpad/stock-rating-system/internal/infrastructure/server/middleware/pagination"
	"github.com/rubenpad/stock-rating-system/internal/infrastructure/server/middleware/search"

	"github.com/gin-gonic/gin"
)

type StockRatingController struct {
	stockRatingService *service.StockRatingService
}

func NewStockRatingController(stockRatingService *service.StockRatingService) *StockRatingController {
	return &StockRatingController{stockRatingService}
}

func (src *StockRatingController) GetStockDetails(ctx *gin.Context) {
	ticker := ctx.Param("ticker")

	stockDetails, err := src.stockRatingService.GetStockDetails(ctx, ticker)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": "error processing the request",
		})
		return
	}

	ctx.JSON(http.StatusOK, stockDetails)
}

func (src *StockRatingController) GetStockRatings(ctx *gin.Context) {
	search := ctx.GetString(search.SearchKey)
	nextPage := ctx.GetString(pagination.NextPageKey)
	pageSize := ctx.GetInt(pagination.PageSizeKey)

	stockRatings, err := src.stockRatingService.GetStockRatings(ctx, nextPage, pageSize, search)
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
	pageSize := ctx.GetInt(pagination.PageSizeKey)

	stockRecommendations, err := src.stockRatingService.GetStockRecommendations(ctx, pageSize)
	if err != nil {
		slog.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": "error processing the request",
		})
		return
	}

	ctx.JSON(http.StatusOK, stockRecommendations)
}

func (src *StockRatingController) LoadStockRatingData(ctx *gin.Context) {
	go src.stockRatingService.LoadStockRatingsData(ctx)

	ctx.JSON(http.StatusAccepted, gin.H{})
}
