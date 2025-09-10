package mocks

import (
	"contract-analysis-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// ClassificationRepositoryMock is a mock implementation of ClassificationRepository.
type ClassificationRepositoryMock struct {
	mock.Mock
}

// Create mocks the Create method.
func (m *ClassificationRepositoryMock) Create(c *models.ClassificationRecord) error {
	args := m.Called(c)
	return args.Error(0)
}

// GetByID mocks the GetByID method.
func (m *ClassificationRepositoryMock) GetByID(id string) (*models.ClassificationRecord, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ClassificationRecord), args.Error(1)
}

// Update mocks the Update method.
func (m *ClassificationRepositoryMock) Update(c *models.ClassificationRecord) error {
	args := m.Called(c)
	return args.Error(0)
}

// Delete mocks the Delete method.
func (m *ClassificationRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// List mocks the List method.
func (m *ClassificationRepositoryMock) List() ([]*models.ClassificationRecord, error) {
	args := m.Called()
	return args.Get(0).([]*models.ClassificationRecord), args.Error(1)
}

// GetByContractID mocks the GetByContractID method.
func (m *ClassificationRepositoryMock) GetByContractID(contractID string) ([]*models.ClassificationRecord, error) {
	args := m.Called(contractID)
	return args.Get(0).([]*models.ClassificationRecord), args.Error(1)
}
