package repositories

import (
	"contract-analysis-service/internal/models"
	"errors"
)

// Common repository errors
var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")
	// ErrStaleData is returned when trying to update a stale record
	ErrStaleData = errors.New("stale data: the record has been updated by another process")
)

type ContractRepository interface {
	Create(c *models.Contract) error
	GetByID(id string) (*models.Contract, error)
	Update(c *models.Contract) error
	Delete(id string) error
	List() ([]*models.Contract, error)
}

type MilestoneRepository interface {
	Create(m *models.Milestone) error
	GetByID(id string) (*models.Milestone, error)
	Update(m *models.Milestone) error
	Delete(id string) error
	List() ([]*models.Milestone, error)
}

type RiskAssessmentRepository interface {
	Create(r *models.RiskAssessment) error
	GetByID(id string) (*models.RiskAssessment, error)
	Update(r *models.RiskAssessment) error
	Delete(id string) error
	List() ([]*models.RiskAssessment, error)
}

type KnowledgeEntryRepository interface {
	Create(k *models.KnowledgeEntry) error
	GetByID(id string) (*models.KnowledgeEntry, error)
	Update(k *models.KnowledgeEntry) error
	Delete(id string) error
	List() ([]*models.KnowledgeEntry, error)
	GetByIndustry(industry string) ([]*models.KnowledgeEntry, error)
}

type ValidationRepository interface {
	Create(v *models.ValidationRecord) error
	GetByID(id string) (*models.ValidationRecord, error)
	Update(v *models.ValidationRecord) error
	Delete(id string) error
	List() ([]*models.ValidationRecord, error)
	GetByContractID(contractID string) ([]*models.ValidationRecord, error)
	GetByUserID(userID string) ([]*models.ValidationRecord, error)
	GetByType(validationType string) ([]*models.ValidationRecord, error)
}

type ValidationAuditRepository interface {
	Create(a *models.ValidationAuditLog) error
	GetByID(id string) (*models.ValidationAuditLog, error)
	List() ([]*models.ValidationAuditLog, error)
	GetByValidationID(validationID string) ([]*models.ValidationAuditLog, error)
	GetByUserID(userID string) ([]*models.ValidationAuditLog, error)
}

type ValidationFeedbackRepository interface {
	Create(f *models.ValidationFeedback) error
	GetByID(id string) (*models.ValidationFeedback, error)
	Update(f *models.ValidationFeedback) error
	Delete(id string) error
	List() ([]*models.ValidationFeedback, error)
	GetByValidationID(validationID string) ([]*models.ValidationFeedback, error)
	GetByUserID(userID string) ([]*models.ValidationFeedback, error)
}

type ClassificationRepository interface {
	Create(c *models.ClassificationRecord) error
	GetByID(id string) (*models.ClassificationRecord, error)
	Update(c *models.ClassificationRecord) error
	Delete(id string) error
	List() ([]*models.ClassificationRecord, error)
	GetByContractID(contractID string) ([]*models.ClassificationRecord, error)
}
