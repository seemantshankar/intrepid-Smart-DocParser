package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seemantshankar/intrepid-smart-docparser/internal/repository"
)

func TestContractRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repos := testDB.CreateRepositories()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		contract := &repository.Contract{
			Title:       "Test Contract",
			Description: "This is a test contract",
			Status:      "active",
		}

		err := repos.Contract.Create(ctx, contract)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, contract.ID)
		assert.NotZero(t, contract.CreatedAt)
		assert.NotZero(t, contract.UpdatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		// First create a contract
		contract := &repository.Contract{
			Title:       "GetByID Test",
			Description: "Contract for GetByID test",
			Status:      "draft",
		}
		err := repos.Contract.Create(ctx, contract)
		require.NoError(t, err)

		// Retrieve the contract
		retrieved, err := repos.Contract.GetByID(ctx, contract.ID)
		require.NoError(t, err)
		assert.Equal(t, contract.ID, retrieved.ID)
		assert.Equal(t, contract.Title, retrieved.Title)
		assert.Equal(t, contract.Description, retrieved.Description)
		assert.Equal(t, contract.Status, retrieved.Status)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repos.Contract.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Update", func(t *testing.T) {
		// Create initial contract
		contract := &repository.Contract{
			Title:       "Update Test",
			Description: "Original description",
			Status:      "draft",
		}
		err := repos.Contract.Create(ctx, contract)
		require.NoError(t, err)

		originalUpdatedAt := contract.UpdatedAt

		// Update contract
		time.Sleep(time.Millisecond) // Ensure time difference
		contract.Title = "Updated Title"
		contract.Description = "Updated description"
		contract.Status = "active"

		err = repos.Contract.Update(ctx, contract)
		require.NoError(t, err)
		assert.True(t, contract.UpdatedAt.After(originalUpdatedAt))

		// Verify update
		retrieved, err := repos.Contract.GetByID(ctx, contract.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, "Updated description", retrieved.Description)
		assert.Equal(t, "active", retrieved.Status)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create contract for deletion
		contract := &repository.Contract{
			Title:  "Delete Test",
			Status: "draft",
		}
		err := repos.Contract.Create(ctx, contract)
		require.NoError(t, err)

		// Delete contract
		err = repos.Contract.Delete(ctx, contract.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repos.Contract.GetByID(ctx, contract.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("List", func(t *testing.T) {
		// Clear existing contracts
		contracts, err := repos.Contract.List(ctx, 0, 0)
		require.NoError(t, err)

		initialCount := len(contracts)

		// Create multiple contracts
		for i := 0; i < 3; i++ {
			contract := &repository.Contract{
				Title:  "List Test " + string(rune(i+'1')),
				Status: "active",
			}
			err := repos.Contract.Create(ctx, contract)
			require.NoError(t, err)
		}

		// List all contracts
		allContracts, err := repos.Contract.List(ctx, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, initialCount+3, len(allContracts))

		// List with limit
		limitedContracts, err := repos.Contract.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, len(limitedContracts))

		// List with offset
		offsetContracts, err := repos.Contract.List(ctx, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, 2, len(offsetContracts))
	})

	t.Run("GetByStatus", func(t *testing.T) {
		// Create contracts with different statuses
		statuses := []string{"active", "draft", "completed"}
		for _, status := range statuses {
			contract := &repository.Contract{
				Title:  "Status Test " + status,
				Status: status,
			}
			err := repos.Contract.Create(ctx, contract)
			require.NoError(t, err)
		}

		// Get contracts by status
		activeContracts, err := repos.Contract.GetByStatus(ctx, "active", 0, 0)
		require.NoError(t, err)
		assert.True(t, len(activeContracts) >= 1)

		for _, contract := range activeContracts {
			assert.Equal(t, "active", contract.Status)
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		// Create contract
		contract := &repository.Contract{
			Title:  "Status Update Test",
			Status: "draft",
		}
		err := repos.Contract.Create(ctx, contract)
		require.NoError(t, err)

		// Update status
		err = repos.Contract.UpdateStatus(ctx, contract.ID, "active")
		require.NoError(t, err)

		// Verify status update
		retrieved, err := repos.Contract.GetByID(ctx, contract.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", retrieved.Status)
	})

	t.Run("UpdateStatus_NotFound", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := repos.Contract.UpdateStatus(ctx, nonExistentID, "active")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
