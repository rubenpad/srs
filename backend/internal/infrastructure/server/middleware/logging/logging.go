package logging

import (
	"log/slog"
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

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		slog.Info("response",
			slog.Duration("timestamp", time.Duration(timestamp.Unix())),
			slog.Int("statusCode", statusCode),
			slog.Duration("latency", time.Duration(latency.Milliseconds())),
			slog.String("clientIP", clientIP),
			slog.String("method", method),
			slog.String("path", path),
		)
	}
}
