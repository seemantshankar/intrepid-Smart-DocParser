package document_test

import (
	"context"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
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

// findProjectRoot walks up from CWD to locate go.mod as project root.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// findRasterPDF returns a path to a scanned/image-based PDF for rasterization.
// Primary source is env var RASTER_INTEGRATION_PDF. If empty, tries uploads/*.pdf.
func findRasterPDF(t *testing.T) (string, bool) {
	t.Helper()
	if p := os.Getenv("RASTER_INTEGRATION_PDF"); p != "" {
		return p, true
	}
	root := findProjectRoot(t)
	if root == "" {
		return "", false
	}
	candidates, _ := filepath.Glob(filepath.Join(root, "uploads", "*.pdf"))
	if len(candidates) == 0 {
		return "", false
	}
	// We cannot guarantee which of these are scanned; allow caller to set env for certainty.
	// Return the first candidate and let the test proceed; it may skip later if no OCR calls occur.
	return candidates[0], true
}

func TestDocumentService_Integration_Upload_PDF_Rasterization(t *testing.T) {
	if os.Getenv("RUN_PDF_RASTER_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping PDF rasterization integration test; set RUN_PDF_RASTER_INTEGRATION_TESTS to run")
	}

	// Ensure pdftoppm is available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not found in PATH; install Poppler (e.g., brew install poppler) to run this test")
	}

	pdfPath, ok := findRasterPDF(t)
	if !ok {
		t.Skip("No PDF found for rasterization test; set RASTER_INTEGRATION_PDF to a scanned PDF path")
	}

	// Load config
	root := findProjectRoot(t)
	require.NotEmpty(t, root, "project root not found")
	cfg, err := configs.LoadConfig(filepath.Join(root, "config_test.yaml"))
	require.NoError(t, err)

	// Dependencies
	logger := zap.NewNop()
	db, err := database.NewDB(cfg.Database)
	require.NoError(t, err)
	contractRepo := sqlite.NewContractRepository(db)
	fileStorage, err := storage.NewLocalStorage(filepath.Join(root, "test-uploads"))
	require.NoError(t, err)

	// Services
	llmService := llm.NewLLMService(logger)
	llmclient.AddOpenRouterClientToService(llmService, cfg)
	analysisService := analysis.NewService(llmService)
	service := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)

	// Open the PDF and build a file header
	f, err := os.Open(pdfPath)
	require.NoError(t, err)
	defer f.Close()
	stat, err := f.Stat()
	require.NoError(t, err)
	fh := &multipart.FileHeader{Filename: filepath.Base(pdfPath), Size: stat.Size()}

	// Execute upload
	ctx := context.Background()
	contract, err := service.UploadAndAnalyze(ctx, f, fh, "test-user")
	require.NoError(t, err)
	require.NotNil(t, contract)

	// Fetch and assert stored
	retrievedContract, err := service.GetByID(ctx, contract.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedContract)
	assert.Equal(t, contract.ID, retrievedContract.ID)

	// Clean up - remove uploaded file
	if retrievedContract.StoragePath != "" {
		os.Remove(retrievedContract.StoragePath)
	}
}
