package classification

import (
	"context"
	"testing"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/repositories/mocks"
	llm_mocks "contract-analysis-service/internal/services/llm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestClassificationService_ClassifyContract(t *testing.T) {
	mockRepo := &mocks.ClassificationRepositoryMock{}
	mockLLM := &llm_mocks.Service{}

	logger := zap.NewNop()
	service := NewClassificationService(mockLLM, logger, mockRepo)

	// Mock LLM response
	responseBody := []byte(`{
		"choices": [{
			"message": {
				"content": "{\"primary_type\": \"Service Agreement\", \"sub_type\": \"Software License\", \"industry\": \"Technology\", \"complexity\": \"moderate\", \"risk_level\": \"medium\", \"jurisdiction\": \"California\", \"contract_value\": \"medium\", \"duration\": \"long_term\", \"party_types\": [\"B2B\"], \"special_clauses\": [\"intellectual_property\", \"confidentiality\"], \"confidence\": 0.85, \"metadata\": {}}"
			}
		}]
	}`)

	mockLLM.On("ExecuteRequest", mock.Anything, "openrouter", mock.AnythingOfType("*external.Request")).Return(&external.Response{
		Body: responseBody,
	}, nil)

	result, err := service.ClassifyContract(context.Background(), "Sample contract text")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Service Agreement", result.PrimaryType)
	assert.Equal(t, "Software License", result.SubType)
	assert.Equal(t, "Technology", result.Industry)
	assert.Equal(t, "moderate", result.Complexity)
	assert.Equal(t, "medium", result.RiskLevel)
	assert.Equal(t, 0.85, result.Confidence)
	assert.Contains(t, result.PartyTypes, "B2B")
	assert.Contains(t, result.SpecialClauses, "intellectual_property")

	mockLLM.AssertExpectations(t)
}

func TestClassificationService_GetContractComplexity(t *testing.T) {
	mockRepo := &mocks.ClassificationRepositoryMock{}
	mockLLM := &llm_mocks.Service{}

	logger := zap.NewNop()
	service := NewClassificationService(mockLLM, logger, mockRepo)

	// Mock LLM response
	responseBody := []byte(`{
		"choices": [{
			"message": {
				"content": "{\"level\": \"complex\", \"score\": 0.75, \"factors\": [\"multiple_parties\", \"complex_terms\"], \"clause_count\": 25, \"page_count\": 15, \"legal_term_count\": 45, \"cross_references\": 8, \"external_references\": 3}"
			}
		}]
	}`)

	mockLLM.On("ExecuteRequest", mock.Anything, "openrouter", mock.AnythingOfType("*external.Request")).Return(&external.Response{
		Body: responseBody,
	}, nil)

	result, err := service.GetContractComplexity(context.Background(), "Complex contract text")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "complex", result.Level)
	assert.Equal(t, 0.75, result.Score)
	assert.Equal(t, 25, result.ClauseCount)
	assert.Equal(t, 15, result.PageCount)
	assert.Contains(t, result.Factors, "multiple_parties")

	mockLLM.AssertExpectations(t)
}

func TestClassificationService_ClassifyByIndustry(t *testing.T) {
	mockRepo := &mocks.ClassificationRepositoryMock{}
	mockLLM := &llm_mocks.Service{}

	logger := zap.NewNop()
	service := NewClassificationService(mockLLM, logger, mockRepo)

	// Mock LLM response
	responseBody := []byte(`{
		"choices": [{
			"message": {
				"content": "{\"primary_industry\": \"Healthcare\", \"secondary_industry\": \"Technology\", \"industry_code\": \"621111\", \"regulations\": [\"HIPAA\", \"FDA\"], \"standards\": [\"HL7\", \"FHIR\"], \"compliance\": {\"privacy\": \"required\", \"security\": \"high\"}, \"confidence\": 0.92}"
			}
		}]
	}`)

	mockLLM.On("ExecuteRequest", mock.Anything, "openrouter", mock.AnythingOfType("*external.Request")).Return(&external.Response{
		Body: responseBody,
	}, nil)

	result, err := service.ClassifyByIndustry(context.Background(), "Healthcare contract text")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Healthcare", result.PrimaryIndustry)
	assert.Equal(t, "Technology", result.SecondaryIndustry)
	assert.Equal(t, "621111", result.IndustryCode)
	assert.Contains(t, result.Regulations, "HIPAA")
	assert.Contains(t, result.Standards, "HL7")
	assert.Equal(t, 0.92, result.Confidence)

	mockLLM.AssertExpectations(t)
}

func TestClassificationService_StoreClassification(t *testing.T) {
	mockRepo := &mocks.ClassificationRepositoryMock{}
	mockLLM := &llm_mocks.Service{}

	logger := zap.NewNop()
	service := NewClassificationService(mockLLM, logger, mockRepo)

	classification := &models.ContractClassification{
		PrimaryType: "Service Agreement",
		SubType:     "Consulting",
		Industry:    "Technology",
		Complexity:  "moderate",
		RiskLevel:   "medium",
		Confidence:  0.88,
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.ClassificationRecord")).Return(nil)

	err := service.StoreClassification(context.Background(), "contract-123", classification)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestClassificationService_GetClassificationHistory(t *testing.T) {
	mockRepo := &mocks.ClassificationRepositoryMock{}
	mockLLM := &llm_mocks.Service{}

	logger := zap.NewNop()
	service := NewClassificationService(mockLLM, logger, mockRepo)

	expectedRecords := []*models.ClassificationRecord{
		{
			ID:         "record-1",
			ContractID: "contract-123",
			Classification: &models.ContractClassification{
				PrimaryType: "Service Agreement",
				Industry:    "Technology",
			},
			CreatedAt: time.Now(),
			Version:   1,
		},
	}

	mockRepo.On("GetByContractID", "contract-123").Return(expectedRecords, nil)

	result, err := service.GetClassificationHistory(context.Background(), "contract-123")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "record-1", result[0].ID)
	assert.Equal(t, "contract-123", result[0].ContractID)
	assert.Equal(t, "Service Agreement", result[0].Classification.PrimaryType)

	mockRepo.AssertExpectations(t)
}