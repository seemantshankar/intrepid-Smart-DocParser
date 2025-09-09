package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"contract-analysis-service/configs"
	"contract-analysis-service/internal/handlers"
	"contract-analysis-service/internal/middleware"
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/database"
	"contract-analysis-service/internal/pkg/storage"
	"contract-analysis-service/internal/repositories/sqlite"
	"contract-analysis-service/internal/services/document"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"contract-analysis-service/internal/services/llm"
	"contract-analysis-service/internal/services/analysis"
	"contract-analysis-service/internal/services/validation"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	cfg, err := configs.LoadConfig("../../config_test.yaml")
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	
	logger := zap.NewNop()
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	
	contractRepo := sqlite.NewContractRepository(db)
	fileStorage, err := storage.NewLocalStorage("../../test-uploads")
	if err != nil {
		panic("Failed to create storage: " + err.Error())
	}
	
	llmService := llm.NewLLMService(logger)
	llmclient.AddOpenRouterClientToService(llmService, cfg)
	analysisService := analysis.NewService(llmService)
	
	// Create validation repositories
	validationRepo := sqlite.NewValidationRepository(db)
	validationAuditRepo := sqlite.NewValidationAuditRepository(db)
	validationFeedbackRepo := sqlite.NewValidationFeedbackRepository(db)
	validationService := validation.NewValidationService(llmService, logger, validationRepo, validationAuditRepo, validationFeedbackRepo)
	docService := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)

	r := gin.New()
	m := middleware.NewMiddleware(logger)
	r.Use(m.JWT("test-secret-key"))

	h := handlers.NewDocumentHandler(docService, validationService, logger)
	r.POST("/upload", h.Upload)
	r.GET("/documents/:id", h.Get)
	r.DELETE("/documents/:id", h.Delete)
	// Add other routes as needed

	return r
}

// generateToken creates a test JWT token
func generateToken(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

func TestDocumentHandler_Integration_UploadAndGet_Authorized(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test")
	}

	r := setupTestRouter()

	// Generate JWT token for user1
	token, _ := generateToken("user1", "test-secret-key")
	token = "Bearer " + token

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = io.WriteString(part, "test content")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	documentID := resp["document_id"]
	require.NotEmpty(t, documentID)

	// Get with same user
	reqGet := httptest.NewRequest("GET", "/documents/"+documentID, nil)
	reqGet.Header.Set("Authorization", token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, reqGet)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentHandler_Integration_UploadAndGet_Unauthorized(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test")
	}

	r := setupTestRouter()

	// Generate JWT token for user1
	token1, _ := generateToken("user1", "test-secret-key")
	token1 = "Bearer " + token1
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = io.WriteString(part, "test content")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token1)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	documentID := resp["document_id"]
	require.NotEmpty(t, documentID)

	// Generate JWT token for user2 (different user)
	token2, _ := generateToken("user2", "test-secret-key")
	token2 = "Bearer " + token2
	
	reqGet := httptest.NewRequest("GET", "/documents/"+documentID, nil)
	reqGet.Header.Set("Authorization", token2)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, reqGet)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// Remove misplaced imports and comments inside function
func TestDocumentHandler_Integration_LifecycleCleanup(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test")
	}
	// Setup service as in setupTestRouter
	cfg, err := configs.LoadConfig("../../config_test.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	logger := zap.NewNop()
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	
	contractRepo := sqlite.NewContractRepository(db)
	fileStorage, err2 := storage.NewLocalStorage("../../test-uploads")
	if err2 != nil {
		t.Fatalf("Failed to create storage: %v", err2)
	}
	
	llmService := llm.NewLLMService(logger)
	llmclient.AddOpenRouterClientToService(llmService, cfg)
	analysisService := analysis.NewService(llmService)
	docService := document.NewDocumentService(contractRepo, fileStorage, analysisService, logger)
	
	// Create a contract with old CreatedAt
	oldTime := time.Now().Add(-time.Duration(366*24) * time.Hour) // older than 365 days
	contract := &models.Contract{
		ID:           uuid.New().String(),
		UserID:       "test-user",
		Filename:     "old-contract.pdf",
		StoragePath:  "/tmp/old-contract.pdf",
		CreatedAt:    oldTime,
		RetentionDays: 365,
	}
	err3 := contractRepo.Create(contract)
	require.NoError(t, err3)
	// Call cleanup
	err = docService.CleanupExpiredDocuments(context.Background())
	assert.NoError(t, err)
	// Verify deleted
	_, err = contractRepo.GetByID(contract.ID)
	assert.Error(t, err)
}

// Implement tests for file types: upload PDF, expect success; invalid type, expect 400
// For edge cases: upload file > max size, expect error; zero size, etc.
// Move or adjust lifecycle test to service_integration_test.go if mismatched package
// For now, keep and test service directly
func TestDocumentHandler_Integration_VariousFileTypes(t *testing.T) {
	if os.Getenv("RUN_DOCUMENT_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test")
	}

	r := setupTestRouter()
	token, _ := generateToken("user2", "test-secret-key")

	// Test PDF
	bodyPDF := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyPDF)
	part, _ := writer.CreateFormFile("file", "test.pdf")
	_, _ = io.WriteString(part, "%PDF-1.4 test")
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", bodyPDF)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test invalid type, e.g., .exe
	bodyInvalid := &bytes.Buffer{}
	writer = multipart.NewWriter(bodyInvalid)
	part, _ = writer.CreateFormFile("file", "test.exe")
	_, _ = io.WriteString(part, "invalid")
	writer.Close()
	req = httptest.NewRequest("POST", "/upload", bodyInvalid)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Assuming validation returns 400
}

func TestDocumentHandler_Integration_EdgeCases(t *testing.T) {
	// Similar setup
	// Large file: create body > maxFileSize, expect error
	// Zero size: empty file, expect success or error based on logic
}