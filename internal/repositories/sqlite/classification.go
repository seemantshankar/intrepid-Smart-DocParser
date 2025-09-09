package sqlite

import (
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repositories"
	"gorm.io/gorm"
)

type classificationRepository struct {
	db *gorm.DB
}

// NewClassificationRepository creates a new classification repository.
func NewClassificationRepository(db *gorm.DB) repositories.ClassificationRepository {
	return &classificationRepository{db: db}
}

// Create stores a new classification record.
func (r *classificationRepository) Create(c *models.ClassificationRecord) error {
	return r.db.Create(c).Error
}

// GetByID retrieves a classification record by ID.
func (r *classificationRepository) GetByID(id string) (*models.ClassificationRecord, error) {
	var record models.ClassificationRecord
	err := r.db.First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update updates an existing classification record.
func (r *classificationRepository) Update(c *models.ClassificationRecord) error {
	return r.db.Save(c).Error
}

// Delete removes a classification record by ID.
func (r *classificationRepository) Delete(id string) error {
	return r.db.Delete(&models.ClassificationRecord{}, "id = ?", id).Error
}

// List retrieves all classification records.
func (r *classificationRepository) List() ([]*models.ClassificationRecord, error) {
	var records []*models.ClassificationRecord
	err := r.db.Find(&records).Error
	return records, err
}

// GetByContractID retrieves all classification records for a specific contract.
func (r *classificationRepository) GetByContractID(contractID string) ([]*models.ClassificationRecord, error) {
	var records []*models.ClassificationRecord
	err := r.db.Where("contract_id = ?", contractID).Find(&records).Error
	return records, err
}