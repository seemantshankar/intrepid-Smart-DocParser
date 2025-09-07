package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
)

// ContractAnalyzer handles contract analysis using LLM APIs
type ContractAnalyzer struct {
	service    Service
	promptEngine *PromptEngine
}

// NewContractAnalyzer creates a new contract analyzer
func NewContractAnalyzer(service Service, promptEngine *PromptEngine) *ContractAnalyzer {
	return &ContractAnalyzer{
		service:      service,
		promptEngine: promptEngine,
	}
}

// AnalyzeContract performs comprehensive contract analysis
func (c *ContractAnalyzer) AnalyzeContract(ctx context.Context, provider, contractText string) (*models.ContractAnalysis, error) {
	// Build analysis prompt
	prompt := c.promptEngine.BuildContractAnalysisPrompt(contractText)
	
	// Create request payload for LLM
	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a legal document analysis expert. Always respond with valid JSON.",
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
	
	// Execute request through LLM service
	resp, err := c.service.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions", // This is the endpoint path, not the full URL
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}
	
	// Parse response
	return c.parseAnalysisResponse(resp.Body)
}

// parseAnalysisResponse parses the LLM API response into ContractAnalysis
func (c *ContractAnalyzer) parseAnalysisResponse(body []byte) (*models.ContractAnalysis, error) {
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
	
	var analysis models.ContractAnalysis
	content := openAIResp.Choices[0].Message.Content
	
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse contract analysis: %w", err)
	}
	
	return &analysis, nil
}
