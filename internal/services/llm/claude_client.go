package llm

import (
	"context"

	"contract-analysis-service/internal/pkg/external"
)

// ClaudeClient implements Client for Anthropic Claude API
type ClaudeClient struct {
	client external.Client
	apiKey string
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(client external.Client, apiKey string) *ClaudeClient {
	return &ClaudeClient{
		client: client,
		apiKey: apiKey,
	}
}

// ExecuteRequest executes a request to Claude API
func (c *ClaudeClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	// Set Claude-specific headers
	req.Headers["x-api-key"] = c.apiKey
	req.Headers["anthropic-version"] = "2023-06-01"
	req.Headers["Content-Type"] = "application/json"

	return c.client.ExecuteRequest(ctx, req)
}
