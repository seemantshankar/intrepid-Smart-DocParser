package document

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"contract-analysis-service/internal/models"
	pdfutil "contract-analysis-service/internal/pkg/pdf"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/ocr"
	"contract-analysis-service/internal/services/validation"
	"contract-analysis-service/internal/services/analysis"

	"github.com/google/uuid"
	"go.uber.org/zap"
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

// UploadAndAnalyze securely stores the file and performs one-step image-based LLM analysis.
// Returns the stored document ID and the analysis JSON.
func (s *documentService) UploadAndAnalyze(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, *models.ContractAnalysis, error) {
    if fileHeader.Size > maxFileSize {
        return "", nil, fmt.Errorf("file size %d exceeds the limit of %d bytes", fileHeader.Size, maxFileSize)
    }

    buf := make([]byte, 512)
    n, err := file.Read(buf)
    if err != nil && err != io.EOF {
        return "", nil, fmt.Errorf("failed to read file for validation: %w", err)
    }
    combinedReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

    ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
    mimeType := getMimeType(buf, ext)
    if !allowedMimeTypes[mimeType] {
        return "", nil, fmt.Errorf("file type '%s' is not allowed", mimeType)
    }

    // Save the file securely
    filePath, err := s.storage.Save(combinedReader, fileHeader.Filename)
    if err != nil {
        return "", nil, fmt.Errorf("failed to save file: %w", err)
    }

    var analysisResult *models.ContractAnalysis
    if ext == ".pdf" {
        images, rasterErr := pdfutil.RasterizeToJPEGs(filePath, 10)
        if rasterErr == nil && len(images) > 0 {
            analysisResult, err = s.analysisService.AnalyzeContractFromImages(ctx, images)
            if err != nil {
                return "", nil, fmt.Errorf("analysis failed: %w", err)
            }
        } else {
            // Minimal fallback: send header bytes as text
            analysisResult, err = s.analysisService.AnalyzeContract(ctx, string(buf[:n]))
            if err != nil {
                return "", nil, fmt.Errorf("analysis failed (text fallback): %w", err)
            }
        }
    } else {
        // Non-PDF fallback
        analysisResult, err = s.analysisService.AnalyzeContract(ctx, string(buf[:n]))
        if err != nil {
            return "", nil, fmt.Errorf("analysis failed (non-pdf): %w", err)
        }
    }

    // Persist a Contract record with summary fields populated
    contract := &models.Contract{
        ID:       uuid.New().String(),
        FilePath: filePath,
        Status:   models.Analyzed,
        Summary: &models.ContractSummary{
            BuyerName:     analysisResult.Buyer,
            BuyerAddress:  analysisResult.BuyerAddress,
            BuyerCountry:  analysisResult.BuyerCountry,
            SellerName:    analysisResult.Seller,
            SellerAddress: analysisResult.SellerAddress,
            SellerCountry: analysisResult.SellerCountry,
            TotalValue:    analysisResult.TotalValue,
            Currency:      analysisResult.Currency,
        },
    }
    if err := s.contractRepo.Create(contract); err != nil {
        return "", nil, fmt.Errorf("failed to create contract record: %w", err)
    }

    return contract.ID, analysisResult, nil
}

type Service interface {
	Upload(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, error)
	UploadAndAnalyze(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, *models.ContractAnalysis, error)
	GetByID(ctx context.Context, id string) (*models.Contract, error)
	Delete(ctx context.Context, id string) error
}

type documentService struct {
	logger            *zap.Logger
	storage           storage.FileStorage
	contractRepo      repositories.ContractRepository
	validationService validation.Service
	ocrService        ocr.Service
	analysisService   analysis.Service
}

func NewDocumentService(logger *zap.Logger, storage storage.FileStorage, contractRepo repositories.ContractRepository, validationService validation.Service, ocrService ocr.Service, analysisService analysis.Service) Service {
	return &documentService{
		logger:            logger,
		storage:           storage,
		contractRepo:      contractRepo,
		validationService: validationService,
		ocrService:        ocrService,
		analysisService:   analysisService,
	}
}

func (s *documentService) Upload(ctx context.Context, file io.Reader, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxFileSize {
		return "", fmt.Errorf("file size %d exceeds the limit of %d bytes", fileHeader.Size, maxFileSize)
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for validation: %w", err)
	}

	combinedReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	mimeType := getMimeType(buf, ext)

	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("file type '%s' is not allowed", mimeType)
	}

	s.logger.Info("File validated successfully", zap.String("filename", fileHeader.Filename), zap.String("mime_type", mimeType))

	filePath, err := s.storage.Save(combinedReader, fileHeader.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	var docText string
	if ext == ".pdf" {
		if text, ok := pdfutil.ExtractText(filePath); ok {
			docText = text
			s.logger.Info("PDF text extracted directly", zap.String("path", filePath), zap.Int("length", len(text)))
		} else {
			images, rasterErr := pdfutil.RasterizeToJPEGs(filePath, 10)
			if rasterErr != nil {
				s.logger.Error("failed to rasterize PDF; falling back to header bytes", zap.String("path", filePath), zap.Error(rasterErr))
				docText = string(buf[:n])
			} else {
				var b strings.Builder
				for _, img := range images {
					res, oerr := s.ocrService.ExtractTextFromImage(ctx, img)
					if oerr != nil {
						s.logger.Error("OCR failed for page image", zap.String("image", img), zap.Error(oerr))
						continue
					}
					if res != nil && res.Text != "" {
						b.WriteString(res.Text)
						b.WriteString("\n\n")
					}
				}
				docText = strings.TrimSpace(b.String())
				if docText == "" {
					docText = string(buf[:n])
				}
			}
		}
	} else {
		docText = string(buf[:n])
	}

	validationResult, err := s.validationService.ValidateContract(ctx, docText)
	if err != nil {
		return "", fmt.Errorf("failed to validate contract: %w", err)
	}

	newContract := &models.Contract{
		ID:           uuid.New().String(),
		FilePath:     filePath,
		Status:       models.Validated,
		ContractType: validationResult.ContractType,
		Validation:   validationResult,
	}

	if err := s.contractRepo.Create(newContract); err != nil {
		return "", fmt.Errorf("failed to create contract record: %w", err)
	}

	return newContract.ID, nil
}

func (s *documentService) GetByID(ctx context.Context, id string) (*models.Contract, error) {
	return s.contractRepo.GetByID(id)
}

func (s *documentService) Delete(ctx context.Context, id string) error {
	contract, err := s.contractRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get contract for deletion: %w", err)
	}

	if err := s.storage.Delete(contract.FilePath); err != nil {
		s.logger.Error("failed to delete file from storage", zap.String("path", contract.FilePath), zap.Error(err))
	}

	return s.contractRepo.Delete(id)
}

func getMimeType(buffer []byte, extension string) string {
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
