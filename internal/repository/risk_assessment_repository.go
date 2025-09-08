package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresRiskAssessmentRepository implements RiskAssessmentRepository
type PostgresRiskAssessmentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgresRiskAssessmentRepository creates a new PostgreSQL risk assessment repository
func NewPostgresRiskAssessmentRepository(db *gorm.DB, logger *zap.Logger) RiskAssessmentRepository {
	return &PostgresRiskAssessmentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresRiskAssessmentRepository) Create(ctx context.Context, assessment *RiskAssessment) error {
	if assessment == nil {
		return fmt.Errorf("risk assessment cannot be nil")
	}

	if err := r.db.WithContext(ctx).Create(assessment).Error; err != nil {
		r.logger.Error("Failed to create risk assessment", zap.Error(err), zap.String("contract_id", assessment.ContractID.String()), zap.String("risk_level", assessment.RiskLevel))
		return fmt.Errorf("failed to create risk assessment: %w", err)
	}

	r.logger.Info("Risk assessment created successfully", zap.String("id", assessment.ID.String()), zap.String("contract_id", assessment.ContractID.String()))
	return nil
}

func (r *PostgresRiskAssessmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*RiskAssessment, error) {
	var assessment RiskAssessment
	if err := r.db.WithContext(ctx).First(&assessment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Risk assessment not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("risk assessment not found")
		}
		r.logger.Error("Failed to get risk assessment by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, fmt.Errorf("failed to get risk assessment: %w", err)
	}

	return &assessment, nil
}

func (r *PostgresRiskAssessmentRepository) Update(ctx context.Context, assessment *RiskAssessment) error {
	if assessment == nil {
		return fmt.Errorf("risk assessment cannot be nil")
	}

	assessment.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(assessment).Error; err != nil {
		r.logger.Error("Failed to update risk assessment", zap.Error(err), zap.String("id", assessment.ID.String()))
		return fmt.Errorf("failed to update risk assessment: %w", err)
	}

	r.logger.Info("Risk assessment updated successfully", zap.String("id", assessment.ID.String()))
	return nil
}

func (r *PostgresRiskAssessmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&RiskAssessment{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete risk assessment", zap.Error(result.Error), zap.String("id", id.String()))
		return fmt.Errorf("failed to delete risk assessment: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Risk assessment not found for deletion", zap.String("id", id.String()))
		return fmt.Errorf("risk assessment not found")
	}

	r.logger.Info("Risk assessment deleted successfully", zap.String("id", id.String()))
	return nil
}

func (r *PostgresRiskAssessmentRepository) List(ctx context.Context, limit, offset int) ([]*RiskAssessment, error) {
	var assessments []*RiskAssessment
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&assessments).Error; err != nil {
		r.logger.Error("Failed to list risk assessments", zap.Error(err), zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, fmt.Errorf("failed to list risk assessments: %w", err)
	}

	r.logger.Debug("Risk assessments listed successfully", zap.Int("count", len(assessments)))
	return assessments, nil
}

func (r *PostgresRiskAssessmentRepository) GetByContractID(ctx context.Context, contractID uuid.UUID, limit, offset int) ([]*RiskAssessment, error) {
	var assessments []*RiskAssessment
	query := r.db.WithContext(ctx).Where("contract_id = ?", contractID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&assessments).Error; err != nil {
		r.logger.Error("Failed to get risk assessments by contract ID", zap.Error(err), zap.String("contract_id", contractID.String()))
		return nil, fmt.Errorf("failed to get risk assessments by contract ID: %w", err)
	}

	r.logger.Debug("Risk assessments retrieved by contract ID", zap.String("contract_id", contractID.String()), zap.Int("count", len(assessments)))
	return assessments, nil
}

func (r *PostgresRiskAssessmentRepository) GetByRiskLevel(ctx context.Context, level string, limit, offset int) ([]*RiskAssessment, error) {
	var assessments []*RiskAssessment
	query := r.db.WithContext(ctx).Where("risk_level = ?", level)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&assessments).Error; err != nil {
		r.logger.Error("Failed to get risk assessments by risk level", zap.Error(err), zap.String("risk_level", level))
		return nil, fmt.Errorf("failed to get risk assessments by risk level: %w", err)
	}

	r.logger.Debug("Risk assessments retrieved by risk level", zap.String("risk_level", level), zap.Int("count", len(assessments)))
	return assessments, nil
}
