package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresKnowledgeEntryRepository implements KnowledgeEntryRepository
type PostgresKnowledgeEntryRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgresKnowledgeEntryRepository creates a new PostgreSQL knowledge entry repository
func NewPostgresKnowledgeEntryRepository(db *gorm.DB, logger *zap.Logger) KnowledgeEntryRepository {
	return &PostgresKnowledgeEntryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresKnowledgeEntryRepository) Create(ctx context.Context, entry *KnowledgeEntry) error {
	if entry == nil {
		return fmt.Errorf("knowledge entry cannot be nil")
	}

	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		r.logger.Error("Failed to create knowledge entry", zap.Error(err), zap.String("title", entry.Title))
		return fmt.Errorf("failed to create knowledge entry: %w", err)
	}

	r.logger.Info("Knowledge entry created successfully", zap.String("id", entry.ID.String()), zap.String("title", entry.Title))
	return nil
}

func (r *PostgresKnowledgeEntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*KnowledgeEntry, error) {
	var entry KnowledgeEntry
	if err := r.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Knowledge entry not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("knowledge entry not found")
		}
		r.logger.Error("Failed to get knowledge entry by ID", zap.Error(err), zap.String("id", id.String()))
		return nil, fmt.Errorf("failed to get knowledge entry: %w", err)
	}

	return &entry, nil
}

func (r *PostgresKnowledgeEntryRepository) Update(ctx context.Context, entry *KnowledgeEntry) error {
	if entry == nil {
		return fmt.Errorf("knowledge entry cannot be nil")
	}

	entry.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(entry).Error; err != nil {
		r.logger.Error("Failed to update knowledge entry", zap.Error(err), zap.String("id", entry.ID.String()))
		return fmt.Errorf("failed to update knowledge entry: %w", err)
	}

	r.logger.Info("Knowledge entry updated successfully", zap.String("id", entry.ID.String()))
	return nil
}

func (r *PostgresKnowledgeEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&KnowledgeEntry{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete knowledge entry", zap.Error(result.Error), zap.String("id", id.String()))
		return fmt.Errorf("failed to delete knowledge entry: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Knowledge entry not found for deletion", zap.String("id", id.String()))
		return fmt.Errorf("knowledge entry not found")
	}

	r.logger.Info("Knowledge entry deleted successfully", zap.String("id", id.String()))
	return nil
}

func (r *PostgresKnowledgeEntryRepository) List(ctx context.Context, limit, offset int) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	query := r.db.WithContext(ctx)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entries).Error; err != nil {
		r.logger.Error("Failed to list knowledge entries", zap.Error(err), zap.Int("limit", limit), zap.Int("offset", offset))
		return nil, fmt.Errorf("failed to list knowledge entries: %w", err)
	}

	r.logger.Debug("Knowledge entries listed successfully", zap.Int("count", len(entries)))
	return entries, nil
}

func (r *PostgresKnowledgeEntryRepository) SearchByTags(ctx context.Context, tags []string, limit, offset int) ([]*KnowledgeEntry, error) {
	if len(tags) == 0 {
		return r.List(ctx, limit, offset)
	}

	var entries []*KnowledgeEntry
	query := r.db.WithContext(ctx)

	// Build the query to search for entries that contain any of the specified tags
	tagConditions := make([]string, len(tags))
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		tagConditions[i] = "tags @> ARRAY[?]::text[]"
		args[i] = tag
	}

	condition := strings.Join(tagConditions, " OR ")
	if err := query.Where(condition, args...).
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		r.logger.Error("Failed to search knowledge entries by tags", zap.Error(err), zap.Strings("tags", tags))
		return nil, fmt.Errorf("failed to search knowledge entries by tags: %w", err)
	}

	r.logger.Debug("Knowledge entries searched by tags", zap.Strings("tags", tags), zap.Int("count", len(entries)))
	return entries, nil
}

func (r *PostgresKnowledgeEntryRepository) SearchByContent(ctx context.Context, query string, limit, offset int) ([]*KnowledgeEntry, error) {
	if query == "" {
		return r.List(ctx, limit, offset)
	}

	var entries []*KnowledgeEntry
	searchQuery := "%" + strings.ToLower(query) + "%"

	if err := r.db.WithContext(ctx).
		Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", searchQuery, searchQuery).
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		r.logger.Error("Failed to search knowledge entries by content", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("failed to search knowledge entries by content: %w", err)
	}

	r.logger.Debug("Knowledge entries searched by content", zap.String("query", query), zap.Int("count", len(entries)))
	return entries, nil
}
