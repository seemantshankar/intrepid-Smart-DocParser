package analysis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/services/llm"
	"golang.org/x/image/draw"
)

// Service defines analysis capabilities over contract text
//
//go:generate mockery --name=Service --output=./mocks --filename=service_mock.go
type Service interface {
	AnalyzeContract(ctx context.Context, text string) (*models.ContractAnalysis, error)
	AnalyzeContractFromImages(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error)
}

type service struct {
	llmService   llm.Service
	promptEngine *llm.PromptEngine
}

func NewService(llmService llm.Service) Service {
	return &service{llmService: llmService, promptEngine: llm.NewPromptEngine()}
}

func (s *service) AnalyzeContract(ctx context.Context, text string) (*models.ContractAnalysis, error) {
	prompt := s.promptEngine.BuildContractAnalysisPrompt(text)

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, "openrouter", &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    b,
	})
	if err != nil {
		return nil, fmt.Errorf("analysis request failed: %w", err)
	}

	return parseAnalysisResponse(resp.Body)
}

// AnalyzeContractFromImages performs direct image analysis using GPT-5 Nano's large context window for optimal speed and cost efficiency.
func (s *service) AnalyzeContractFromImages(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error) {
	fmt.Printf("🔍 STARTING DOCUMENT ANALYSIS - Found %d pages to process\n", len(imagePaths))
	
	// With GPT-5 Nano's 400k token context window, use direct image analysis for better speed and cost efficiency
	fmt.Printf("🚀 DIRECT IMAGE ANALYSIS - Using GPT-5 Nano's large context window for optimal processing\n")
	result, err := s.AnalyzeContractFromImagesDirectly(ctx, imagePaths)
	if err != nil {
		return nil, fmt.Errorf("direct image analysis failed: %w", err)
	}
	fmt.Printf("✅ ANALYSIS COMPLETE - Contract processing finished successfully\n")
	return result, nil
}

func parseAnalysisResponse(body []byte) (*models.ContractAnalysis, error) {
	// Check if response body is empty
	if len(body) == 0 {
		return nil, fmt.Errorf("received empty response from LLM API")
	}
	
	// Debug: log the raw response for troubleshooting (truncate if too long)
	responseStr := string(body)
	if len(responseStr) > 1000 {
		fmt.Printf("Debug: Raw LLM response (first 1000 chars): %s...\n", responseStr[:1000])
	} else {
		fmt.Printf("Debug: Raw LLM response: %s\n", responseStr)
	}
	
	// Check if response looks like valid JSON
	if !strings.HasPrefix(strings.TrimSpace(responseStr), "{") {
		maxLen := 200
		if len(responseStr) < maxLen {
			maxLen = len(responseStr)
		}
		return nil, fmt.Errorf("response does not appear to be JSON: %s", responseStr[:maxLen])
	}
	
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string      `json:"message"`
			Type    string      `json:"type"`
			Code    interface{} `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse analysis wrapper (response length: %d): %w", len(body), err)
	}
	
	// Check for API errors first
	if wrapper.Error != nil {
		return nil, fmt.Errorf("API error: %s (type: %s, code: %v)", wrapper.Error.Message, wrapper.Error.Type, wrapper.Error.Code)
	}
	
	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in analysis response - raw response: %s", string(body))
	}
	var result models.ContractAnalysis
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse analysis json content: %w", err)
	}
	return &result, nil
}

// compressImage resizes and compresses an image to reduce token usage
func (s *service) compressImage(imagePath string) ([]byte, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate new dimensions targeting ~1MB file size with good readability
	// Higher resolution for better text recognition with GPT-5 Nano's vision capabilities
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	
	// Use larger dimensions (1024px max) for better text legibility
	maxDimension := 1024
	if width > maxDimension || height > maxDimension {
		if width > height {
			height = height * maxDimension / width
			width = maxDimension
		} else {
			width = width * maxDimension / height
			height = maxDimension
		}
	}

	// Create a new image with the target size
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Resize the image using high-quality scaling
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Start with higher quality and adjust if needed to stay under 1MB
	targetSize := 1024 * 1024 // 1MB target
	quality := 85 // Start with higher quality for better text recognition
	
	for quality >= 60 {
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode compressed image: %w", err)
		}
		
		// If under target size, use this quality
		if buf.Len() <= targetSize {
			fmt.Printf("📸 Compressed image: %d bytes (quality: %d%%)\n", buf.Len(), quality)
			return buf.Bytes(), nil
		}
		
		// Reduce quality and try again
		quality -= 5
	}
	
	// Final fallback with minimum quality
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 60})
	if err != nil {
		return nil, fmt.Errorf("failed to encode compressed image: %w", err)
	}
	
	fmt.Printf("📸 Compressed image: %d bytes (final quality: 60%%)\n", buf.Len())
	return buf.Bytes(), nil
}

// AnalyzeContractFromImagesWithOCR processes images page-by-page using OCR, then analyzes the concatenated text
// DEPRECATED: With GPT-5 Nano's 400k token context window, direct image analysis is faster and more cost-effective
func (s *service) AnalyzeContractFromImagesWithOCR(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no images provided for analysis")
	}

	// Step 1: Extract text from each page using high-quality OCR
	fmt.Printf("🔄 PREPARING OCR EXTRACTION - Processing %d pages\n", len(imagePaths))
	
	// Limit to first 10 pages to avoid timeouts
	maxPages := len(imagePaths)
	if maxPages > 10 {
		fmt.Printf("⚠️  LIMITING PAGES - Processing first 10 of %d total pages to avoid timeout\n", maxPages)
		imagePaths = imagePaths[:10]
	}
	
	var allText []string
	for i, imagePath := range imagePaths {
		// Read the image with full quality (no compression for OCR)
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image %s: %w", imagePath, err)
		}

		// Use OCR service to extract text from this page
		fmt.Printf("📖 PROCESSING PAGE %d/%d - Converting to text (%d bytes)\n", i+1, len(imagePaths), len(imageData))
		
		// Create timeout context for each OCR call (30 seconds max per page)
		ocrCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		extractedText, err := s.extractTextFromImage(ocrCtx, imageData, fmt.Sprintf("Page %d", i+1))
		cancel()
		if err != nil {
			fmt.Printf("❌ PAGE %d FAILED: %v\n", i+1, err)
			return nil, fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}
		fmt.Printf("✅ PAGE %d COMPLETE - Extracted %d characters\n", i+1, len(extractedText))

		if extractedText != "" {
			allText = append(allText, fmt.Sprintf("=== PAGE %d ===\n%s\n", i+1, extractedText))
		}
	}

	if len(allText) == 0 {
		return nil, fmt.Errorf("no text could be extracted from any page")
	}

	// Step 2: Concatenate all extracted text
	fmt.Printf("📝 COMBINING TEXT - Merging content from %d pages\n", len(allText))
	concatenatedText := fmt.Sprintf("Contract Document Text (extracted from %d pages):\n\n%s", 
		len(imagePaths), strings.Join(allText, "\n"))

	// Step 3: Analyze the concatenated text using LLM (text-based, not image-based)
	fmt.Printf("🧠 ANALYZING CONTRACT - Sending to AI for financial analysis\n")
	return s.AnalyzeContractFromText(ctx, concatenatedText)
}

// extractTextFromImage uses OCR to extract text from a single image
func (s *service) extractTextFromImage(ctx context.Context, imageData []byte, pageLabel string) (string, error) {
	// Convert image to base64 for OCR service
	b64 := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", b64)

	// Create OCR request payload
	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano", // Use vision model for OCR
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text", 
						"text": "Extract all text from this document image. Return ONLY the extracted text, no analysis or commentary.",
					},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{"url": imageURL, "detail": "high"}, // High quality for OCR
					},
				},
			},
		},
		"temperature": 0.0, // Low temperature for consistent OCR
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OCR payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, "openrouter", &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    b,
	})
	if err != nil {
		return "", fmt.Errorf("OCR request failed: %w", err)
	}

	// Parse OCR response
	var ocrResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string      `json:"message"`
			Type    string      `json:"type"`
			Code    interface{} `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(resp.Body, &ocrResponse); err != nil {
		fmt.Printf("OCR response parse error: %v\nRaw response: %s\n", err, string(resp.Body))
		return "", fmt.Errorf("failed to parse OCR response: %w", err)
	}

	// Check for API errors
	if ocrResponse.Error != nil {
		return "", fmt.Errorf("OCR API error: %s (type: %s, code: %v)", ocrResponse.Error.Message, ocrResponse.Error.Type, ocrResponse.Error.Code)
	}

	if len(ocrResponse.Choices) == 0 {
		fmt.Printf("No OCR choices in response: %s\n", string(resp.Body))
		return "", fmt.Errorf("no OCR response choices")
	}

	return ocrResponse.Choices[0].Message.Content, nil
}

// AnalyzeContractFromText analyzes contract text (not images) using LLM
func (s *service) AnalyzeContractFromText(ctx context.Context, contractText string) (*models.ContractAnalysis, error) {
	instruction := s.promptEngine.BuildTextBasedContractAnalysisPrompt(contractText)

	// Create analysis request payload
	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": instruction,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.1,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal text analysis payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, "openrouter", &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    b,
	})
	if err != nil {
		return nil, fmt.Errorf("text analysis request failed: %w", err)
	}

	return parseAnalysisResponse(resp.Body)
}

// AnalyzeContractFromImagesDirectly uses compressed images directly (fallback method)
func (s *service) AnalyzeContractFromImagesDirectly(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no images provided for analysis")
	}

	fmt.Printf("🖼️  DIRECT IMAGE ANALYSIS - Processing %d compressed images\n", len(imagePaths))

	instruction := s.promptEngine.BuildComprehensiveContractAnalysisPrompt()

	// Start conservative with page limits while debugging, then can increase
	maxPages := len(imagePaths)
	if maxPages > 10 {
		fmt.Printf("⚠️  LIMITING TO 10 PAGES - Found %d total, processing first 10 for testing\n", maxPages)
		imagePaths = imagePaths[:10]
	} else {
		fmt.Printf("📄 PROCESSING ALL PAGES - Found %d total pages\n", maxPages)
	}

	var content []interface{}
	content = append(content, map[string]interface{}{"type": "text", "text": instruction})
	
	for i, p := range imagePaths {
		fmt.Printf("🗜️  COMPRESSING IMAGE %d/%d - Preparing for AI analysis\n", i+1, len(imagePaths))
		compressedData, err := s.compressImage(p)
		if err != nil {
			fmt.Printf("❌ COMPRESSION FAILED for image %d: %v\n", i+1, err)
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(compressedData)
		url := fmt.Sprintf("data:image/jpeg;base64,%s", b64)
		content = append(content, map[string]interface{}{
			"type":      "image_url",
			"image_url": map[string]string{"url": url, "detail": "high"},
		})
		fmt.Printf("✅ IMAGE %d READY - Compressed and encoded\n", i+1)
	}

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": content,
			},
		},
		"stream": false,
	}
	
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal direct analysis payload: %w", err)
	}

	// Log payload size for debugging
	payloadSizeKB := len(b) / 1024
	fmt.Printf("📦 PAYLOAD SIZE - %d KB (%d bytes)\n", payloadSizeKB, len(b))
	fmt.Printf("🚀 SENDING TO AI - Requesting contract analysis with %d images\n", len(imagePaths))
	
	// Create a timeout context for large image analysis (5 minutes)
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	
	resp, err := s.llmService.ExecuteRequest(timeoutCtx, "openrouter", &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"HTTP-Referer":   "https://smart-docparser.com", // Optional: Site URL for OpenRouter rankings
			"X-Title":        "Smart DocParser",              // Optional: Site title for OpenRouter rankings
		},
		Body: b,
	})
	if err != nil {
		fmt.Printf("❌ AI REQUEST FAILED: %v\n", err)
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			return nil, fmt.Errorf("analysis request timed out (payload: %d KB) - try reducing image count: %w", payloadSizeKB, err)
		}
		return nil, fmt.Errorf("direct analysis request failed: %w", err)
	}
	
	// Check response status code
	if resp.StatusCode != 200 {
		fmt.Printf("❌ AI API ERROR - Status: %d, Body: %s\n", resp.StatusCode, string(resp.Body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(resp.Body))
	}
	
	fmt.Printf("📋 PARSING RESULTS - Processing AI response (status: %d)\n", resp.StatusCode)
	return parseAnalysisResponse(resp.Body)
}
