package document_test

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"

	"contract-analysis-service/internal/models"
	repo_mocks "contract-analysis-service/internal/repositories/mocks"
	"contract-analysis-service/internal/services/document"
	analysis_mocks "contract-analysis-service/internal/services/analysis/mocks"
	storage_mocks "contract-analysis-service/internal/pkg/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDocumentService_Upload(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	analysisMock := new(analysis_mocks.Service)
	service := document.NewDocumentService(contractRepoMock, storageMock, analysisMock, logger)

	fileContent := "this is a test file"
	fileHeader := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     int64(len(fileContent)),
	}

	analysisMock.On("AnalyzeContract", mock.Anything, "this is a test file").Return(&models.ContractAnalysis{}, nil)  // Adjust return type
	storageMock.On("Save", mock.Anything, "test.txt").Return("/path/to/file.txt", nil)
	contractRepoMock.On("Create", mock.AnythingOfType("*models.Contract")).Return(nil)

	contract, err := service.UploadAndAnalyze(context.Background(), strings.NewReader(fileContent), fileHeader, "test-user")
	assert.NoError(t, err)
	assert.NotNil(t, contract)
	storageMock.AssertExpectations(t)
	contractRepoMock.AssertExpectations(t)
	analysisMock.AssertExpectations(t)
}

func TestDocumentService_GetByID(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	analysisMock := new(analysis_mocks.Service)

	service := document.NewDocumentService(contractRepoMock, storageMock, analysisMock, logger)

	expectedContract := &models.Contract{ID: "test-id"}
	contractRepoMock.On("GetByID", "test-id").Return(expectedContract, nil)

	contract, err := service.GetByID(context.Background(), "test-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedContract, contract)
	contractRepoMock.AssertExpectations(t)
}

func TestDocumentService_Delete(t *testing.T) {
	logger := zap.NewNop()
	storageMock := new(storage_mocks.FileStorage)
	contractRepoMock := new(repo_mocks.ContractRepository)
	analysisMock := new(analysis_mocks.Service)

	service := document.NewDocumentService(contractRepoMock, storageMock, analysisMock, logger)

	expectedContract := &models.Contract{ID: "test-id", StoragePath: "/path/to/file.txt"}
	contractRepoMock.On("GetByID", "test-id").Return(expectedContract, nil)
	storageMock.On("Delete", "/path/to/file.txt").Return(nil)
	contractRepoMock.On("Delete", "test-id").Return(nil)

	err := service.Delete(context.Background(), "test-id")

	assert.NoError(t, err)
	storageMock.AssertExpectations(t)
	contractRepoMock.AssertExpectations(t)
}
