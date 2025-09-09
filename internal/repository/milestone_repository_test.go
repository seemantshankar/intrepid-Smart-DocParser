//go:build repo_integration
// +build repo_integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"contract-analysis-service/internal/repository"
)

func TestMilestoneRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repos := testDB.CreateRepositories()
	ctx := context.Background()

	// Create a test contract for milestones
	testContract := &repository.Contract{
		Title:  "Test Contract for Milestones",
		Status: "active",
	}
	err := repos.Contract.Create(ctx, testContract)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		dueDate := time.Now().AddDate(0, 0, 7)
		milestone := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Test Milestone",
			DueDate:    &dueDate,
			Status:     "pending",
		}

		err := repos.Milestone.Create(ctx, milestone)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, milestone.ID)
		assert.NotZero(t, milestone.CreatedAt)
		assert.NotZero(t, milestone.UpdatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		// Create milestone
		milestone := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "GetByID Test",
			Status:     "pending",
		}
		err := repos.Milestone.Create(ctx, milestone)
		require.NoError(t, err)

		// Retrieve milestone
		retrieved, err := repos.Milestone.GetByID(ctx, milestone.ID)
		require.NoError(t, err)
		assert.Equal(t, milestone.ID, retrieved.ID)
		assert.Equal(t, milestone.ContractID, retrieved.ContractID)
		assert.Equal(t, milestone.Name, retrieved.Name)
		assert.Equal(t, milestone.Status, retrieved.Status)
	})

	t.Run("GetByContractID", func(t *testing.T) {
		testDB.Cleanup() // Clean up any existing data

		// Create test milestones for the contract
		now := time.Now()
		milestone1 := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Milestone 1",
			DueDate:    &now,
			Status:     "pending",
		}
		milestone2 := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Milestone 2",
			DueDate:    &now,
			Status:     "pending",
		}

		err = repos.Milestone.Create(ctx, milestone1)
		require.NoError(t, err)
		err = repos.Milestone.Create(ctx, milestone2)
		require.NoError(t, err)

		// Create another contract and milestone to test isolation
		otherContract := &repository.Contract{
			Title:  "Other Contract",
			Status: "active",
		}
		err = repos.Contract.Create(ctx, otherContract)
		require.NoError(t, err)

		otherMilestone := &repository.Milestone{
			ContractID: otherContract.ID,
			Name:       "Other milestone",
			DueDate:    &now,
			Status:     "pending",
		}
		err = repos.Milestone.Create(ctx, otherMilestone)
		require.NoError(t, err)

		// Test getting milestones by contract ID
		milestones, err := repos.Milestone.GetByContractID(ctx, testContract.ID, 0, 0)
		require.NoError(t, err)
		assert.Len(t, milestones, 2)

		for _, milestone := range milestones {
			assert.Equal(t, testContract.ID, milestone.ContractID)
		}

		// Verify other contract's milestones
		otherMilestones, err := repos.Milestone.GetByContractID(ctx, otherContract.ID, 0, 0)
		require.NoError(t, err)
		assert.Len(t, otherMilestones, 1)
		assert.Equal(t, otherContract.ID, otherMilestones[0].ContractID)
	})

	t.Run("GetDueSoon", func(t *testing.T) {
		// Create milestones with different due dates
		pastDue := time.Now().AddDate(0, 0, -1)
		dueSoon := time.Now().AddDate(0, 0, 2)
		future := time.Now().AddDate(0, 0, 10)

		milestones := []*repository.Milestone{
			{
				ContractID: testContract.ID,
				Name:       "Past Due Milestone",
				DueDate:    &pastDue,
				Status:     "pending",
			},
			{
				ContractID: testContract.ID,
				Name:       "Due Soon Milestone",
				DueDate:    &dueSoon,
				Status:     "pending",
			},
			{
				ContractID: testContract.ID,
				Name:       "Future Milestone",
				DueDate:    &future,
				Status:     "pending",
			},
		}

		for _, milestone := range milestones {
			err := repos.Milestone.Create(ctx, milestone)
			require.NoError(t, err)
		}

		// Get milestones due within 7 days
		dueSoonMilestones, err := repos.Milestone.GetDueSoon(ctx, 7)
		require.NoError(t, err)

		// Should include past due and due soon, but not future
		foundDueSoon := false
		foundPastDue := false
		for _, milestone := range dueSoonMilestones {
			if milestone.Name == "Due Soon Milestone" {
				foundDueSoon = true
			}
			if milestone.Name == "Past Due Milestone" {
				foundPastDue = true
			}
			if milestone.Name == "Future Milestone" {
				t.Error("Future milestone should not be included")
			}
		}

		assert.True(t, foundDueSoon || foundPastDue, "Should find at least one due soon or past due milestone")
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		// Create milestone
		milestone := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Status Update Test",
			Status:     "pending",
		}
		err := repos.Milestone.Create(ctx, milestone)
		require.NoError(t, err)

		// Update status
		err = repos.Milestone.UpdateStatus(ctx, milestone.ID, "completed")
		require.NoError(t, err)

		// Verify status update
		retrieved, err := repos.Milestone.GetByID(ctx, milestone.ID)
		require.NoError(t, err)
		assert.Equal(t, "completed", retrieved.Status)
	})

	t.Run("Update", func(t *testing.T) {
		// Create milestone
		milestone := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Update Test",
			Status:     "pending",
		}
		err := repos.Milestone.Create(ctx, milestone)
		require.NoError(t, err)

		// Update milestone
		newDueDate := time.Now().AddDate(0, 0, 14)
		milestone.Name = "Updated Milestone"
		milestone.DueDate = &newDueDate
		milestone.Status = "in_progress"

		err = repos.Milestone.Update(ctx, milestone)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repos.Milestone.GetByID(ctx, milestone.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Milestone", retrieved.Name)
		assert.Equal(t, "in_progress", retrieved.Status)
		assert.True(t, retrieved.UpdatedAt.After(milestone.CreatedAt))
	})

	t.Run("Delete", func(t *testing.T) {
		// Create milestone
		milestone := &repository.Milestone{
			ContractID: testContract.ID,
			Name:       "Delete Test",
			Status:     "pending",
		}
		err := repos.Milestone.Create(ctx, milestone)
		require.NoError(t, err)

		// Delete milestone
		err = repos.Milestone.Delete(ctx, milestone.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repos.Milestone.GetByID(ctx, milestone.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("List", func(t *testing.T) {
		// Get current count
		allMilestones, err := repos.Milestone.List(ctx, 0, 0)
		require.NoError(t, err)
		initialCount := len(allMilestones)

		// Create additional milestones
		for i := 0; i < 3; i++ {
			milestone := &repository.Milestone{
				ContractID: testContract.ID,
				Name:       "List Test " + string(rune(i+'1')),
				Status:     "pending",
			}
			err := repos.Milestone.Create(ctx, milestone)
			require.NoError(t, err)
		}

		// List all milestones
		allMilestones, err = repos.Milestone.List(ctx, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, initialCount+3, len(allMilestones))

		// List with limit
		limitedMilestones, err := repos.Milestone.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(limitedMilestones))
	})
}
