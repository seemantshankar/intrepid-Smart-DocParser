package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
)

// ComplianceChecker handles compliance checking using LLM
type ComplianceChecker struct {
	service      Service
	promptEngine *PromptEngine
}

// NewComplianceChecker creates a new compliance checker
func NewComplianceChecker(service Service, promptEngine *PromptEngine) *ComplianceChecker {
	return &ComplianceChecker{
		service:      service,
		promptEngine: promptEngine,
	}
}

// CheckCompliance performs compliance analysis
func (c *ComplianceChecker) CheckCompliance(ctx context.Context, provider, contractText, jurisdiction string) (*models.AnalysisComplianceReport, error) {
	prompt := fmt.Sprintf(`You are a legal compliance expert. Analyze the following contract for compliance with %s jurisdiction requirements.

CONTRACT TEXT:
"""
%s
"""

INSTRUCTIONS:
1. Identify required legal clauses for this jurisdiction
2. Check if all required clauses are present
3. Flag any missing regulatory requirements
4. Suggest standard clause additions
5. Assess overall compliance level

Return a JSON object with compliance analysis:
{
  "jurisdiction": "string",
  "required_clauses": ["string"],
  "missing_clauses": ["string"],
  "compliance_level": "full|partial|minimal|non-compliant",
  "recommendations": ["string"],
  "risk_level": "low|medium|high|critical"
}

Only return the JSON, no additional text.`, jurisdiction, contractText)
	
	// Create request payload
	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a legal compliance expert. Always respond with valid JSON.",
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
	resp, err := c.service.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}
	
	// Parse response
	return c.parseComplianceResponse(resp.Body)
}

// parseComplianceResponse parses the compliance response
func (c *ComplianceChecker) parseComplianceResponse(body []byte) (*models.AnalysisComplianceReport, error) {
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
	
	var report models.AnalysisComplianceReport
	content := openAIResp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), &report); err != nil {
		return nil, fmt.Errorf("failed to parse compliance report: %w", err)
	}

	return &report, nil
}
