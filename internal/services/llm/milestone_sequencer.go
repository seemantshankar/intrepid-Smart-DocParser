package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
)

// MilestoneSequencer handles milestone sequencing using LLM
type MilestoneSequencer struct {
	service      Service
	promptEngine *PromptEngine
}

// NewMilestoneSequencer creates a new milestone sequencer
func NewMilestoneSequencer(service Service, promptEngine *PromptEngine) *MilestoneSequencer {
	return &MilestoneSequencer{
		service:      service,
		promptEngine: promptEngine,
	}
}

// SequenceMilestones sequences milestones chronologically
func (s *MilestoneSequencer) SequenceMilestones(ctx context.Context, provider string, milestones []models.AnalysisMilestone) ([]models.SequencedMilestone, error) {
	// Build sequencing prompt
	prompt := s.promptEngine.BuildMilestoneSequencingPrompt(milestones)

	// Create request payload
	payload := map[string]interface{}{
		"model": "openai/gpt-5-nano",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a project management expert. Always respond with valid JSON arrays.",
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
	resp, err := s.service.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	// Parse response
	return s.parseSequencingResponse(resp.Body)
}

// parseSequencingResponse parses the sequencing response
func (s *MilestoneSequencer) parseSequencingResponse(body []byte) ([]models.SequencedMilestone, error) {
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

	var sequenced []models.SequencedMilestone
	content := openAIResp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), &sequenced); err != nil {
		return nil, fmt.Errorf("failed to parse sequenced milestones: %w", err)
	}

	return sequenced, nil
}
