package service

import (
	"context"
	"fmt"

	"contract-analysis-service/internal/repository"
	"github.com/google/uuid"
)

type KnowledgeService interface {
	CreateKnowledge(ctx context.Context, title, content, category, source string, tags []string, metadata map[string]any) error
	GetKnowledge(ctx context.Context, id string) (*repository.KnowledgeEntry, error)
	UpdateKnowledge(ctx context.Context, id string, title, content, category string, tags []string, metadata map[string]any) error
	DeleteKnowledge(ctx context.Context, id string) error
	SearchKnowledge(ctx context.Context, query, category string) ([]*repository.KnowledgeEntry, error)
	ListKnowledge(ctx context.Context, limit, offset int) ([]*repository.KnowledgeEntry, error)
}

type knowledgeService struct {
	repo repository.KnowledgeRepository
}

func NewKnowledgeService(repo repository.KnowledgeRepository) KnowledgeService {
	return &knowledgeService{repo: repo}
}

func (s *knowledgeService) CreateKnowledge(ctx context.Context, title, content, category, source string, tags []string, metadata map[string]any) error {
	id := uuid.New()
	entry := &repository.KnowledgeEntry{
		ID:       id,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     tags,
		Source:   source,
		Metadata: metadata,
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

func (s *knowledgeService) UpdateKnowledge(ctx context.Context, idStr, title, content, category string, tags []string, metadata map[string]any) error {
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
	entry.Tags = tags
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
