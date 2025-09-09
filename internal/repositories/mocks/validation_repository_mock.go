package mocks

import (
	"contract-analysis-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// ValidationRepository is a mock implementation of repositories.ValidationRepository
type ValidationRepository struct {
	mock.Mock
}

func (m *ValidationRepository) Create(record *models.ValidationRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *ValidationRepository) GetByID(id string) (*models.ValidationRecord, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ValidationRecord), args.Error(1)
}

func (m *ValidationRepository) GetByContractID(contractID string) ([]*models.ValidationRecord, error) {
	args := m.Called(contractID)
	return args.Get(0).([]*models.ValidationRecord), args.Error(1)
}

func (m *ValidationRepository) GetByUserID(userID string) ([]*models.ValidationRecord, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.ValidationRecord), args.Error(1)
}

func (m *ValidationRepository) Update(record *models.ValidationRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *ValidationRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *ValidationRepository) GetByType(validationType string) ([]*models.ValidationRecord, error) {
	args := m.Called(validationType)
	return args.Get(0).([]*models.ValidationRecord), args.Error(1)
}

func (m *ValidationRepository) List() ([]*models.ValidationRecord, error) {
	args := m.Called()
	return args.Get(0).([]*models.ValidationRecord), args.Error(1)
}

// ValidationAuditRepository is a mock implementation of repositories.ValidationAuditRepository
type ValidationAuditRepository struct {
	mock.Mock
}

func (m *ValidationAuditRepository) Create(log *models.ValidationAuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *ValidationAuditRepository) GetByValidationID(validationID string) ([]*models.ValidationAuditLog, error) {
	args := m.Called(validationID)
	return args.Get(0).([]*models.ValidationAuditLog), args.Error(1)
}

func (m *ValidationAuditRepository) GetByUserID(userID string) ([]*models.ValidationAuditLog, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.ValidationAuditLog), args.Error(1)
}

func (m *ValidationAuditRepository) GetByID(id string) (*models.ValidationAuditLog, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ValidationAuditLog), args.Error(1)
}

func (m *ValidationAuditRepository) List() ([]*models.ValidationAuditLog, error) {
	args := m.Called()
	return args.Get(0).([]*models.ValidationAuditLog), args.Error(1)
}

// ValidationFeedbackRepository is a mock implementation of repositories.ValidationFeedbackRepository
type ValidationFeedbackRepository struct {
	mock.Mock
}

func (m *ValidationFeedbackRepository) Create(feedback *models.ValidationFeedback) error {
	args := m.Called(feedback)
	return args.Error(0)
}

func (m *ValidationFeedbackRepository) GetByValidationID(validationID string) ([]*models.ValidationFeedback, error) {
	args := m.Called(validationID)
	return args.Get(0).([]*models.ValidationFeedback), args.Error(1)
}

func (m *ValidationFeedbackRepository) GetByUserID(userID string) ([]*models.ValidationFeedback, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.ValidationFeedback), args.Error(1)
}

func (m *ValidationFeedbackRepository) GetByType(feedbackType string) ([]*models.ValidationFeedback, error) {
	args := m.Called(feedbackType)
	return args.Get(0).([]*models.ValidationFeedback), args.Error(1)
}

func (m *ValidationFeedbackRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *ValidationFeedbackRepository) GetByID(id string) (*models.ValidationFeedback, error) {
	args := m.Called(id)
	return args.Get(0).(*models.ValidationFeedback), args.Error(1)
}

func (m *ValidationFeedbackRepository) List() ([]*models.ValidationFeedback, error) {
	args := m.Called()
	return args.Get(0).([]*models.ValidationFeedback), args.Error(1)
}

func (m *ValidationFeedbackRepository) Update(feedback *models.ValidationFeedback) error {
	args := m.Called(feedback)
	return args.Error(0)
}