package client

import (
	"context"

	"contract-analysis-service/internal/pkg/external"
)

const (
	OpenRouterProvider = "openrouter"
	DefaultModel     = "qwen/qwen-2.5-vl-72b-instruct:free"
)

// OpenRouterClient is a client for the OpenRouter API.
// It wraps the generic external client to provide specific functionality for OpenRouter.
type OpenRouterClient struct {
	client external.Client
	apiKey string
	model  string
}

// NewOpenRouterClient creates a new OpenRouter client.
func NewOpenRouterClient(client external.Client, apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		client: client,
		apiKey: apiKey,
		model:  DefaultModel,
	}
}

// SetModel allows overriding the default model.
func (c *OpenRouterClient) SetModel(model string) {
	c.model = model
}

// ExecuteRequest prepares and executes a request to the OpenRouter API.
// This method is not part of the external.Client interface, but is a specific implementation detail.
func (c *OpenRouterClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	// Set the auth header for OpenRouter
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	req.Headers["Authorization"] = "Bearer " + c.apiKey

	// Here you would typically structure the body of the request according to OpenRouter's API.
	// For now, we will pass it through as is.

	return c.client.ExecuteRequest(ctx, req)
}
