package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckHandler returns an HTTP handler to perform health checks.
func HealthCheck(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Server up and running!")
}
