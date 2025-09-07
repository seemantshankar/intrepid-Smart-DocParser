package mocks

import (
	"context"

	"contract-analysis-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// Service is a mock implementation of the Service interface.
type Service struct {
	mock.Mock
}

// ValidateContract mocks the ValidateContract method.
func (m *Service) ValidateContract(ctx context.Context, documentText string) (*models.ValidationResult, error) {
	args := m.Called(ctx, documentText)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ValidationResult), args.Error(1)
}
