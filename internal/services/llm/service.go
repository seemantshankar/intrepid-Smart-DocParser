package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"go.uber.org/zap"
)

// Service defines the interface for interacting with a large language model.
// It abstracts the underlying implementation of the LLM provider.
type Service interface {
	AnalyzeContract(ctx context.Context, provider, contractText string) (*models.ContractAnalysis, error)
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
