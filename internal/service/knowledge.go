package service

import (
	"context"
	"fmt"

	"contract-analysis-service/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type KnowledgeService interface {
	CreateKnowledge(ctx context.Context, title, content, category, source string, tags []string, metadata string) error
	GetKnowledge(ctx context.Context, id string) (*repository.KnowledgeEntry, error)
	UpdateKnowledge(ctx context.Context, id string, title, content, category string, tags []string, metadata string) error
	UpdateKnowledgeWithConflictDetection(ctx context.Context, id string, title, content, category string, tags []string, metadata string, version int) error
	DeleteKnowledge(ctx context.Context, id string) error
	SearchKnowledge(ctx context.Context, query, category string) ([]*repository.KnowledgeEntry, error)
	SearchKnowledgeAdvanced(ctx context.Context, filter repository.KnowledgeSearchFilter) ([]*repository.KnowledgeEntry, error)
	ListKnowledge(ctx context.Context, limit, offset int) ([]*repository.KnowledgeEntry, error)
	
	// Versioning methods
	GetVersionHistory(ctx context.Context, id string) ([]*repository.KnowledgeEntry, error)
	GetLatestVersion(ctx context.Context, id string) (*repository.KnowledgeEntry, error)
	
	// Advanced search methods
	FindSimilarContent(ctx context.Context, content string, threshold float64, limit int) ([]*repository.KnowledgeEntry, error)
	FullTextSearch(ctx context.Context, query string, limit, offset int) ([]*repository.KnowledgeEntry, error)
}

type knowledgeService struct {
	repo repository.KnowledgeRepository
}

func NewKnowledgeService(repo repository.KnowledgeRepository) KnowledgeService {
	return &knowledgeService{repo: repo}
}

func (s *knowledgeService) CreateKnowledge(ctx context.Context, title, content, category, source string, tags []string, metadata string) error {
	id := uuid.New()
	entry := &repository.KnowledgeEntry{
		ID:       id,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     pq.StringArray(tags),
		Source:   source,
		Metadata: metadata,
		Version:  1,
	}
	return s.repo.CreateKnowledge(ctx, entry)
}

func (s *knowledgeService) GetKnowledge(ctx context.Context, idStr string) (*repository.KnowledgeEntry, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid knowledge ID: %w", err)
	}
	return s.repo.GetKnowledgeByID(ctx, id)
}

func (s *knowledgeService) UpdateKnowledge(ctx context.Context, idStr, title, content, category string, tags []string, metadata string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid knowledge ID: %w", err)
	}
	entry, err := s.repo.GetKnowledgeByID(ctx, id)
	if err != nil {
		return err
	}
	entry.Title = title
	entry.Content = content
	entry.Category = category
	entry.Tags = pq.StringArray(tags)
	entry.Metadata = metadata
	return s.repo.UpdateKnowledge(ctx, entry)
}

func (s *knowledgeService) DeleteKnowledge(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid knowledge ID: %w", err)
	}
	return s.repo.DeleteKnowledge(ctx, id)
}

func (s *knowledgeService) SearchKnowledge(ctx context.Context, query, category string) ([]*repository.KnowledgeEntry, error) {
	return s.repo.SearchKnowledge(ctx, query, category)
}

func (s *knowledgeService) ListKnowledge(ctx context.Context, limit, offset int) ([]*repository.KnowledgeEntry, error) {
	return s.repo.ListKnowledge(ctx, limit, offset)
}

// UpdateKnowledgeWithConflictDetection updates a knowledge entry with version conflict detection
func (s *knowledgeService) UpdateKnowledgeWithConflictDetection(ctx context.Context, idStr, title, content, category string, tags []string, metadata string, version int) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid knowledge ID: %w", err)
	}
	entry := &repository.KnowledgeEntry{
		ID:       id,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     pq.StringArray(tags),
		Metadata: metadata,
		Version:  version,
	}
	return s.repo.UpdateKnowledgeWithConflictDetection(ctx, entry)
}

// SearchKnowledgeAdvanced performs advanced search with filtering
func (s *knowledgeService) SearchKnowledgeAdvanced(ctx context.Context, filter repository.KnowledgeSearchFilter) ([]*repository.KnowledgeEntry, error) {
	return s.repo.SearchKnowledgeAdvanced(ctx, filter)
}

// GetVersionHistory returns the version history of a knowledge entry
func (s *knowledgeService) GetVersionHistory(ctx context.Context, idStr string) ([]*repository.KnowledgeEntry, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid knowledge ID: %w", err)
	}
	return s.repo.GetVersionHistory(ctx, id)
}

// GetLatestVersion returns the latest version of a knowledge entry
func (s *knowledgeService) GetLatestVersion(ctx context.Context, idStr string) (*repository.KnowledgeEntry, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid knowledge ID: %w", err)
	}
	return s.repo.GetLatestVersion(ctx, id)
}

// FindSimilarContent finds entries with similar content
func (s *knowledgeService) FindSimilarContent(ctx context.Context, content string, threshold float64, limit int) ([]*repository.KnowledgeEntry, error) {
	return s.repo.FindSimilarContent(ctx, content, threshold, limit)
}

// FullTextSearch performs ranked full-text search
func (s *knowledgeService) FullTextSearch(ctx context.Context, query string, limit, offset int) ([]*repository.KnowledgeEntry, error) {
	return s.repo.FullTextSearch(ctx, query, limit, offset)
}
