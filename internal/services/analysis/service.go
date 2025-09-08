package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"encoding/base64"

	"contract-analysis-service/internal/models"
	"contract-analysis-service/internal/pkg/external"
	"contract-analysis-service/internal/services/llm"
)

// Service defines analysis capabilities over contract text
//go:generate mockery --name=Service --output=./mocks --filename=service_mock.go
type Service interface {
	AnalyzeContract(ctx context.Context, text string) (*models.ContractAnalysis, error)
	AnalyzeContractFromImages(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error)
}

type service struct {
	llmService  llm.Service
	promptEngine *llm.PromptEngine
}

func NewService(llmService llm.Service) Service {
	return &service{llmService: llmService, promptEngine: llm.NewPromptEngine()}
}

func (s *service) AnalyzeContract(ctx context.Context, text string) (*models.ContractAnalysis, error) {
	prompt := s.promptEngine.BuildContractAnalysisPrompt(text)

	payload := map[string]interface{}{
		"model": "qwen/qwen-2.5-vl-72b-instruct:free",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis payload: %w", err)
	}

	resp, err := s.llmService.ExecuteRequest(ctx, "openrouter", &external.Request{
		Method:  "POST",
		URL:     "/chat/completions",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    b,
	})
	if err != nil {
		return nil, fmt.Errorf("analysis request failed: %w", err)
	}

	return parseAnalysisResponse(resp.Body)
}

// AnalyzeContractFromImages performs one-step multimodal extraction+analysis directly from page images.
func (s *service) AnalyzeContractFromImages(ctx context.Context, imagePaths []string) (*models.ContractAnalysis, error) {
    if len(imagePaths) == 0 {
        return nil, fmt.Errorf("no images provided for analysis")
    }

    // Instruction adapted from PromptEngine JSON structure but for images, now including addresses and countries
    instruction := "You are a legal document analysis expert. Analyze the following contract from the provided page images and extract key information in JSON format with keys: buyer, buyer_address, buyer_country, seller, seller_address, seller_country, total_value (number), currency (string), milestones (array of {description, amount, percentage}), risk_factors (array of {type, description, severity}). Only return JSON."

    // Build multimodal content: one text block + multiple image_url blocks using data URLs
    var content []interface{}
    content = append(content, map[string]interface{}{"type": "text", "text": instruction})
    for _, p := range imagePaths {
        data, err := os.ReadFile(p)
        if err != nil {
            return nil, fmt.Errorf("failed to read image %s: %w", p, err)
        }
        b64 := base64.StdEncoding.EncodeToString(data)
        url := fmt.Sprintf("data:image/jpeg;base64,%s", b64)
        content = append(content, map[string]interface{}{
            "type": "image_url",
            "image_url": map[string]string{"url": url},
        })
    }

    payload := map[string]interface{}{
        "model": "qwen/qwen-2.5-vl-72b-instruct:free",
        "messages": []interface{}{
            map[string]interface{}{
                "role": "user",
                "content": content,
            },
        },
        "response_format": map[string]string{"type": "json_object"},
        "stream": false,
    }
    b, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal multimodal analysis payload: %w", err)
    }

    resp, err := s.llmService.ExecuteRequest(ctx, "openrouter", &external.Request{
        Method:  "POST",
        URL:     "/chat/completions",
        Headers: map[string]string{"Content-Type": "application/json"},
        Body:    b,
    })
    if err != nil {
        return nil, fmt.Errorf("multimodal analysis request failed: %w", err)
    }
    return parseAnalysisResponse(resp.Body)
}

func parseAnalysisResponse(body []byte) (*models.ContractAnalysis, error) {
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse analysis wrapper: %w", err)
	}
	if len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("no choices in analysis response")
	}
	var result models.ContractAnalysis
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse analysis json content: %w", err)
	}
	return &result, nil
}
