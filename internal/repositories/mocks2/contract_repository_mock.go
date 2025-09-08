package mocks

import (
	"contract-analysis-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// ContractRepository is a mock implementation of the ContractRepository interface.
type ContractRepository struct {
	mock.Mock
}

// Create mocks the Create method.
func (m *ContractRepository) Create(c *models.Contract) error {
	args := m.Called(c)
	return args.Error(0)
}

// GetByID mocks the GetByID method.
func (m *ContractRepository) GetByID(id string) (*models.Contract, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contract), args.Error(1)
}

// Update mocks the Update method.
func (m *ContractRepository) Update(c *models.Contract) error {
	args := m.Called(c)
	return args.Error(0)
}

// Delete mocks the Delete method.
func (m *ContractRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// List mocks the List method.
func (m *ContractRepository) List() ([]*models.Contract, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Contract), args.Error(1)
}
