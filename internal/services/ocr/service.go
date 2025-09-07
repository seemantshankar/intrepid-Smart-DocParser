package ocr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/pkg/metrics"
)

// Service defines the interface for the OCR service.
//go:generate mockery --name=Service --output=./mocks --filename=service_mock.go
type Service interface {
	ExtractTextFromImage(ctx context.Context, imagePath string) (*OCRResult, error)
}

// OCRResult represents the outcome of an OCR operation.
type OCRResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// ocrService handles document text extraction
type ocrService struct {
	client         external.Client
	apiKey         string
	model          string
	fallbackModels []string
	apiURL         string
	validator      Validator
	metrics        *metrics.OCRMetrics
}

// NewOCRService creates a new OCR service instance
func NewOCRService(client external.Client, apiKey string, fallbackModels []string, validator Validator, metrics *metrics.OCRMetrics) Service {
	return &ocrService{
		client:         client,
		apiKey:         apiKey,
		model:          "qwen/qwen2.5-vl-32b-instruct:free",
		fallbackModels: fallbackModels,
		// Use a relative path; the external HTTP client will prepend BaseURL
		apiURL:         "/chat/completions",
		validator:      validator,
		metrics:        metrics,
	}
}

// ExtractTextFromImage extracts text from images, trying the primary model first, then any configured fallback models.
func (s *ocrService) ExtractTextFromImage(ctx context.Context, imagePath string) (*OCRResult, error) {
	var imageURL string
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		imageURL = imagePath
	} else {
		// Read and encode image once
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image: %w", err)
		}
		base64Image := base64.StdEncoding.EncodeToString(imageData)
		imageURL = fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)
	}

	// Try primary model first
	modelsToTry := append([]string{s.model}, s.fallbackModels...)
	var lastErr error

	for _, model := range modelsToTry {
		result, err := s.tryExtractWithModel(ctx, model, imageURL)
		if err == nil {
			if validationErr := s.validator.Validate(result); validationErr == nil {
				return result, nil // Success
			}
		}
		lastErr = err // Store the last error
	}

	return nil, fmt.Errorf("all OCR models failed: %w", lastErr)
}

// tryExtractWithModel attempts to extract text using a specific model.
func (s *ocrService) tryExtractWithModel(ctx context.Context, model, imageURL string) (*OCRResult, error) {
	s.metrics.RequestsTotal.WithLabelValues(model).Inc()
	startTime := time.Now()
	defer func() {
		s.metrics.RequestDuration.WithLabelValues(model).Observe(time.Since(startTime).Seconds())
	}()

	// Prepare payload
	payload := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "Extract only the stock names that are in bold from this image. List them in a comma-separated format. Respond with a JSON object containing two keys: 'text' for the extracted stock names, and 'confidence' (a float between 0.0 and 1.0) for your confidence in the extraction accuracy."},
					map[string]interface{}{"type": "image_url", "image_url": map[string]string{"url": imageURL}},
				},
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"stream":          true,
	}

	// Marshal JSON payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Execute request
	resp, err := s.client.ExecuteRequest(ctx, &external.Request{
		Method:      "POST",
		URL:         s.apiURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + s.apiKey,
			"Content-Type":  "application/json",
			"HTTP-Referer":  "https://smart-docparser.app",
			"X-Title":       "Smart-DocParser OCR Service",
			"User-Agent":    "Smart-DocParser/1.0",
		},
		Body:        payloadBytes,
		IsStreaming: true,
	})
	if err != nil {
		s.metrics.ErrorsTotal.WithLabelValues(model).Inc()
		return nil, fmt.Errorf("OCR API request failed: %w", err)
	}

	// Parse response
	return parseOCRResponse(resp.Body)
}

// parseOCRResponse extracts text from OCR API response
func parseOCRResponse(body []byte) (*OCRResult, error) {
	var result OCRResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse OCR result: %w", err)
	}
	return &result, nil
}
