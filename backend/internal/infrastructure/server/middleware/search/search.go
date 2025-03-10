package search

import (
	"github.com/gin-gonic/gin"
)

const SearchKey = "search"

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		search := ctx.DefaultQuery(SearchKey, "")
		ctx.Set(SearchKey, search)
		ctx.Next()
	}
}
