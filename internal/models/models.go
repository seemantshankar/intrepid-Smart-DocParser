package models

import (
	"time"
	"github.com/shopspring/decimal"
)

type Contract struct {
	ID           string             `json:"id" gorm:"primaryKey"`
	UserID       string             `json:"user_id" gorm:"index"`
	Filename     string             `json:"filename"`
	StoragePath  string             `json:"storage_path"`
	CreatedAt    time.Time          `json:"created_at" gorm:"autoCreateTime"`
	RetentionDays int               `json:"retention_days" gorm:"default:365"` // Default 1 year
	Analysis     ContractAnalysis   `json:"analysis" gorm:"embedded"`
}

type ContractSummary struct {
	BuyerName    string `json:"buyer_name"`
	BuyerAddress string `json:"buyer_address"`
	BuyerCountry string `json:"buyer_country"`
	SellerName   string `json:"seller_name"`
	SellerAddress string `json:"seller_address"`
	SellerCountry string `json:"seller_country"`
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

type ContractStatus string

const (
	Validated ContractStatus = "validated"
	Analyzed  ContractStatus = "analyzed"
)

// ContractAnalysis represents the structured analysis result from an LLM.
// This is a DTO used for parsing the raw LLM output.
type ContractAnalysis struct {
	Buyer         string              `json:"buyer"`
	BuyerAddress  string              `json:"buyer_address"`
	BuyerCountry  string              `json:"buyer_country"`
	Seller        string              `json:"seller"`
	SellerAddress string              `json:"seller_address"`
	SellerCountry string              `json:"seller_country"`
	TotalValue    decimal.Decimal     `json:"total_value" gorm:"type:decimal(20,8)"`
	Currency      string              `json:"currency"`
	Milestones    []AnalysisMilestone `json:"milestones" gorm:"-"`
	RiskFactors   []AnalysisRisk      `json:"risk_factors" gorm:"-"`
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

// ContractElementsResult represents detailed contract element detection results.
type ContractElementsResult struct {
	Parties      []ContractParty      `json:"parties"`
	Obligations  []ContractObligation `json:"obligations"`
	Terms        []ContractTerm       `json:"terms"`
	Confidence   float64              `json:"confidence"`
}

// ContractParty represents a party identified in the contract.
type ContractParty struct {
	Name    string `json:"name"`
	Role    string `json:"role"` // e.g., "buyer", "seller", "contractor", "client"
	Address string `json:"address,omitempty"`
	Contact string `json:"contact,omitempty"`
}

// ContractObligation represents an obligation identified in the contract.
type ContractObligation struct {
	Party       string `json:"party"`
	Description string `json:"description"`
	Type        string `json:"type"` // e.g., "payment", "delivery", "performance"
	Deadline    string `json:"deadline,omitempty"`
}

// ContractTerm represents a term or condition identified in the contract.
type ContractTerm struct {
	Type        string `json:"type"` // e.g., "payment", "termination", "warranty"
	Description string `json:"description"`
	Value       string `json:"value,omitempty"` // e.g., amount, duration
}

// ValidationRecord stores validation results with audit trail
type ValidationRecord struct {
	ID          string           `json:"id" gorm:"primaryKey"`
	ContractID  string           `json:"contract_id" gorm:"index"`
	UserID      string           `json:"user_id" gorm:"index"`
	ValidationType string        `json:"validation_type"` // e.g., "contract_validation", "element_detection"
	Result      ValidationResult `json:"result" gorm:"embedded"`
	Elements    *ContractElementsResult `json:"elements,omitempty" gorm:"embedded;embeddedPrefix:elements_"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	Version     int              `json:"version" gorm:"default:1"`
}

// ClassificationRecord stores classification results with audit trail.
type ClassificationRecord struct {
	ID             string                   `json:"id" gorm:"primaryKey"`
	ContractID     string                   `json:"contract_id" gorm:"index"`
	Classification *ContractClassification `json:"classification" gorm:"embedded"`
	Complexity     *ContractComplexity     `json:"complexity" gorm:"embedded;embeddedPrefix:complexity_"`
	Industry       *IndustryClassification `json:"industry" gorm:"embedded;embeddedPrefix:industry_"`
	CreatedAt      time.Time               `json:"created_at" gorm:"autoCreateTime"`
	Version        int                     `json:"version" gorm:"default:1"`
}

// ContractClassification represents the comprehensive classification of a contract.
type ContractClassification struct {
	PrimaryType    string                 `json:"primary_type"`    // e.g., "Sale of Goods", "Service Agreement"
	SubType        string                 `json:"sub_type"`        // e.g., "Software License", "Consulting Agreement"
	Industry       string                 `json:"industry"`        // e.g., "Technology", "Manufacturing"
	Complexity     string                 `json:"complexity"`      // Simple, Moderate, Complex, Highly Complex
	RiskLevel      string                 `json:"risk_level"`      // Low, Medium, High, Critical
	Jurisdiction   string                 `json:"jurisdiction"`    // Legal jurisdiction
	ContractValue  string                 `json:"contract_value"`  // Value range classification
	Duration       string                 `json:"duration"`        // Short-term, Long-term, etc.
	PartyTypes     []string               `json:"party_types"`     // B2B, B2C, Government, etc.
	SpecialClauses []string               `json:"special_clauses"`  // IP, Non-compete, etc.
	Confidence     float64                `json:"confidence"`      // Classification confidence
	Metadata       map[string]interface{} `json:"metadata"`        // Additional classification data
}

// ContractComplexity represents the complexity analysis of a contract.
type ContractComplexity struct {
	Level              string  `json:"level"`
	Score              float64 `json:"score"`              // 0.0-1.0
	Factors            []string `json:"factors"`            // Factors contributing to complexity
	ClauseCount        int     `json:"clause_count"`       // Number of clauses
	PageCount          int     `json:"page_count"`         // Estimated page count
	LegalTermCount     int     `json:"legal_term_count"`   // Number of legal terms
	CrossReferences    int     `json:"cross_references"`   // Internal references
	ExternalReferences int     `json:"external_references"` // External document references
}

// IndustryClassification represents industry-specific classification.
type IndustryClassification struct {
	PrimaryIndustry   string            `json:"primary_industry"`
	SecondaryIndustry string            `json:"secondary_industry,omitempty"`
	IndustryCode      string            `json:"industry_code,omitempty"` // NAICS or SIC code
	Regulations       []string          `json:"regulations"`             // Applicable regulations
	Standards         []string          `json:"standards"`               // Industry standards
	Compliance        map[string]string `json:"compliance"`              // Compliance requirements
	Confidence        float64           `json:"confidence"`
}

// ValidationAuditLog tracks changes to validation records
type ValidationAuditLog struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	ValidationID     string    `json:"validation_id" gorm:"index"`
	UserID           string    `json:"user_id" gorm:"index"`
	Action           string    `json:"action"` // e.g., "created", "updated", "reviewed"
	PreviousVersion  int       `json:"previous_version"`
	CurrentVersion   int       `json:"current_version"`
	Changes          string    `json:"changes" gorm:"type:text"` // JSON string of changes
	Reason           string    `json:"reason,omitempty"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ValidationFeedback stores user feedback on validation results
type ValidationFeedback struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	ValidationID string    `json:"validation_id" gorm:"index"`
	UserID       string    `json:"user_id" gorm:"index"`
	FeedbackType string    `json:"feedback_type"` // e.g., "accuracy", "completeness", "suggestion"
	Rating       int       `json:"rating"` // 1-5 scale
	Comment      string    `json:"comment,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
