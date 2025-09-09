package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"contract-analysis-service/configs"

	"contract-analysis-service/internal/pkg/database"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/pkg/metrics"
	"contract-analysis-service/internal/pkg/pdf"
	"contract-analysis-service/internal/repositories/sqlite"
	"contract-analysis-service/internal/services/llm"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"contract-analysis-service/internal/services/ocr"
	"contract-analysis-service/internal/services/validation"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// MockValidator implements the ocr.Validator interface for demo
type MockValidator struct{}

func (m *MockValidator) Validate(result *ocr.OCRResult) error {
	if result.Confidence < 0.1 {
		return fmt.Errorf("OCR confidence too low: %.2f", result.Confidence)
	}
	return nil
}

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Setup logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Check for PDF file argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/ocr-demo/main.go <path-to-pdf>")
	}
	pdfPath := os.Args[1]

	// Check if PDF exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatalf("PDF file not found: %s", pdfPath)
	}

	// Check file size
	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		log.Fatalf("Error checking file: %v", err)
	}
	// Check file size (limit to 10MB to avoid excessive API usage)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		log.Fatalf("PDF file too large: %d bytes (max 10MB)", fileInfo.Size())
	}

	log.Printf("Processing PDF: %s (%d bytes)", pdfPath, fileInfo.Size())

	// Step 1: Convert PDF to images
	log.Println("Step 1: Converting PDF to images...")
	images, err := pdf.RasterizeToJPEGs(pdfPath, 3) // Process max 3 pages
	if err != nil {
		log.Fatalf("Failed to rasterize PDF: %v", err)
	}
	if len(images) == 0 {
		log.Fatal("No images generated from PDF")
	}
	log.Printf("Generated %d images", len(images))

	// Cleanup images on exit
	defer func() {
		for _, img := range images {
			os.Remove(img)
		}
		if len(images) > 0 {
			tempDir := filepath.Dir(images[0])
			os.RemoveAll(tempDir)
		}
	}()

	// Step 2: Setup OCR service
	log.Println("Step 2: Setting up OCR service...")
	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterAPIKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	retryConfig := external.RetryConfig{
		MaxRetries:      3,
		InitialInterval: time.Second,
		MaxInterval:     10 * time.Second,
	}
	httpClient := external.NewHTTPClient("https://openrouter.ai/api/v1", "ocr-demo", retryConfig, 30*time.Second)
	ocrMetrics := metrics.NewOCRMetrics()
	ocrValidator := &MockValidator{}
	ocrService := ocr.NewOCRService(httpClient, openRouterAPIKey, []string{"qwen/qwen2.5-vl-32b-instruct:free"}, ocrValidator, ocrMetrics)

	// Step 3: Extract text from images using OCR
	log.Println("Step 3: Extracting text from images...")
	var extractedTexts []string
	for i, imagePath := range images {
		log.Printf("Processing image %d/%d: %s", i+1, len(images), filepath.Base(imagePath))
		ocrResult, err := ocrService.ExtractTextFromImage(context.Background(), imagePath)
		if err != nil {
			log.Printf("OCR failed for image %d: %v", i+1, err)
			continue
		}
		if ocrResult.Text != "" {
			extractedTexts = append(extractedTexts, ocrResult.Text)
			log.Printf("OCR confidence for image %d: %.2f", i+1, ocrResult.Confidence)
			log.Printf("Extracted text length: %d characters", len(ocrResult.Text))
		}
	}

	if len(extractedTexts) == 0 {
		log.Fatal("No text extracted from any images")
	}

	// Step 4: Combine extracted text
	combinedText := strings.Join(extractedTexts, "\n\n")
	log.Printf("Step 4: Combined extracted text length: %d characters", len(combinedText))
	log.Printf("First 300 characters: %s", func() string {
		if len(combinedText) > 300 {
			return combinedText[:300] + "..."
		}
		return combinedText
	}())

	// Step 5: Setup validation service
	log.Println("Step 5: Setting up validation service...")

	// Setup database
	dbConfig := &configs.DatabaseConfig{
		Dialect: "sqlite3",
		Name:    ":memory:",
		LogMode: true,
	}

	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Setup repositories
	validationRepo := sqlite.NewValidationRepository(db)
	auditRepo := sqlite.NewValidationAuditRepository(db)
	feedbackRepo := sqlite.NewValidationFeedbackRepository(db)

	// Setup LLM service
	llmService := llm.NewLLMService(logger)
	config := &configs.Config{
		LLM: configs.LLMConfig{
			OpenRouter: configs.LLMProviderConfig{
				BaseURL:          "https://openrouter.ai/api/v1",
				APIKey:           openRouterAPIKey,
				Timeout:          30 * time.Second,
				RetryCount:       3,
				RetryWaitTime:    time.Second,
				RetryMaxInterval: 10 * time.Second,
			},
		},
	}
	llmclient.AddOpenRouterClientToService(llmService, config)

	// Setup validation service
	validationService := validation.NewValidationService(llmService, logger, validationRepo, auditRepo, feedbackRepo)

	// Step 6: Validate the extracted text as a contract
	log.Println("Step 6: Validating extracted text as contract...")
	result, err := validationService.ValidateContract(context.Background(), combinedText)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Step 7: Display results
	log.Println("\n=== VALIDATION RESULTS ===")
	log.Printf("Is Valid Contract: %v", result.IsValidContract)
	log.Printf("Contract Type: %s", result.ContractType)
	log.Printf("Confidence: %.2f", result.Confidence)
	log.Printf("Reason: %s", result.Reason)

	if len(result.MissingElements) > 0 {
		log.Println("\nMissing Elements:")
		for i, elem := range result.MissingElements {
			log.Printf("  %d. %s", i+1, elem)
		}
	}

	if len(result.DetectedElements) > 0 {
		log.Println("\nDetected Elements:")
		for i, elem := range result.DetectedElements {
			log.Printf("  %d. %s", i+1, elem)
		}
	}

	log.Println("\n=== OCR DEMO COMPLETED ===")
	log.Printf("Successfully processed scanned PDF through OCR pipeline")
}