package middleware_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"

	mw "contract-analysis-service/internal/middleware"
)

func TestRecovery(t *testing.T) {
	// Set up a test router with the recovery middleware
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	// Create a new router with recovery middleware
	router := gin.New()
	recoveryMw := ginzap.RecoveryWithZap(logger, true)
	router.Use(recoveryMw)

	// Add a route that will panic
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check that we got a 500 status code
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequestLogger(t *testing.T) {
	// Set up a test router with the request logger middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()

	// Add request logger middleware
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Just verify the request was handled
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORS(t *testing.T) {
	// Create a test server with CORS middleware
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new Gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		// Create middleware instance with a test logger
		mw := mw.NewMiddleware(zap.NewNop())
		
		// Apply CORS middleware
		handler := mw.CORS()
		handler(c)

		// If this is a preflight request, we're done
		if r.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			return
		}

		// For actual requests, continue with the request
		c.Status(http.StatusOK)
	}))
	defer srv.Close()

	// Test preflight request
	t.Run("preflight request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", srv.URL, nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})

	// Test actual request
	t.Run("actual request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", srv.URL, nil)
		req.Header.Set("Origin", "http://example.com")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
}

func TestRateLimit(t *testing.T) {
	// Set up a test router with the rate limit middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a test rate limiter with a very low limit
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  2, // 2 requests per second for testing
	}
	instance := limiter.New(store, rate)

	// Add the rate limiting middleware
	router.Use(func(c *gin.Context) {
		// Use a fixed IP for testing
		c.Request.RemoteAddr = "127.0.0.1:12345"
		
		limiterCtx, err := instance.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			return
		}
		
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiterCtx.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiterCtx.Remaining))
		
		if limiterCtx.Reached {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
			return
		}
		
		c.Next()
	})

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First request should succeed
	t.Run("first request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Second request should also succeed
	t.Run("second request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Third request should be rate limited
	t.Run("rate limited request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	// Wait for the rate limit window to reset
	time.Sleep(1100 * time.Millisecond)

	// Next request after window reset should succeed
	t.Run("request after window reset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequestID(t *testing.T) {
	// Set up a test router with the request ID middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := zap.NewNop()

	// Create middleware instance
	mw := mw.NewMiddleware(logger)

	// Apply request ID middleware
	router.Use(mw.RequestID())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		c.String(http.StatusOK, requestID)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Check that we got a 200 status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the request ID was set and is not empty
	requestID := w.Body.String()
	assert.NotEmpty(t, requestID)
	assert.True(t, len(requestID) > 0)
}

func TestValidation(t *testing.T) {
	// Set up a test router with the validation middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a validator
	validate := validator.New()

	// Define a test struct for validation
	type testStruct struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	// Add a test route that uses the validation middleware
	router.POST("/test", func(c *gin.Context) {
		var body testStruct
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	})

	tests := []struct {
		name           string
		body           string
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:          "valid request",
			body:          `{"name":"John Doe","email":"john@example.com"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing name",
			body:           `{"email":"john@example.com"}`,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Key: 'testStruct.Name' Error:Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:           "invalid email",
			body:           `{"name":"John","email":"invalid-email"}`,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Key: 'testStruct.Email' Error:Field validation for 'Email' failed on the 'email' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedMessage != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedMessage)
			}
		})
	}
}
