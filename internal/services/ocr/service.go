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
					map[string]interface{}{"type": "text", "text": "Extract the full readable text from this page image. Respond ONLY as a JSON object with keys: 'text' (string with extracted text) and 'confidence' (float 0.0-1.0 indicating extraction confidence)."},
					map[string]interface{}{"type": "image_url", "image_url": map[string]string{"url": imageURL}},
				},
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"stream":          false,
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
		IsStreaming: false,
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
	// OpenRouter chat completion response with response_format json_object returns
	// a JSON string in choices[0].message.content that contains our OCRResult fields
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse OCR wrapper: %w", err)
	}
	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OCR response")
	}
	var result OCRResult
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse OCR content: %w", err)
	}
	return &result, nil
}
