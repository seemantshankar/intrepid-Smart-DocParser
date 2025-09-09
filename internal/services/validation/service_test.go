package validation_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/repositories/mocks"
	llm_mocks "contract-analysis-service/internal/services/llm/mocks"
	"contract-analysis-service/internal/services/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestValidationService_ValidateContract(t *testing.T) {
	logger := zap.NewNop()
	llmMock := new(llm_mocks.Service)
	validationRepoMock := new(mocks.ValidationRepository)
	auditRepoMock := new(mocks.ValidationAuditRepository)
	feedbackRepoMock := new(mocks.ValidationFeedbackRepository)

	service := validation.NewValidationService(llmMock, logger, validationRepoMock, auditRepoMock, feedbackRepoMock)

	// Mock the LLM response
	validationResult := &models.ValidationResult{
		IsValidContract: true,
		Confidence:      0.95,
		ContractType:    "Sale of Goods",
	}
	validationResultJSON, _ := json.Marshal(validationResult)

	// Construct the full expected response from the LLM
	llmResponseContent := struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: string(validationResultJSON),
				},
			},
		},
	}
	responseBody, _ := json.Marshal(llmResponseContent)

	llmResponse := &external.Response{
		StatusCode: 200,
		Body:       responseBody,
	}

	llmMock.On("ExecuteRequest", mock.Anything, "openrouter", mock.Anything).Return(llmResponse, nil)

	// Call the method
	result, err := service.ValidateContract(context.Background(), "some document text")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsValidContract)
	assert.Equal(t, "Sale of Goods", result.ContractType)
	llmMock.AssertExpectations(t)
}

func TestValidationService_ValidateContract_Error(t *testing.T) {
	logger := zap.NewNop()
	llmMock := new(llm_mocks.Service)
	validationRepoMock := new(mocks.ValidationRepository)
	auditRepoMock := new(mocks.ValidationAuditRepository)
	feedbackRepoMock := new(mocks.ValidationFeedbackRepository)

	service := validation.NewValidationService(llmMock, logger, validationRepoMock, auditRepoMock, feedbackRepoMock)

	// Mock the LLM error
	expectedError := errors.New("LLM API error")
	llmMock.On("ExecuteRequest", mock.Anything, "openrouter", mock.Anything).Return(nil, expectedError)

	// Call the method
	result, err := service.ValidateContract(context.Background(), "some document text")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, expectedError.Error())
	llmMock.AssertExpectations(t)
}

func TestValidationService_CalculateConfidenceScore(t *testing.T) {
	llmMock := &llm_mocks.Service{}
	logger := zap.NewNop()
	validationRepoMock := &mocks.ValidationRepository{}
	auditRepoMock := &mocks.ValidationAuditRepository{}
	feedbackRepoMock := &mocks.ValidationFeedbackRepository{}

	validationService := validation.NewValidationService(llmMock, logger, validationRepoMock, auditRepoMock, feedbackRepoMock)

	// Test case 1: High confidence with all required elements
	result := &models.ValidationResult{
		IsValidContract: true,
		Confidence:      0.9,
		ContractType:    "Sale of Goods",
		DetectedElements: []string{
			"parties_identification", "offer_and_acceptance", "consideration",
			"legal_capacity", "mutual_consent", "lawful_purpose",
		},
		MissingElements: []string{},
	}

	elements := &models.ContractElementsResult{
		Confidence: 0.85,
	}

	confidenceScore := validationService.CalculateConfidenceScore(context.Background(), result, elements)

	// Should be high confidence (weighted average of factors)
	assert.Greater(t, confidenceScore, 0.8)
	assert.LessOrEqual(t, confidenceScore, 1.0)

	// Test case 2: Lower confidence with missing elements
	result2 := &models.ValidationResult{
		IsValidContract: true,
		Confidence:      0.6,
		ContractType:    "Service Agreement",
		DetectedElements: []string{"parties_identification", "consideration"},
		MissingElements:  []string{"offer_and_acceptance", "legal_capacity"},
	}

	confidenceScore2 := validationService.CalculateConfidenceScore(context.Background(), result2, elements)

	// Should be lower confidence due to missing elements
	assert.Less(t, confidenceScore2, confidenceScore)
	assert.Greater(t, confidenceScore2, 0.0)
}

func TestValidationService_GetConfidenceFactors(t *testing.T) {
	llmMock := &llm_mocks.Service{}
	logger := zap.NewNop()
	validationRepoMock := &mocks.ValidationRepository{}
	auditRepoMock := &mocks.ValidationAuditRepository{}
	feedbackRepoMock := &mocks.ValidationFeedbackRepository{}

	validationService := validation.NewValidationService(llmMock, logger, validationRepoMock, auditRepoMock, feedbackRepoMock)

	result := &models.ValidationResult{
		IsValidContract: true,
		Confidence:      0.85,
		DetectedElements: []string{"parties_identification", "consideration", "legal_capacity"},
		MissingElements:  []string{"offer_and_acceptance"},
	}

	elements := &models.ContractElementsResult{
		Confidence: 0.8,
	}

	factors := validationService.GetConfidenceFactors(context.Background(), result, elements)

	// Check that all expected factors are present
	assert.Contains(t, factors, "llm_confidence")
	assert.Contains(t, factors, "element_completeness")
	assert.Contains(t, factors, "contract_structure")
	assert.Contains(t, factors, "content_quality")

	// Check factor values
	assert.Equal(t, 0.85, factors["llm_confidence"])
	assert.Equal(t, 0.5, factors["element_completeness"]) // 3 out of 6 required elements
	assert.Equal(t, 0.9, factors["contract_structure"])   // 1 missing element = 1.0 - 0.1
	assert.Equal(t, 0.8, factors["content_quality"])     // From elements confidence
}

func TestValidationService_UpdateConfidenceBasedOnFeedback(t *testing.T) {
	llmMock := &llm_mocks.Service{}
	logger := zap.NewNop()
	validationRepoMock := &mocks.ValidationRepository{}
	auditRepoMock := &mocks.ValidationAuditRepository{}
	feedbackRepoMock := &mocks.ValidationFeedbackRepository{}

	validationService := validation.NewValidationService(llmMock, logger, validationRepoMock, auditRepoMock, feedbackRepoMock)

	validationID := "test-validation-id"
	originalConfidence := 0.8

	// Mock validation record
	validationRecord := &models.ValidationRecord{
		ID: validationID,
		Result: models.ValidationResult{
			Confidence: originalConfidence,
		},
		Version: 1,
	}

	// Mock feedback with high ratings (should increase confidence)
	feedbacks := []*models.ValidationFeedback{
		{
			ID:           "feedback-1",
			ValidationID: validationID,
			FeedbackType: "accuracy",
			Rating:       5,
		},
		{
			ID:           "feedback-2",
			ValidationID: validationID,
			FeedbackType: "accuracy",
			Rating:       4,
		},
	}

	// Setup mocks
	validationRepoMock.On("GetByID", validationID).Return(validationRecord, nil)
	feedbackRepoMock.On("GetByValidationID", validationID).Return(feedbacks, nil)
	validationRepoMock.On("Update", mock.AnythingOfType("*models.ValidationRecord")).Return(nil)
	auditRepoMock.On("Create", mock.AnythingOfType("*models.ValidationAuditLog")).Return(nil)

	err := validationService.UpdateConfidenceBasedOnFeedback(context.Background(), validationID)

	assert.NoError(t, err)

	// Verify that Update was called with increased confidence
	validationRepoMock.AssertCalled(t, "Update", mock.MatchedBy(func(v *models.ValidationRecord) bool {
		// Average rating is 4.5, adjustment should be (4.5 - 3.0) * 0.1 = 0.15
		// New confidence should be 0.8 + 0.15 = 0.95
		return v.Result.Confidence > originalConfidence && v.Version == 2
	}))

	validationRepoMock.AssertExpectations(t)
	feedbackRepoMock.AssertExpectations(t)
	auditRepoMock.AssertExpectations(t)
}
