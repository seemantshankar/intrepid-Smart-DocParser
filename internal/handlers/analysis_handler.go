package handlers

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	pdfutil "contract-analysis-service/internal/pkg/pdf"
	"contract-analysis-service/internal/services/analysis"
	"contract-analysis-service/internal/services/ocr"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AnalysisHandler handles ad-hoc analysis requests
type AnalysisHandler struct {
	logger          *zap.Logger
	ocrService      ocr.Service
	analysisService analysis.Service
}

func NewAnalysisHandler(logger *zap.Logger, ocrSvc ocr.Service, analysisSvc analysis.Service) *AnalysisHandler {
	return &AnalysisHandler{logger: logger, ocrService: ocrSvc, analysisService: analysisSvc}
}

const maxAnalyzeFileSize = 10 * 1024 * 1024 // 10MB

// Analyze handles file upload and performs one-step multimodal LLM analysis.
// For PDFs, it rasterizes up to 10 pages to JPEGs and feeds images directly to the LLM
// for extraction + analysis in a single request. If rasterization fails, it falls back
// to a lightweight text-based prompt using header bytes.
// @Summary Analyze a contract document
// @Description Upload a document (PDF preferred). If scanned, pages are rasterized and OCR'd. The extracted text is analyzed using LLM prompts to return structured JSON.
// @Tags Analysis
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "The document to analyze"
// @Success 200 {object} map[string]interface{} "Structured analysis JSON"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /contracts/analyze [post]
func (h *AnalysisHandler) Analyze(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fh.Size > maxAnalyzeFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	file, err := fh.Open()
	if err != nil {
		h.logger.Error("failed to open uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	// Peek header for mime inference and to allow re-read
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	// Only implement PDF path now; other types can be added later
	if ext == ".pdf" {
		// Buffer the file to a temp PDF path
		tmp, tmpErr := os.CreateTemp("", "analyze-*.pdf")
		if tmpErr != nil {
			h.logger.Error("failed to create temp file", zap.Error(tmpErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer func() { _ = os.Remove(tmp.Name()) }()
		if _, copyErr := io.Copy(tmp, reader); copyErr != nil {
			h.logger.Error("failed to buffer upload", zap.Error(copyErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		_ = tmp.Close()

		images, rerr := pdfutil.RasterizeToJPEGs(tmp.Name(), 10)
		if rerr == nil && len(images) > 0 {
			// Preferred path: one-step multimodal analysis on images
			result, aerr := h.analysisService.AnalyzeContractFromImages(c.Request.Context(), images)
			if aerr != nil {
				h.logger.Error("multimodal analysis failed", zap.Error(aerr))
				c.JSON(http.StatusInternalServerError, gin.H{"error": aerr.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
			return
		}

		// Fallback: minimal text using header bytes
		result, aerr := h.analysisService.AnalyzeContract(c.Request.Context(), string(buf[:n]))
		if aerr != nil {
			h.logger.Error("text-based analysis failed", zap.Error(aerr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": aerr.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// Non-PDF fallback: use header bytes
	result, err := h.analysisService.AnalyzeContract(c.Request.Context(), string(buf[:n]))
	if err != nil {
		h.logger.Error("analysis failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
