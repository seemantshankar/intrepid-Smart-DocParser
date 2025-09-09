package sqlite

import (
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repositories"
	"gorm.io/gorm"
)

// validationRepository implements ValidationRepository interface
type validationRepository struct {
	db *gorm.DB
}

// NewValidationRepository creates a new validation repository
func NewValidationRepository(db *gorm.DB) repositories.ValidationRepository {
	return &validationRepository{db: db}
}

func (r *validationRepository) Create(v *models.ValidationRecord) error {
	return r.db.Create(v).Error
}

func (r *validationRepository) GetByID(id string) (*models.ValidationRecord, error) {
	var validation models.ValidationRecord
	err := r.db.Where("id = ?", id).First(&validation).Error
	if err != nil {
		return nil, err
	}
	return &validation, nil
}

func (r *validationRepository) Update(v *models.ValidationRecord) error {
	return r.db.Save(v).Error
}

func (r *validationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.ValidationRecord{}).Error
}

func (r *validationRepository) List() ([]*models.ValidationRecord, error) {
	var validations []*models.ValidationRecord
	err := r.db.Find(&validations).Error
	return validations, err
}

func (r *validationRepository) GetByContractID(contractID string) ([]*models.ValidationRecord, error) {
	var validations []*models.ValidationRecord
	err := r.db.Where("contract_id = ?", contractID).Find(&validations).Error
	return validations, err
}

func (r *validationRepository) GetByUserID(userID string) ([]*models.ValidationRecord, error) {
	var validations []*models.ValidationRecord
	err := r.db.Where("user_id = ?", userID).Find(&validations).Error
	return validations, err
}

func (r *validationRepository) GetByType(validationType string) ([]*models.ValidationRecord, error) {
	var validations []*models.ValidationRecord
	err := r.db.Where("validation_type = ?", validationType).Find(&validations).Error
	return validations, err
}

// validationAuditRepository implements ValidationAuditRepository interface
type validationAuditRepository struct {
	db *gorm.DB
}

// NewValidationAuditRepository creates a new validation audit repository
func NewValidationAuditRepository(db *gorm.DB) repositories.ValidationAuditRepository {
	return &validationAuditRepository{db: db}
}

func (r *validationAuditRepository) Create(a *models.ValidationAuditLog) error {
	return r.db.Create(a).Error
}

func (r *validationAuditRepository) GetByID(id string) (*models.ValidationAuditLog, error) {
	var audit models.ValidationAuditLog
	err := r.db.Where("id = ?", id).First(&audit).Error
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

func (r *validationAuditRepository) List() ([]*models.ValidationAuditLog, error) {
	var audits []*models.ValidationAuditLog
	err := r.db.Find(&audits).Error
	return audits, err
}

func (r *validationAuditRepository) GetByValidationID(validationID string) ([]*models.ValidationAuditLog, error) {
	var audits []*models.ValidationAuditLog
	err := r.db.Where("validation_id = ?", validationID).Find(&audits).Error
	return audits, err
}

func (r *validationAuditRepository) GetByUserID(userID string) ([]*models.ValidationAuditLog, error) {
	var audits []*models.ValidationAuditLog
	err := r.db.Where("user_id = ?", userID).Find(&audits).Error
	return audits, err
}

// validationFeedbackRepository implements ValidationFeedbackRepository interface
type validationFeedbackRepository struct {
	db *gorm.DB
}

// NewValidationFeedbackRepository creates a new validation feedback repository
func NewValidationFeedbackRepository(db *gorm.DB) repositories.ValidationFeedbackRepository {
	return &validationFeedbackRepository{db: db}
}

func (r *validationFeedbackRepository) Create(f *models.ValidationFeedback) error {
	return r.db.Create(f).Error
}

func (r *validationFeedbackRepository) GetByID(id string) (*models.ValidationFeedback, error) {
	var feedback models.ValidationFeedback
	err := r.db.Where("id = ?", id).First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *validationFeedbackRepository) Update(f *models.ValidationFeedback) error {
	return r.db.Save(f).Error
}

func (r *validationFeedbackRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.ValidationFeedback{}).Error
}

func (r *validationFeedbackRepository) List() ([]*models.ValidationFeedback, error) {
	var feedbacks []*models.ValidationFeedback
	err := r.db.Find(&feedbacks).Error
	return feedbacks, err
}

func (r *validationFeedbackRepository) GetByValidationID(validationID string) ([]*models.ValidationFeedback, error) {
	var feedbacks []*models.ValidationFeedback
	err := r.db.Where("validation_id = ?", validationID).Find(&feedbacks).Error
	return feedbacks, err
}

func (r *validationFeedbackRepository) GetByUserID(userID string) ([]*models.ValidationFeedback, error) {
	var feedbacks []*models.ValidationFeedback
	err := r.db.Where("user_id = ?", userID).Find(&feedbacks).Error
	return feedbacks, err
}