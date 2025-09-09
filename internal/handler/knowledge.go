package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"contract-analysis-service/internal/service"
)

type KnowledgeHandler struct {
	service service.KnowledgeService
}

func NewKnowledgeHandler(svc service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{service: svc}
}

func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	var req struct {
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Category string   `json:"category"`
		Source   string   `json:"source"`
		Tags     []string `json:"tags"`
		Metadata map[string]any `json:"metadata"`
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
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Metadata map[string]any `json:"metadata"`
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