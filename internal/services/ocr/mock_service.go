package ocr

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockOCRService is a mock implementation of OCR Service
type MockOCRService struct {
	mock.Mock
}

// ExtractTextFromImage mocks the ExtractTextFromImage method
func (m *MockOCRService) ExtractTextFromImage(ctx context.Context, imagePath string) (*OCRResult, error) {
	args := m.Called(ctx, imagePath)
	if v := args.Get(0); v != nil {
		return v.(*OCRResult), args.Error(1)
	}
	return nil, args.Error(1)
}
