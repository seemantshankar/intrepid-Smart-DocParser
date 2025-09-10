package analysis

import (
	"context"
	"errors"
	"testing"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repository"
	"contract-analysis-service/internal/services/llm"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockKnowledgeRepository is a mock implementation of the KnowledgeRepository interface
type MockKnowledgeRepository struct {
	mock.Mock
}

func (m *MockKnowledgeRepository) CreateKnowledge(ctx context.Context, entry *repository.KnowledgeEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockKnowledgeRepository) GetKnowledgeByID(ctx context.Context, id uuid.UUID) (*repository.KnowledgeEntry, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repository.KnowledgeEntry), args.Error(1)
}

func (m *MockKnowledgeRepository) UpdateKnowledge(ctx context.Context, entry *repository.KnowledgeEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockKnowledgeRepository) DeleteKnowledge(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockKnowledgeRepository) SearchKnowledge(ctx context.Context, query string, category string) ([]*repository.KnowledgeEntry, error) {
	args := m.Called(ctx, query, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.KnowledgeEntry), args.Error(1)
}

func (m *MockKnowledgeRepository) ListKnowledge(ctx context.Context, limit, offset int) ([]*repository.KnowledgeEntry, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*repository.KnowledgeEntry), args.Error(1)
}

func TestContractAnalysisService_AnalyzeContract(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := `
	SALE OF GOODS CONTRACT
	
	This contract is between Acme Corp (Buyer) located at 123 Main St, New York, NY, USA
	and Global Suppliers Ltd (Seller) located at 456 Oak Ave, London, UK.
	
	Total contract value: $100,000 USD
	
	Payment Terms:
	- 30% deposit upon signing: $30,000
	- 50% upon delivery: $50,000  
	- 20% upon final acceptance: $20,000
	`

	// Expected results
	expectedSummary := &models.ContractSummary{
		BuyerName:     "Acme Corp",
		BuyerAddress:  "123 Main St, New York, NY, USA",
		BuyerCountry:  "USA",
		SellerName:    "Global Suppliers Ltd",
		SellerAddress: "456 Oak Ave, London, UK",
		SellerCountry: "UK",
		GoodsNature:   "physical",
		TotalValue:    decimal.NewFromFloat(100000),
		Currency:      "USD",
		Jurisdiction:  "New York",
	}

	expectedMilestones := []models.AnalysisMilestone{
		{
			Description: "Deposit upon signing",
			Amount:      models.FlexibleString("30000"),
			Percentage:  30.0,
		},
		{
			Description: "Payment upon delivery",
			Amount:      models.FlexibleString("50000"),
			Percentage:  50.0,
		},
		{
			Description: "Final payment upon acceptance",
			Amount:      models.FlexibleString("20000"),
			Percentage:  20.0,
		},
	}

	expectedRiskAssessment := &models.AnalysisRiskAssessment{
		MissingClauses: []string{"Force majeure clause", "Dispute resolution clause"},
		Risks: []models.AnalysisIndividualRisk{
			{
				Party:          "buyer",
				Type:           "payment",
				Severity:       "medium",
				Description:    "High upfront payment percentage",
				Recommendation: "Consider reducing deposit to 20%",
			},
		},
		ComplianceScore: 0.75,
		Suggestions:     []string{"Add termination clauses", "Include warranty provisions"},
	}

	expectedAnalysis := &models.ContractAnalysis{
		Buyer:         "Acme Corp",
		BuyerAddress:  "123 Main St, New York, NY, USA",
		BuyerCountry:  "USA",
		Seller:        "Global Suppliers Ltd",
		SellerAddress: "456 Oak Ave, London, UK",
		SellerCountry: "UK",
		TotalValue:    models.FlexibleString("100000"),
		Currency:      "USD",
		Milestones:    expectedMilestones,
	}

	// Setup mock expectations
	mockLLMService.On("ExtractContractSummary", mock.Anything, "openai", contractText).
		Return(expectedSummary, nil)

	mockLLMService.On("IdentifyPaymentObligations", mock.Anything, "openai", contractText).
		Return(expectedMilestones, nil)

	mockLLMService.On("CalculatePercentageBasedPayments", float64(100000), expectedMilestones).
		Return(expectedMilestones, nil)

	mockKnowledgeRepo.On("SearchKnowledge", mock.Anything, "contract standards best practices", "").
		Return([]*repository.KnowledgeEntry{
			{
				ID:      uuid.New(),
				Title:   "General Contract Standards",
				Content: "Standard contract clauses include force majeure, dispute resolution, termination provisions...",
			},
		}, nil)

	mockLLMService.On("AssessContractRisks", mock.Anything, "openai", contractText, mock.AnythingOfType("string")).
		Return(expectedRiskAssessment, nil)

	mockLLMService.On("AnalyzeContract", mock.Anything, "openai", contractText).
		Return(expectedAnalysis, nil)

	mockLLMService.On("ValidateAnalysisConfidence", expectedAnalysis).
		Return(0.85, []string{}, nil)

	ctx := context.Background()
	result, err := service.AnalyzeContract(ctx, "contract-123", contractText)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "contract-123", result.ContractID)
	assert.Equal(t, expectedSummary.BuyerName, result.Summary.BuyerName)
	assert.Equal(t, expectedSummary.SellerName, result.Summary.SellerName)
	assert.True(t, expectedSummary.TotalValue.Equal(result.Summary.TotalValue))
	assert.Len(t, result.PaymentObligations, 3)
	assert.Equal(t, 0.75, result.RiskAssessment.ComplianceScore)
	assert.Equal(t, 0.85, result.ConfidenceScore)
	assert.Empty(t, result.ValidationIssues)
	assert.WithinDuration(t, time.Now(), result.ProcessedAt, time.Minute)

	mockLLMService.AssertExpectations(t)
	mockKnowledgeRepo.AssertExpectations(t)
}

func TestContractAnalysisService_ExtractContractSummary(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := "Sample contract text"
	expectedSummary := &models.ContractSummary{
		BuyerName:    "Test Buyer",
		SellerName:   "Test Seller",
		GoodsNature:  "services",
		TotalValue:   decimal.NewFromFloat(50000),
		Currency:     "USD",
		Jurisdiction: "California",
	}

	mockLLMService.On("ExtractContractSummary", mock.Anything, "openai", contractText).
		Return(expectedSummary, nil)

	ctx := context.Background()
	result, err := service.ExtractContractSummary(ctx, contractText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSummary.BuyerName, result.BuyerName)
	assert.Equal(t, expectedSummary.SellerName, result.SellerName)
	assert.True(t, expectedSummary.TotalValue.Equal(result.TotalValue))

	mockLLMService.AssertExpectations(t)
}

func TestContractAnalysisService_IdentifyPaymentObligations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := "Sample contract text with payment terms"
	expectedObligations := []models.AnalysisMilestone{
		{
			Description: "Initial payment",
			Amount:      models.FlexibleString("25000"),
			Percentage:  50.0,
		},
		{
			Description: "Final payment",
			Amount:      models.FlexibleString("25000"),
			Percentage:  50.0,
		},
	}

	mockLLMService.On("IdentifyPaymentObligations", mock.Anything, "openai", contractText).
		Return(expectedObligations, nil)

	ctx := context.Background()
	result, err := service.IdentifyPaymentObligations(ctx, contractText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Initial payment", result[0].Description)
	assert.Equal(t, models.FlexibleString("25000"), result[0].Amount)
	assert.Equal(t, 50.0, result[0].Percentage)

	mockLLMService.AssertExpectations(t)
}

func TestContractAnalysisService_CalculatePercentageBasedPayments(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	totalValue := 100000.0
	inputMilestones := []models.AnalysisMilestone{
		{
			Description: "First payment",
			Percentage:  30.0,
		},
		{
			Description: "Second payment",
			Amount:      models.FlexibleString("70000"),
		},
	}

	expectedMilestones := []models.AnalysisMilestone{
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
	}

	mockLLMService.On("CalculatePercentageBasedPayments", totalValue, inputMilestones).
		Return(expectedMilestones, nil)

	result, err := service.CalculatePercentageBasedPayments(totalValue, inputMilestones)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, models.FlexibleString("30000"), result[0].Amount)
	assert.Equal(t, 30.0, result[0].Percentage)
	assert.Equal(t, models.FlexibleString("70000"), result[1].Amount)
	assert.Equal(t, 70.0, result[1].Percentage)

	mockLLMService.AssertExpectations(t)
}

func TestContractAnalysisService_AssessContractRisks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := "Sample contract text"
	industries := []string{"technology", "manufacturing"}

	knowledgeEntries := []*repository.KnowledgeEntry{
		{
			ID:      uuid.New(),
			Title:   "Technology Contract Standards",
			Content: "Technology industry standards for contracts include...",
		},
		{
			ID:      uuid.New(),
			Title:   "Manufacturing Contract Standards",
			Content: "Manufacturing industry standards include...",
		},
	}

	expectedAssessment := &models.AnalysisRiskAssessment{
		MissingClauses:  []string{"IP ownership clause"},
		ComplianceScore: 0.8,
		Risks: []models.AnalysisIndividualRisk{
			{
				Party:       "seller",
				Type:        "intellectual property",
				Severity:    "high",
				Description: "IP ownership not clearly defined",
			},
		},
		Suggestions: []string{"Add IP ownership clause"},
	}

	mockKnowledgeRepo.On("SearchKnowledge", mock.Anything, "technology contract standards", "technology").
		Return([]*repository.KnowledgeEntry{knowledgeEntries[0]}, nil)

	mockKnowledgeRepo.On("SearchKnowledge", mock.Anything, "manufacturing contract standards", "manufacturing").
		Return([]*repository.KnowledgeEntry{knowledgeEntries[1]}, nil)

	mockLLMService.On("AssessContractRisks", mock.Anything, "openai", contractText, mock.AnythingOfType("string")).
		Return(expectedAssessment, nil)

	ctx := context.Background()
	result, err := service.AssessContractRisks(ctx, contractText, industries...)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.8, result.ComplianceScore)
	assert.Len(t, result.Risks, 1)
	assert.Equal(t, "intellectual property", result.Risks[0].Type)
	assert.Equal(t, "high", result.Risks[0].Severity)
	assert.Contains(t, result.MissingClauses, "IP ownership clause")

	mockLLMService.AssertExpectations(t)
	mockKnowledgeRepo.AssertExpectations(t)
}

func TestContractAnalysisService_ValidateAnalysisConfidence(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	analysis := &models.ContractAnalysis{
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
	}

	expectedConfidence := 0.9
	expectedIssues := []string{}

	mockLLMService.On("ValidateAnalysisConfidence", analysis).
		Return(expectedConfidence, expectedIssues, nil)

	confidence, issues, err := service.ValidateAnalysisConfidence(analysis)

	assert.NoError(t, err)
	assert.Equal(t, expectedConfidence, confidence)
	assert.Equal(t, expectedIssues, issues)

	mockLLMService.AssertExpectations(t)
}

func TestContractAnalysisService_AnalyzeContract_LLMErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := "Sample contract text"

	// Test summary extraction error
	mockLLMService.On("ExtractContractSummary", mock.Anything, "openai", contractText).
		Return((*models.ContractSummary)(nil), errors.New("LLM service error"))

	ctx := context.Background()
	result, err := service.AnalyzeContract(ctx, "contract-123", contractText)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to extract contract summary")

	mockLLMService.AssertExpectations(t)
}

func TestContractAnalysisService_AnalyzeContract_ValidationIssues(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockLLMService := &llm.MockLLMService{}
	mockKnowledgeRepo := &MockKnowledgeRepository{}

	service := NewContractAnalysisService(mockLLMService, mockKnowledgeRepo, logger, "openai")

	contractText := "Sample contract text"

	// Create analysis with missing fields to trigger validation issues
	incompleteAnalysis := &models.ContractAnalysis{
		Buyer:      "", // Missing buyer name
		Seller:     "Test Seller",
		TotalValue: models.FlexibleString(""),                 // Missing total value
		Currency:   "",                           // Missing currency
		Milestones: []models.AnalysisMilestone{}, // No milestones
	}

	expectedSummary := &models.ContractSummary{
		BuyerName:  "Test Buyer",
		SellerName: "Test Seller",
		TotalValue: decimal.NewFromFloat(50000),
		Currency:   "USD",
	}

	expectedRiskAssessment := &models.AnalysisRiskAssessment{
		ComplianceScore: 0.6,
		Risks:           []models.AnalysisIndividualRisk{},
	}

	expectedIssues := []string{"buyer name is missing", "total contract value is missing or zero", "currency is missing", "no payment milestones identified"}

	// Setup mocks
	mockLLMService.On("ExtractContractSummary", mock.Anything, "openai", contractText).
		Return(expectedSummary, nil)

	mockLLMService.On("IdentifyPaymentObligations", mock.Anything, "openai", contractText).
		Return([]models.AnalysisMilestone{}, nil)

	mockLLMService.On("CalculatePercentageBasedPayments", float64(50000), []models.AnalysisMilestone{}).
		Return([]models.AnalysisMilestone{}, nil)

	mockKnowledgeRepo.On("SearchKnowledge", mock.Anything, mock.AnythingOfType("string"), "").
		Return([]*repository.KnowledgeEntry{}, nil)

	mockLLMService.On("AssessContractRisks", mock.Anything, "openai", contractText, "").
		Return(expectedRiskAssessment, nil)

	mockLLMService.On("AnalyzeContract", mock.Anything, "openai", contractText).
		Return(incompleteAnalysis, nil)

	mockLLMService.On("ValidateAnalysisConfidence", incompleteAnalysis).
		Return(0.3, expectedIssues, nil)

	ctx := context.Background()
	result, err := service.AnalyzeContract(ctx, "contract-123", contractText)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.3, result.ConfidenceScore)
	assert.Len(t, result.ValidationIssues, 4)
	assert.Contains(t, result.ValidationIssues, "buyer name is missing")
	assert.Contains(t, result.ValidationIssues, "no payment milestones identified")

	mockLLMService.AssertExpectations(t)
	mockKnowledgeRepo.AssertExpectations(t)
}
