package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"contract-analysis-service/internal/pkg/external"
)

// SearchResult represents a search result from Google Search API
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"` // Domain name for credibility scoring
}

// SearchClient defines the interface for web search operations
type SearchClient interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

// GoogleSearchConfig holds configuration for Google Search API
type GoogleSearchConfig struct {
	APIKey         string                  `json:"api_key"`
	SearchEngineID string                  `json:"search_engine_id"`
	BaseURL        string                  `json:"base_url"`
	MaxResults     int                     `json:"max_results"`
	Timeout        time.Duration           `json:"timeout"`
	RetryConfig    external.RetryConfig    `json:"retry_config"`
}

// googleSearchClient implements SearchClient for Google Custom Search API
type googleSearchClient struct {
	httpClient     external.Client
	apiKey         string
	searchEngineID string
	baseURL        string
	maxResults     int
}

// GoogleSearchResponse represents the response from Google Search API
type GoogleSearchResponse struct {
	Items []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Snippet     string `json:"snippet"`
		DisplayLink string `json:"displayLink"`
	} `json:"items"`
	SearchInformation struct {
		SearchTime   float64 `json:"searchTime"`
		TotalResults string  `json:"totalResults"`
	} `json:"searchInformation"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error"`
}

// NewGoogleSearchClient creates a new Google Search API client
func NewGoogleSearchClient(config GoogleSearchConfig) SearchClient {
	// Set default values if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://www.googleapis.com/customsearch/v1"
	}
	if config.MaxResults == 0 {
		config.MaxResults = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryConfig.MaxRetries == 0 {
		config.RetryConfig = external.DefaultRetryConfig()
	}

	// Create HTTP client with resilience patterns
	httpClient := external.NewHTTPClient(
		config.BaseURL,
		"google-search-client",
		config.RetryConfig,
		config.Timeout,
	)

	return &googleSearchClient{
		httpClient:     httpClient,
		apiKey:         config.APIKey,
		searchEngineID: config.SearchEngineID,
		baseURL:        config.BaseURL,
		maxResults:     config.MaxResults,
	}
}

// Search performs a search query using Google Custom Search API
func (c *googleSearchClient) Search(ctx context.Context, query string) ([]SearchResult, error) {
	// Validate input
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Build search URL
	searchURL := c.buildSearchURL(trimmedQuery)

	// Create request
	req := &external.Request{
		Method: "GET",
		URL:    searchURL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: nil,
	}

	// Execute request
	resp, err := c.httpClient.ExecuteRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var searchResp GoogleSearchResponse
	if err := json.Unmarshal(resp.Body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Check for API error in response
	if searchResp.Error != nil {
		return nil, fmt.Errorf("search API error: %s (code: %d)", searchResp.Error.Message, searchResp.Error.Code)
	}

	// Convert to SearchResult format
	results := make([]SearchResult, len(searchResp.Items))
	for i, item := range searchResp.Items {
		results[i] = SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
			Source:  item.DisplayLink,
		}
	}

	return results, nil
}

// buildSearchURL constructs the Google Search API URL with query parameters
func (c *googleSearchClient) buildSearchURL(query string) string {
	// Create URL with base path
	u, err := url.Parse(c.baseURL)
	if err != nil {
		// Fallback to simple string concatenation if URL parsing fails
		return fmt.Sprintf("%s?key=%s&cx=%s&q=%s&num=%d",
			c.baseURL,
			url.QueryEscape(c.apiKey),
			url.QueryEscape(c.searchEngineID),
			url.QueryEscape(query),
			c.maxResults,
		)
	}

	// Set query parameters
	values := url.Values{}
	values.Set("key", c.apiKey)
	values.Set("cx", c.searchEngineID)
	values.Set("q", query)
	values.Set("num", fmt.Sprintf("%d", c.maxResults))

	u.RawQuery = values.Encode()
	return u.String()
}
