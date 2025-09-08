package document_test

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"

	"contract-analysis-service/internal/models"
	repo_mocks "contract-analysis-service/internal/repositories/mocks"
	"contract-analysis-service/internal/services/document"
	validation_mocks "contract-analysis-service/internal/services/validation/mocks"
	storage_mocks "contract-analysis-service/internal/pkg/storage/mocks"
	ocrsvc "contract-analysis-service/internal/services/ocr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDocumentService_Upload(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	validationMock := new(validation_mocks.Service)
	ocrMock := new(ocrsvc.MockOCRService)

	service := document.NewDocumentService(logger, storageMock, contractRepoMock, validationMock, ocrMock)

	fileContent := "this is a test file"
	fileHeader := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     int64(len(fileContent)),
	}

	validationMock.On("ValidateContract", mock.Anything, "this is a test file").Return(&models.ValidationResult{IsValidContract: true, ContractType: "Sale of Goods"}, nil)
	storageMock.On("Save", mock.Anything, "test.txt").Return("/path/to/file.txt", nil)
	contractRepoMock.On("Create", mock.AnythingOfType("*models.Contract")).Return(nil)

	documentID, err := service.Upload(context.Background(), strings.NewReader(fileContent), fileHeader)

	assert.NoError(t, err)
	assert.NotEmpty(t, documentID)
	storageMock.AssertExpectations(t)
	contractRepoMock.AssertExpectations(t)
	validationMock.AssertExpectations(t)
	ocrMock.AssertExpectations(t)
}

func TestDocumentService_GetByID(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	validationMock := new(validation_mocks.Service)
	ocrMock := new(ocrsvc.MockOCRService)

	service := document.NewDocumentService(logger, storageMock, contractRepoMock, validationMock, ocrMock)

	expectedContract := &models.Contract{ID: "test-id"}
	contractRepoMock.On("GetByID", "test-id").Return(expectedContract, nil)

	contract, err := service.GetByID(context.Background(), "test-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedContract, contract)
	contractRepoMock.AssertExpectations(t)
	ocrMock.AssertExpectations(t)
}

func TestDocumentService_Delete(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	validationMock := new(validation_mocks.Service)
	ocrMock := new(ocrsvc.MockOCRService)

	service := document.NewDocumentService(logger, storageMock, contractRepoMock, validationMock, ocrMock)

	contract := &models.Contract{ID: "test-id", FilePath: "/path/to/file.txt"}

	contractRepoMock.On("GetByID", "test-id").Return(contract, nil)
	storageMock.On("Delete", "/path/to/file.txt").Return(nil)
	contractRepoMock.On("Delete", "test-id").Return(nil)

	err := service.Delete(context.Background(), "test-id")

	assert.NoError(t, err)
	storageMock.AssertExpectations(t)
	contractRepoMock.AssertExpectations(t)
	ocrMock.AssertExpectations(t)
}
