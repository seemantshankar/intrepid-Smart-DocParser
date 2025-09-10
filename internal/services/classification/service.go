package classification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/llm"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines the interface for contract classification.
type Service interface {
	ClassifyContract(ctx context.Context, documentText string) (*models.ContractClassification, error)
	GetContractComplexity(ctx context.Context, documentText string) (*models.ContractComplexity, error)
	ClassifyByIndustry(ctx context.Context, documentText string) (*models.IndustryClassification, error)
	StoreClassification(ctx context.Context, contractID string, classification *models.ContractClassification) error
	GetClassificationHistory(ctx context.Context, contractID string) ([]*models.ClassificationRecord, error)
}

// classificationService implements the Service interface.
type classificationService struct {
	llmService llm.Service
	logger     *zap.Logger
	repo       repositories.ClassificationRepository
}

// NewClassificationService creates a new classification service instance.
func NewClassificationService(llmService llm.Service, logger *zap.Logger, repo repositories.ClassificationRepository) Service {
	return &classificationService{
		llmService: llmService,
		logger:     logger,
		repo:       repo,
	}
}

// ClassifyContract performs comprehensive contract classification.
func (s *classificationService) ClassifyContract(ctx context.Context, documentText string) (*models.ContractClassification, error) {
	prompt := buildClassificationPrompt(documentText)
	provider := "openrouter"

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal classification payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("contract classification request failed: %w", err)
	}

	return parseClassificationResponse(resp.Body)
}

// GetContractComplexity analyzes the complexity of a contract.
func (s *classificationService) GetContractComplexity(ctx context.Context, documentText string) (*models.ContractComplexity, error) {
	prompt := buildComplexityPrompt(documentText)
	provider := "openrouter"

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal complexity payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("complexity analysis request failed: %w", err)
	}

	return parseComplexityResponse(resp.Body)
}

// ClassifyByIndustry performs industry-specific classification.
func (s *classificationService) ClassifyByIndustry(ctx context.Context, documentText string) (*models.IndustryClassification, error) {
	prompt := buildIndustryPrompt(documentText)
	provider := "openrouter"

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal industry payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("industry classification request failed: %w", err)
	}

	return parseIndustryResponse(resp.Body)
}

// StoreClassification stores classification results.
func (s *classificationService) StoreClassification(ctx context.Context, contractID string, classification *models.ContractClassification) error {
	record := &models.ClassificationRecord{
		ID:             uuid.New().String(),
		ContractID:     contractID,
		Classification: classification,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	return s.repo.Create(record)
}

// GetClassificationHistory retrieves classification history for a contract.
func (s *classificationService) GetClassificationHistory(ctx context.Context, contractID string) ([]*models.ClassificationRecord, error) {
	return s.repo.GetByContractID(contractID)
}

func buildClassificationPrompt(documentText string) string {
	return fmt.Sprintf(`Analyze the following contract document and provide comprehensive classification.

Classify the contract across multiple dimensions:

1. PRIMARY TYPE: Main contract category (e.g., "Sale of Goods", "Service Agreement", "Employment Contract", "Lease Agreement", "Partnership Agreement", "License Agreement")
2. SUB TYPE: Specific subtype (e.g., "Software License", "Consulting Agreement", "Equipment Lease", "Non-Disclosure Agreement")
3. INDUSTRY: Primary industry sector (e.g., "Technology", "Manufacturing", "Healthcare", "Finance", "Real Estate")
4. COMPLEXITY: Contract complexity level ("simple", "moderate", "complex", "highly_complex")
5. RISK LEVEL: Overall risk assessment ("low", "medium", "high", "critical")
6. JURISDICTION: Legal jurisdiction or governing law
7. CONTRACT VALUE: Value range ("micro", "small", "medium", "large", "enterprise")
8. DURATION: Time classification ("short_term", "medium_term", "long_term", "perpetual")
9. PARTY TYPES: Types of contracting parties (e.g., ["B2B"], ["B2C"], ["Government", "Private"])
10. SPECIAL CLAUSES: Notable clauses present (e.g., ["intellectual_property", "non_compete", "confidentiality", "force_majeure"])

Respond with a JSON object containing these keys:
- 'primary_type' (string)
- 'sub_type' (string)
- 'industry' (string)
- 'complexity' (string)
- 'risk_level' (string)
- 'jurisdiction' (string)
- 'contract_value' (string)
- 'duration' (string)
- 'party_types' (array of strings)
- 'special_clauses' (array of strings)
- 'confidence' (float, 0.0-1.0)
- 'metadata' (object with additional classification details)

Document:

%s`, documentText)
}

func buildComplexityPrompt(documentText string) string {
	return fmt.Sprintf(`Analyze the complexity of the following contract document.

Evaluate complexity based on:
1. Number and sophistication of clauses
2. Legal terminology density
3. Cross-references and dependencies
4. Conditional logic complexity
5. Multi-party arrangements
6. International elements
7. Regulatory compliance requirements
8. Technical specifications

Respond with a JSON object containing:
- 'level' (string): "simple", "moderate", "complex", or "highly_complex"
- 'score' (float): complexity score from 0.0 to 1.0
- 'factors' (array): factors contributing to complexity
- 'clause_count' (integer): estimated number of distinct clauses
- 'page_count' (integer): estimated page count
- 'legal_term_count' (integer): count of legal/technical terms
- 'cross_references' (integer): internal document references
- 'external_references' (integer): references to external documents

Document:

%s`, documentText)
}

func buildIndustryPrompt(documentText string) string {
	return fmt.Sprintf(`Analyze the following contract document for industry-specific classification.

Identify:
1. PRIMARY INDUSTRY: Main industry sector
2. SECONDARY INDUSTRY: Secondary industry if applicable
3. INDUSTRY CODE: NAICS or SIC code if determinable
4. REGULATIONS: Applicable industry regulations
5. STANDARDS: Relevant industry standards
6. COMPLIANCE: Specific compliance requirements

Respond with a JSON object containing:
- 'primary_industry' (string)
- 'secondary_industry' (string, optional)
- 'industry_code' (string, optional)
- 'regulations' (array of strings)
- 'standards' (array of strings)
- 'compliance' (object with requirement mappings)
- 'confidence' (float, 0.0-1.0)

Document:

%s`, documentText)
}

func parseClassificationResponse(body []byte) (*models.ContractClassification, error) {
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse classification response: %w", err)
	}

	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in classification response")
	}

	var result models.ContractClassification
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse classification from content: %w", err)
	}

	return &result, nil
}

func parseComplexityResponse(body []byte) (*models.ContractComplexity, error) {
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse complexity response: %w", err)
	}

	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in complexity response")
	}

	var result models.ContractComplexity
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse complexity from content: %w", err)
	}

	return &result, nil
}

func parseIndustryResponse(body []byte) (*models.IndustryClassification, error) {
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse industry response: %w", err)
	}

	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in industry response")
	}

	var result models.IndustryClassification
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse industry classification from content: %w", err)
	}

	return &result, nil
}
