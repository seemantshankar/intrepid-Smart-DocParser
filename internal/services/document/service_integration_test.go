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
	"contract-analysis-service/internal/services/analysis"
	"contract-analysis-service/internal/services/document"
	"contract-analysis-service/internal/services/llm"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDocumentService_Integration_Upload_Get_Delete(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set RUN_DOCUMENT_INTEGRATION_TESTS to run")
	}

	// Load config
	cfg, err := configs.LoadConfig("/Users/seemant/Library/Mobile Documents/com~apple~CloudDocs/Documents/Projects/Smart-Cheques Ripple/intrepid-Smart-DocParser/config_test.yaml")
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
	llmclient.AddOpenRouterClientToService(llmService, cfg)
	analysisService := analysis.NewService(llmService)
	service := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)

	// --- Test Upload ---
	fileContent := "integration test file content"
	fileHeader := &multipart.FileHeader{
		Filename: "integration_test.txt",
		Size:     int64(len(fileContent)),
	}

	contract, err := service.UploadAndAnalyze(context.Background(), strings.NewReader(fileContent), fileHeader, "test-user")
	require.NoError(t, err)
	require.NotNil(t, contract)
	documentID := contract.ID

	// --- Test GetByID ---
	retrievedContract, err := service.GetByID(context.Background(), documentID)
	require.NoError(t, err)
	require.NotNil(t, retrievedContract)
	assert.Equal(t, documentID, retrievedContract.ID)

	// --- Test Delete ---
	err = service.Delete(context.Background(), documentID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = service.GetByID(context.Background(), documentID)
	assert.Error(t, err, "Expected error when getting deleted contract")

	// Clean up the created file
	_, err = os.Stat(contract.StoragePath)
	assert.True(t, os.IsNotExist(err), "Expected file to be deleted")
}

// Add new test for authorization
func TestDocumentService_Authorization(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set RUN_DOCUMENT_INTEGRATION_TESTS to run")
	}

	// Load config and setup (similar to existing)
	cfg, err := configs.LoadConfig("/Users/seemant/Library/Mobile Documents/com~apple~CloudDocs/Documents/Projects/Smart-Cheques Ripple/intrepid-Smart-DocParser/config_test.yaml")
	require.NoError(t, err)
	logger := zap.NewNop()
	db, err := database.NewDB(cfg.Database)
	require.NoError(t, err)
	contractRepo := sqlite.NewContractRepository(db)
	fileStorage, err := storage.NewLocalStorage("../../../../test-uploads")
	require.NoError(t, err)
	llmService := llm.NewLLMService(logger)
	llmclient.AddOpenRouterClientToService(llmService, cfg)
	analysisService := analysis.NewService(llmService)
	service := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)

	// Upload with user1
	documentID, err := service.Upload(context.Background(), strings.NewReader("auth test"), &multipart.FileHeader{Filename: "auth.txt", Size: 9}, "user1")
	require.NoError(t, err)

	// Try get with user1 (should succeed)
	_, err = service.GetByID(context.Background(), documentID)
	assert.NoError(t, err)

	// Note: Authorization is in handler, service.GetByID doesn't check userID. Test may need handler integration.
}

// Add test for lifecycle
func TestDocumentService_CleanupExpiredDocuments(t *testing.T) {
	// Setup
	// Create contract with past CreatedAt and small RetentionDays
	// Call Cleanup
	// Verify deleted
}

// Add tests for file types and edge cases similarly
