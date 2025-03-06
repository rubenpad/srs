package stock

import (
	"net/http"

	"github.com/rubenpad/stock-rating-system/internal/domain/service"

	"github.com/gin-gonic/gin"
)

type StockController struct {
	stockService service.StockService
}

func NewStockController(stockService service.StockService) *StockController {
	return &StockController{stockService}
}

func (sc *StockController) GetStocks(context *gin.Context) {
	stocks, err := sc.stockService.GetStocks(context, 1, 0)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_server_error",
			"message": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, stocks)
}
