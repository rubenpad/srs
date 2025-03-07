package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware is a gin.HandlerFunc that logs some information
// of the incoming request and the consequent response.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		// Request URL path
		path := c.Request.URL.Path

		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// Process request
		c.Next()

		// Results
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		logger.Info("response",
			slog.Duration("timestamp", time.Duration(timestamp.Unix())),
			slog.Int("statusCode", statusCode),
			slog.Duration("latency", latency),
			slog.String("clientIP", clientIP),
			slog.String("method", method),
			slog.String("path", path),
		)
	}
}
