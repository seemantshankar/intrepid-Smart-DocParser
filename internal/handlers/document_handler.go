package handlers

import (
	"net/http"

	"contract-analysis-service/internal/services/document"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DocumentHandler handles HTTP requests for document operations.
type DocumentHandler struct {
	service document.Service
	logger  *zap.Logger
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

    documentID, analysis, err := h.service.UploadAndAnalyze(c.Request.Context(), file, fileHeader)
    if err != nil {
        h.logger.Error("Failed to upload and analyze document", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "document_id": documentID,
        "analysis":    analysis,
    })
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(service document.Service, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{
		service: service,
		logger:  logger,
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

	documentID, err := h.service.Upload(c.Request.Context(), file, fileHeader)
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
	id := c.Param("id")

	document, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get document", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	c.JSON(http.StatusOK, document)
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
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document"})
		return
	}

	c.Status(http.StatusNoContent)
}
