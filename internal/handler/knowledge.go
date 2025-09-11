package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"contract-analysis-service/internal/repository"
	"contract-analysis-service/internal/service"
	"github.com/gin-gonic/gin"
)

type KnowledgeHandler struct {
	service service.KnowledgeService
}

func NewKnowledgeHandler(svc service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{service: svc}
}

func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	var req struct {
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Category string   `json:"category"`
		Source   string   `json:"source"`
		Tags     []string `json:"tags"`
		Metadata string   `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.CreateKnowledge(c.Request.Context(), req.Title, req.Content, req.Category, req.Source, req.Tags, req.Metadata); err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusCreated, gin.H{"message": "Knowledge created successfully"})
}

func (h *KnowledgeHandler) GetKnowledge(c *gin.Context) {
	id := c.Param("id")
	entry, err := h.service.GetKnowledge(c.Request.Context(), id)
	if err != nil {
		h.sendError(c, http.StatusNotFound, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entry)
}

func (h *KnowledgeHandler) UpdateKnowledge(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Metadata string   `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.UpdateKnowledge(c.Request.Context(), id, req.Title, req.Content, req.Category, req.Tags, req.Metadata); err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, gin.H{"message": "Knowledge updated successfully"})
}

func (h *KnowledgeHandler) DeleteKnowledge(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteKnowledge(c.Request.Context(), id); err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusNoContent, nil)
}

func (h *KnowledgeHandler) SearchKnowledge(c *gin.Context) {
	query := c.Query("q")
	category := c.Query("category")
	entries, err := h.service.SearchKnowledge(c.Request.Context(), query, category)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entries)
}

func (h *KnowledgeHandler) ListKnowledge(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit == 0 {
		limit = 10
	}
	if offset == 0 {
		offset = 0
	}
	entries, err := h.service.ListKnowledge(c.Request.Context(), limit, offset)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entries)
}

func (h *KnowledgeHandler) sendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

func (h *KnowledgeHandler) sendSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// UpdateKnowledgeWithConflictDetection updates knowledge with version conflict detection
func (h *KnowledgeHandler) UpdateKnowledgeWithConflictDetection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title    string   `json:"title" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Metadata string   `json:"metadata"`
		Version  int      `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.UpdateKnowledgeWithConflictDetection(c.Request.Context(), id, req.Title, req.Content, req.Category, req.Tags, req.Metadata, req.Version); err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			h.sendError(c, http.StatusConflict, err.Error())
		} else {
			h.sendError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.sendSuccess(c, http.StatusOK, gin.H{"message": "Knowledge updated successfully"})
}

// SearchKnowledgeAdvanced performs advanced search with filters
func (h *KnowledgeHandler) SearchKnowledgeAdvanced(c *gin.Context) {
	var req struct {
		Query         string     `json:"query"`
		Category      string     `json:"category"`
		Tags          []string   `json:"tags"`
		Source        string     `json:"source"`
		CreatedAfter  *time.Time `json:"created_after"`
		CreatedBefore *time.Time `json:"created_before"`
		Limit         int        `json:"limit"`
		Offset        int        `json:"offset"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err.Error())
		return
	}
	
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	filter := repository.KnowledgeSearchFilter{
		Query:         req.Query,
		Category:      req.Category,
		Tags:          req.Tags,
		Source:        req.Source,
		CreatedAfter:  req.CreatedAfter,
		CreatedBefore: req.CreatedBefore,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}
	
	entries, err := h.service.SearchKnowledgeAdvanced(c.Request.Context(), filter)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entries)
}

// GetVersionHistory returns version history of a knowledge entry
func (h *KnowledgeHandler) GetVersionHistory(c *gin.Context) {
	id := c.Param("id")
	versions, err := h.service.GetVersionHistory(c.Request.Context(), id)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, versions)
}

// GetLatestVersion returns the latest version of a knowledge entry
func (h *KnowledgeHandler) GetLatestVersion(c *gin.Context) {
	id := c.Param("id")
	entry, err := h.service.GetLatestVersion(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendError(c, http.StatusNotFound, err.Error())
		} else {
			h.sendError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.sendSuccess(c, http.StatusOK, entry)
}

// FindSimilarContent finds entries with similar content
func (h *KnowledgeHandler) FindSimilarContent(c *gin.Context) {
	var req struct {
		Content   string  `json:"content" binding:"required"`
		Threshold float64 `json:"threshold"`
		Limit     int     `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err.Error())
		return
	}
	
	// Set defaults
	if req.Threshold <= 0 {
		req.Threshold = 0.3
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	entries, err := h.service.FindSimilarContent(c.Request.Context(), req.Content, req.Threshold, req.Limit)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entries)
}

// FullTextSearch performs ranked full-text search
func (h *KnowledgeHandler) FullTextSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		h.sendError(c, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}
	
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	
	if limit <= 0 {
		limit = 10
	}
	
	entries, err := h.service.FullTextSearch(c.Request.Context(), query, limit, offset)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendSuccess(c, http.StatusOK, entries)
}

