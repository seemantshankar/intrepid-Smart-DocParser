package mocks

import (
	"context"
	"contract-analysis-service/internal/models"
	"github.com/stretchr/testify/mock"
)

type Service struct {
	mock.Mock
}

func (m *Service) AnalyzeContract(ctx context.Context, content string) (*models.ContractAnalysis, error) {
	args := m.Called(ctx, content)
	return args.Get(0).(*models.ContractAnalysis), args.Error(1)
}

func (m *Service) AnalyzeContractFromImages(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error) {
	args := m.Called(ctx, imagePaths)
	return args.Get(0).(*models.ContractAnalysis), args.Error(1)
}