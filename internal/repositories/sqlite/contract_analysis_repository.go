package sqlite

import (
	"context"
	"time"

	"contract-analysis-service/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// contractAnalysisRepository implements ContractAnalysisRepository for SQLite/PostgreSQL
type contractAnalysisRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewContractAnalysisRepository creates a new contract analysis repository
func NewContractAnalysisRepository(db *gorm.DB, logger *zap.Logger) repository.ContractAnalysisRepository {
	// Auto-migrate the table
	if err := db.AutoMigrate(&repository.ContractAnalysisRecord{}); err != nil {
		logger.Error("failed to auto-migrate contract analysis table", zap.Error(err))
	}

	return &contractAnalysisRepository{
		db:     db,
		logger: logger,
	}
}

func (r *contractAnalysisRepository) CreateAnalysis(ctx context.Context, record *repository.ContractAnalysisRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *contractAnalysisRepository) GetAnalysisByID(ctx context.Context, id string) (*repository.ContractAnalysisRecord, error) {
	var record repository.ContractAnalysisRecord
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *contractAnalysisRepository) GetAnalysisByContractID(ctx context.Context, contractID string) (*repository.ContractAnalysisRecord, error) {
	var record repository.ContractAnalysisRecord
	err := r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *contractAnalysisRepository) GetAnalysesByUserID(ctx context.Context, userID string, limit, offset int) ([]*repository.ContractAnalysisRecord, error) {
	var records []*repository.ContractAnalysisRecord
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *contractAnalysisRepository) UpdateAnalysis(ctx context.Context, record *repository.ContractAnalysisRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *contractAnalysisRepository) DeleteAnalysis(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&repository.ContractAnalysisRecord{}, "id = ?", id).Error
}

func (r *contractAnalysisRepository) ListAnalyses(ctx context.Context, limit, offset int) ([]*repository.ContractAnalysisRecord, error) {
	var records []*repository.ContractAnalysisRecord
	query := r.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *contractAnalysisRepository) GetAnalysesByConfidenceRange(ctx context.Context, minConfidence, maxConfidence float64, limit, offset int) ([]*repository.ContractAnalysisRecord, error) {
	var records []*repository.ContractAnalysisRecord
	query := r.db.WithContext(ctx).
		Where("confidence_score >= ? AND confidence_score <= ?", minConfidence, maxConfidence)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("confidence_score DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *contractAnalysisRepository) GetAnalysesCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*repository.ContractAnalysisRecord, error) {
	var records []*repository.ContractAnalysisRecord
	query := r.db.WithContext(ctx).Where("created_at > ?", after)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}