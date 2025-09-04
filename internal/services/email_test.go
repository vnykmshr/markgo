package services

import (
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/models"
)

func TestNewEmailService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name   string
		config config.EmailConfig
	}{
		{
			name: "Complete configuration",
			config: config.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				To:       "recipient@gmail.com",
				UseSSL:   true,
			},
		},
		{
			name: "Configuration without credentials",
			config: config.EmailConfig{
				Host: "smtp.gmail.com",
				Port: 587,
				From: "test@gmail.com",
				To:   "recipient@gmail.com",
			},
		},
		{
			name:   "Empty configuration",
			config: config.EmailConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			assert.NotNil(t, service)
			assert.Equal(t, tt.config, service.config)
			assert.Equal(t, logger, service.logger)

			if tt.config.Username != "" && tt.config.Password != "" {
				assert.NotNil(t, service.auth)
			} else {
				assert.Nil(t, service.auth)
			}
		})
	}
}

func TestEmailService_SendContactMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name        string
		config      config.EmailConfig
		message     *models.ContactMessage
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration and message",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       "recipient@example.com",
			},
			message: &models.ContactMessage{
				Name:    "John Doe",
				Email:   "john@example.com",
				Subject: "Test Subject",
				Message: "This is a test message",
			},
			expectError: true, // Will fail because we can't actually send emails in tests
		},
		{
			name: "No credentials configured",
			config: config.EmailConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "test@example.com",
				To:   "recipient@example.com",
			},
			message: &models.ContactMessage{
				Name:    "John Doe",
				Email:   "john@example.com",
				Subject: "Test Subject",
				Message: "This is a test message",
			},
			expectError: true,
			errorMsg:    "email service not configured",
		},
		{
			name: "Empty credentials",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "",
				Password: "",
				From:     "test@example.com",
				To:       "recipient@example.com",
			},
			message: &models.ContactMessage{
				Name:    "John Doe",
				Email:   "john@example.com",
				Subject: "Test Subject",
				Message: "This is a test message",
			},
			expectError: true,
			errorMsg:    "email service not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			err := service.SendContactMessage(tt.message)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_SendNotification(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name        string
		config      config.EmailConfig
		to          string
		subject     string
		body        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid notification",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       "recipient@example.com",
			},
			to:          "recipient@example.com",
			subject:     "Test Notification",
			body:        "This is a test notification",
			expectError: true, // Will fail because we can't actually send emails in tests
		},
		{
			name: "No credentials configured",
			config: config.EmailConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "test@example.com",
				To:   "recipient@example.com",
			},
			to:          "recipient@example.com",
			subject:     "Test Notification",
			body:        "This is a test notification",
			expectError: true,
			errorMsg:    "email service not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			err := service.SendNotification(tt.to, tt.subject, tt.body)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_BuildEmailMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{
		Host: "smtp.example.com",
		Port: 587,
		From: "test@example.com",
		To:   "recipient@example.com",
	}
	service := NewEmailService(config, logger)

	from := "sender@example.com"
	to := "recipient@example.com"
	subject := "Test Subject"
	body := "<h1>Test Body</h1><p>This is a test email.</p>"

	msg := service.buildEmailMessage(from, to, subject, body)

	// Check message components
	assert.Contains(t, msg, "From: "+from)
	assert.Contains(t, msg, "To: "+to)
	assert.Contains(t, msg, "Subject: "+subject)
	assert.Contains(t, msg, "MIME-Version: 1.0")
	assert.Contains(t, msg, "Content-Type: text/html; charset=UTF-8")
	assert.Contains(t, msg, "Date:")
	assert.Contains(t, msg, body)

	// Check proper line endings
	assert.Contains(t, msg, "\r\n")
}

func TestEmailService_GenerateContactEmailBody(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{}
	service := NewEmailService(config, logger)

	message := &models.ContactMessage{
		Name:    "John Doe",
		Email:   "john@example.com",
		Subject: "Test Subject",
		Message: "This is a test message with\nmultiple lines\nof content.",
	}

	body, err := service.generateContactEmailBody(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Check that all message fields are included
	assert.Contains(t, body, message.Name)
	assert.Contains(t, body, message.Email)
	assert.Contains(t, body, message.Subject)
	assert.Contains(t, body, message.Message)

	// Check HTML structure
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "<html>")
	assert.Contains(t, body, "</html>")
	assert.Contains(t, body, "New Contact Form Submission")

	// Check CSS styling is included
	assert.Contains(t, body, "<style>")
	assert.Contains(t, body, "</style>")

	// Check timestamp is included
	assert.Contains(t, body, "Received:")
}

func TestEmailService_TestConnection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name        string
		config      config.EmailConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "No credentials",
			config: config.EmailConfig{
				Host: "smtp.example.com",
				Port: 587,
			},
			expectError: true,
			errorMsg:    "email credentials not configured",
		},
		{
			name: "Empty username",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "",
				Password: "password",
			},
			expectError: true,
			errorMsg:    "email credentials not configured",
		},
		{
			name: "Empty password",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "",
			},
			expectError: true,
			errorMsg:    "email credentials not configured",
		},
		{
			name: "Valid credentials but invalid server",
			config: config.EmailConfig{
				Host:     "invalid.smtp.server",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
			},
			expectError: true, // Will fail to connect to invalid server
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			err := service.TestConnection()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_SendTestEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name        string
		config      config.EmailConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: config.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password",
				From:     "test@example.com",
				To:       "recipient@example.com",
			},
			expectError: true, // Will fail because we can't actually send emails in tests
		},
		{
			name: "No credentials",
			config: config.EmailConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "test@example.com",
				To:   "recipient@example.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			err := service.SendTestEmail()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_GenerateTestEmailBody(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       "recipient@example.com",
	}
	service := NewEmailService(config, logger)

	body := service.generateTestEmailBody()
	assert.NotEmpty(t, body)

	// Check HTML structure
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "<html>")
	assert.Contains(t, body, "</html>")

	// Check content
	assert.Contains(t, body, "Email Service Test")
	assert.Contains(t, body, "This is a test email")
	assert.Contains(t, body, config.From)
	assert.Contains(t, body, config.Host)
	assert.Contains(t, body, "587") // Port as string

	// Check timestamp is included
	now := time.Now()
	year := now.Format("2006")
	assert.Contains(t, body, year)
}

func TestEmailService_ValidateConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		config         config.EmailConfig
		expectedErrors []string
	}{
		{
			name: "Valid configuration",
			config: config.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				To:       "recipient@gmail.com",
			},
			expectedErrors: []string{},
		},
		{
			name:   "Empty configuration",
			config: config.EmailConfig{},
			expectedErrors: []string{
				"SMTP host is required",
				"SMTP port is required",
				"SMTP username is required",
				"SMTP password is required",
				"From email address is required",
				"To email address is required",
			},
		},
		{
			name: "Invalid email addresses",
			config: config.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "invalid-email",
				To:       "another-invalid-email",
			},
			expectedErrors: []string{
				"From email address is invalid",
				"To email address is invalid",
			},
		},
		{
			name: "Missing host",
			config: config.EmailConfig{
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				To:       "recipient@gmail.com",
			},
			expectedErrors: []string{
				"SMTP host is required",
			},
		},
		{
			name: "Missing port",
			config: config.EmailConfig{
				Host:     "smtp.gmail.com",
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				To:       "recipient@gmail.com",
			},
			expectedErrors: []string{
				"SMTP port is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			errors := service.ValidateConfig()

			assert.Equal(t, len(tt.expectedErrors), len(errors))

			for _, expectedError := range tt.expectedErrors {
				found := slices.Contains(errors, expectedError)
				assert.True(t, found, "Expected error not found: %s", expectedError)
			}
		})
	}
}

func TestEmailService_IsValidEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{}
	service := NewEmailService(config, logger)

	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user@domain.org", true},
		{"name.surname@company.co.uk", true},
		{"invalid-email", false},
		{"@domain.com", true}, // Basic validation only checks for @ and .
		{"user@", false},
		{"", false},
		{"user@domain", false},     // No TLD
		{"user.domain.com", false}, // No @
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := service.isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailService_GetConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name   string
		config config.EmailConfig
	}{
		{
			name: "Complete configuration",
			config: config.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				To:       "recipient@gmail.com",
				UseSSL:   true,
			},
		},
		{
			name: "Configuration without credentials",
			config: config.EmailConfig{
				Host:   "smtp.gmail.com",
				Port:   587,
				From:   "test@gmail.com",
				To:     "recipient@gmail.com",
				UseSSL: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, logger)
			configMap := service.GetConfig()

			// Check that all expected fields are present
			assert.Equal(t, tt.config.Host, configMap["host"])
			assert.Equal(t, tt.config.Port, configMap["port"])
			assert.Equal(t, tt.config.From, configMap["from"])
			assert.Equal(t, tt.config.To, configMap["to"])
			assert.Equal(t, tt.config.UseSSL, configMap["use_ssl"])

			// Check authentication status
			hasAuth := tt.config.Username != "" && tt.config.Password != ""
			assert.Equal(t, hasAuth, configMap["has_auth"])

			// Ensure sensitive data is not exposed
			_, hasUsername := configMap["username"]
			_, hasPassword := configMap["password"]
			assert.False(t, hasUsername, "Username should not be exposed")
			assert.False(t, hasPassword, "Password should not be exposed")
		})
	}
}

func TestEmailService_SendEmail_InvalidSMTPConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{
		Host:     "invalid.smtp.server.that.does.not.exist",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		To:       "recipient@example.com",
	}
	service := NewEmailService(config, logger)

	err := service.sendEmail("recipient@example.com", "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "dial")
}

func TestEmailService_InterfaceCompliance(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{}

	// This test ensures EmailService implements EmailServiceInterface
	var _ EmailServiceInterface = NewEmailService(config, logger)
}

func BenchmarkEmailService_BuildEmailMessage(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{}
	service := NewEmailService(config, logger)

	from := "sender@example.com"
	to := "recipient@example.com"
	subject := "Benchmark Test Subject"
	body := "<h1>Benchmark Test</h1><p>This is a benchmark test email body.</p>"

	for b.Loop() {
		msg := service.buildEmailMessage(from, to, subject, body)
		_ = msg
	}
}

func BenchmarkEmailService_GenerateContactEmailBody(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{}
	service := NewEmailService(config, logger)

	message := &models.ContactMessage{
		Name:    "Benchmark User",
		Email:   "benchmark@example.com",
		Subject: "Benchmark Test",
		Message: "This is a benchmark test message with some content to test performance.",
	}

	for b.Loop() {
		body, err := service.generateContactEmailBody(message)
		if err != nil {
			b.Fatal(err)
		}
		_ = body
	}
}

func BenchmarkEmailService_ValidateConfig(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.EmailConfig{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "test@gmail.com",
		Password: "password",
		From:     "test@gmail.com",
		To:       "recipient@gmail.com",
	}
	service := NewEmailService(config, logger)

	for b.Loop() {
		errors := service.ValidateConfig()
		_ = errors
	}
}
