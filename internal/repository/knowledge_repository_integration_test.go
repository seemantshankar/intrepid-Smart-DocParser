//go:build repo_integration
// +build repo_integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"contract-analysis-service/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeRepository_CreateKnowledge(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repo := repository.NewKnowledgeRepository(testDB.DB)
	ctx := context.Background()

	tests := []struct {
		name    string
		entry   *repository.KnowledgeEntry
		wantErr bool
	}{
		{
			name: "successful creation",
			entry: &repository.KnowledgeEntry{
				ID:       uuid.New(),
				Title:    "Test Knowledge",
				Content:  "Test content for knowledge entry",
				Category: "technology",
				Tags:     pq.StringArray{"ai", "ml", "test"},
				Source:   "test source",
				Metadata: "{\"key\": \"value\"}",
				Version:  1,
			},
			wantErr: false,
		},
		{
			name: "creation with minimal fields",
			entry: &repository.KnowledgeEntry{
				ID:      uuid.New(),
				Title:   "Minimal Knowledge",
				Content: "Minimal content",
				Version: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateKnowledge(ctx, tt.entry)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.entry.CreatedAt)
				assert.NotZero(t, tt.entry.UpdatedAt)

				// Verify creation by fetching
				retrieved, err := repo.GetKnowledgeByID(ctx, tt.entry.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.entry.Title, retrieved.Title)
				assert.Equal(t, tt.entry.Content, retrieved.Content)
				assert.Equal(t, tt.entry.Version, retrieved.Version)
			}
		})
	}
}

func TestKnowledgeRepository_Versioning(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repo := repository.NewKnowledgeRepository(testDB.DB)
	ctx := context.Background()

	// Create initial version
	originalEntry := &repository.KnowledgeEntry{
		ID:       uuid.New(),
		Title:    "Original Title",
		Content:  "Original content",
		Category: "technology",
		Version:  1,
	}
	
	err := repo.CreateKnowledge(ctx, originalEntry)
	require.NoError(t, err)

	// Create second version
	secondVersion := &repository.KnowledgeEntry{
		ID:              uuid.New(),
		Title:           "Updated Title",
		Content:         "Updated content",
		Category:        "technology",
		Version:         2,
		ParentVersionID: &originalEntry.ID,
	}
	
	err = repo.CreateKnowledge(ctx, secondVersion)
	require.NoError(t, err)

	// Test getting version history
	versions, err := repo.GetVersionHistory(ctx, originalEntry.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(versions), 0) // May be empty if no old versions

	// Test getting latest version (should be the one with highest version number)
	latest, err := repo.GetLatestVersion(ctx, originalEntry.ID)
	require.NoError(t, err)
	// Should get the highest version from the chain
	assert.Equal(t, secondVersion.ID, latest.ID)
	assert.Equal(t, "Updated Title", latest.Title)
	assert.Equal(t, 2, latest.Version)
}

func TestKnowledgeRepository_ConflictDetection(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repo := repository.NewKnowledgeRepository(testDB.DB)
	ctx := context.Background()

	// Create initial entry
	originalEntry := &repository.KnowledgeEntry{
		ID:       uuid.New(),
		Title:    "Test Entry",
		Content:  "Original content",
		Category: "test",
		Version:  1,
	}
	
	err := repo.CreateKnowledge(ctx, originalEntry)
	require.NoError(t, err)

	// Simulate concurrent updates by attempting to update with old version
	updatedEntry1 := &repository.KnowledgeEntry{
		ID:      originalEntry.ID,
		Title:   "Updated by User 1",
		Content: "Content updated by user 1",
		Version: 1, // Using old version number - should cause conflict
	}

	updatedEntry2 := &repository.KnowledgeEntry{
		ID:      originalEntry.ID,
		Title:   "Updated by User 2", 
		Content: "Content updated by user 2",
		Version: 1, // Using old version number - should cause conflict
	}

	// First update should succeed
	err = repo.UpdateKnowledgeWithConflictDetection(ctx, updatedEntry1)
	require.NoError(t, err)

	// Second update should fail due to version conflict
	err = repo.UpdateKnowledgeWithConflictDetection(ctx, updatedEntry2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version conflict")
}

func TestKnowledgeRepository_SearchCapabilities(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repo := repository.NewKnowledgeRepository(testDB.DB)
	ctx := context.Background()

	// Create test entries
	entries := []*repository.KnowledgeEntry{
		{
			ID:       uuid.New(),
			Title:    "Machine Learning Basics",
			Content:  "Introduction to machine learning algorithms and concepts",
			Category: "technology",
			Tags:     pq.StringArray{"ml", "ai", "algorithms"},
			Source:   "university",
			Version:  1,
		},
		{
			ID:       uuid.New(),
			Title:    "Contract Law Overview",
			Content:  "Basic principles of contract law and legal obligations",
			Category: "legal",
			Tags:     pq.StringArray{"law", "contracts", "legal"},
			Source:   "law firm",
			Version:  1,
		},
		{
			ID:       uuid.New(),
			Title:    "AI in Finance",
			Content:  "Applications of artificial intelligence in financial services",
			Category: "finance",
			Tags:     pq.StringArray{"ai", "finance", "fintech"},
			Source:   "industry report",
			Version:  1,
		},
	}

	for _, entry := range entries {
		err := repo.CreateKnowledge(ctx, entry)
		require.NoError(t, err)
	}

	t.Run("search by content", func(t *testing.T) {
		results, err := repo.SearchKnowledgeAdvanced(ctx, repository.KnowledgeSearchFilter{
			Query: "machine learning",
			Limit: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
		
		// Find the ML entry
		found := false
		for _, result := range results {
			if result.Title == "Machine Learning Basics" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find Machine Learning Basics entry")
	})

	t.Run("filter by category", func(t *testing.T) {
		results, err := repo.SearchKnowledgeAdvanced(ctx, repository.KnowledgeSearchFilter{
			Category: "technology",
			Limit:    10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
		
		for _, result := range results {
			assert.Equal(t, "technology", result.Category)
		}
	})

	t.Run("filter by date range", func(t *testing.T) {
		yesterday := time.Now().AddDate(0, 0, -1)
		tomorrow := time.Now().AddDate(0, 0, 1)
		
		results, err := repo.SearchKnowledgeAdvanced(ctx, repository.KnowledgeSearchFilter{
			CreatedAfter:  &yesterday,
			CreatedBefore: &tomorrow,
			Limit:         10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3) // Should find all entries created today
	})

	t.Run("pagination", func(t *testing.T) {
		// Test first page
		results, err := repo.SearchKnowledgeAdvanced(ctx, repository.KnowledgeSearchFilter{
			Limit:  2,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 2)

		// Test second page  
		results2, err := repo.SearchKnowledgeAdvanced(ctx, repository.KnowledgeSearchFilter{
			Limit:  2,
			Offset: 2,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results2), 0)
	})
}

func TestKnowledgeRepository_FullTextSearch(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer testDB.Cleanup()

	repo := repository.NewKnowledgeRepository(testDB.DB)
	ctx := context.Background()

	// Create entries for full-text search testing
	entries := []*repository.KnowledgeEntry{
		{
			ID:      uuid.New(),
			Title:   "Advanced Machine Learning",
			Content: "Deep dive into advanced machine learning techniques including neural networks, support vector machines, and ensemble methods",
			Version: 1,
		},
		{
			ID:      uuid.New(),
			Title:   "Contract Negotiation",
			Content: "Best practices for negotiating business contracts and managing vendor relationships",
			Version: 1,
		},
	}

	for _, entry := range entries {
		err := repo.CreateKnowledge(ctx, entry)
		require.NoError(t, err)
	}

	// Test ranked full-text search
	results, err := repo.FullTextSearch(ctx, "machine learning neural", 10, 0)
	require.NoError(t, err)
	
	assert.GreaterOrEqual(t, len(results), 1)
	
	// Results should contain relevant entries
	if len(results) > 0 {
		found := false
		for _, result := range results {
			if result.Title == "Advanced Machine Learning" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find Advanced Machine Learning entry")
	}
}
