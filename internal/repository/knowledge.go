package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type KnowledgeEntry struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title           string          `gorm:"type:varchar(255);not null" json:"title"`
	Content         string          `gorm:"type:text;not null" json:"content"`
	Category        string          `gorm:"type:varchar(100)" json:"category"`
	Tags            pq.StringArray  `gorm:"type:text[]" json:"tags"`
	Source          string          `gorm:"type:varchar(500)" json:"source"`
	Metadata        string          `gorm:"type:text" json:"metadata"`
	Version         int             `gorm:"not null;default:1" json:"version"`
	ParentVersionID *uuid.UUID      `gorm:"type:uuid" json:"parent_version_id"`
	IsLatest        bool            `gorm:"default:true" json:"is_latest"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at"`
}

// KnowledgeSearchFilter provides advanced search filtering options
type KnowledgeSearchFilter struct {
	Query         string     `json:"query"`          // Text search in title and content
	Category      string     `json:"category"`       // Filter by category
	Tags          pq.StringArray `json:"tags"`       // Filter by tags (any match)
	Source        string     `json:"source"`         // Filter by source
	CreatedAfter  *time.Time `json:"created_after"`  // Filter entries created after this date
	CreatedBefore *time.Time `json:"created_before"` // Filter entries created before this date
	Limit         int        `json:"limit"`          // Maximum number of results
	Offset        int        `json:"offset"`         // Pagination offset
}

type KnowledgeRepository interface {
	CreateKnowledge(ctx context.Context, entry *KnowledgeEntry) error
	GetKnowledgeByID(ctx context.Context, id uuid.UUID) (*KnowledgeEntry, error)
	UpdateKnowledge(ctx context.Context, entry *KnowledgeEntry) error
	UpdateKnowledgeWithConflictDetection(ctx context.Context, entry *KnowledgeEntry) error
	DeleteKnowledge(ctx context.Context, id uuid.UUID) error
	SearchKnowledge(ctx context.Context, query string, category string) ([]*KnowledgeEntry, error)
	SearchKnowledgeAdvanced(ctx context.Context, filter KnowledgeSearchFilter) ([]*KnowledgeEntry, error)
	ListKnowledge(ctx context.Context, limit, offset int) ([]*KnowledgeEntry, error)
	
	// Versioning methods
	GetVersionHistory(ctx context.Context, entryID uuid.UUID) ([]*KnowledgeEntry, error)
	GetLatestVersion(ctx context.Context, entryID uuid.UUID) (*KnowledgeEntry, error)
	
	// Advanced search methods
	FindSimilarContent(ctx context.Context, content string, threshold float64, limit int) ([]*KnowledgeEntry, error)
	FullTextSearch(ctx context.Context, query string, limit, offset int) ([]*KnowledgeEntry, error)
}

type knowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) KnowledgeRepository {
	return &knowledgeRepository{db: db}
}

func (r *knowledgeRepository) CreateKnowledge(ctx context.Context, entry *KnowledgeEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *knowledgeRepository) GetKnowledgeByID(ctx context.Context, id uuid.UUID) (*KnowledgeEntry, error) {
	var entry KnowledgeEntry
	err := r.db.WithContext(ctx).First(&entry, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge entry not found")
		}
		return nil, err
	}
	return &entry, nil
}

func (r *knowledgeRepository) UpdateKnowledge(ctx context.Context, entry *KnowledgeEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *knowledgeRepository) UpdateKnowledgeWithConflictDetection(ctx context.Context, entry *KnowledgeEntry) error {
	// Start transaction for atomic version checking and updating
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current version from database
	var currentEntry KnowledgeEntry
	if err := tx.First(&currentEntry, "id = ?", entry.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return fmt.Errorf("knowledge entry not found")
		}
		tx.Rollback()
		return err
	}

	// Check version conflict
	if currentEntry.Version != entry.Version {
		tx.Rollback()
		return fmt.Errorf("version conflict: expected version %d, found version %d", entry.Version, currentEntry.Version)
	}

	// Increment version for update
	entry.Version = currentEntry.Version + 1
	entry.UpdatedAt = time.Now()

	// Update entry
	if err := tx.Save(entry).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *knowledgeRepository) DeleteKnowledge(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&KnowledgeEntry{}, "id = ?", id).Error
}

func (r *knowledgeRepository) SearchKnowledge(ctx context.Context, query, category string) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	db := r.db.WithContext(ctx).Where("title ILIKE ? OR content ILIKE ?", "%"+query+"%", "%"+query+"%")
	if category != "" {
		db = db.Where("category = ?", category)
	}
	err := db.Find(&entries).Error
	return entries, err
}

func (r *knowledgeRepository) ListKnowledge(ctx context.Context, limit, offset int) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at desc").Find(&entries).Error
	return entries, err
}

// GetVersionHistory returns all versions of a knowledge entry except the latest
func (r *knowledgeRepository) GetVersionHistory(ctx context.Context, entryID uuid.UUID) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	err := r.db.WithContext(ctx).
		Where("id = ? OR parent_version_id = ?", entryID, entryID).
		Where("is_latest = ?", false).
		Order("version desc").
		Find(&entries).Error
	return entries, err
}

// GetLatestVersion returns the latest version of a knowledge entry
func (r *knowledgeRepository) GetLatestVersion(ctx context.Context, entryID uuid.UUID) (*KnowledgeEntry, error) {
	var entry KnowledgeEntry
	// Find the entry with highest version number in the version chain
	err := r.db.WithContext(ctx).
		Where("id = ? OR parent_version_id = ?", entryID, entryID).
		Order("version DESC").
		First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("knowledge entry not found")
		}
		return nil, err
	}
	return &entry, nil
}

// SearchKnowledgeAdvanced performs advanced search with multiple filters
func (r *knowledgeRepository) SearchKnowledgeAdvanced(ctx context.Context, filter KnowledgeSearchFilter) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	query := r.db.WithContext(ctx).Where("is_latest = ?", true)

	// Text search in title and content
	if filter.Query != "" {
		searchQuery := "%" + filter.Query + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchQuery, searchQuery)
	}

	// Filter by category
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	// Filter by source
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}

	// Filter by tags (any match)
	if len(filter.Tags) > 0 {
		// PostgreSQL array overlap operator
		for _, tag := range filter.Tags {
			query = query.Where("? = ANY(tags)", tag)
			break // For now, just match the first tag
		}
	}

	// Filter by created date range
	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filter.CreatedBefore)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Order by relevance (created_at desc for now, could be enhanced with ranking)
	query = query.Order("created_at desc")

	err := query.Find(&entries).Error
	return entries, err
}

// FindSimilarContent uses PostgreSQL's similarity function to find similar content
func (r *knowledgeRepository) FindSimilarContent(ctx context.Context, content string, threshold float64, limit int) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	
	// Using PostgreSQL's similarity function (requires pg_trgm extension)
	err := r.db.WithContext(ctx).Raw(`
		SELECT * FROM knowledge_entries 
		WHERE is_latest = ? 
		AND similarity(content, ?) > ? 
		ORDER BY similarity(content, ?) DESC 
		LIMIT ?
	`, true, content, threshold, content, limit).Find(&entries).Error
	
	if err != nil {
		// Fallback to basic text matching if similarity function is not available
		searchQuery := "%" + content + "%"
		err = r.db.WithContext(ctx).
			Where("is_latest = ? AND content ILIKE ?", true, searchQuery).
			Limit(limit).
			Find(&entries).Error
	}
	
	return entries, err
}

// FullTextSearch performs ranked full-text search using PostgreSQL's full-text search
func (r *knowledgeRepository) FullTextSearch(ctx context.Context, query string, limit, offset int) ([]*KnowledgeEntry, error) {
	var entries []*KnowledgeEntry
	
	// Using PostgreSQL's full-text search with ranking
	err := r.db.WithContext(ctx).Raw(`
		SELECT *, ts_rank_cd(to_tsvector('english', title || ' ' || content), plainto_tsquery('english', ?)) AS rank
		FROM knowledge_entries 
		WHERE is_latest = ? 
		AND to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', ?)
		ORDER BY rank DESC, created_at DESC
		LIMIT ? OFFSET ?
	`, query, true, query, limit, offset).Find(&entries).Error
	
	if err != nil {
		// Fallback to ILIKE search if full-text search fails
		searchQuery := "%" + query + "%"
		err = r.db.WithContext(ctx).
			Where("is_latest = ? AND (title ILIKE ? OR content ILIKE ?)", true, searchQuery, searchQuery).
			Order("created_at desc").
			Limit(limit).
			Offset(offset).
			Find(&entries).Error
	}
	
	return entries, err
}
