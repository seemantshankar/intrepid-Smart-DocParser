package models

import (
	"time"
	"github.com/shopspring/decimal"
)

type Contract struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	FilePath    string                 `json:"file_path"`
	Hash        string                 `json:"contract_hash"`
	Summary     *ContractSummary       `json:"summary" gorm:"embedded"`
	Milestones  []*Milestone           `json:"milestones" gorm:"foreignKey:ContractID"`
	Risks       []*RiskAssessment      `json:"risks" gorm:"foreignKey:ContractID"`
	Compliance  *ComplianceReport      `json:"compliance" gorm:"embedded"`
	Validation  *ValidationResult      `json:"validation" gorm:"embedded"`
	KnowledgeID string                 `json:"knowledge_id"`
	Confidence  float64                `json:"confidence_score"`
	Status        ContractStatus         `json:"status" gorm:"type:varchar(50)"`
	ContractType  string                 `json:"contract_type,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ContractSummary struct {
	BuyerName    string `json:"buyer_name"`
	SellerName   string `json:"seller_name"`
	GoodsNature  string `json:"goods_nature"`
	TotalValue   decimal.Decimal `json:"total_value" gorm:"type:decimal(20,8)"`
	Currency     string `json:"currency"`
	Jurisdiction string `json:"jurisdiction"`
}

type Milestone struct {
	ID             string            `json:"id" gorm:"primaryKey"`
	ContractID     string            `json:"contract_id" gorm:"index"`
	Description    string            `json:"description"`
	Amount         decimal.Decimal   `json:"amount" gorm:"type:decimal(20,8)"`
	Percentage     float64           `json:"percentage"`
	Trigger        string            `json:"trigger_condition"`
	SequenceOrder  int               `json:"sequence_order"`
	Dependencies   []string          `json:"dependencies" gorm:"type:text;serializer:json"`
	Category       string            `json:"category"`
	Verification   VerificationMethod `json:"verification_method" gorm:"type:varchar(50)"`
	OracleConfig   *OracleConfig     `json:"oracle_config,omitempty" gorm:"embedded"`
}

type VerificationMethod string

const (
	Manual   VerificationMethod = "manual"
	Oracle   VerificationMethod = "oracle"
	Hybrid   VerificationMethod = "hybrid"
	API      VerificationMethod = "api"
)

type OracleConfig struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key" gorm:"type:text"`
}

type RiskAssessment struct {
	ID        string `json:"id" gorm:"primaryKey"`
	ContractID string `json:"contract_id" gorm:"index"`
	Type      string `json:"type"`
	Severity  Severity `json:"severity" gorm:"type:varchar(20)"`
	Description string `json:"description"`
	Recommendation string `json:"recommendation"`
	IndustryRef string `json:"industry_reference"`
}

type Severity string

const (
	Low      Severity = "low"
	Medium   Severity = "medium"
	High     Severity = "high"
	Critical Severity = "critical"
)

type ComplianceReport struct {
	MissingClauses []string `json:"missing_clauses" gorm:"type:text;serializer:json"`
	Suggestions    []string `json:"suggestions" gorm:"type:text;serializer:json"`
	IsCompliant    bool     `json:"is_compliant"`
	Report         string   `json:"report" gorm:"type:text"`
}

type SmartChequeConfig struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	ContractID  string         `json:"contract_id" gorm:"index"`
	PayerID     string         `json:"payer_id"`
	PayeeID     string         `json:"payee_id"`
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(20,8)"`
	Currency    string         `json:"currency"`
	Milestones  []*Milestone   `json:"milestones" gorm:"foreignKey:SmartChequeID"`
	ContractHash string        `json:"contract_hash"`
	Status      ChequeStatus   `json:"status" gorm:"type:varchar(50)"`
	DisputePath DisputePath    `json:"dispute_path" gorm:"embedded"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ChequeStatus string

const (
	Created     ChequeStatus = "created"
	Locked      ChequeStatus = "locked"
	InProgress  ChequeStatus = "in_progress"
	Completed   ChequeStatus = "completed"
	Disputed    ChequeStatus = "disputed"
)

type DisputePath struct {
	Method     string   `json:"method"`
	Priority   string   `json:"priority"`
	Category   string   `json:"category"`
	Transitions []string `json:"state_transitions" gorm:"type:text;serializer:json"`
}

type KnowledgeEntry struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Industry     string    `json:"industry" gorm:"index"`
	Jurisdiction string    `json:"jurisdiction" gorm:"index"`
	Type         string    `json:"type"`
	Content      string    `json:"content" gorm:"type:text"`
	Version      int       `json:"version" gorm:"version"`
	UpdatedAt    time.Time `json:"updated_at"`
	Source       string    `json:"source"`
}

type ContractStatus string

const (
	Validated ContractStatus = "validated"
	Analyzed  ContractStatus = "analyzed"
)

// ContractAnalysis represents the structured analysis result from an LLM.
// This is a DTO used for parsing the raw LLM output.
type ContractAnalysis struct {
	Buyer         string              `json:"buyer"`
	Seller        string              `json:"seller"`
	TotalValue    decimal.Decimal     `json:"total_value"`
	Currency      string              `json:"currency"`
	Milestones    []AnalysisMilestone `json:"milestones"`
	RiskFactors   []AnalysisRisk      `json:"risk_factors"`
}

// AnalysisMilestone is a simplified milestone structure for LLM parsing.
type AnalysisMilestone struct {
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Percentage  float64         `json:"percentage"`
}

// AnalysisRisk is a simplified risk structure for LLM parsing.
type AnalysisRisk struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// SequencedMilestone represents a milestone with sequencing information from the LLM.
// This is a DTO used for parsing the raw LLM output.
type SequencedMilestone struct {
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	SequenceOrder  int      `json:"sequence_order"`
	Category       string   `json:"category"`
	Dependencies   []string `json:"dependencies"`
	Percentage     float64  `json:"percentage"`
}

// AnalysisComplianceReport represents a compliance analysis report from the LLM.
// This is a DTO used for parsing the raw LLM output.
type AnalysisComplianceReport struct {
	Jurisdiction     string   `json:"jurisdiction"`
	RequiredClauses  []string `json:"required_clauses"`
	MissingClauses   []string `json:"missing_clauses"`
	ComplianceLevel  string   `json:"compliance_level"`
	Recommendations  []string `json:"recommendations"`
	RiskLevel        string   `json:"risk_level"`
}

// AnalysisRiskAssessment represents a comprehensive risk assessment from the LLM.
// This is a DTO used for parsing the raw LLM output.
type AnalysisRiskAssessment struct {
	MissingClauses []string               `json:"missing_clauses"`
	Risks          []AnalysisIndividualRisk `json:"risks"`
	ComplianceScore float64              `json:"compliance_score"`
	Suggestions     []string               `json:"suggestions"`
}

// AnalysisIndividualRisk represents an individual risk identified by the LLM.
type AnalysisIndividualRisk struct {
	Party        string `json:"party"`
	Type         string `json:"type"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// ValidationResult represents the outcome of a contract validation check.
type ValidationResult struct {
	IsValidContract  bool     `json:"is_valid_contract"`
	Reason           string   `json:"reason,omitempty"`
	Confidence       float64  `json:"confidence"`
	ContractType     string   `json:"contract_type,omitempty"`
	MissingElements  []string `json:"missing_elements,omitempty" gorm:"type:text;serializer:json"`
	DetectedElements []string `json:"detected_elements,omitempty" gorm:"type:text;serializer:json"`
}
