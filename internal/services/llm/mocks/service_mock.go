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
