package llm

import (
	"context"
	"encoding/json"
	"testing"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockClient is a mock implementation of the external.Client interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) ExecuteRequest(ctx context.Context, req *external.Request) (*external.Response, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*external.Response), args.Error(1)
}

func TestLLMService_ExtractContractSummary(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)
	mockClient := &MockClient{}
	service.AddClient("test-provider", mockClient)

	contractText := `
	CONTRACT BETWEEN PARTIES
	Buyer: TechCorp Inc, 123 Silicon Valley, California, USA
	Seller: Manufacturing Ltd, 456 Industrial Ave, Ontario, Canada
	Total Value: $150,000 CAD
	Goods: Custom software development services
	Jurisdiction: Ontario, Canada
	`

	expectedResponse := map[string]interface{}{
		"buyer_name":     "TechCorp Inc",
		"buyer_address":  "123 Silicon Valley, California, USA",
		"buyer_country":  "USA",
		"seller_name":    "Manufacturing Ltd",
		"seller_address": "456 Industrial Ave, Ontario, Canada",
		"seller_country": "Canada",
		"goods_nature":   "services",
		"total_value":    150000,
		"currency":       "CAD",
		"jurisdiction":   "Ontario, Canada",
	}

	openAIResponse := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": toJSONString(expectedResponse),
				},
			},
		},
	}

	mockResponse := &external.Response{
		StatusCode: 200,
		Body:       []byte(toJSONString(openAIResponse)),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.MatchedBy(func(req *external.Request) bool {
		return req.Method == "POST" && req.URL == "/chat/completions"
	})).Return(mockResponse, nil)

	ctx := context.Background()
	result, err := service.ExtractContractSummary(ctx, "test-provider", contractText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "TechCorp Inc", result.BuyerName)
	assert.Equal(t, "123 Silicon Valley, California, USA", result.BuyerAddress)
	assert.Equal(t, "USA", result.BuyerCountry)
	assert.Equal(t, "Manufacturing Ltd", result.SellerName)
	assert.Equal(t, "456 Industrial Ave, Ontario, Canada", result.SellerAddress)
	assert.Equal(t, "Canada", result.SellerCountry)
	assert.Equal(t, "services", result.GoodsNature)
	assert.True(t, result.TotalValue.Equal(decimal.NewFromFloat(150000)))
	assert.Equal(t, "CAD", result.Currency)
	assert.Equal(t, "Ontario, Canada", result.Jurisdiction)

	mockClient.AssertExpectations(t)
}

func TestLLMService_IdentifyPaymentObligations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)
	mockClient := &MockClient{}
	service.AddClient("test-provider", mockClient)

	contractText := `
	PAYMENT TERMS:
	1. 25% deposit upon contract signing
	2. 50% progress payment upon delivery of Phase 1
	3. 25% final payment upon project completion
	Total contract value: $200,000
	`

	expectedMilestones := []map[string]interface{}{
		{
			"description":       "Deposit upon contract signing",
			"amount":            50000,
			"percentage":        25.0,
			"trigger_condition": "contract signing",
		},
		{
			"description":       "Progress payment upon delivery of Phase 1",
			"amount":            100000,
			"percentage":        50.0,
			"trigger_condition": "delivery of Phase 1",
		},
		{
			"description":       "Final payment upon project completion",
			"amount":            50000,
			"percentage":        25.0,
			"trigger_condition": "project completion",
		},
	}

	openAIResponse := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": toJSONString(expectedMilestones),
				},
			},
		},
	}

	mockResponse := &external.Response{
		StatusCode: 200,
		Body:       []byte(toJSONString(openAIResponse)),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.MatchedBy(func(req *external.Request) bool {
		return req.Method == "POST" && req.URL == "/chat/completions"
	})).Return(mockResponse, nil)

	ctx := context.Background()
	result, err := service.IdentifyPaymentObligations(ctx, "test-provider", contractText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Check first milestone
	assert.Equal(t, "Deposit upon contract signing", result[0].Description)
	assert.Equal(t, models.FlexibleString("50000"), result[0].Amount)
	assert.Equal(t, 25.0, result[0].Percentage)

	// Check second milestone
	assert.Equal(t, "Progress payment upon delivery of Phase 1", result[1].Description)
	assert.Equal(t, models.FlexibleString("100000"), result[1].Amount)
	assert.Equal(t, 50.0, result[1].Percentage)

	// Check third milestone
	assert.Equal(t, "Final payment upon project completion", result[2].Description)
	assert.Equal(t, models.FlexibleString("50000"), result[2].Amount)
	assert.Equal(t, 25.0, result[2].Percentage)

	mockClient.AssertExpectations(t)
}

func TestLLMService_CalculatePercentageBasedPayments(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)

	totalValue := 100000.0

	tests := []struct {
		name        string
		milestones  []models.AnalysisMilestone
		expected    []models.AnalysisMilestone
		expectError bool
	}{
		{
			name: "Calculate amounts from percentages",
			milestones: []models.AnalysisMilestone{
				{
					Description: "First payment",
					Percentage:  30.0,
				},
				{
					Description: "Second payment",
					Percentage:  70.0,
				},
			},
			expected: []models.AnalysisMilestone{
				{
					Description: "First payment",
					Amount:      models.FlexibleString("30000"),
					Percentage:  30.0,
				},
				{
					Description: "Second payment",
					Amount:      models.FlexibleString("70000"),
					Percentage:  70.0,
				},
			},
			expectError: false,
		},
		{
			name: "Calculate percentages from amounts",
			milestones: []models.AnalysisMilestone{
				{
					Description: "First payment",
					Amount:      models.FlexibleString("40000"),
				},
				{
					Description: "Second payment",
					Amount:      models.FlexibleString("60000"),
				},
			},
			expected: []models.AnalysisMilestone{
				{
					Description: "First payment",
					Amount:      models.FlexibleString("40000"),
					Percentage:  40.0,
				},
				{
					Description: "Second payment",
					Amount:      models.FlexibleString("60000"),
					Percentage:  60.0,
				},
			},
			expectError: false,
		},
		{
			name:        "Error with zero total value",
			milestones:  []models.AnalysisMilestone{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testTotalValue float64
			if tt.expectError && tt.name == "Error with zero total value" {
				testTotalValue = 0
			} else {
				testTotalValue = totalValue
			}

			result, err := service.CalculatePercentageBasedPayments(testTotalValue, tt.milestones)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, len(tt.expected))

				for i, expected := range tt.expected {
					assert.Equal(t, expected.Description, result[i].Description)
					assert.Equal(t, expected.Amount, result[i].Amount)
					assert.Equal(t, expected.Percentage, result[i].Percentage)
				}
			}
		})
	}
}

func TestLLMService_AssessContractRisks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)
	mockClient := &MockClient{}
	service.AddClient("test-provider", mockClient)

	contractText := "Sample contract text"
	industryStandards := "Industry standards for technology contracts"

	expectedRiskAssessment := map[string]interface{}{
		"missing_clauses": []string{"Force majeure clause", "Data protection clause"},
		"risks": []map[string]interface{}{
			{
				"party":          "buyer",
				"type":           "data security",
				"severity":       "high",
				"description":    "No data protection measures specified",
				"recommendation": "Add comprehensive data protection clause",
			},
			{
				"party":          "seller",
				"type":           "liability",
				"severity":       "medium",
				"description":    "Unlimited liability exposure",
				"recommendation": "Cap liability to contract value",
			},
		},
		"compliance_score": 0.65,
		"suggestions":      []string{"Add termination clauses", "Include IP ownership provisions"},
	}

	openAIResponse := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": toJSONString(expectedRiskAssessment),
				},
			},
		},
	}

	mockResponse := &external.Response{
		StatusCode: 200,
		Body:       []byte(toJSONString(openAIResponse)),
	}

	mockClient.On("ExecuteRequest", mock.Anything, mock.MatchedBy(func(req *external.Request) bool {
		return req.Method == "POST" && req.URL == "/chat/completions"
	})).Return(mockResponse, nil)

	ctx := context.Background()
	result, err := service.AssessContractRisks(ctx, "test-provider", contractText, industryStandards)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.MissingClauses, 2)
	assert.Contains(t, result.MissingClauses, "Force majeure clause")
	assert.Contains(t, result.MissingClauses, "Data protection clause")

	assert.Len(t, result.Risks, 2)
	assert.Equal(t, "buyer", result.Risks[0].Party)
	assert.Equal(t, "data security", result.Risks[0].Type)
	assert.Equal(t, "high", result.Risks[0].Severity)

	assert.Equal(t, "seller", result.Risks[1].Party)
	assert.Equal(t, "liability", result.Risks[1].Type)
	assert.Equal(t, "medium", result.Risks[1].Severity)

	assert.Equal(t, 0.65, result.ComplianceScore)
	assert.Len(t, result.Suggestions, 2)

	mockClient.AssertExpectations(t)
}

func TestLLMService_ValidateAnalysisConfidence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)

	tests := []struct {
		name               string
		analysis           *models.ContractAnalysis
		expectedConfidence float64
		expectedIssues     []string
	}{
		{
			name: "High confidence - complete analysis",
			analysis: &models.ContractAnalysis{
				Buyer:      "Test Buyer",
				Seller:     "Test Seller",
				TotalValue: models.FlexibleString("100000"),
				Currency:   "USD",
				Milestones: []models.AnalysisMilestone{
					{
						Description: "Payment 1",
						Amount:      models.FlexibleString("50000"),
						Percentage:  50.0,
					},
					{
						Description: "Payment 2",
						Amount:      models.FlexibleString("50000"),
						Percentage:  50.0,
					},
				},
			},
			expectedConfidence: 1.0,
			expectedIssues:     []string{},
		},
		{
			name: "Medium confidence - missing buyer",
			analysis: &models.ContractAnalysis{
				Buyer:      "", // Missing buyer
				Seller:     "Test Seller",
				TotalValue: models.FlexibleString("100000"),
				Currency:   "USD",
				Milestones: []models.AnalysisMilestone{
					{
						Description: "Payment 1",
						Amount:      models.FlexibleString("100000"),
						Percentage:  100.0,
					},
				},
			},
			expectedConfidence: 0.8,
			expectedIssues:     []string{"buyer name is missing"},
		},
		{
			name: "Low confidence - multiple issues",
			analysis: &models.ContractAnalysis{
				Buyer:      "",                           // Missing buyer
				Seller:     "",                           // Missing seller
				TotalValue: models.FlexibleString("0"),                 // Missing total value
				Currency:   "",                           // Missing currency
				Milestones: []models.AnalysisMilestone{}, // No milestones
			},
			expectedConfidence: 0.0,
			expectedIssues: []string{
				"buyer name is missing",
				"seller name is missing",
				"total contract value is missing or zero",
				"currency is missing",
				"no payment milestones identified",
			},
		},
		{
			name: "Percentage sum issue",
			analysis: &models.ContractAnalysis{
				Buyer:      "Test Buyer",
				Seller:     "Test Seller",
				TotalValue: models.FlexibleString("100000"),
				Currency:   "USD",
				Milestones: []models.AnalysisMilestone{
					{
						Description: "Payment 1",
						Amount:      models.FlexibleString("30000"),
						Percentage:  30.0,
					},
					{
						Description: "Payment 2",
						Amount:      models.FlexibleString("40000"),
						Percentage:  40.0,
					},
					// Total percentage = 70%, not 100%
				},
			},
			expectedConfidence: 0.85, // 1.0 - 0.15 for percentage issue
			expectedIssues:     []string{"milestone percentages sum to 70.00%, not 100%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, issues, err := service.ValidateAnalysisConfidence(tt.analysis)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedConfidence, confidence)
			assert.Equal(t, len(tt.expectedIssues), len(issues))

			for _, expectedIssue := range tt.expectedIssues {
				assert.Contains(t, issues, expectedIssue)
			}
		})
	}
}

func TestLLMService_ValidateAnalysisConfidence_NilAnalysis(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)

	confidence, issues, err := service.ValidateAnalysisConfidence(nil)

	assert.Error(t, err)
	assert.Equal(t, 0.0, confidence)
	assert.Contains(t, issues, "analysis is nil")
	assert.Contains(t, err.Error(), "analysis cannot be nil")
}

func TestLLMService_UnsupportedProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewLLMService(logger)

	ctx := context.Background()
	contractText := "Sample contract text"

	// Test ExtractContractSummary with unsupported provider
	summary, err := service.ExtractContractSummary(ctx, "unsupported-provider", contractText)
	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "unsupported LLM provider")

	// Test IdentifyPaymentObligations with unsupported provider
	obligations, err := service.IdentifyPaymentObligations(ctx, "unsupported-provider", contractText)
	assert.Error(t, err)
	assert.Nil(t, obligations)
	assert.Contains(t, err.Error(), "unsupported LLM provider")

	// Test AssessContractRisks with unsupported provider
	risks, err := service.AssessContractRisks(ctx, "unsupported-provider", contractText, "")
	assert.Error(t, err)
	assert.Nil(t, risks)
	assert.Contains(t, err.Error(), "unsupported LLM provider")
}

// Helper function to convert objects to JSON strings
func toJSONString(obj interface{}) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}
