package handlers

import (
	"net/http"

	"contract-analysis-service/internal/services/document"
	"contract-analysis-service/internal/services/validation"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DocumentHandler handles HTTP requests for document operations.
type DocumentHandler struct {
	service           document.Service
	validationService validation.Service
	logger            *zap.Logger
}

// UploadAndAnalyze handles the upload + analyze endpoint.
// @Summary Upload and analyze a contract in one step
// @Description Uploads a document, securely stores it, performs one-step multimodal analysis for scanned PDFs, and returns the stored document_id along with structured analysis JSON.
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "The document to upload and analyze"
// @Success 201 {object} map[string]interface{} "Returns document_id and analysis JSON"
// @Failure 400 {object} map[string]string "Bad request if the file is missing, invalid, or too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/upload-analyze [post]
func (h *DocumentHandler) UploadAndAnalyze(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
    fileHeader, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
        return
    }

    file, err := fileHeader.Open()
    if err != nil {
        h.logger.Error("Failed to open uploaded file", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
        return
    }
    defer file.Close()

    contract, err := h.service.UploadAndAnalyze(c.Request.Context(), file, fileHeader, userIDStr)
    if err != nil {
        h.logger.Error("Failed to upload and analyze document", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "document_id": contract.ID,
        "analysis":    contract.Analysis,
    })
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(service document.Service, validationService validation.Service, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{
		service:           service,
		validationService: validationService,
		logger:            logger,
	}
}

// Upload handles the document upload endpoint.
// @Summary Upload a document for analysis
// @Description Upload a single document (PDF, DOCX, TXT, JPG, PNG, TIFF) with a 10MB size limit.
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "The document to upload"
// @Success 201 {object} map[string]string "Returns the ID of the uploaded document"
// @Failure 400 {object} map[string]string "Bad request if the file is missing, invalid, or too large"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/upload [post]
func (h *DocumentHandler) Upload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.logger.Error("Failed to open uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	documentID, err := h.service.Upload(c.Request.Context(), file, fileHeader, userIDStr)
	if err != nil {
		h.logger.Error("Failed to upload document", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"document_id": documentID})
}

// Get retrieves a document by its ID.
// @Summary Get a document by ID
// @Description Get a single document's metadata by its unique ID.
// @Tags Documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} models.Contract
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/{id} [get]
func (h *DocumentHandler) Get(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	id := c.Param("id")
	document, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil || document == nil {
		h.logger.Error("Failed to get document", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if document.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to document"})
		return
	}
	c.JSON(http.StatusOK, document)
}

// RetrieveAnalysis retrieves the analysis results for a document by its ID.
// @Summary Get analysis results for a document
// @Description Retrieves the structured analysis JSON for a previously analyzed document.
// @Tags Documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} map[string]interface{} "Analysis results"
// @Failure 404 {object} map[string]string "Document not found or no analysis available"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/{id}/analysis [get]
func (h *DocumentHandler) RetrieveAnalysis(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	id := c.Param("id")
	document, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil || document == nil {
		h.logger.Error("Failed to get document for analysis", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if document.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to analysis"})
		return
	}
	analysis, err := h.service.RetrieveAnalysis(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve analysis", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}

	// Extract key fields from analysis
	contractID := analysis["contract_id"]
	contractName := analysis["contract_name"]
	parties := analysis["parties_involved"]
	effectiveDate := analysis["effective_date"]
	terminationDate := analysis["termination_date"]

	// Create response with extracted fields
	response := gin.H{
		"contract_id":      contractID,
		"contract_name":    contractName,
		"parties_involved": parties,
		"effective_date":   effectiveDate,
		"termination_date": terminationDate,
	}

	c.JSON(http.StatusOK, response)
}

// Delete handles the document deletion endpoint.
// @Summary Delete a document by ID
// @Description Delete a single document by its unique ID. This will remove the file from storage and the record from the database.
// @Tags Documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/{id} [delete]
func (h *DocumentHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	id := c.Param("id")
	document, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil || document == nil {
		h.logger.Error("Failed to get document for deletion", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if document.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized to delete document"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DetectElements detects contract elements (parties, obligations, terms) from a document.
// @Summary Detect contract elements
// @Description Analyzes a stored document to extract detailed contract elements including parties, obligations, and terms
// @Tags Documents
// @Produce json
// @Param document_id path string true "Document ID"
// @Success 200 {object} models.ContractElementsResult "Contract elements detection result"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /contracts/{document_id}/elements [get]
func (h *DocumentHandler) DetectElements(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_id is required"})
		return
	}

	// Get the document content
	document, err := h.service.GetDocument(c.Request.Context(), documentID, userIDStr)
	if err != nil {
		h.logger.Error("Failed to get document", zap.Error(err), zap.String("document_id", documentID))
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// Read document content
	content, err := h.service.GetDocumentContent(c.Request.Context(), document.StoragePath)
	if err != nil {
		h.logger.Error("Failed to read document content", zap.Error(err), zap.String("document_id", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read document content"})
		return
	}

	// Detect contract elements
	elements, err := h.validationService.DetectContractElements(c.Request.Context(), string(content))
	if err != nil {
		h.logger.Error("Failed to detect contract elements", zap.Error(err), zap.String("document_id", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to detect contract elements"})
		return
	}

	c.JSON(http.StatusOK, elements)
}
