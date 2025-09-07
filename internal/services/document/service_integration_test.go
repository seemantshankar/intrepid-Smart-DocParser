package document_test

import (
	"context"
	"mime/multipart"
	"os"
	"strings"
	"testing"

	"contract-analysis-service/configs"
	"contract-analysis-service/internal/pkg/database"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/repositories/sqlite"
	"contract-analysis-service/internal/services/document"
	"contract-analysis-service/internal/services/llm"
	"contract-analysis-service/internal/services/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDocumentService_Integration_Upload_Get_Delete(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set RUN_DOCUMENT_INTEGRATION_TESTS to run")
	}

	// Load config
	cfg, err := configs.LoadConfig("../../../../config_test.yaml")
	require.NoError(t, err, "Failed to load test config")

	// Dependencies
	logger := zap.NewNop()
	db, err := database.NewDB(cfg.Database)
	require.NoError(t, err, "Failed to connect to test database")
	contractRepo := sqlite.NewContractRepository(db)
	fileStorage, err := storage.NewLocalStorage("../../../../test-uploads")
	require.NoError(t, err, "Failed to create test storage")

	// Create services
	llmService := llm.NewLLMService(logger)
	validationService := validation.NewValidationService(llmService, logger)
	service := document.NewDocumentService(logger, fileStorage, contractRepo, validationService)

	// --- Test Upload ---
	fileContent := "integration test file content"
	fileHeader := &multipart.FileHeader{
		Filename: "integration_test.txt",
		Size:     int64(len(fileContent)),
	}

	documentID, err := service.Upload(context.Background(), strings.NewReader(fileContent), fileHeader)
	require.NoError(t, err)
	require.NotEmpty(t, documentID)

	// --- Test GetByID ---
	contract, err := service.GetByID(context.Background(), documentID)
	require.NoError(t, err)
	require.NotNil(t, contract)
	assert.Equal(t, documentID, contract.ID)

	// --- Test Delete ---
	err = service.Delete(context.Background(), documentID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = service.GetByID(context.Background(), documentID)
	assert.Error(t, err, "Expected error when getting deleted contract")

	// Clean up the created file
	_, err = os.Stat(contract.FilePath)
	assert.True(t, os.IsNotExist(err), "Expected file to be deleted")
}
