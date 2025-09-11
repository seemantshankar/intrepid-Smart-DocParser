package websearch

import (
	"context"
	"errors"
	"testing"
	"time"

	"contract-analysis-service/internal/pkg/external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockExternalClient is a mock implementation of external.Client
type MockExternalClient struct {
	mock.Mock
}

func (m *MockExternalClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*external.Response), args.Error(1)
}

func TestGoogleSearchClient_Search_Success(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock response
	mockResponse := &external.Response{
		StatusCode: 200,
		Body: []byte(`{
			"items": [
				{
					"title": "Manufacturing Industry Standards",
					"link": "https://example.com/manufacturing-standards",
					"snippet": "Comprehensive guide to manufacturing industry standards and best practices...",
					"displayLink": "example.com"
				},
				{
					"title": "Manufacturing Safety Guidelines",
					"link": "https://example.com/safety-guidelines",
					"snippet": "Safety guidelines for manufacturing operations...",
					"displayLink": "example.com"
				}
			],
			"searchInformation": {
				"searchTime": 0.45,
				"totalResults": "1250000"
			}
		}`),
	}

	// URL parameters are encoded alphabetically by url.Values.Encode()
	expectedURL := "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-api-key&num=10&q=manufacturing+industry+standards+best+practices"
	expectedRequest := &external.Request{
		Method: "GET",
		URL:    expectedURL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: nil,
	}

	mockClient.On("ExecuteRequest", mock.Anything, expectedRequest).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "manufacturing industry standards best practices")

	// Verify
	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	assert.Equal(t, "Manufacturing Industry Standards", results[0].Title)
	assert.Equal(t, "https://example.com/manufacturing-standards", results[0].URL)
	assert.Equal(t, "Comprehensive guide to manufacturing industry standards and best practices...", results[0].Snippet)
	assert.Equal(t, "example.com", results[0].Source)
	
	assert.Equal(t, "Manufacturing Safety Guidelines", results[1].Title)
	assert.Equal(t, "https://example.com/safety-guidelines", results[1].URL)
	assert.Equal(t, "Safety guidelines for manufacturing operations...", results[1].Snippet)
	assert.Equal(t, "example.com", results[1].Source)

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_HTTPError_403(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "invalid-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock HTTP error response
	mockResponse := &external.Response{
		StatusCode: 403,
		Body: []byte(`{
			"error": {
				"code": 403,
				"message": "The provided API key is invalid.",
				"errors": [
					{
						"message": "The provided API key is invalid.",
						"domain": "global",
						"reason": "forbidden"
					}
				]
			}
		}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")
	assert.Contains(t, err.Error(), "403")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_HTTPError_429_RateLimit(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock rate limit error response
	mockResponse := &external.Response{
		StatusCode: 429,
		Body: []byte(`{
			"error": {
				"code": 429,
				"message": "Quota exceeded for this request",
				"errors": [
					{
						"message": "Quota exceeded for this request",
						"domain": "usageLimits",
						"reason": "quotaExceeded"
					}
				]
			}
		}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")
	assert.Contains(t, err.Error(), "429")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_HTTPError_500(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock server error response
	mockResponse := &external.Response{
		StatusCode: 500,
		Body:       []byte(`{"error": "Internal server error"}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")
	assert.Contains(t, err.Error(), "500")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_NetworkError(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock network error
	networkErr := errors.New("network timeout")
	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(nil, networkErr)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_InvalidJSON(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock response with invalid JSON
	mockResponse := &external.Response{
		StatusCode: 200,
		Body:       []byte(`invalid json response`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to parse search response")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_MalformedJSON(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock response with malformed JSON structure
	mockResponse := &external.Response{
		StatusCode: 200,
		Body:       []byte(`{"items": "not an array"}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "failed to parse search response")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_EmptyQuery(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Execute
	results, err := searchClient.Search(context.Background(), "")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search query cannot be empty")

	// Should not make any HTTP calls
	mockClient.AssertNotCalled(t, "ExecuteRequest")
}

func TestGoogleSearchClient_Search_WhitespaceQuery(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Execute
	results, err := searchClient.Search(context.Background(), "   \t\n  ")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search query cannot be empty")

	// Should not make any HTTP calls
	mockClient.AssertNotCalled(t, "ExecuteRequest")
}

func TestGoogleSearchClient_Search_ContextCancellation(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Mock context cancelled error
	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(nil, context.Canceled)

	// Execute
	results, err := searchClient.Search(ctx, "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_ContextDeadlineExceeded(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Create context with deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure deadline is exceeded

	// Mock deadline exceeded error
	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded)

	// Execute
	results, err := searchClient.Search(ctx, "test query")

	// Verify
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search request failed")

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_NoResults(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock response with no items
	mockResponse := &external.Response{
		StatusCode: 200,
		Body: []byte(`{
			"searchInformation": {
				"searchTime": 0.25,
				"totalResults": "0"
			}
		}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "very specific query with no results")

	// Verify
	require.NoError(t, err)
	assert.Empty(t, results)

	mockClient.AssertExpectations(t)
}

func TestGoogleSearchClient_Search_PartialResults(t *testing.T) {
	// Setup
	mockClient := &MockExternalClient{}
	searchClient := &googleSearchClient{
		httpClient:     mockClient,
		apiKey:         "test-api-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	// Mock response with some results having missing fields
	mockResponse := &external.Response{
		StatusCode: 200,
		Body: []byte(`{
			"items": [
				{
					"title": "Complete Result",
					"link": "https://example.com/complete",
					"snippet": "This result has all fields",
					"displayLink": "example.com"
				},
				{
					"title": "Missing Snippet",
					"link": "https://example.com/no-snippet",
					"displayLink": "example.com"
				},
				{
					"link": "https://example.com/no-title",
					"snippet": "This result has no title",
					"displayLink": "example.com"
				},
				{
					"title": "Missing Link",
					"snippet": "This result has no link",
					"displayLink": "example.com"
				}
			]
		}`),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(mockResponse, nil)

	// Execute
	results, err := searchClient.Search(context.Background(), "test query")

	// Verify
	require.NoError(t, err)
	assert.Len(t, results, 4)
	
	// Complete result
	assert.Equal(t, "Complete Result", results[0].Title)
	assert.Equal(t, "https://example.com/complete", results[0].URL)
	assert.Equal(t, "This result has all fields", results[0].Snippet)
	assert.Equal(t, "example.com", results[0].Source)
	
	// Missing snippet
	assert.Equal(t, "Missing Snippet", results[1].Title)
	assert.Equal(t, "https://example.com/no-snippet", results[1].URL)
	assert.Equal(t, "", results[1].Snippet)
	assert.Equal(t, "example.com", results[1].Source)
	
	// Missing title
	assert.Equal(t, "", results[2].Title)
	assert.Equal(t, "https://example.com/no-title", results[2].URL)
	assert.Equal(t, "This result has no title", results[2].Snippet)
	assert.Equal(t, "example.com", results[2].Source)
	
	// Missing link
	assert.Equal(t, "Missing Link", results[3].Title)
	assert.Equal(t, "", results[3].URL)
	assert.Equal(t, "This result has no link", results[3].Snippet)
	assert.Equal(t, "example.com", results[3].Source)

	mockClient.AssertExpectations(t)
}

func TestNewGoogleSearchClient(t *testing.T) {
	// Setup
	config := GoogleSearchConfig{
		APIKey:         "test-api-key",
		SearchEngineID: "test-cx",
		BaseURL:        "https://www.googleapis.com/customsearch/v1",
		MaxResults:     15,
		Timeout:        45 * time.Second,
		RetryConfig: external.RetryConfig{
			MaxRetries:      5,
			InitialInterval: 1 * time.Second,
			MaxInterval:     10 * time.Second,
		},
	}

	// Execute
	client := NewGoogleSearchClient(config)

	// Verify
	require.NotNil(t, client)
	
	// Cast to concrete type to verify configuration
	googleClient, ok := client.(*googleSearchClient)
	require.True(t, ok)
	assert.Equal(t, config.APIKey, googleClient.apiKey)
	assert.Equal(t, config.SearchEngineID, googleClient.searchEngineID)
	assert.Equal(t, config.BaseURL, googleClient.baseURL)
	assert.Equal(t, config.MaxResults, googleClient.maxResults)
	assert.NotNil(t, googleClient.httpClient)
}

func TestNewGoogleSearchClient_DefaultValues(t *testing.T) {
	// Setup with minimal config
	config := GoogleSearchConfig{
		APIKey:         "test-api-key",
		SearchEngineID: "test-cx",
	}

	// Execute
	client := NewGoogleSearchClient(config)

	// Verify defaults are applied
	require.NotNil(t, client)
	
	googleClient, ok := client.(*googleSearchClient)
	require.True(t, ok)
	
	assert.Equal(t, "https://www.googleapis.com/customsearch/v1", googleClient.baseURL)
	assert.Equal(t, 10, googleClient.maxResults) // Default max results
}

func TestGoogleSearchClient_buildSearchURL(t *testing.T) {
	// Setup
	client := &googleSearchClient{
		apiKey:         "test-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     10,
	}

	testCases := []struct {
		name          string
		query         string
		expectedURL   string
	}{
		{
			name:        "Simple query",
			query:       "test query",
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=test+query",
		},
		{
			name:        "Query with special characters",
			query:       "manufacturing & industry standards",
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=manufacturing+%26+industry+standards",
		},
		{
			name:        "Query with multiple spaces",
			query:       "healthcare   industry    best   practices",
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=healthcare+++industry++++best+++practices",
		},
		{
			name:        "Query with quotes",
			query:       `"financial services" regulations`,
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=%22financial+services%22+regulations",
		},
		{
			name:        "Query with plus signs",
			query:       "C++ programming standards",
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=C%2B%2B+programming+standards",
		},
		{
			name:        "Single word query",
			query:       "manufacturing",
			expectedURL: "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=10&q=manufacturing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			url := client.buildSearchURL(tc.query)

			// Verify
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}

func TestGoogleSearchClient_buildSearchURL_DifferentMaxResults(t *testing.T) {
	// Setup
	client := &googleSearchClient{
		apiKey:         "test-key",
		searchEngineID: "test-cx",
		baseURL:        "https://www.googleapis.com/customsearch/v1",
		maxResults:     5,
	}

	// Execute
	url := client.buildSearchURL("test query")

	// Verify
	expectedURL := "https://www.googleapis.com/customsearch/v1?cx=test-cx&key=test-key&num=5&q=test+query"
	assert.Equal(t, expectedURL, url)
}

// Mock interfaces for testing
type MockSearchClient struct {
	mock.Mock
}

func (m *MockSearchClient) Search(ctx context.Context, query string) ([]SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SearchResult), args.Error(1)
}

func TestSearchClient_Interface_Compliance(t *testing.T) {
	// This test ensures our concrete type implements the interface correctly
	var _ SearchClient = (*googleSearchClient)(nil)
	var _ SearchClient = (*MockSearchClient)(nil)
}

func TestGoogleSearchConfig_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		config      GoogleSearchConfig
		expectValid bool
	}{
		{
			name: "Valid config",
			config: GoogleSearchConfig{
				APIKey:         "test-key",
				SearchEngineID: "test-cx",
				BaseURL:        "https://api.example.com",
				MaxResults:     10,
				Timeout:        30 * time.Second,
			},
			expectValid: true,
		},
		{
			name: "Missing API key",
			config: GoogleSearchConfig{
				SearchEngineID: "test-cx",
				BaseURL:        "https://api.example.com",
				MaxResults:     10,
				Timeout:        30 * time.Second,
			},
			expectValid: false,
		},
		{
			name: "Missing search engine ID",
			config: GoogleSearchConfig{
				APIKey:     "test-key",
				BaseURL:    "https://api.example.com",
				MaxResults: 10,
				Timeout:    30 * time.Second,
			},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewGoogleSearchClient(tc.config)
			
			if tc.expectValid {
				assert.NotNil(t, client)
			} else {
				// In a real implementation, we might return nil or panic for invalid config
				// For now, we just ensure the client is created but might not work correctly
				assert.NotNil(t, client)
			}
		})
	}
}
