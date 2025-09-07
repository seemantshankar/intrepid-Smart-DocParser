package ocr

import (
	"errors"
	"fmt"
)

const (
	minConfidenceThreshold = 0.75
	minTextLength          = 10
)

// Validator defines the interface for validating OCR results.
type Validator interface {
	Validate(result *OCRResult) error
}

// ocrResultValidator implements the Validator interface.
type ocrResultValidator struct{}

// NewValidator creates a new OCR result validator.
func NewValidator() Validator {
	return &ocrResultValidator{}
}

// Validate checks the quality of the OCR result.
func (v *ocrResultValidator) Validate(result *OCRResult) error {
	if result == nil {
		return errors.New("ocr result cannot be nil")
	}

	if result.Confidence < minConfidenceThreshold {
		return fmt.Errorf("ocr confidence score %.2f is below threshold of %.2f", result.Confidence, minConfidenceThreshold)
	}

	if len(result.Text) < minTextLength {
		return fmt.Errorf("extracted text length %d is below minimum of %d", len(result.Text), minTextLength)
	}

	return nil
}
