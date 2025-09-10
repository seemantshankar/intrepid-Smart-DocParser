package ocr_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"contract-analysis-service/configs"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/pkg/metrics"
	"contract-analysis-service/internal/services/ocr"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Find project root directory
	projectRoot, err := findProjectRoot()
	if err == nil {
		// Load environment variables from .env file
		_ = godotenv.Load(projectRoot + "/.env")

		// Copy OPENROUTER_API_KEY to OPENROUTER_API_KEY_TEST if it exists
		// This ensures the test config can use the same API key
		if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
			os.Setenv("OPENROUTER_API_KEY_TEST", apiKey)
		}
	}

	// Run tests
	os.Exit(m.Run())
}

// findProjectRoot attempts to find the project root by looking for go.mod file
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root")
}

// maskAPIKey masks an API key for safe logging
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "[too short to mask]"
	}

	// Show first 4 and last 4 characters, mask the rest
	return key[:4] + "..." + key[len(key)-4:]
}

// debuggingClient wraps an external.Client to log request/response details
type debuggingClient struct {
	baseClient external.Client
	t          *testing.T
}

// ExecuteRequest implements external.Client interface
func (d *debuggingClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	// Log request details
	d.t.Logf("Request URL: %s", req.URL)
	d.t.Logf("Request Method: %s", req.Method)

	// Log request headers
	d.t.Log("Request Headers:")
	for k, v := range req.Headers {
		if k == "Authorization" {
			d.t.Logf("  %s: Bearer %s", k, maskAPIKey(v[7:])) // Skip "Bearer " prefix
		} else {
			d.t.Logf("  %s: %s", k, v)
		}
	}

	// Log request body
	d.t.Logf("Request Body (first 1000 chars): %s", truncateString(string(req.Body), 1000))

	// Execute the request
	resp, err := d.baseClient.ExecuteRequest(ctx, req)

	// Log response or error
	if err != nil {
		d.t.Logf("Request failed with error: %v", err)
		return nil, err
	}

	// Log response details
	d.t.Logf("Response Status: %d", resp.StatusCode)
	d.t.Logf("Response Body (first 1000 chars): %s", truncateString(string(resp.Body), 1000))

	return resp, nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func TestOCRService_Integration_ExtractTextFromImage(t *testing.T) {
	if os.Getenv("RUN_OCR_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set RUN_OCR_INTEGRATION_TESTS to run")
	}

	// Check for OCR_TEST_IMAGE_URL
	imageURL := os.Getenv("OCR_TEST_IMAGE_URL")
	if imageURL == "" {
		imageURL = "https://www.mattmahoney.net/ocr/stock_gs200.jpg"
		t.Log("Using default test image URL: https://www.mattmahoney.net/ocr/stock_gs200.jpg")
	}

	// Find project root and load config
	projectRoot, err := findProjectRoot()
	require.NoError(t, err, "Failed to find project root")

	cfg, err := configs.LoadConfig(filepath.Join(projectRoot, "config_test.yaml"))
	require.NoError(t, err, "Failed to load test config")

	// Check if API key is loaded from environment variable
	apiKeyFromEnv := os.Getenv("OPENROUTER_API_KEY")
	t.Logf("API Key from env: %s", maskAPIKey(apiKeyFromEnv))
	t.Logf("API Key from config: %s", maskAPIKey(cfg.OCR.APIKey))

	// Override config with environment variable if it exists
	if apiKeyFromEnv != "" {
		cfg.OCR.APIKey = apiKeyFromEnv
		t.Log("Using API key from environment variable")
	}

	// Create a debugging client wrapper
	debugClient := &debuggingClient{
		baseClient: external.NewHTTPClient(
			cfg.LLM.OpenRouter.BaseURL,
			"OCR_IntegrationTest",
			external.RetryConfig{
				MaxRetries:      cfg.LLM.OpenRouter.RetryCount,
				InitialInterval: cfg.LLM.OpenRouter.RetryWaitTime,
				MaxInterval:     cfg.LLM.OpenRouter.RetryMaxInterval,
			},
			cfg.LLM.OpenRouter.Timeout,
		),
		t: t,
	}
	validator := ocr.NewValidator()
	ocrMetrics := metrics.NewOCRMetrics()

	// Create service
	service := ocr.NewOCRService(debugClient, cfg.OCR.APIKey, cfg.OCR.FallbackModels, validator, ocrMetrics)

	// Run test
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := service.ExtractTextFromImage(ctx, imageURL)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Text)
	assert.Greater(t, result.Confidence, 0.0)

	t.Logf("Extracted Text: %s", result.Text)
	t.Logf("Confidence: %.2f", result.Confidence)
}
