package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/1mb-dev/markgo/internal/errors"
	"github.com/1mb-dev/markgo/internal/models"
	"github.com/1mb-dev/markgo/internal/services"
)

// ContactHandler handles contact form display and submission.
type ContactHandler struct {
	*BaseHandler
	emailService services.EmailServiceInterface
}

// NewContactHandler creates a new contact handler.
func NewContactHandler(base *BaseHandler, emailService services.EmailServiceInterface) *ContactHandler {
	return &ContactHandler{
		BaseHandler:  base,
		emailService: emailService,
	}
}

// Submit handles contact form submissions.
func (h *ContactHandler) Submit(c *gin.Context) {
	var form struct {
		Name    string `json:"name" binding:"required"`
		Email   string `json:"email" binding:"required,email"`
		Subject string `json:"subject" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	contactMsg := &models.ContactMessage{
		Name:    form.Name,
		Email:   form.Email,
		Subject: form.Subject,
		Message: form.Message,
	}

	if err := h.emailService.SendContactMessage(contactMsg); err != nil {
		if errors.Is(err, apperrors.ErrEmailNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Contact form temporarily unavailable",
				"message": "Email service is not configured. Please try again later or contact us through alternative means.",
				"status":  "unavailable",
			})
			return
		}

		h.handleError(c, err, "Failed to send contact message")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact message sent successfully",
		"status":  "success",
	})
}
