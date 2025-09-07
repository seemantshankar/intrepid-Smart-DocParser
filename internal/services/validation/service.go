package validation

import (
	"context"
	"encoding/json"
	"fmt"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/services/llm"
	"go.uber.org/zap"
)

// Service defines the interface for the contract validation service.
type Service interface {
	ValidateContract(ctx context.Context, documentText string) (*models.ValidationResult, error)
}

// validationService implements the Service interface.
type validationService struct {
	llmService llm.Service
	logger     *zap.Logger
}

// NewValidationService creates a new validation service instance.
func NewValidationService(llmService llm.Service, logger *zap.Logger) Service {
	return &validationService{
		llmService: llmService,
		logger:     logger,
	}
}

// ValidateContract uses the LLM service to determine if a document is a valid contract.
func (s *validationService) ValidateContract(ctx context.Context, documentText string) (*models.ValidationResult, error) {
	prompt := buildValidationPrompt(documentText)

	// For simplicity, we'll use the default provider. A more advanced implementation
	// could allow for provider selection.
	provider := "openrouter"

	// Prepare payload for the LLM request
	payload := map[string]interface{}{
		"model": "qwen/qwen-2.5-vl-72b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation payload: %w", err)
	}

	// Execute the request using the LLM service
	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("contract validation request failed: %w", err)
	}

	// Parse the validation response
	return parseValidationResponse(resp.Body)
}

func buildValidationPrompt(documentText string) string {
	return fmt.Sprintf(`Analyze the following document and determine if it is a valid legal contract. Respond with a JSON object containing these keys: 'is_valid_contract' (boolean), 'reason' (string, if not valid), 'confidence' (float, 0.0-1.0), 'contract_type' (string, e.g., 'Sale of Goods', 'Service Agreement'), 'missing_elements' (array of strings), and 'detected_elements' (array of strings). Document:\n\n%s`, documentText)
}

func parseValidationResponse(body []byte) (*models.ValidationResult, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in validation response")
	}

	var result models.ValidationResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse validation result from content: %w", err)
	}

	return &result, nil
}
