package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresMilestoneRepository implements MilestoneRepository
type PostgresMilestoneRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgresMilestoneRepository creates a new PostgreSQL milestone repository
func NewPostgresMilestoneRepository(db *gorm.DB, logger *zap.Logger) MilestoneRepository {
	return &PostgresMilestoneRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresMilestoneRepository) Create(ctx context.Context, milestone *Milestone) error {
	if milestone == nil {
		return fmt.Errorf("milestone cannot be nil")
	}

	if err := r.db.WithContext(ctx).Create(milestone).Error; err != nil {
		r.logger.Error("Failed to create milestone", zap.Error(err), zap.String("contract_id", milestone.ContractID.String()), zap.String("name", milestone.Name))
		return fmt.Errorf("failed to create milestone: %w", err)
	}

	r.logger.Info("Milestone created successfully", zap.String("id", milestone.ID.String()), zap.String("contract_id", milestone.ContractID.String()))
	return nil
}

func (r *PostgresMilestoneRepository) GetByID(ctx context.Context, id uuid.UUID) (*Milestone, error) {
	var milestone Milestone
	if err := r.db.WithContext(ctx).First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Milestone not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("milestone not found")
		}
		r.logger.Error("Failed to get milestone by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	return &milestone, nil
}

func (r *PostgresMilestoneRepository) Update(ctx context.Context, milestone *Milestone) error {
	if milestone == nil {
		return fmt.Errorf("milestone cannot be nil")
	}

	milestone.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(milestone).Error; err != nil {
		r.logger.Error("Failed to update milestone", zap.Error(err), zap.String("id", milestone.ID.String()))
		return fmt.Errorf("failed to update milestone: %w", err)
	}

	r.logger.Info("Milestone updated successfully", zap.String("id", milestone.ID.String()))
	return nil
}

func (r *PostgresMilestoneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Milestone{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete milestone", zap.Error(result.Error), zap.String("id", id.String()))
		return fmt.Errorf("failed to delete milestone: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Milestone not found for deletion", zap.String("id", id.String()))
		return fmt.Errorf("milestone not found")
	}

	r.logger.Info("Milestone deleted successfully", zap.String("id", id.String()))
	return nil
}

func (r *PostgresMilestoneRepository) List(ctx context.Context, limit, offset int) ([]*Milestone, error) {
	var milestones []*Milestone
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&milestones).Error; err != nil {
		r.logger.Error("Failed to list milestones", zap.Error(err), zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, fmt.Errorf("failed to list milestones: %w", err)
	}

	r.logger.Debug("Milestones listed successfully", zap.Int("count", len(milestones)))
	return milestones, nil
}

func (r *PostgresMilestoneRepository) GetByContractID(ctx context.Context, contractID uuid.UUID, limit, offset int) ([]*Milestone, error) {
	var milestones []*Milestone
	query := r.db.WithContext(ctx).Where("contract_id = ?", contractID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&milestones).Error; err != nil {
		r.logger.Error("Failed to get milestones by contract ID", zap.Error(err), zap.String("contract_id", contractID.String()))
		return nil, fmt.Errorf("failed to get milestones by contract ID: %w", err)
	}

	r.logger.Debug("Milestones retrieved by contract ID", zap.String("contract_id", contractID.String()), zap.Int("count", len(milestones)))
	return milestones, nil
}

func (r *PostgresMilestoneRepository) GetDueSoon(ctx context.Context, days int) ([]*Milestone, error) {
	var milestones []*Milestone
	dueDate := time.Now().AddDate(0, 0, days)

	if err := r.db.WithContext(ctx).
		Where("due_date <= ? AND due_date >= ?", dueDate, time.Now()).
		Find(&milestones).Error; err != nil {
		r.logger.Error("Failed to get milestones due soon", zap.Error(err), zap.Int("days", days))
		return nil, fmt.Errorf("failed to get milestones due soon: %w", err)
	}

	r.logger.Debug("Milestones due soon retrieved", zap.Int("days", days), zap.Int("count", len(milestones)))
	return milestones, nil
}

func (r *PostgresMilestoneRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&Milestone{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		r.logger.Error("Failed to update milestone status", zap.Error(result.Error), zap.String("id", id.String()), zap.String("status", status))
		return fmt.Errorf("failed to update milestone status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Milestone not found for status update", zap.String("id", id.String()))
		return fmt.Errorf("milestone not found")
	}

	r.logger.Info("Milestone status updated successfully", zap.String("id", id.String()), zap.String("status", status))
	return nil
}
