package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
)

// RiskAssessor handles risk assessment using LLM
type RiskAssessor struct {
	service      Service
	promptEngine *PromptEngine
}

// NewRiskAssessor creates a new risk assessor
func NewRiskAssessor(service Service, promptEngine *PromptEngine) *RiskAssessor {
	return &RiskAssessor{
		service:      service,
		promptEngine: promptEngine,
	}
}

// AssessRisks performs comprehensive risk assessment
func (r *RiskAssessor) AssessRisks(ctx context.Context, provider, contractText, industryStandards string) (*models.AnalysisRiskAssessment, error) {
	// Build risk assessment prompt
	prompt := r.promptEngine.BuildRiskAssessmentPrompt(contractText, industryStandards)

	// Create request payload
	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a risk management and legal expert. Always respond with valid JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Execute request
	resp, err := r.service.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	// Parse response
	return r.parseRiskResponse(resp.Body)
}

// parseRiskResponse parses the risk assessment response
func (r *RiskAssessor) parseRiskResponse(body []byte) (*models.AnalysisRiskAssessment, error) {
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
