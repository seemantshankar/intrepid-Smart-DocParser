package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"contract-analysis-service/internal/models"
)

// PromptEngine handles prompt engineering for contract analysis
type PromptEngine struct{}

// NewPromptEngine creates a new prompt engine
func NewPromptEngine() *PromptEngine {
	return &PromptEngine{}
}

// BuildContractAnalysisPrompt creates a structured prompt for contract analysis
func (p *PromptEngine) BuildContractAnalysisPrompt(contractText string) string {
	return fmt.Sprintf(`You are a legal document analysis expert. Analyze the following contract and extract key information in JSON format.

CONTRACT TEXT:
"""
%s
"""

INSTRUCTIONS:
1. Extract the buyer name and seller name
2. Extract the full mailing address of both buyer and seller, including country
2. Identify the total contract value and currency
3. List all payment obligations with amounts, percentages, and trigger conditions
4. Identify any risk factors or concerns
5. Determine the nature of goods/services (physical, digital, services)

Return a JSON object with this structure:
{
  "buyer": "string",
  "buyer_address": "string",
  "buyer_country": "string",
  "seller": "string", 
  "seller_address": "string",
  "seller_country": "string",
  "total_value": number,
  "currency": "string",
  "milestones": [
    {
      "description": "string",
      "amount": number,
      "percentage": number,
      "trigger_condition": "string"
    }
  ],
  "risk_factors": [
    {
      "type": "string",
      "description": "string",
      "severity": "low|medium|high|critical"
    }
  ],
  "goods_nature": "physical|digital|services"
}

Only return the JSON, no additional text.`, strings.TrimSpace(contractText))
}

// BuildMilestoneSequencingPrompt creates a prompt for milestone sequencing
func (p *PromptEngine) BuildMilestoneSequencingPrompt(milestones []models.AnalysisMilestone) string {
	milestonesJSON, _ := json.Marshal(milestones)

	return fmt.Sprintf(`You are a project management expert. Sequence the following contract milestones in chronological and logical order.

MILESTONES:
%s

INSTRUCTIONS:
1. Analyze the trigger conditions for each milestone
2. Sequence them chronologically based on contract timeline
3. Identify any dependencies between milestones
4. Group related milestones by functional categories
5. Ensure the total percentages sum to 100%%

Return a JSON array of sequenced milestones with dependencies and categories:
[
  {
    "id": "string",
    "description": "string",
    "sequence_order": number,
    "category": "string",
    "dependencies": ["milestone_id1", "milestone_id2"],
    "percentage": number
  }
]

Only return the JSON array, no additional text.`, string(milestonesJSON))
}

// BuildRiskAssessmentPrompt creates a prompt for risk assessment
func (p *PromptEngine) BuildRiskAssessmentPrompt(contractText, industryStandards string) string {
	return fmt.Sprintf(`You are a risk management expert. Assess the following contract for potential risks and vulnerabilities.

CONTRACT TEXT:
"""
%s
"""

INDUSTRY STANDARDS:
"""
%s
"""

INSTRUCTIONS:
1. Compare the contract against industry best practices
2. Identify missing contractual elements or clauses
3. Assess risks for both buyer and seller
4. Suggest specific improvements with legal reasoning
5. Categorize risks by severity

Return a JSON object with risk assessment:
{
  "missing_clauses": ["string"],
  "risks": [
    {
      "party": "buyer|seller",
      "type": "string",
      "severity": "low|medium|high|critical",
      "description": "string",
      "recommendation": "string"
    }
  ],
  "compliance_score": number,
  "suggestions": ["string"]
}

Only return the JSON, no additional text.`, contractText, industryStandards)
}

// BuildComprehensiveContractAnalysisPrompt creates a detailed prompt for comprehensive contract analysis with financial calculations
func (p *PromptEngine) BuildComprehensiveContractAnalysisPrompt() string {
	return `You are a legal document analysis expert with financial calculation capabilities. Analyze the following contract from the provided page images and extract key information in JSON format.

CRITICAL REQUIREMENTS:
1. Extract ALL payment amounts, royalties, advances, percentages, and financial obligations
2. Convert written amounts (like "fifty thousand" or "Rs. 50,000") to numbers
3. Apply factual tax rates based on seller's country (e.g., India GST 18% for services)
4. Perform your own calculations for percentages, tax rates, and totals
5. Compare your calculations with contract figures and identify discrepancies

REQUIRED JSON STRUCTURE - Use these exact field names and types:
{
  "buyer": "string",
  "buyer_address": "string", 
  "buyer_country": "string",
  "seller": "string",
  "seller_address": "string",
  "seller_country": "string", 
  "total_value": "50000",
  "currency": "USD",
  "contract_stated_total": "50000",
  "milestones": [{"description": "string", "amount": "25000", "percentage": 50.0, "calculated_amount": "25000", "due_date": "string"}],
  "payment_obligations": [{"party": "string", "description": "string", "amount": "25000", "calculated_amount": "25000", "due_date": "string", "conditions": "string"}],
  "tax_analysis": {"applicable_tax_rate": "18", "tax_amount": "8100", "net_amount": "45000", "tax_basis": "services"},
  "calculations": {"base_amounts": ["45000"], "percentages_applied": ["18"], "tax_calculations": ["8100"], "final_totals": ["53100"]},
  "discrepancies": [{"type": "string", "contract_amount": "string", "calculated_amount": "string", "difference": "string", "explanation": "string"}],
  "risk_factors": [{"type": "string", "description": "string", "severity": "string"}]
}

IMPORTANT: Return ONLY valid JSON matching this exact structure. All numeric amounts must be strings. No additional text or explanation outside the JSON structure.`
}

// BuildTextBasedContractAnalysisPrompt creates a prompt for analyzing contract text (non-image) with financial calculations
func (p *PromptEngine) BuildTextBasedContractAnalysisPrompt(contractText string) string {
	return fmt.Sprintf(`You are a legal document analysis expert with financial calculation capabilities. Analyze the following contract text and extract key information in JSON format.

CONTRACT TEXT:
"""
%s
"""

CRITICAL REQUIREMENTS:
1. Extract ALL payment amounts, royalties, advances, percentages, and financial obligations
2. Convert written amounts (like "fifty thousand" or "Rs. 50,000") to numbers
3. Apply factual tax rates based on seller's country (e.g., India GST 18%% for services)
4. Perform your own calculations for percentages, tax rates, and totals
5. Compare your calculations with contract figures and identify discrepancies

Extract these keys:
- buyer: person/entity receiving goods/services
- buyer_address: full address
- buyer_country: country name
- seller: person/entity providing goods/services  
- seller_address: full address
- seller_country: country name
- total_value: total contract value as number (your calculated amount)
- currency: currency code (INR, USD, etc.)
- contract_stated_total: total as stated in contract document
- milestones: array of {description, amount, percentage, calculated_amount, due_date}
- payment_obligations: array of {party, description, amount, calculated_amount, due_date, conditions}
- tax_analysis: {applicable_tax_rate, tax_amount, net_amount, tax_basis}
- calculations: {base_amounts: [], percentages_applied: [], tax_calculations: [], final_totals: []}
- discrepancies: array of {type, contract_amount, calculated_amount, difference, explanation}
- risk_factors: array of {type, description, severity}

CALCULATION INSTRUCTIONS:
- For India: Apply 18%% GST on services, 12%% on goods (unless exempted)
- For USA: Apply state sales tax where applicable
- Calculate royalties as percentage of stated base amounts
- Verify all percentage calculations shown in document
- Cross-check totals, subtotals, and tax computations

Pay special attention to:
- Royalty rates and their calculation basis
- Advance payments and adjustments
- Revenue sharing percentages
- Tax calculations and compliance
- Payment schedules with amounts
- Penalty clauses with financial impact

Only return valid JSON with all calculations performed.`, strings.TrimSpace(contractText))
}
