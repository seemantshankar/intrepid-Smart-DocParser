package validation_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	llm_mocks "contract-analysis-service/internal/services/llm/mocks"
	"contract-analysis-service/internal/services/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestValidationService_ValidateContract(t *testing.T) {
	logger := zap.NewNop()
	llmMock := new(llm_mocks.Service)

	service := validation.NewValidationService(llmMock, logger)

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

	service := validation.NewValidationService(llmMock, logger)

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
