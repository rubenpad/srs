package pagination

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const minPageSize = 1
const maxPageSize = 100
const NextPageKey = "nextPage"
const PageSizeKey = "pageSize"

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nextPage := ctx.DefaultQuery(NextPageKey, "")
		pageSizeStrValue := ctx.DefaultQuery(PageSizeKey, "10")

		pageSizeValue, err := strconv.Atoi(pageSizeStrValue)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s parameter must be an integer", PageSizeKey)})
			return
		}

		if err := validatePageSize(pageSizeValue); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.Set(NextPageKey, nextPage)
		ctx.Set(PageSizeKey, pageSizeValue)

		ctx.Next()
	}
}

func validatePageSize(size int) error {
	if size < minPageSize || size > maxPageSize {
		return fmt.Errorf("%s must be between %d and %d", PageSizeKey, minPageSize, maxPageSize)
	}

	return nil
}
