package llm

import (
	"context"

	"contract-analysis-service/internal/pkg/external"
)

// OpenAIClient implements Client for OpenAI API
type OpenAIClient struct {
	client external.Client
	apiKey string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(client external.Client, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: client,
		apiKey: apiKey,
	}
}

// ExecuteRequest executes a request to OpenAI API
func (c *OpenAIClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	// Set OpenAI-specific headers
	req.Headers["Authorization"] = "Bearer " + c.apiKey
	req.Headers["Content-Type"] = "application/json"

	return c.client.ExecuteRequest(ctx, req)
}
