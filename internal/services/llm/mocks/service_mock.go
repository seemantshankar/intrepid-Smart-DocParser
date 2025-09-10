package mocks

import (
	"context"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"github.com/stretchr/testify/mock"
)

// Service is a mock implementation of the llm.Service interface.
type Service struct {
	mock.Mock
}

// AnalyzeContract mocks the AnalyzeContract method.
func (m *Service) AnalyzeContract(ctx context.Context, provider, contractText string) (*models.ContractAnalysis, error) {
	args := m.Called(ctx, provider, contractText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ContractAnalysis), args.Error(1)
}

// AddClient mocks the AddClient method.
func (m *Service) AddClient(provider string, client external.Client) {
	m.Called(provider, client)
}

// ExecuteRequest mocks the ExecuteRequest method.
func (m *Service) ExecuteRequest(ctx context.Context, provider string, req *external.Request) (*external.Response, error) {
	args := m.Called(ctx, provider, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*external.Response), args.Error(1)
}

// ExtractContractSummary mocks the ExtractContractSummary method
func (m *Service) ExtractContractSummary(ctx context.Context, provider, contractText string) (*models.ContractSummary, error) {
	args := m.Called(ctx, provider, contractText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ContractSummary), args.Error(1)
}

// IdentifyPaymentObligations mocks the IdentifyPaymentObligations method
func (m *Service) IdentifyPaymentObligations(ctx context.Context, provider, contractText string) ([]models.AnalysisMilestone, error) {
	args := m.Called(ctx, provider, contractText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AnalysisMilestone), args.Error(1)
}

// CalculatePercentageBasedPayments mocks the CalculatePercentageBasedPayments method
func (m *Service) CalculatePercentageBasedPayments(totalValue float64, milestones []models.AnalysisMilestone) ([]models.AnalysisMilestone, error) {
	args := m.Called(totalValue, milestones)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AnalysisMilestone), args.Error(1)
}

// AssessContractRisks mocks the AssessContractRisks method
func (m *Service) AssessContractRisks(ctx context.Context, provider, contractText, industryStandards string) (*models.AnalysisRiskAssessment, error) {
	args := m.Called(ctx, provider, contractText, industryStandards)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalysisRiskAssessment), args.Error(1)
}

// ValidateAnalysisConfidence mocks the ValidateAnalysisConfidence method
func (m *Service) ValidateAnalysisConfidence(analysis *models.ContractAnalysis) (float64, []string, error) {
	args := m.Called(analysis)
	if args.Get(1) == nil {
		return args.Get(0).(float64), nil, args.Error(2)
	}
	return args.Get(0).(float64), args.Get(1).([]string), args.Error(2)
}
