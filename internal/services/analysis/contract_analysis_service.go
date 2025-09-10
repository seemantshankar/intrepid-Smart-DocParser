package analysis

import (
	"context"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/repository"
	"contract-analysis-service/internal/services/llm"
	"go.uber.org/zap"
)

// ContractAnalysisService provides comprehensive contract analysis capabilities
type ContractAnalysisService struct {
	llmService         llm.Service
	knowledgeRepo      repository.KnowledgeRepository
	logger             *zap.Logger
	defaultLLMProvider string
}

// NewContractAnalysisService creates a new contract analysis service
func NewContractAnalysisService(
	llmService llm.Service,
	knowledgeRepo repository.KnowledgeRepository,
	logger *zap.Logger,
	defaultLLMProvider string,
) *ContractAnalysisService {
	return &ContractAnalysisService{
		llmService:         llmService,
		knowledgeRepo:      knowledgeRepo,
		logger:             logger,
		defaultLLMProvider: defaultLLMProvider,
	}
}

// AnalyzeContractResult represents the comprehensive analysis result
type AnalyzeContractResult struct {
	ContractID         string                         `json:"contract_id"`
	Summary            *models.ContractSummary        `json:"summary"`
	Analysis           *models.ContractAnalysis       `json:"analysis"`
	PaymentObligations []models.AnalysisMilestone     `json:"payment_obligations"`
	RiskAssessment     *models.AnalysisRiskAssessment `json:"risk_assessment"`
	ConfidenceScore    float64                        `json:"confidence_score"`
	ValidationIssues   []string                       `json:"validation_issues"`
	ProcessedAt        time.Time                      `json:"processed_at"`
}

// AnalyzeContract performs comprehensive contract analysis including summary extraction,
// payment obligation identification, percentage-based payment calculation, and risk assessment
func (s *ContractAnalysisService) AnalyzeContract(ctx context.Context, contractID, contractText string) (*AnalyzeContractResult, error) {
	startTime := time.Now()

	s.logger.Info("Starting comprehensive contract analysis",
		zap.String("contract_id", contractID),
		zap.Int("text_length", len(contractText)))

	result := &AnalyzeContractResult{
		ContractID:  contractID,
		ProcessedAt: startTime,
	}

	// Step 1: Extract contract summary (buyer, seller, goods, total value)
	summary, err := s.llmService.ExtractContractSummary(ctx, s.defaultLLMProvider, contractText)
	if err != nil {
		s.logger.Error("Failed to extract contract summary",
			zap.String("contract_id", contractID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to extract contract summary: %w", err)
	}
	result.Summary = summary

	s.logger.Info("Contract summary extracted successfully",
		zap.String("contract_id", contractID),
		zap.String("buyer", summary.BuyerName),
		zap.String("seller", summary.SellerName),
		zap.String("total_value", summary.TotalValue.String()))

	// Step 2: Identify payment obligations
	paymentObligations, err := s.llmService.IdentifyPaymentObligations(ctx, s.defaultLLMProvider, contractText)
	if err != nil {
		s.logger.Error("Failed to identify payment obligations",
			zap.String("contract_id", contractID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to identify payment obligations: %w", err)
	}

	s.logger.Info("Payment obligations identified",
		zap.String("contract_id", contractID),
		zap.Int("obligation_count", len(paymentObligations)))

	// Step 3: Calculate percentage-based payments
	totalValueFloat, _ := summary.TotalValue.Float64()
	if totalValueFloat > 0 {
		calculatedObligations, err := s.llmService.CalculatePercentageBasedPayments(totalValueFloat, paymentObligations)
		if err != nil {
			s.logger.Warn("Failed to calculate percentage-based payments",
				zap.String("contract_id", contractID),
				zap.Error(err))
			// Continue with original obligations if calculation fails
		} else {
			paymentObligations = calculatedObligations
		}
	}
	result.PaymentObligations = paymentObligations

	// Step 4: Perform risk assessment with industry standards
	industryStandards := s.getIndustryStandards(ctx, summary.BuyerName, summary.SellerName)

	riskAssessment, err := s.llmService.AssessContractRisks(ctx, s.defaultLLMProvider, contractText, industryStandards)
	if err != nil {
		s.logger.Error("Failed to assess contract risks",
			zap.String("contract_id", contractID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to assess contract risks: %w", err)
	}
	result.RiskAssessment = riskAssessment

	s.logger.Info("Risk assessment completed",
		zap.String("contract_id", contractID),
		zap.Float64("compliance_score", riskAssessment.ComplianceScore),
		zap.Int("risk_count", len(riskAssessment.Risks)))

	// Step 5: Perform overall analysis for backward compatibility
	analysis, err := s.llmService.AnalyzeContract(ctx, s.defaultLLMProvider, contractText)
	if err != nil {
		s.logger.Error("Failed to perform overall contract analysis",
			zap.String("contract_id", contractID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to perform contract analysis: %w", err)
	}
	result.Analysis = analysis

	// Step 6: Validate analysis confidence and calculate overall confidence score
	confidence, issues, err := s.llmService.ValidateAnalysisConfidence(analysis)
	if err != nil {
		s.logger.Error("Failed to validate analysis confidence",
			zap.String("contract_id", contractID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to validate analysis confidence: %w", err)
	}
	result.ConfidenceScore = confidence
	result.ValidationIssues = issues

	if len(issues) > 0 {
		s.logger.Warn("Analysis validation issues found",
			zap.String("contract_id", contractID),
			zap.Strings("issues", issues),
			zap.Float64("confidence_score", confidence))
	}

	duration := time.Since(startTime)
	s.logger.Info("Contract analysis completed successfully",
		zap.String("contract_id", contractID),
		zap.Duration("duration", duration),
		zap.Float64("confidence_score", confidence))

	return result, nil
}

// ExtractContractSummary extracts detailed contract summary (buyer, seller, goods, total value)
func (s *ContractAnalysisService) ExtractContractSummary(ctx context.Context, contractText string) (*models.ContractSummary, error) {
	s.logger.Info("Extracting contract summary", zap.Int("text_length", len(contractText)))

	summary, err := s.llmService.ExtractContractSummary(ctx, s.defaultLLMProvider, contractText)
	if err != nil {
		s.logger.Error("Failed to extract contract summary", zap.Error(err))
		return nil, fmt.Errorf("failed to extract contract summary: %w", err)
	}

	s.logger.Info("Contract summary extracted successfully",
		zap.String("buyer", summary.BuyerName),
		zap.String("seller", summary.SellerName))

	return summary, nil
}

// IdentifyPaymentObligations extracts all payment obligations and milestones
func (s *ContractAnalysisService) IdentifyPaymentObligations(ctx context.Context, contractText string) ([]models.AnalysisMilestone, error) {
	s.logger.Info("Identifying payment obligations", zap.Int("text_length", len(contractText)))

	obligations, err := s.llmService.IdentifyPaymentObligations(ctx, s.defaultLLMProvider, contractText)
	if err != nil {
		s.logger.Error("Failed to identify payment obligations", zap.Error(err))
		return nil, fmt.Errorf("failed to identify payment obligations: %w", err)
	}

	s.logger.Info("Payment obligations identified", zap.Int("count", len(obligations)))
	return obligations, nil
}

// CalculatePercentageBasedPayments calculates absolute amounts from percentages
func (s *ContractAnalysisService) CalculatePercentageBasedPayments(totalValue float64, milestones []models.AnalysisMilestone) ([]models.AnalysisMilestone, error) {
	s.logger.Info("Calculating percentage-based payments",
		zap.Float64("total_value", totalValue),
		zap.Int("milestone_count", len(milestones)))

	calculated, err := s.llmService.CalculatePercentageBasedPayments(totalValue, milestones)
	if err != nil {
		s.logger.Error("Failed to calculate percentage-based payments", zap.Error(err))
		return nil, fmt.Errorf("failed to calculate percentage-based payments: %w", err)
	}

	s.logger.Info("Percentage-based payments calculated successfully")
	return calculated, nil
}

// AssessContractRisks performs comprehensive risk assessment using industry standards
func (s *ContractAnalysisService) AssessContractRisks(ctx context.Context, contractText string, industries ...string) (*models.AnalysisRiskAssessment, error) {
	s.logger.Info("Assessing contract risks",
		zap.Int("text_length", len(contractText)),
		zap.Strings("industries", industries))

	// Get relevant industry standards
	industryStandards := s.getIndustryStandardsForIndustries(ctx, industries...)

	assessment, err := s.llmService.AssessContractRisks(ctx, s.defaultLLMProvider, contractText, industryStandards)
	if err != nil {
		s.logger.Error("Failed to assess contract risks", zap.Error(err))
		return nil, fmt.Errorf("failed to assess contract risks: %w", err)
	}

	s.logger.Info("Risk assessment completed",
		zap.Float64("compliance_score", assessment.ComplianceScore),
		zap.Int("risk_count", len(assessment.Risks)))

	return assessment, nil
}

// ValidateAnalysisConfidence validates analysis results and calculates confidence score
func (s *ContractAnalysisService) ValidateAnalysisConfidence(analysis *models.ContractAnalysis) (float64, []string, error) {
	s.logger.Info("Validating analysis confidence")

	confidence, issues, err := s.llmService.ValidateAnalysisConfidence(analysis)
	if err != nil {
		s.logger.Error("Failed to validate analysis confidence", zap.Error(err))
		return 0, nil, fmt.Errorf("failed to validate analysis confidence: %w", err)
	}

	s.logger.Info("Analysis confidence validated",
		zap.Float64("confidence_score", confidence),
		zap.Int("issue_count", len(issues)))

	return confidence, issues, nil
}

// getIndustryStandards retrieves relevant industry standards from the knowledge database
func (s *ContractAnalysisService) getIndustryStandards(ctx context.Context, buyerName, sellerName string) string {
	// Try to identify industries from buyer/seller names or contract content
	// This is a simplified implementation - in production, this would be more sophisticated

	// Search for general contract standards
	entries, err := s.knowledgeRepo.SearchKnowledge(ctx, "contract standards best practices", "")
	if err != nil {
		s.logger.Warn("Failed to retrieve industry standards from knowledge base", zap.Error(err))
		return ""
	}

	if len(entries) == 0 {
		return ""
	}

	// Combine relevant standards
	var standards string
	for i, entry := range entries {
		if i >= 3 { // Limit to top 3 most relevant
			break
		}
		standards += entry.Content + "\n\n"
	}

	return standards
}

// getIndustryStandardsForIndustries retrieves industry standards for specific industries
func (s *ContractAnalysisService) getIndustryStandardsForIndustries(ctx context.Context, industries ...string) string {
	if len(industries) == 0 {
		return s.getIndustryStandards(ctx, "", "")
	}

	var allStandards string
	for _, industry := range industries {
		entries, err := s.knowledgeRepo.SearchKnowledge(ctx, industry+" contract standards", industry)
		if err != nil {
			s.logger.Warn("Failed to retrieve industry standards for industry",
				zap.String("industry", industry),
				zap.Error(err))
			continue
		}

		for i, entry := range entries {
			if i >= 2 { // Limit per industry
				break
			}
			allStandards += fmt.Sprintf("Industry: %s\n%s\n\n", industry, entry.Content)
		}
	}

	if allStandards == "" {
		return s.getIndustryStandards(ctx, "", "")
	}

	return allStandards
}
