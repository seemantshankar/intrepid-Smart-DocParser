package client

import (
	"contract-analysis-service/configs"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/services/llm"
)

// AddOpenRouterClientToService creates a new OpenRouter client and adds it to the LLM service.
// This function encapsulates the client creation and registration logic.
func AddOpenRouterClientToService(service llm.Service, cfg *configs.Config) {
	// Create a new resilient HTTP client for the OpenRouter API.
	resilientClient := external.NewHTTPClient(
		cfg.LLM.OpenRouter.BaseURL,
		"OpenRouter",
		external.RetryConfig{
			MaxRetries:      cfg.LLM.OpenRouter.RetryCount,
			InitialInterval: cfg.LLM.OpenRouter.RetryWaitTime,
			MaxInterval:     cfg.LLM.OpenRouter.RetryMaxInterval,
		},
		cfg.LLM.OpenRouter.Timeout,
	)

	// Create the OpenRouter-specific client.
	openRouterClient := NewOpenRouterClient(resilientClient, cfg.LLM.OpenRouter.APIKey)

	// Add the client to the LLM service.
	service.AddClient(OpenRouterProvider, openRouterClient)
}
