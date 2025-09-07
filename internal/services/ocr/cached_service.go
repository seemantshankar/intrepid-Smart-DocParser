package ocr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// cachedOCRService is a decorator for the OCR service that adds a caching layer.
// It uses Redis to cache OCR results to avoid redundant API calls.
type cachedOCRService struct {
	next        Service
	redisClient *redis.Client
	logger      *zap.Logger
	ttl         time.Duration
}

// NewCachedOCRService creates a new cached OCR service.
func NewCachedOCRService(next Service, redisClient *redis.Client, logger *zap.Logger, ttl time.Duration) Service {
	return &cachedOCRService{
		next:        next,
		redisClient: redisClient,
		logger:      logger,
		ttl:         ttl,
	}
}

// ExtractTextFromImage first checks the cache for a result. If not found, it calls the next service
// and stores the result in the cache.
func (s *cachedOCRService) ExtractTextFromImage(ctx context.Context, imagePath string) (*OCRResult, error) {
	// Use the image path as the cache key. For a more robust implementation,
	// a hash of the image content could be used.
	cacheKey := "ocr:" + imagePath

	// Try to get the result from the cache
	cachedResult, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		s.logger.Info("OCR result found in cache", zap.String("key", cacheKey))
		var result OCRResult
		if err := json.Unmarshal([]byte(cachedResult), &result); err == nil {
			return &result, nil
		}
		// If there's an error unmarshalling, we'll proceed to fetch from the service
		s.logger.Warn("Failed to unmarshal cached OCR result", zap.String("key", cacheKey), zap.Error(err))
	}

	// If not in cache, call the next service
	s.logger.Info("OCR result not in cache, calling next service", zap.String("key", cacheKey))
	result, err := s.next.ExtractTextFromImage(ctx, imagePath)
	if err != nil {
		return nil, err
	}

	// Store the result in the cache
	serializedResult, err := json.Marshal(result)
	if err != nil {
		s.logger.Error("Failed to marshal OCR result for caching", zap.Error(err))
		return result, nil // Return the result even if caching fails
	}

	if err := s.redisClient.Set(ctx, cacheKey, serializedResult, s.ttl).Err(); err != nil {
		s.logger.Error("Failed to cache OCR result", zap.String("key", cacheKey), zap.Error(err))
	}

	return result, nil
}
