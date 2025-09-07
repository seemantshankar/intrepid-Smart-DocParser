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
2. Identify the total contract value and currency
3. List all payment obligations with amounts, percentages, and trigger conditions
4. Identify any risk factors or concerns
5. Determine the nature of goods/services (physical, digital, services)

Return a JSON object with this structure:
{
  "buyer": "string",
  "seller": "string", 
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
