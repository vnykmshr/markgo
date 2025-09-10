package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Performance logs request timing and basic metrics
func Performance(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log slow requests (over 1 second)
		if duration > time.Second {
			logger.Warn("Slow request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"duration", duration,
				"status", c.Writer.Status(),
			)
		}

		// Add timing header
		c.Header("X-Response-Time", duration.String())
	}
}
