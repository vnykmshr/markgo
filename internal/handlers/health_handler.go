package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vnykmshr/markgo/internal/constants"
)

// HealthHandler handles health check and metrics requests.
type HealthHandler struct {
	*BaseHandler
	startTime time.Time
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(base *BaseHandler, startTime time.Time) *HealthHandler {
	return &HealthHandler{
		BaseHandler: base,
		startTime:   startTime,
	}
}

// Health handles health check requests.
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	health := map[string]any{
		"status":      "healthy",
		"timestamp":   time.Now().Unix(),
		"uptime":      uptime.String(),
		"version":     constants.AppVersion,
		"environment": h.config.Environment,
		"services": map[string]any{
			"articles": "healthy",
			"cache":    "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}
