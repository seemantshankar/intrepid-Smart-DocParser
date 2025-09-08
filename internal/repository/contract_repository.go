package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresContractRepository implements ContractRepository
type PostgresContractRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgresContractRepository creates a new PostgreSQL contract repository
func NewPostgresContractRepository(db *gorm.DB, logger *zap.Logger) ContractRepository {
	return &PostgresContractRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresContractRepository) Create(ctx context.Context, contract *Contract) error {
	if contract == nil {
		return fmt.Errorf("contract cannot be nil")
	}

	if err := r.db.WithContext(ctx).Create(contract).Error; err != nil {
		r.logger.Error("Failed to create contract", zap.Error(err), zap.String("title", contract.Title))
		return fmt.Errorf("failed to create contract: %w", err)
	}

	r.logger.Info("Contract created successfully", zap.String("id", contract.ID.String()), zap.String("title", contract.Title))
	return nil
}

func (r *PostgresContractRepository) GetByID(ctx context.Context, id uuid.UUID) (*Contract, error) {
	var contract Contract
	if err := r.db.WithContext(ctx).First(&contract, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Contract not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("contract not found")
		}
		r.logger.Error("Failed to get contract by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

func (r *PostgresContractRepository) Update(ctx context.Context, contract *Contract) error {
	if contract == nil {
		return fmt.Errorf("contract cannot be nil")
	}

	contract.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(contract).Error; err != nil {
		r.logger.Error("Failed to update contract", zap.Error(err), zap.String("id", contract.ID.String()))
		return fmt.Errorf("failed to update contract: %w", err)
	}

	r.logger.Info("Contract updated successfully", zap.String("id", contract.ID.String()))
	return nil
}

func (r *PostgresContractRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Contract{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete contract", zap.Error(result.Error), zap.String("id", id.String()))
		return fmt.Errorf("failed to delete contract: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Contract not found for deletion", zap.String("id", id.String()))
		return fmt.Errorf("contract not found")
	}

	r.logger.Info("Contract deleted successfully", zap.String("id", id.String()))
	return nil
}

func (r *PostgresContractRepository) List(ctx context.Context, limit, offset int) ([]*Contract, error) {
	var contracts []*Contract
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&contracts).Error; err != nil {
		r.logger.Error("Failed to list contracts", zap.Error(err), zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, fmt.Errorf("failed to list contracts: %w", err)
	}

	r.logger.Debug("Contracts listed successfully", zap.Int("count", len(contracts)))
	return contracts, nil
}

func (r *PostgresContractRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*Contract, error) {
	var contracts []*Contract
	query := r.db.WithContext(ctx).Where("status = ?", status)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&contracts).Error; err != nil {
		r.logger.Error("Failed to get contracts by status", zap.Error(err), zap.String("status", status))
		return nil, fmt.Errorf("failed to get contracts by status: %w", err)
	}

	r.logger.Debug("Contracts retrieved by status", zap.String("status", status), zap.Int("count", len(contracts)))
	return contracts, nil
}

func (r *PostgresContractRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&Contract{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		r.logger.Error("Failed to update contract status", zap.Error(result.Error), zap.String("id", id.String()), zap.String("status", status))
		return fmt.Errorf("failed to update contract status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Contract not found for status update", zap.String("id", id.String()))
		return fmt.Errorf("contract not found")
	}

	r.logger.Info("Contract status updated successfully", zap.String("id", id.String()), zap.String("status", status))
	return nil
}
