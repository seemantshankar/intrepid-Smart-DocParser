package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/repositories"
	"contract-analysis-service/internal/services/llm"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service defines the interface for the contract validation service.
type Service interface {
	ValidateContract(ctx context.Context, documentText string) (*models.ValidationResult, error)
	DetectContractElements(ctx context.Context, documentText string) (*models.ContractElementsResult, error)
	StoreValidationResult(ctx context.Context, contractID, userID string, result *models.ValidationResult, elements *models.ContractElementsResult) (*models.ValidationRecord, error)
	GetValidationHistory(ctx context.Context, contractID string) ([]*models.ValidationRecord, error)
	AddValidationFeedback(ctx context.Context, validationID, userID string, feedbackType string, rating int, comment string) error
	GetValidationAuditTrail(ctx context.Context, validationID string) ([]*models.ValidationAuditLog, error)
	// Enhanced confidence scoring methods
	CalculateConfidenceScore(ctx context.Context, result *models.ValidationResult, elements *models.ContractElementsResult) float64
	GetConfidenceFactors(ctx context.Context, result *models.ValidationResult, elements *models.ContractElementsResult) map[string]float64
	UpdateConfidenceBasedOnFeedback(ctx context.Context, validationID string) error
}

// validationService implements the Service interface.
type validationService struct {
	llmService         llm.Service
	logger             *zap.Logger
	validationRepo     repositories.ValidationRepository
	auditRepo          repositories.ValidationAuditRepository
	feedbackRepo       repositories.ValidationFeedbackRepository
}

// NewValidationService creates a new validation service instance.
func NewValidationService(llmService llm.Service, logger *zap.Logger, validationRepo repositories.ValidationRepository, auditRepo repositories.ValidationAuditRepository, feedbackRepo repositories.ValidationFeedbackRepository) Service {
	return &validationService{
		llmService:     llmService,
		logger:         logger,
		validationRepo: validationRepo,
		auditRepo:      auditRepo,
		feedbackRepo:   feedbackRepo,
	}
}

// ValidateContract uses the LLM service to determine if a document is a valid contract.
func (s *validationService) ValidateContract(ctx context.Context, documentText string) (*models.ValidationResult, error) {
	prompt := buildValidationPrompt(documentText)

	// For simplicity, we'll use the default provider. A more advanced implementation
	// could allow for provider selection.
	provider := "openrouter"

	// Prepare payload for the LLM request
	payload := map[string]interface{}{
		"model": "qwen/qwen-2.5-vl-72b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal validation payload: %w", err)
	}

	// Execute the request using the LLM service
	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("contract validation request failed: %w", err)
	}

	// Parse the validation response
	return parseValidationResponse(resp.Body)
}

func buildValidationPrompt(documentText string) string {
	return fmt.Sprintf(`Analyze the following document and determine if it is a valid legal contract. 

Perform comprehensive contract element detection and validation:

1. VALIDATION: Determine if this is a valid legal contract
2. ELEMENT DETECTION: Identify key contract elements present in the document
3. MISSING ELEMENTS: Identify essential contract elements that are missing

Respond with a JSON object containing these keys:
- 'is_valid_contract' (boolean): true if this is a valid legal contract
- 'reason' (string): explanation if not valid, omit if valid
- 'confidence' (float, 0.0-1.0): confidence level in the validation
- 'contract_type' (string): type of contract (e.g., 'Sale of Goods', 'Service Agreement', 'Employment Contract', 'Lease Agreement')
- 'detected_elements' (array of strings): contract elements found in the document from this list:
  ["parties_identification", "offer_and_acceptance", "consideration", "legal_capacity", "mutual_consent", "lawful_purpose", "payment_terms", "delivery_terms", "performance_obligations", "termination_clauses", "dispute_resolution", "governing_law", "signatures", "dates", "contact_information", "warranties", "liability_limitations", "force_majeure", "confidentiality", "intellectual_property"]
- 'missing_elements' (array of strings): essential elements missing from the document using the same element names as above

Document:\n\n%s`, documentText)
}

func parseValidationResponse(body []byte) (*models.ValidationResult, error) {
	// Log the response body for debugging
	log.Printf("LLM Response Body: %s", string(body))

	// OpenRouter chat completion response with response_format json_object returns
	// a JSON string in choices[0].message.content that contains our ValidationResult fields
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in validation response")
	}

	var result models.ValidationResult
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse validation result from content: %w", err)
	}

	return &result, nil
}

// DetectContractElements extracts detailed contract elements like parties, obligations, and terms.
func (s *validationService) DetectContractElements(ctx context.Context, documentText string) (*models.ContractElementsResult, error) {
	prompt := buildElementDetectionPrompt(documentText)

	// Use the same provider as validation
	provider := "openrouter"

	// Prepare payload for the LLM request
	payload := map[string]interface{}{
		"model": "qwen/qwen-2.5-vl-72b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal element detection payload: %w", err)
	}

	// Execute the request using the LLM service
	resp, err := s.llmService.ExecuteRequest(ctx, provider, &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    payloadBytes,
	})
	if err != nil {
		s.logger.Error("Failed to analyze contract for element detection", zap.Error(err))
		return nil, fmt.Errorf("failed to analyze contract for element detection: %w", err)
	}

	// Parse the response
	result, err := parseElementDetectionResponse(resp.Body)
	if err != nil {
		s.logger.Error("Failed to parse element detection response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse element detection response: %w", err)
	}

	return result, nil
}

func buildElementDetectionPrompt(documentText string) string {
	return fmt.Sprintf(`Extract detailed contract elements from the following document. 

Analyze and extract:
1. PARTIES: All parties involved in the contract
2. OBLIGATIONS: Specific obligations and responsibilities of each party
3. TERMS: Key terms and conditions

Respond with a JSON object containing these keys:
- 'parties' (array of objects): Each party with 'name', 'role' (buyer/seller/contractor/client), 'address', 'contact'
- 'obligations' (array of objects): Each obligation with 'party', 'description', 'type' (payment/delivery/performance), 'deadline'
- 'terms' (array of objects): Each term with 'type' (payment/termination/warranty), 'description', 'value'
- 'confidence' (float, 0.0-1.0): confidence level in the extraction

Document:\n\n%s`, documentText)
}

func parseElementDetectionResponse(body []byte) (*models.ContractElementsResult, error) {
	// Log the response body for debugging
	log.Printf("Element Detection Response Body: %s", string(body))

	// OpenRouter chat completion response with response_format json_object returns
	// a JSON string in choices[0].message.content that contains our ContractElementsResult fields
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse element detection response: %w", err)
	}

	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in element detection response")
	}

	// Parse the content as JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal element detection response: %w", err)
	}

	// Extract detected elements
	result := &models.ContractElementsResult{}
	
	if parties, ok := response["parties"].([]interface{}); ok {
		for _, party := range parties {
			if partyMap, ok := party.(map[string]interface{}); ok {
				contractParty := models.ContractParty{
					Name:    getString(partyMap, "name"),
					Role:    getString(partyMap, "role"),
					Address: getString(partyMap, "address"),
					Contact: getString(partyMap, "contact"),
				}
				result.Parties = append(result.Parties, contractParty)
			}
		}
	}

	if obligations, ok := response["obligations"].([]interface{}); ok {
		for _, obligation := range obligations {
			if obligationMap, ok := obligation.(map[string]interface{}); ok {
				contractObligation := models.ContractObligation{
					Party:       getString(obligationMap, "party"),
					Description: getString(obligationMap, "description"),
					Deadline:    getString(obligationMap, "deadline"),
					Type:        getString(obligationMap, "type"),
				}
				result.Obligations = append(result.Obligations, contractObligation)
			}
		}
	}

	if terms, ok := response["terms"].([]interface{}); ok {
		for _, term := range terms {
			if termMap, ok := term.(map[string]interface{}); ok {
				contractTerm := models.ContractTerm{
					Type:        getString(termMap, "type"),
					Description: getString(termMap, "description"),
					Value:       getString(termMap, "value"),
				}
				result.Terms = append(result.Terms, contractTerm)
			}
		}
	}

	return result, nil
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// StoreValidationResult stores a validation result with audit trail
func (s *validationService) StoreValidationResult(ctx context.Context, contractID, userID string, result *models.ValidationResult, elements *models.ContractElementsResult) (*models.ValidationRecord, error) {
	validationRecord := &models.ValidationRecord{
		ID:             uuid.New().String(),
		ContractID:     contractID,
		UserID:         userID,
		ValidationType: "contract_validation",
		Result:         *result,
		Elements:       elements,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}

	if elements != nil {
		validationRecord.ValidationType = "element_detection"
	}

	err := s.validationRepo.Create(validationRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to store validation result: %w", err)
	}

	// Create audit log entry
	auditLog := &models.ValidationAuditLog{
		ID:              uuid.New().String(),
		ValidationID:    validationRecord.ID,
		UserID:          userID,
		Action:          "created",
		PreviousVersion: 0,
		CurrentVersion:  1,
		Changes:         "{\"action\": \"initial_creation\"}",
		Reason:          "Initial validation result storage",
		CreatedAt:       time.Now(),
	}

	if err := s.auditRepo.Create(auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
		// Don't fail the main operation for audit log failure
	}

	return validationRecord, nil
}

// GetValidationHistory retrieves validation history for a contract
func (s *validationService) GetValidationHistory(ctx context.Context, contractID string) ([]*models.ValidationRecord, error) {
	return s.validationRepo.GetByContractID(contractID)
}

// AddValidationFeedback adds user feedback for a validation result
func (s *validationService) AddValidationFeedback(ctx context.Context, validationID, userID string, feedbackType string, rating int, comment string) error {
	feedback := &models.ValidationFeedback{
		ID:           uuid.New().String(),
		ValidationID: validationID,
		UserID:       userID,
		FeedbackType: feedbackType,
		Rating:       rating,
		Comment:      comment,
		CreatedAt:    time.Now(),
	}

	err := s.feedbackRepo.Create(feedback)
	if err != nil {
		return fmt.Errorf("failed to add validation feedback: %w", err)
	}

	// Create audit log entry for feedback
	auditLog := &models.ValidationAuditLog{
		ID:              uuid.New().String(),
		ValidationID:    validationID,
		UserID:          userID,
		Action:          "feedback_added",
		PreviousVersion: 0,
		CurrentVersion:  0,
		Changes:         fmt.Sprintf("{\"feedback_type\": \"%s\", \"rating\": %d}", feedbackType, rating),
		Reason:          "User feedback added",
		CreatedAt:       time.Now(),
	}

	if err := s.auditRepo.Create(auditLog); err != nil {
		s.logger.Error("Failed to create audit log for feedback", zap.Error(err))
	}

	return nil
}

// GetValidationAuditTrail retrieves the audit trail for a validation
func (s *validationService) GetValidationAuditTrail(ctx context.Context, validationID string) ([]*models.ValidationAuditLog, error) {
	return s.auditRepo.GetByValidationID(validationID)
}

// CalculateConfidenceScore computes an enhanced confidence score based on multiple factors
func (s *validationService) CalculateConfidenceScore(ctx context.Context, result *models.ValidationResult, elements *models.ContractElementsResult) float64 {
	factors := s.GetConfidenceFactors(ctx, result, elements)
	
	// Weighted average of confidence factors
	weights := map[string]float64{
		"llm_confidence":      0.4,  // Base LLM confidence
		"element_completeness": 0.3,  // How many required elements are present
		"contract_structure":  0.2,  // Structural integrity of the contract
		"content_quality":     0.1,  // Text quality and clarity
	}
	
	totalScore := 0.0
	totalWeight := 0.0
	
	for factor, score := range factors {
		if weight, exists := weights[factor]; exists {
			totalScore += score * weight
			totalWeight += weight
		}
	}
	
	if totalWeight == 0 {
		return result.Confidence // Fallback to original confidence
	}
	
	finalScore := totalScore / totalWeight
	
	// Ensure score is within valid range [0.0, 1.0]
	if finalScore < 0.0 {
		return 0.0
	}
	if finalScore > 1.0 {
		return 1.0
	}
	
	return finalScore
}

// GetConfidenceFactors returns individual confidence factors for transparency
func (s *validationService) GetConfidenceFactors(ctx context.Context, result *models.ValidationResult, elements *models.ContractElementsResult) map[string]float64 {
	factors := make(map[string]float64)
	
	// Base LLM confidence
	factors["llm_confidence"] = result.Confidence
	
	// Element completeness factor
	requiredElements := []string{
		"parties_identification", "offer_and_acceptance", "consideration",
		"legal_capacity", "mutual_consent", "lawful_purpose",
	}
	
	detectedCount := 0
	for _, required := range requiredElements {
		for _, detected := range result.DetectedElements {
			if required == detected {
				detectedCount++
				break
			}
		}
	}
	
	factors["element_completeness"] = float64(detectedCount) / float64(len(requiredElements))
	
	// Contract structure factor (based on missing elements)
	missingCount := len(result.MissingElements)
	if missingCount == 0 {
		factors["contract_structure"] = 1.0
	} else {
		// Penalize based on number of missing elements
		factors["contract_structure"] = 1.0 - (float64(missingCount) * 0.1)
		if factors["contract_structure"] < 0.0 {
			factors["contract_structure"] = 0.0
		}
	}
	
	// Content quality factor (based on elements confidence if available)
	if elements != nil {
		factors["content_quality"] = elements.Confidence
	} else {
		// Default to moderate confidence if no elements data
		factors["content_quality"] = 0.7
	}
	
	return factors
}

// UpdateConfidenceBasedOnFeedback adjusts confidence scores based on user feedback
func (s *validationService) UpdateConfidenceBasedOnFeedback(ctx context.Context, validationID string) error {
	// Get validation record
	validation, err := s.validationRepo.GetByID(validationID)
	if err != nil {
		return fmt.Errorf("failed to get validation record: %w", err)
	}
	
	// Get feedback for this validation
	feedbacks, err := s.feedbackRepo.GetByValidationID(validationID)
	if err != nil {
		return fmt.Errorf("failed to get feedback: %w", err)
	}
	
	if len(feedbacks) == 0 {
		return nil // No feedback to process
	}
	
	// Calculate feedback adjustment
	totalRating := 0
	feedbackCount := 0
	
	for _, feedback := range feedbacks {
		if feedback.FeedbackType == "accuracy" {
			totalRating += feedback.Rating
			feedbackCount++
		}
	}
	
	if feedbackCount == 0 {
		return nil // No accuracy feedback
	}
	
	// Calculate average rating (1-5 scale)
	averageRating := float64(totalRating) / float64(feedbackCount)
	
	// Convert to confidence adjustment (-0.2 to +0.2 range)
	// Rating 3 = no change, Rating 5 = +0.2, Rating 1 = -0.2
	adjustment := (averageRating - 3.0) * 0.1
	
	// Apply adjustment to confidence
	newConfidence := validation.Result.Confidence + adjustment
	
	// Ensure confidence stays within valid range
	if newConfidence < 0.0 {
		newConfidence = 0.0
	}
	if newConfidence > 1.0 {
		newConfidence = 1.0
	}
	
	// Update validation record
	validation.Result.Confidence = newConfidence
	validation.Version++
	
	err = s.validationRepo.Update(validation)
	if err != nil {
		return fmt.Errorf("failed to update validation record: %w", err)
	}
	
	// Create audit log
	auditLog := &models.ValidationAuditLog{
		ID:              uuid.New().String(),
		ValidationID:    validationID,
		UserID:          "system", // System-generated update
		Action:          "confidence_updated",
		PreviousVersion: validation.Version - 1,
		CurrentVersion:  validation.Version,
		Changes:         fmt.Sprintf("Confidence updated from %.3f to %.3f based on %d feedback(s)", validation.Result.Confidence-adjustment, newConfidence, feedbackCount),
		Reason:          "Feedback-based confidence adjustment",
		CreatedAt:       time.Now(),
	}
	
	return s.auditRepo.Create(auditLog)
}
