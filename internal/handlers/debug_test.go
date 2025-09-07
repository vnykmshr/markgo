package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDebugEndpoints(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)
	handlers, _, _, _ := createTestHandlers()

	// Override config for development mode
	handlers.config.Environment = "development"

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		handler        gin.HandlerFunc
	}{
		{
			name:           "Debug Memory",
			method:         "GET",
			path:           "/debug/memory",
			expectedStatus: http.StatusOK,
			handler:        handlers.DebugMemory,
		},
		{
			name:           "Debug Runtime",
			method:         "GET",
			path:           "/debug/runtime",
			expectedStatus: http.StatusOK,
			handler:        handlers.DebugRuntime,
		},
		{
			name:           "Debug Config",
			method:         "GET",
			path:           "/debug/config",
			expectedStatus: http.StatusOK,
			handler:        handlers.DebugConfig,
		},
		{
			name:           "Debug Requests",
			method:         "GET",
			path:           "/debug/requests",
			expectedStatus: http.StatusOK,
			handler:        handlers.DebugRequests,
		},
		{
			name:           "Set Log Level - Valid",
			method:         "POST",
			path:           "/debug/log-level",
			body:           `{"level": "debug"}`,
			expectedStatus: http.StatusOK,
			handler:        handlers.SetLogLevel,
		},
		{
			name:           "Set Log Level - Invalid",
			method:         "POST",
			path:           "/debug/log-level",
			body:           `{"level": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
			handler:        handlers.SetLogLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Handle(tt.method, tt.path, tt.handler)

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// For successful requests, verify JSON response structure
			if recorder.Code == http.StatusOK {
				bodyBytes := recorder.Body.Bytes()
				if len(bodyBytes) == 0 {
					t.Logf("%s returned empty response", tt.name)
					return
				}

				var response map[string]interface{}
				err := json.Unmarshal(bodyBytes, &response)
				if err != nil {
					t.Logf("%s response body: %s", tt.name, string(bodyBytes))
					assert.NoError(t, err)
					return
				}

				// Different endpoints have different response structures
				switch tt.name {
				case "Debug Memory":
					// Memory debug has timestamp in nested current object
					if current, ok := response["current"].(map[string]interface{}); ok {
						assert.Contains(t, current, "timestamp")
					}
				case "Debug Requests":
					// Debug Requests should have current_request and performance data
					assert.Contains(t, response, "current_request")
					assert.Contains(t, response, "performance")
					assert.Contains(t, response, "timestamp")
				default:
					// Other endpoints should have timestamp at root level
					assert.Contains(t, response, "timestamp")
				}
			}
		})
	}
}

func TestDebugEndpointsProductionMode(t *testing.T) {
	// Test that debug endpoints are restricted in production
	handlers, _, _, _ := createTestHandlers()
	handlers.config.Environment = "production"

	restrictedEndpoints := []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{"Debug Requests - Production", "GET", "/debug/requests", handlers.DebugRequests},
		{"Set Log Level - Production", "POST", "/debug/log-level", handlers.SetLogLevel},
		{"Pprof Index - Production", "GET", "/debug/pprof/", handlers.PprofIndex},
		{"Pprof Profile - Production", "GET", "/debug/pprof/profile", handlers.PprofProfile},
	}

	for _, tt := range restrictedEndpoints {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Handle(tt.method, tt.path, tt.handler)

			var req *http.Request
			if tt.method == "POST" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(`{"level": "debug"}`))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusForbidden, recorder.Code)

			var response map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "development")
		})
	}
}

func TestDebugConfigResponseStructure(t *testing.T) {
	handlers, _, _, _ := createTestHandlers()
	handlers.config.Environment = "development"

	router := gin.New()
	router.GET("/debug/config", handlers.DebugConfig)

	req := httptest.NewRequest("GET", "/debug/config", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify expected configuration fields are present
	expectedFields := []string{
		"environment", "port", "log_level", "cache_ttl",
		"posts_per_page", "rate_limit", "timestamp",
	}

	for _, field := range expectedFields {
		assert.Contains(t, response, field, "Response should contain %s field", field)
	}

	// Verify sensitive fields are NOT present
	sensitiveFields := []string{
		"password", "secret", "key", "token", "username",
	}

	for _, field := range sensitiveFields {
		assert.NotContains(t, response, field, "Response should not contain sensitive field %s", field)
	}

	// Verify rate limit structure
	rateLimit, ok := response["rate_limit"].(map[string]interface{})
	assert.True(t, ok, "rate_limit should be a map")
	assert.Contains(t, rateLimit, "general_requests")
	assert.Contains(t, rateLimit, "general_window")
	assert.Contains(t, rateLimit, "contact_requests")
	assert.Contains(t, rateLimit, "contact_window")
}
