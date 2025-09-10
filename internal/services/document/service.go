package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/analysis"

	"github.com/google/uuid"
	"go.uber.org/zap"
	// Remove: "encoding/json"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10 MB
)

var allowedMimeTypes = map[string]bool{
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain": true,
	"image/jpeg": true,
	"image/png":  true,
	"image/tiff": true,
}

type Service interface {
	UploadAndAnalyze(ctx context.Context, content io.Reader, header *multipart.FileHeader, userID string) (*models.Contract, error)
	Upload(ctx context.Context, content io.Reader, header *multipart.FileHeader, userID string) (string, error)
	GetByID(ctx context.Context, id string) (*models.Contract, error)
	GetDocument(ctx context.Context, id string, userID string) (*models.Contract, error)
	GetDocumentContent(ctx context.Context, storagePath string) ([]byte, error)
	RetrieveAnalysis(ctx context.Context, id string) (map[string]interface{}, error)
	Delete(ctx context.Context, id string) error
	CleanupExpiredDocuments(ctx context.Context) error
	// Add other methods as needed
}

type documentService struct {
	repo     repositories.ContractRepository
	storage  storage.FileStorage
	analyzer analysis.Service
	logger   *zap.Logger
}

// NewDocumentService creates a new document service implementing the Service interface.
func NewDocumentService(repo repositories.ContractRepository, storage storage.FileStorage, analyzer analysis.Service, logger *zap.Logger) Service {
	return &documentService{
		repo:     repo,
		storage:  storage,
		analyzer: analyzer,
		logger:   logger,
	}
}

const defaultRetentionDays = 365

type closableBytesReader struct {
	*bytes.Reader
}

func (c *closableBytesReader) Close() error {
	return nil
}

func (s *documentService) UploadAndAnalyze(ctx context.Context, content io.Reader, header *multipart.FileHeader, userID string) (*models.Contract, error) {
	if header.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit")
	}

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}

	mimeType := getMimeType(strings.ToLower(filepath.Ext(header.Filename)))
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("invalid file type")
	}

	// Create a reader for storage
	fileReader := &closableBytesReader{bytes.NewReader(contentBytes)}

	analysisResult, err := s.analyzer.AnalyzeContract(ctx, string(contentBytes))
	if err != nil {
		return nil, err
	}

	storagePath, err := s.storage.Save(fileReader, header.Filename)
	if err != nil {
		return nil, err
	}

	contract := &models.Contract{
		ID:            uuid.New().String(),
		UserID:        userID,
		Filename:      header.Filename,
		StoragePath:   storagePath,
		CreatedAt:     time.Now(),
		RetentionDays: defaultRetentionDays,
		Analysis:      *analysisResult,
	}

	if err := s.repo.Create(contract); err != nil {
		return nil, err
	}

	return contract, nil
}
func (s *documentService) Upload(ctx context.Context, content io.Reader, header *multipart.FileHeader, userID string) (string, error) {
	// Read content
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	// Validate file
	if len(contentBytes) > maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size")
	}

	mimeType := getMimeType(filepath.Ext(header.Filename))
	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("unsupported file type: %s", mimeType)
	}

	// Create contract record
	contract := &models.Contract{
		ID:        uuid.New().String(),
		UserID:    userID,
		Filename:  header.Filename,
		CreatedAt: time.Now(),
	}

	// Store file
	fileReader := bytes.NewReader(contentBytes)
	storagePath, err := s.storage.Save(fileReader, header.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to store file: %w", err)
	}
	contract.StoragePath = storagePath

	// Save to repository
	if err := s.repo.Create(contract); err != nil {
		return "", fmt.Errorf("failed to save contract: %w", err)
	}

	return contract.ID, nil
}

func (s *documentService) GetByID(ctx context.Context, id string) (*models.Contract, error) {
	return s.repo.GetByID(id)
}

func (s *documentService) GetDocument(ctx context.Context, id string, userID string) (*models.Contract, error) {
	contract, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if contract.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to document")
	}
	return contract, nil
}

func (s *documentService) GetDocumentContent(ctx context.Context, storagePath string) ([]byte, error) {
	return s.storage.Read(storagePath)
}

func (s *documentService) RetrieveAnalysis(ctx context.Context, id string) (map[string]interface{}, error) {
	contract, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	// Check if analysis has any data by checking if buyer field is empty
	if contract.Analysis.Buyer == "" {
		return nil, fmt.Errorf("no analysis found for document")
	}
	// Convert struct to map for compatibility with handler
	analysisMap := map[string]interface{}{
		"buyer":          contract.Analysis.Buyer,
		"buyer_address":  contract.Analysis.BuyerAddress,
		"buyer_country":  contract.Analysis.BuyerCountry,
		"seller":         contract.Analysis.Seller,
		"seller_address": contract.Analysis.SellerAddress,
		"seller_country": contract.Analysis.SellerCountry,
		"total_value":    contract.Analysis.TotalValue,
		"currency":       contract.Analysis.Currency,
		"milestones":     contract.Analysis.Milestones,
		"risk_factors":   contract.Analysis.RiskFactors,
		"contract_id":    contract.ID,
	}
	return analysisMap, nil
}

func (s *documentService) Delete(ctx context.Context, id string) error {
	contract, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(contract.StoragePath); err != nil {
		s.logger.Warn("Failed to delete file from storage", zap.Error(err), zap.String("path", contract.StoragePath))
	}
	return s.repo.Delete(id)
}

//		return &contract.Analysis, nil
//	}
func (s *documentService) CleanupExpiredDocuments(ctx context.Context) error {
	contracts, err := s.repo.List()
	if err != nil {
		return err
	}
	for _, contract := range contracts {
		if time.Since(contract.CreatedAt) > time.Duration(contract.RetentionDays)*24*time.Hour {
			if err := s.repo.Delete(contract.ID); err != nil {
				// log error
				continue
			}
			if err := s.storage.Delete(contract.StoragePath); err != nil {
				// log error
				continue
			}
		}
	}
	return nil
}
func getMimeType(extension string) string {
	switch extension {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	case ".jpeg", ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".tiff":
		return "image/tiff"
	}
	return "application/octet-stream"
}
