package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/llm"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Service defines the interface for the knowledge service.
type Service interface {
	ClassifyIndustry(ctx context.Context, contractText string) (string, error)
	QueryByIndustry(ctx context.Context, industry string) ([]*models.KnowledgeEntry, error)
}

// knowledgeService implements the Service interface.
type knowledgeService struct {
	llmService  llm.Service
	logger      *zap.Logger
	repo        repositories.KnowledgeEntryRepository
	redisClient *redis.Client
	ttl         time.Duration
}

// NewKnowledgeService creates a new knowledge service instance.
func NewKnowledgeService(llmService llm.Service, logger *zap.Logger, repo repositories.KnowledgeEntryRepository, redisClient *redis.Client) Service {
	return &knowledgeService{
		llmService:  llmService,
		logger:      logger,
		repo:        repo,
		redisClient: redisClient,
		ttl:         24 * time.Hour, // Cache for 24 hours
	}
}

// ClassifyIndustry uses the LLM service to determine the industry of a contract.
func (s *knowledgeService) ClassifyIndustry(ctx context.Context, contractText string) (string, error) {
	prompt := buildClassificationPrompt(contractText)
	provider := "openrouter"

	payload := map[string]interface{}{
		"model": "qwen/qwen-2.5-vl-72b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal classification payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return "", fmt.Errorf("industry classification request failed: %w", err)
	}

	return parseClassificationResponse(resp.Body)
}

func buildClassificationPrompt(contractText string) string {
	return fmt.Sprintf(`Analyze the following contract text and classify its industry (e.g., 'Technology', 'Manufacturing', 'Finance', 'Healthcare'). Respond with a JSON object containing a single key 'industry'. Document:\n\n%s`, contractText)
}

func parseClassificationResponse(body []byte) (string, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse classification response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in classification response")
	}

	var result struct {
		Industry string `json:"industry"`
	}
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &result); err != nil {
		return "", fmt.Errorf("failed to parse industry from content: %w", err)
	}

	return result.Industry, nil
}

// QueryByIndustry retrieves knowledge entries for a given industry, with caching.
func (s *knowledgeService) QueryByIndustry(ctx context.Context, industry string) ([]*models.KnowledgeEntry, error) {
	cacheKey := "knowledge:" + industry

	// Try to get from cache
	val, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var entries []*models.KnowledgeEntry
		if err := json.Unmarshal([]byte(val), &entries); err == nil {
			s.logger.Info("Knowledge entries found in cache", zap.String("industry", industry))
			return entries, nil
		}
		s.logger.Warn("Failed to unmarshal cached knowledge entries", zap.String("industry", industry), zap.Error(err))
	}

	// If not in cache, get from repository
	s.logger.Info("Knowledge entries not in cache, querying repository", zap.String("industry", industry))
	entries, err := s.repo.GetByIndustry(industry)
	if err != nil {
		return nil, err
	}

	// Store in cache
	serialized, err := json.Marshal(entries)
	if err != nil {
		s.logger.Error("Failed to marshal knowledge entries for caching", zap.Error(err))
		return entries, nil // Return data even if caching fails
	}

	if err := s.redisClient.Set(ctx, cacheKey, serialized, s.ttl).Err(); err != nil {
		s.logger.Error("Failed to cache knowledge entries", zap.String("industry", industry), zap.Error(err))
	}

	return entries, nil
}
