package validation_test

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"contract-analysis-service/configs"
	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/database"
	"contract-analysis-service/internal/repositories/sqlite"
	"contract-analysis-service/internal/services/llm"
	llmclient "contract-analysis-service/internal/services/llm/client"
	"contract-analysis-service/internal/services/validation"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestValidationService_Integration_ValidateContract(t *testing.T) {
	// Load environment variables from .env file
	err := godotenv.Load("../../../.env")
	if err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	// Set test environment variable if not already set
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	// Load configuration
	cfg, err := configs.LoadConfig("../../../config.yaml")
	require.NoError(t, err)

	// Setup database
	dbConfig := &configs.DatabaseConfig{
		Dialect: "sqlite3",
		Name:    ":memory:",
		LogMode: true,
	}

	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&models.ValidationResult{})
	require.NoError(t, err)

	// Create repositories
	validationRepo := sqlite.NewValidationRepository(db)
	auditRepo := sqlite.NewValidationAuditRepository(db)
	feedbackRepo := sqlite.NewValidationFeedbackRepository(db)
	logger := zap.NewNop()
	llmService := llm.NewLLMService(logger)

	// Add OpenRouter client to LLM service
	llmclient.AddOpenRouterClientToService(llmService, cfg)

	// Create validation service
	validationService := validation.NewValidationService(llmService, logger, validationRepo, auditRepo, feedbackRepo)

	tests := []struct {
		name     string
		contract string
	}{
		{
			name: "Valid Sales Contract",
			contract: `SALES CONTRACT

This Sales Agreement is entered into on January 15, 2024, between:

Seller: ABC Corporation, a Delaware corporation
Address: 123 Business Ave, New York, NY 10001

Buyer: XYZ Industries, a California corporation  
Address: 456 Commerce St, Los Angeles, CA 90001

PRODUCT: 1000 units of Model X widgets
PRICE: $50,000 USD
DELIVERY DATE: February 28, 2024
PAYMENT TERMS: Net 30 days

The Seller agrees to deliver the products by the specified delivery date.
The Buyer agrees to pay the full amount within 30 days of delivery.

This contract is governed by the laws of New York.

Seller Signature: ___________________ Date: ___________
Buyer Signature: ____________________ Date: ___________`,
		},
		{
			name: "Valid Service Agreement",
			contract: `SERVICE AGREEMENT

This Service Agreement is made on March 1, 2024, between:

Service Provider: Tech Solutions LLC
Address: 789 Tech Park, Austin, TX 78701

Client: Global Enterprises Inc.
Address: 321 Corporate Blvd, Chicago, IL 60601

SERVICES: Software development and maintenance services
TERM: 12 months starting April 1, 2024
COMPENSATION: $120,000 annually, paid monthly

The Service Provider agrees to:
- Develop custom software solutions
- Provide ongoing technical support
- Maintain system uptime of 99.5%

The Client agrees to:
- Pay monthly fees by the 15th of each month
- Provide necessary access and information
- Give 30 days notice for termination

This agreement is governed by Texas law.

Provider Signature: _________________ Date: ___________
Client Signature: ___________________ Date: ___________`,
		},
		{
			name: "Invalid Document - Just a Letter",
			contract: `Dear John,

I hope this letter finds you well. I wanted to reach out to discuss our upcoming meeting next week.

Please let me know if Tuesday at 2 PM works for you. We can meet at the usual coffee shop downtown.

Looking forward to hearing from you.

Best regards,
Sarah`,
		},
	}

	// Try to add real PDF contract test
	contractPath := "../../../uploads/Publisher agreement.pdf"
	contractData, err := ioutil.ReadFile(contractPath)
	if err == nil {
		// Check if PDF is too large for API (limit to ~100KB to avoid token limits)
		if len(contractData) < 100000 {
			tests = append(tests, struct {
				name     string
				contract string
			}{
				name:     "Real PDF - Publisher Agreement",
				contract: string(contractData),
			})
			t.Logf("Added real PDF test case: %s (size: %d bytes)", contractPath, len(contractData))
		} else {
			t.Logf("Skipping real PDF test case: %s (size: %d bytes - too large for API)", contractPath, len(contractData))
		}
	} else {
		t.Logf("Could not read real PDF file %s: %v", contractPath, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := validationService.ValidateContract(ctx, tt.contract)
			require.NoError(t, err, "Contract validation should not fail")
			require.NotNil(t, result, "Validation result should not be nil")

			// Log detailed results for debugging
			t.Logf("Contract Type: %s", result.ContractType)
			t.Logf("Is Valid: %t", result.IsValidContract)
			t.Logf("Confidence: %.2f", result.Confidence)
			t.Logf("Reason: %s", result.Reason)
			t.Logf("Detected Elements: %v", result.DetectedElements)
			t.Logf("Missing Elements: %v", result.MissingElements)

			// Basic assertions
			require.Greater(t, result.Confidence, 0.0, "Confidence should be greater than 0")
			require.LessOrEqual(t, result.Confidence, 1.0, "Confidence should be less than or equal to 1")
			// Contract type may be empty for invalid documents
			if result.IsValidContract {
				require.NotEmpty(t, result.ContractType, "Valid contract should have a contract type")
			}
			require.NotEmpty(t, result.Reason, "Reason should not be empty")

			// Test-specific assertions
			if strings.Contains(tt.name, "Valid") && !strings.Contains(tt.name, "Invalid") {
				// For valid contracts, expect reasonable confidence and some detected elements
				require.Greater(t, result.Confidence, 0.5, "Valid contracts should have reasonable confidence")
				require.NotEmpty(t, result.DetectedElements, "Valid contracts should have detected elements")
				require.True(t, result.IsValidContract, "Valid contracts should be marked as valid")
			} else if strings.Contains(tt.name, "Invalid") {
				// For invalid documents, just check that the LLM made a determination
				require.False(t, result.IsValidContract, "Invalid documents should be marked as invalid")
			} else if strings.Contains(tt.name, "Real PDF") {
				// For real PDF, expect some elements to be detected
				require.NotEmpty(t, result.DetectedElements, "Real PDF should have some detected elements")
				t.Logf("Real PDF validation completed - Type: %s, Valid: %t, Confidence: %.2f", 
					result.ContractType, result.IsValidContract, result.Confidence)
			}
		})
	}
}