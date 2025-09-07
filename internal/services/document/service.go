package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/validation"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10 MB
)

var allowedMimeTypes = map[string]bool{
	"application/pdf":         true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain":              true,
	"image/jpeg":              true,
	"image/png":               true,
	"image/tiff":              true,
}

// Service defines the interface for the document service.
type Service interface {
	Upload(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, error)
	GetByID(ctx context.Context, id string) (*models.Contract, error)
	Delete(ctx context.Context, id string) error
}

// documentService handles document uploads, validation, and processing.
type documentService struct {
	logger            *zap.Logger
	storage           storage.FileStorage
	contractRepo      repositories.ContractRepository
	validationService validation.Service
}

// NewDocumentService creates a new document service instance.
func NewDocumentService(logger *zap.Logger, storage storage.FileStorage, contractRepo repositories.ContractRepository, validationService validation.Service) Service {
	return &documentService{
		logger:            logger,
		storage:           storage,
		contractRepo:      contractRepo,
		validationService: validationService,
	}
}

// Upload handles a single file upload, validates it, and stores it.
func (s *documentService) Upload(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, error) {
	// Validate file size
	if fileHeader.Size > maxFileSize {
		return "", fmt.Errorf("file size %d exceeds the limit of %d bytes", fileHeader.Size, maxFileSize)
	}

	// For validation, we still need to read a chunk of the file.
	// A better approach would be to wrap the reader to allow for peeking,
	// but for now, we will read and then pass a new reader to the storage.
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for validation: %w", err)
	}

	// Create a new reader that combines the buffer and the rest of the file reader
	// so the storage service gets the full file.
	combinedReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	ext := filepath.Ext(fileHeader.Filename)
	mimeType := getMimeType(buf, ext)

	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("file type '%s' is not allowed", mimeType)
	}

	s.logger.Info("File validated successfully", zap.String("filename", fileHeader.Filename), zap.String("mime_type", mimeType))

	// Save the file
	filePath, err := s.storage.Save(combinedReader, fileHeader.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Validate the contract type
	validationResult, err := s.validationService.ValidateContract(ctx, string(buf[:n]))
	if err != nil {
		return "", fmt.Errorf("failed to validate contract: %w", err)
	}

	// Create a new contract record in the database
	newContract := &models.Contract{
		ID:           uuid.New().String(),
		FilePath:     filePath,
		Status:       models.Validated,
		ContractType: validationResult.ContractType,
		Validation:   validationResult,
		// Other fields like Hash will be populated later
	}

	if err := s.contractRepo.Create(newContract); err != nil {
		return "", fmt.Errorf("failed to create contract record: %w", err)
	}

	return newContract.ID, nil
}

// GetByID retrieves a contract by its ID.
func (s *documentService) GetByID(ctx context.Context, id string) (*models.Contract, error) {
	return s.contractRepo.GetByID(id)
}

// Delete removes a contract and its associated file.
func (s *documentService) Delete(ctx context.Context, id string) error {
	// First, get the contract to find the file path
	contract, err := s.contractRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get contract for deletion: %w", err)
	}

	// Delete the file from storage
	if err := s.storage.Delete(contract.FilePath); err != nil {
		// Log the error but continue to delete the database record
		s.logger.Error("failed to delete file from storage", zap.String("path", contract.FilePath), zap.Error(err))
	}

	// Delete the record from the database
	return s.contractRepo.Delete(id)
}

// getMimeType is a helper to determine the mime type.
// This is a simplified version; a more robust solution would use a library.
func getMimeType(buffer []byte, extension string) string {
	// Simple check based on extension for document types
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
	return "application/octet-stream" // default
}
