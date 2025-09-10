package llm

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestLLMService_AnalyzeContract(t *testing.T) {
	// Create a mock external client
	mockClient := new(external.MockClient)

	// Create a logger
	logger := zap.NewNop()

	// Create the LLM service with the mock client
	service := NewLLMService(logger)
	service.AddClient("test-provider", mockClient)

	// Define the expected response from the LLM
	expectedAnalysis := &models.ContractAnalysis{
		Buyer:      "Test Buyer",
		Seller:     "Test Seller",
		TotalValue: models.FlexibleString("1000.0"),
	}
	responseBody, _ := json.Marshal(expectedAnalysis)

	// Set up the mock client's expected behavior
	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(&external.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)

	// Call the method to be tested
	analysis, err := service.AnalyzeContract(context.Background(), "test-provider", "some contract text")

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, expectedAnalysis.Buyer, analysis.Buyer)
	assert.Equal(t, expectedAnalysis.Seller, analysis.Seller)
	assert.Equal(t, expectedAnalysis.TotalValue, analysis.TotalValue)

	// Assert that the mock was called as expected
	mockClient.AssertExpectations(t)
}

func TestLLMService_AnalyzeContract_Error(t *testing.T) {
	// Create a mock external client
	mockClient := new(external.MockClient)

	// Create a logger
	logger := zap.NewNop()

	// Create the LLM service with the mock client
	service := NewLLMService(logger)
	service.AddClient("test-provider", mockClient)

	// Define the expected error
	expectedError := errors.New("API error")

	// Set up the mock client's expected behavior
	mockClient.On("ExecuteRequest", mock.Anything, mock.Anything).Return(nil, expectedError)

	// Call the method to be tested
	analysis, err := service.AnalyzeContract(context.Background(), "test-provider", "some contract text")

	// Assert the results
	assert.Error(t, err)
	assert.Nil(t, analysis)
	assert.ErrorContains(t, err, expectedError.Error())

	// Assert that the mock was called as expected
	mockClient.AssertExpectations(t)
}
