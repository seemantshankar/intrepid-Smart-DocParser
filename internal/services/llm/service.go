package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Service defines the interface for interacting with a large language model.
// It abstracts the underlying implementation of the LLM provider.
type Service interface {
	AnalyzeContract(ctx context.Context, provider, contractText string) (*models.ContractAnalysis, error)
	ExtractContractSummary(ctx context.Context, provider, contractText string) (*models.ContractSummary, error)
	IdentifyPaymentObligations(ctx context.Context, provider, contractText string) ([]models.AnalysisMilestone, error)
	CalculatePercentageBasedPayments(totalValue float64, milestones []models.AnalysisMilestone) ([]models.AnalysisMilestone, error)
	AssessContractRisks(ctx context.Context, provider, contractText, industryStandards string) (*models.AnalysisRiskAssessment, error)
	ValidateAnalysisConfidence(analysis *models.ContractAnalysis) (float64, []string, error)
	AddClient(provider string, client external.Client)
	ExecuteRequest(ctx context.Context, provider string, req *external.Request) (*external.Response, error)
}

// llmService handles integration with various LLM APIs
type llmService struct {
	clients map[string]external.Client
	logger  *zap.Logger
}

// NewLLMService creates a new LLM service instance
func NewLLMService(logger *zap.Logger) Service {
	return &llmService{
		clients: make(map[string]external.Client),
		logger:  logger,
	}
}

// AddClient adds a new LLM client to the service
func (s *llmService) AddClient(provider string, client external.Client) {
	s.clients[provider] = client
}

// ExecuteRequest executes a request through the specified LLM provider
func (s *llmService) ExecuteRequest(ctx context.Context, provider string, req *external.Request) (*external.Response, error) {
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	logger := s.logger.With(
		zap.String("provider", provider),
		zap.String("url", req.URL),
	)

	logger.Info("Executing LLM API request")
	startTime := time.Now()

	resp, err := client.ExecuteRequest(ctx, req)

	duration := time.Since(startTime)
	logger.Info("LLM API request completed", zap.Duration("duration", duration))

	if err != nil {
		logger.Error("LLM API request failed", zap.Error(err))
		return nil, err
	}

	logger.Info("LLM API request successful", zap.Int("status_code", resp.StatusCode))
	return resp, nil
}

// AnalyzeContract performs contract analysis using the specified LLM
func (s *llmService) AnalyzeContract(ctx context.Context, provider, contractText string) (*models.ContractAnalysis, error) {
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	// The request body for the LLM will be more complex, this is a placeholder
	// for the general structure.
	prompt := buildContractAnalysisPrompt(contractText)

	// The actual request will be handled by the specific client implementation.
	// For now, we are focusing on the service structure.
	resp, err := client.ExecuteRequest(ctx, &external.Request{
		Method: "POST",
		Body:   []byte(prompt), // This will be replaced by a structured request
	})

	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	return parseAnalysisResponse(resp.Body)
}

// buildContractAnalysisPrompt constructs the prompt for contract analysis
func buildContractAnalysisPrompt(contractText string) string {
	return fmt.Sprintf(`Analyze the following contract and extract key information:

Contract:
"""
%s
"""

Instructions:
1. Identify buyer and seller
2. Extract total contract value
3. List payment obligations with amounts/percentages
4. Identify risk factors
5. Output in JSON format`, contractText)
}

// parseAnalysisResponse parses the LLM API response
func parseAnalysisResponse(body []byte) (*models.ContractAnalysis, error) {
	var analysis models.ContractAnalysis
	if err := json.Unmarshal(body, &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	return &analysis, nil
}

// ExtractContractSummary extracts detailed contract summary using LLM
func (s *llmService) ExtractContractSummary(ctx context.Context, provider, contractText string) (*models.ContractSummary, error) {
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	prompt := buildContractSummaryPrompt(contractText)

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a contract analysis expert. Extract detailed contract summary information. Always respond with valid JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  1500,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := client.ExecuteRequest(ctx, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	return parseSummaryResponse(resp.Body)
}

// IdentifyPaymentObligations identifies all payment obligations from contract text
func (s *llmService) IdentifyPaymentObligations(ctx context.Context, provider, contractText string) ([]models.AnalysisMilestone, error) {
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	prompt := buildPaymentObligationsPrompt(contractText)

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a contract analysis expert specializing in payment obligations. Extract all payment-related obligations and milestones. Always respond with valid JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  2000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := client.ExecuteRequest(ctx, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	return parsePaymentObligationsResponse(resp.Body)
}

// CalculatePercentageBasedPayments calculates absolute amounts from percentages
func (s *llmService) CalculatePercentageBasedPayments(totalValue float64, milestones []models.AnalysisMilestone) ([]models.AnalysisMilestone, error) {
	if totalValue <= 0 {
		return nil, fmt.Errorf("total value must be positive")
	}

	var updatedMilestones []models.AnalysisMilestone
	var totalPercentage float64

	for _, milestone := range milestones {
		updated := milestone

		// Parse amount as decimal if possible, otherwise keep as string
		amountStr := milestone.Amount.String()
		amountDecimal, err := decimal.NewFromString(amountStr)
		var amountFloat float64
		if err == nil {
			amountFloat, _ = amountDecimal.Float64()
		}

		// If amount is zero but percentage is provided, calculate amount
		if (amountStr == "" || amountStr == "0") && milestone.Percentage > 0 {
			calculatedAmount := totalValue * milestone.Percentage / 100
			updated.Amount = models.FlexibleString(decimal.NewFromFloat(calculatedAmount).String())
		}

		// If amount is provided but percentage is zero, calculate percentage
		if amountStr != "" && amountStr != "0" && milestone.Percentage == 0 && err == nil {
			updated.Percentage = (amountFloat / totalValue) * 100
		}

		totalPercentage += updated.Percentage
		updatedMilestones = append(updatedMilestones, updated)
	}

	// Validate that percentages sum to approximately 100%
	if totalPercentage < 99.5 || totalPercentage > 100.5 {
		s.logger.Warn("Milestone percentages do not sum to 100%",
			zap.Float64("total_percentage", totalPercentage))
	}

	return updatedMilestones, nil
}

// AssessContractRisks performs comprehensive risk assessment using industry standards
func (s *llmService) AssessContractRisks(ctx context.Context, provider, contractText, industryStandards string) (*models.AnalysisRiskAssessment, error) {
	client, ok := s.clients[provider]
	if !ok {
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	prompt := buildRiskAssessmentPrompt(contractText, industryStandards)

	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a risk assessment expert with deep knowledge of contract law and industry best practices. Analyze contracts for potential risks and compliance issues. Always respond with valid JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  2500,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := client.ExecuteRequest(ctx, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	return parseRiskAssessmentResponse(resp.Body)
}

// ValidateAnalysisConfidence validates analysis results and calculates confidence score
func (s *llmService) ValidateAnalysisConfidence(analysis *models.ContractAnalysis) (float64, []string, error) {
	if analysis == nil {
		return 0.0, []string{"analysis is nil"}, fmt.Errorf("analysis cannot be nil")
	}

	var issues []string
	var confidence float64 = 1.0

	// Check for missing critical fields
	if analysis.Buyer == "" {
		issues = append(issues, "buyer name is missing")
		confidence -= 0.2
	}

	if analysis.Seller == "" {
		issues = append(issues, "seller name is missing")
		confidence -= 0.2
	}

	if analysis.TotalValue.String() == "" || analysis.TotalValue.String() == "0" {
		issues = append(issues, "total contract value is missing or zero")
		confidence -= 0.3
	}

	if analysis.Currency == "" {
		issues = append(issues, "currency is missing")
		confidence -= 0.1
	}

	// Check milestones
	if len(analysis.Milestones) == 0 {
		issues = append(issues, "no payment milestones identified")
		confidence -= 0.4
	} else {
		var totalPercentage float64
		for _, milestone := range analysis.Milestones {
			if milestone.Description == "" {
				issues = append(issues, "milestone missing description")
				confidence -= 0.05
			}
			if (milestone.Amount.String() == "" || milestone.Amount.String() == "0") && milestone.Percentage == 0 {
				issues = append(issues, "milestone missing amount and percentage")
				confidence -= 0.1
			}
			totalPercentage += milestone.Percentage
		}

		// Check if percentages sum to approximately 100%
		if totalPercentage < 95 || totalPercentage > 105 {
			issues = append(issues, fmt.Sprintf("milestone percentages sum to %.2f%%, not 100%%", totalPercentage))
			confidence -= 0.15
		}
	}

	// Ensure confidence doesn't go below 0
	if confidence < 0 {
		confidence = 0
	}

	return confidence, issues, nil
}

// Helper functions for building prompts

func buildContractSummaryPrompt(contractText string) string {
	return fmt.Sprintf(`You are a legal contract analysis expert. Extract detailed contract summary information from the following contract text.

CONTRACT TEXT:
"""
%s
"""

INSTRUCTIONS:
Extract the following information and return as JSON:
1. Buyer name, address, and country
2. Seller name, address, and country  
3. Nature of goods/services (physical, digital, services)
4. Total contract value and currency
5. Legal jurisdiction

Return a JSON object with this exact structure:
{
  "buyer_name": "string",
  "buyer_address": "string",
  "buyer_country": "string",
  "seller_name": "string",
  "seller_address": "string", 
  "seller_country": "string",
  "goods_nature": "physical|digital|services",
  "total_value": number,
  "currency": "string",
  "jurisdiction": "string"
}

Only return the JSON, no additional text.`, contractText)
}

func buildPaymentObligationsPrompt(contractText string) string {
	return fmt.Sprintf(`You are a contract analysis expert specializing in payment obligations. Extract all payment-related obligations and milestones from the following contract text.

CONTRACT TEXT:
"""
%s
"""

INSTRUCTIONS:
1. Identify ALL clauses that trigger monetary transfers
2. Extract payment descriptions, amounts, and trigger conditions
3. Include both fixed amounts and percentage-based payments
4. Include deposit, progress payments, final payments, penalties, etc.
5. Extract the specific conditions that trigger each payment

Return a JSON array of payment obligations:
[
  {
    "description": "string",
    "amount": number,
    "percentage": number,
    "trigger_condition": "string"
  }
]

Only return the JSON array, no additional text.`, contractText)
}

func buildRiskAssessmentPrompt(contractText, industryStandards string) string {
	standards := industryStandards
	if standards == "" {
		standards = "No specific industry standards provided"
	}

	return fmt.Sprintf(`You are a risk assessment expert with deep knowledge of contract law. Assess the following contract for potential risks and vulnerabilities.

CONTRACT TEXT:
"""
%s
"""

INDUSTRY STANDARDS:
"""
%s
"""

INSTRUCTIONS:
1. Compare the contract against general and industry-specific best practices
2. Identify missing contractual elements, clauses, or protections
3. Assess risks for both buyer and seller
4. Suggest specific improvements with legal reasoning
5. Categorize risks by severity and provide actionable recommendations

Return a JSON object with risk assessment:
{
  "missing_clauses": ["string"],
  "risks": [
    {
      "party": "buyer|seller",
      "type": "string",
      "severity": "low|medium|high|critical",
      "description": "string",
      "recommendation": "string"
    }
  ],
  "compliance_score": number,
  "suggestions": ["string"]
}

Only return the JSON, no additional text.`, contractText, standards)
}

// Helper functions for parsing responses

func parseSummaryResponse(body []byte) (*models.ContractSummary, error) {
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	var summary models.ContractSummary
	content := openAIResp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), &summary); err != nil {
		return nil, fmt.Errorf("failed to parse contract summary: %w", err)
	}

	return &summary, nil
}

func parsePaymentObligationsResponse(body []byte) ([]models.AnalysisMilestone, error) {
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	var milestones []models.AnalysisMilestone
	content := openAIResp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), &milestones); err != nil {
		return nil, fmt.Errorf("failed to parse payment obligations: %w", err)
	}

	return milestones, nil
}

func parseRiskAssessmentResponse(body []byte) (*models.AnalysisRiskAssessment, error) {
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	var assessment models.AnalysisRiskAssessment
	content := openAIResp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), &assessment); err != nil {
		return nil, fmt.Errorf("failed to parse risk assessment: %w", err)
	}

	return &assessment, nil
}
