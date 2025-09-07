package ocr

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockOCRService is a mock implementation of OCRService
type MockOCRService struct {
	mock.Mock
}

// ExtractTextFromImage mocks the ExtractTextFromImage method
func (m *MockOCRService) ExtractTextFromImage(ctx context.Context, imagePath string) (string, error) {
	args := m.Called(ctx, imagePath)
	return args.String(0), args.Error(1)
}
