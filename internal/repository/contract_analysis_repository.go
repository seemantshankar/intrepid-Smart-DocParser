package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"contract-analysis-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContractAnalysisRecord represents a complete contract analysis result stored in the database
type ContractAnalysisRecord struct {
	ID                     string                         `json:"id" gorm:"type:varchar(36);primaryKey"`
	ContractID             string                         `json:"contract_id" gorm:"index;not null"`
	UserID                 string                         `json:"user_id" gorm:"index"`
	Summary                *models.ContractSummary        `json:"summary" gorm:"embedded;embeddedPrefix:summary_"`
	Analysis               *models.ContractAnalysis       `json:"analysis" gorm:"embedded;embeddedPrefix:analysis_"`
	PaymentObligationsJSON string                         `json:"payment_obligations_json" gorm:"type:jsonb"`
	PaymentObligations     []models.AnalysisMilestone     `json:"payment_obligations" gorm:"-"`
	RiskAssessmentJSON     string                         `json:"risk_assessment_json" gorm:"type:jsonb"`
	RiskAssessment         *models.AnalysisRiskAssessment `json:"risk_assessment" gorm:"-"`
	ConfidenceScore        float64                        `json:"confidence_score"`
	ValidationIssuesJSON   string                         `json:"validation_issues_json" gorm:"type:jsonb"`
	ValidationIssues       []string                       `json:"validation_issues" gorm:"-"`
	ProcessedAt            time.Time                      `json:"processed_at"`
	CreatedAt              time.Time                      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time                      `json:"updated_at" gorm:"autoUpdateTime"`
	Version                int                            `json:"version" gorm:"default:1"`
}

// TableName returns the table name for GORM
func (ContractAnalysisRecord) TableName() string {
	return "contract_analysis_records"
}

// BeforeCreate generates a UUID for the ID field if it's empty
func (c *ContractAnalysisRecord) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// BeforeSave handles JSON serialization before saving to database
func (c *ContractAnalysisRecord) BeforeSave(tx *gorm.DB) error {
	if c.PaymentObligations != nil {
		data, err := json.Marshal(c.PaymentObligations)
		if err != nil {
			return fmt.Errorf("failed to marshal payment obligations: %w", err)
		}
		c.PaymentObligationsJSON = string(data)
	}

	if c.RiskAssessment != nil {
		data, err := json.Marshal(c.RiskAssessment)
		if err != nil {
			return fmt.Errorf("failed to marshal risk assessment: %w", err)
		}
		c.RiskAssessmentJSON = string(data)
	}

	if c.ValidationIssues != nil {
		data, err := json.Marshal(c.ValidationIssues)
		if err != nil {
			return fmt.Errorf("failed to marshal validation issues: %w", err)
		}
		c.ValidationIssuesJSON = string(data)
	}

	return nil
}

// AfterFind handles JSON deserialization after loading from database
func (c *ContractAnalysisRecord) AfterFind(tx *gorm.DB) error {
	if c.PaymentObligationsJSON != "" {
		if err := json.Unmarshal([]byte(c.PaymentObligationsJSON), &c.PaymentObligations); err != nil {
			return fmt.Errorf("failed to unmarshal payment obligations: %w", err)
		}
	}

	if c.RiskAssessmentJSON != "" {
		if err := json.Unmarshal([]byte(c.RiskAssessmentJSON), &c.RiskAssessment); err != nil {
			return fmt.Errorf("failed to unmarshal risk assessment: %w", err)
		}
	}

	if c.ValidationIssuesJSON != "" {
		if err := json.Unmarshal([]byte(c.ValidationIssuesJSON), &c.ValidationIssues); err != nil {
			return fmt.Errorf("failed to unmarshal validation issues: %w", err)
		}
	}

	return nil
}

// ContractAnalysisRepository defines the interface for contract analysis operations
type ContractAnalysisRepository interface {
	CreateAnalysis(ctx context.Context, record *ContractAnalysisRecord) error
	GetAnalysisByID(ctx context.Context, id string) (*ContractAnalysisRecord, error)
	GetAnalysisByContractID(ctx context.Context, contractID string) (*ContractAnalysisRecord, error)
	GetAnalysesByUserID(ctx context.Context, userID string, limit, offset int) ([]*ContractAnalysisRecord, error)
	UpdateAnalysis(ctx context.Context, record *ContractAnalysisRecord) error
	DeleteAnalysis(ctx context.Context, id string) error
	ListAnalyses(ctx context.Context, limit, offset int) ([]*ContractAnalysisRecord, error)
	GetAnalysesByConfidenceRange(ctx context.Context, minConfidence, maxConfidence float64, limit, offset int) ([]*ContractAnalysisRecord, error)
	GetAnalysesCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*ContractAnalysisRecord, error)
}
