//go:build repo_integration
// +build repo_integration

package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seemantshankar/intrepid-smart-docparser/internal/repository"
)

func TestRiskAssessmentRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repos := testDB.CreateRepositories()
	ctx := context.Background()

	// Create test contracts
	contract1 := &repository.Contract{
		Title:  "Contract 1",
		Status: "active",
	}
	err := repos.Contract.Create(ctx, contract1)
	require.NoError(t, err)

	contract2 := &repository.Contract{
		Title:  "Contract 2",
		Status: "draft",
	}
	err = repos.Contract.Create(ctx, contract2)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		assessment := &repository.RiskAssessment{
			ContractID:  contract1.ID,
			RiskLevel:   "high",
			Description: "High risk due to contract complexity",
		}

		err := repos.RiskAssessment.Create(ctx, assessment)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, assessment.ID)
		assert.NotZero(t, assessment.CreatedAt)
		assert.NotZero(t, assessment.UpdatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		assessment := &repository.RiskAssessment{
			ContractID:  contract1.ID,
			RiskLevel:   "medium",
			Description: "Medium risk assessment",
		}
		err := repos.RiskAssessment.Create(ctx, assessment)
		require.NoError(t, err)

		retrieved, err := repos.RiskAssessment.GetByID(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, assessment.ID, retrieved.ID)
		assert.Equal(t, assessment.ContractID, retrieved.ContractID)
		assert.Equal(t, assessment.RiskLevel, retrieved.RiskLevel)
		assert.Equal(t, assessment.Description, retrieved.Description)
	})

	t.Run("GetByContractID", func(t *testing.T) {
		// Create assessments for both contracts
		assessments := []*repository.RiskAssessment{
			{
				ContractID:  contract1.ID,
				RiskLevel:   "low",
				Description: "Low risk for contract 1",
			},
			{
				ContractID:  contract1.ID,
				RiskLevel:   "high",
				Description: "High risk for contract 1",
			},
			{
				ContractID:  contract2.ID,
				RiskLevel:   "medium",
				Description: "Medium risk for contract 2",
			},
		}

		for _, assessment := range assessments {
			err := repos.RiskAssessment.Create(ctx, assessment)
			require.NoError(t, err)
		}

		// Get assessments for contract 1
		contract1Assessments, err := repos.RiskAssessment.GetByContractID(ctx, contract1.ID, 0, 0)
		require.NoError(t, err)
		// Verify we have the expected number of assessments for contract1
		assert.GreaterOrEqual(t, len(contract1Assessments), 2)

		// Count actual assessments for contract1
		contract1Count := 0
		for _, assessment := range contract1Assessments {
			if assessment.ContractID == contract1.ID {
				contract1Count++
			}
		}
		assert.GreaterOrEqual(t, contract1Count, 2)

		// Get assessments for contract 2
		contract2Assessments, err := repos.RiskAssessment.GetByContractID(ctx, contract2.ID, 0, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contract2Assessments), 1)
		
		// Verify at least one assessment belongs to contract2
		var hasContract2Assessment bool
		for _, assessment := range contract2Assessments {
			if assessment.ContractID == contract2.ID {
				hasContract2Assessment = true
				break
			}
		}
		assert.True(t, hasContract2Assessment, "Expected to find at least one assessment for contract2")
	})

	t.Run("GetByRiskLevel", func(t *testing.T) {
		// Create assessments with different risk levels
		riskLevels := []string{"low", "medium", "high", "critical"}
		for _, level := range riskLevels {
			assessment := &repository.RiskAssessment{
				ContractID:  contract1.ID,
				RiskLevel:   level,
				Description: level + " risk assessment",
			}
			err := repos.RiskAssessment.Create(ctx, assessment)
			require.NoError(t, err)
		}

		// Get assessments by risk level
		highRiskAssessments, err := repos.RiskAssessment.GetByRiskLevel(ctx, "high", 0, 0)
		require.NoError(t, err)
		assert.True(t, len(highRiskAssessments) >= 1)

		for _, assessment := range highRiskAssessments {
			assert.Equal(t, "high", assessment.RiskLevel)
		}
	})

	t.Run("Update", func(t *testing.T) {
		assessment := &repository.RiskAssessment{
			ContractID:  contract1.ID,
			RiskLevel:   "low",
			Description: "Initial assessment",
		}
		err := repos.RiskAssessment.Create(ctx, assessment)
		require.NoError(t, err)

		// Update assessment
		assessment.RiskLevel = "critical"
		assessment.Description = "Updated critical assessment"

		err = repos.RiskAssessment.Update(ctx, assessment)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repos.RiskAssessment.GetByID(ctx, assessment.ID)
		require.NoError(t, err)
		assert.Equal(t, "critical", retrieved.RiskLevel)
		assert.Equal(t, "Updated critical assessment", retrieved.Description)
		assert.True(t, retrieved.UpdatedAt.After(assessment.CreatedAt))
	})

	t.Run("Delete", func(t *testing.T) {
		assessment := &repository.RiskAssessment{
			ContractID:  contract1.ID,
			RiskLevel:   "medium",
			Description: "Assessment to delete",
		}
		err := repos.RiskAssessment.Create(ctx, assessment)
		require.NoError(t, err)

		// Delete assessment
		err = repos.RiskAssessment.Delete(ctx, assessment.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repos.RiskAssessment.GetByID(ctx, assessment.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("List", func(t *testing.T) {
		// Get current count
		allAssessments, err := repos.RiskAssessment.List(ctx, 0, 0)
		require.NoError(t, err)
		initialCount := len(allAssessments)

		// Create additional assessments
		for i := 0; i < 3; i++ {
			assessment := &repository.RiskAssessment{
				ContractID:  contract1.ID,
				RiskLevel:   "low",
				Description: "List test assessment " + string(rune(i+'1')),
			}
			err := repos.RiskAssessment.Create(ctx, assessment)
			require.NoError(t, err)
		}

		// List all assessments
		allAssessments, err = repos.RiskAssessment.List(ctx, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, initialCount+3, len(allAssessments))

		// List with limit
		limitedAssessments, err := repos.RiskAssessment.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(limitedAssessments))
	})
}
