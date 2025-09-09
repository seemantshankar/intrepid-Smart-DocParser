package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KnowledgeEntry struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Category  string         `gorm:"type:varchar(100)" json:"category"`
	Tags      []string       `gorm:"type:jsonb" json:"tags"`
	Source    string         `gorm:"type:varchar(500)" json:"source"`
	Metadata  map[string]any `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type KnowledgeRepository interface {
	CreateKnowledge(ctx context.Context, entry *KnowledgeEntry) error
	GetKnowledgeByID(ctx context.Context, id uuid.UUID) (*KnowledgeEntry, error)
	UpdateKnowledge(ctx context.Context, entry *KnowledgeEntry) error
	DeleteKnowledge(ctx context.Context, id uuid.UUID) error
	SearchKnowledge(ctx context.Context, query string, category string) ([]*KnowledgeEntry, error)
	ListKnowledge(ctx context.Context, limit, offset int) ([]*KnowledgeEntry, error)
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