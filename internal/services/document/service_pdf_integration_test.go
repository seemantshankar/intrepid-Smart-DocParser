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
	"contract-analysis-service/internal/services/document"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"contract-analysis-service/internal/services/llm"
	ocrsvc "contract-analysis-service/internal/services/ocr"
	"contract-analysis-service/internal/services/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	validationService := validation.NewValidationService(llmService, logger)

	ocrMock := new(ocrsvc.MockOCRService)
	// We don't know how many pages; allow any calls and return some text.
	ocrMock.On("ExtractTextFromImage", 
		mock.Anything, 
		mock.AnythingOfType("string")).Return(&ocrsvc.OCRResult{Text: "mock ocr text", Confidence: 0.9}, nil).Maybe()

	service := document.NewDocumentService(logger, fileStorage, contractRepo, validationService, ocrMock)

	// Open the PDF and build a file header
	f, err := os.Open(pdfPath)
	require.NoError(t, err)
	defer f.Close()
	stat, err := f.Stat()
	require.NoError(t, err)
	fh := &multipart.FileHeader{Filename: filepath.Base(pdfPath), Size: stat.Size()}

	// Execute upload
	ctx := context.Background()
	id, err := service.Upload(ctx, f, fh)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// Fetch and assert stored
	contract, err := service.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, contract)
	assert.Equal(t, id, contract.ID)

	// We expect OCR to have been attempted for scanned PDFs; if the test PDF wasn't scanned,
	// the path might have taken direct text extraction and not call OCR. In that case, we skip with info.
	if !ocrMock.AssertCalled(t, "ExtractTextFromImage", mock.Anything, mock.AnythingOfType("string")) {
		t.Skip("The provided PDF did not trigger rasterization/OCR. Provide a scanned PDF via RASTER_INTEGRATION_PDF to fully exercise the path.")
	}
}
