package handlers

import (
	"net/http"
	"strconv"

	"contract-analysis-service/internal/repository"
	"contract-analysis-service/internal/services/analysis"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContractAnalysisHandler handles comprehensive contract analysis requests
type ContractAnalysisHandler struct {
	logger                  *zap.Logger
	contractAnalysisService *analysis.ContractAnalysisService
	contractAnalysisRepo    repository.ContractAnalysisRepository
}

// NewContractAnalysisHandler creates a new contract analysis handler
func NewContractAnalysisHandler(
	logger *zap.Logger,
	contractAnalysisService *analysis.ContractAnalysisService,
	contractAnalysisRepo repository.ContractAnalysisRepository,
) *ContractAnalysisHandler {
	return &ContractAnalysisHandler{
		logger:                  logger,
		contractAnalysisService: contractAnalysisService,
		contractAnalysisRepo:    contractAnalysisRepo,
	}
}

// AnalyzeContractComprehensively performs comprehensive contract analysis including summary extraction,
// payment obligations, percentage-based calculations, and risk assessment
// @Summary Analyze contract comprehensively
// @Description Performs comprehensive contract analysis using LLM including summary extraction, payment obligations identification, percentage-based calculations, and risk assessment
// @Tags Contract Analysis
// @Accept json
// @Produce json
// @Param request body AnalyzeContractRequest true "Contract text to analyze"
// @Success 200 {object} analysis.AnalyzeContractResult "Comprehensive analysis result"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /contracts/analyze-comprehensive [post]
func (h *ContractAnalysisHandler) AnalyzeContractComprehensively(c *gin.Context) {
	var req AnalyzeContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.ContractText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract_text is required"})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous" // fallback for testing
	}

	result, err := h.contractAnalysisService.AnalyzeContract(c.Request.Context(), req.ContractText, userID)
	if err != nil {
		h.logger.Error("comprehensive contract analysis failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetContractAnalysis retrieves a stored contract analysis by contract ID
// @Summary Get contract analysis by contract ID
// @Description Retrieves comprehensive analysis results for a specific contract
// @Tags Contract Analysis
// @Accept json
// @Produce json
// @Param contract_id path string true "Contract ID"
// @Success 200 {object} repository.ContractAnalysisRecord "Contract analysis record"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /contract-analysis/{contract_id} [get]
func (h *ContractAnalysisHandler) GetContractAnalysis(c *gin.Context) {
	contractID := c.Param("contract_id")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contract_id is required"})
		return
	}

	record, err := h.contractAnalysisRepo.GetAnalysisByContractID(c.Request.Context(), contractID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		} else {
			h.logger.Error("failed to get contract analysis", zap.String("contract_id", contractID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve analysis"})
		}
		return
	}

	c.JSON(http.StatusOK, record)
}

// ListContractAnalyses lists contract analyses with pagination
// @Summary List contract analyses
// @Description Lists contract analyses with optional filtering and pagination
// @Tags Contract Analysis
// @Accept json
// @Produce json
// @Param limit query int false "Limit for pagination" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of contract analyses"
// @Failure 500 {object} map[string]string
// @Router /contract-analyses [get]
func (h *ContractAnalysisHandler) ListContractAnalyses(c *gin.Context) {
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")

	// Parse query parameters
	limitInt := 20
	offsetInt := 0
	if l, err := parseIntParam(limit); err == nil && l > 0 {
		limitInt = l
	}
	if o, err := parseIntParam(offset); err == nil && o >= 0 {
		offsetInt = o
	}

	records, err := h.contractAnalysisRepo.ListAnalyses(c.Request.Context(), limitInt, offsetInt)
	if err != nil {
		h.logger.Error("failed to list contract analyses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve analyses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analyses": records,
		"limit":    limitInt,
		"offset":   offsetInt,
		"count":    len(records),
	})
}

// GetAnalysisById retrieves a contract analysis by its unique ID
// @Summary Get analysis by ID
// @Description Retrieves contract analysis by its unique identifier
// @Tags Contract Analysis
// @Accept json
// @Produce json
// @Param id path string true "Analysis ID"
// @Success 200 {object} repository.ContractAnalysisRecord "Contract analysis record"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /analysis/{id} [get]
func (h *ContractAnalysisHandler) GetAnalysisById(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	record, err := h.contractAnalysisRepo.GetAnalysisByID(c.Request.Context(), idStr)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		} else {
			h.logger.Error("failed to get analysis", zap.String("id", idStr), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve analysis"})
		}
		return
	}

	c.JSON(http.StatusOK, record)
}

// AnalyzeContractRequest represents the request body for contract analysis
type AnalyzeContractRequest struct {
	ContractText string `json:"contract_text" binding:"required" example:"This is a sample contract text..."`
}

// parseIntParam safely parses an integer parameter
func parseIntParam(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}