package mocks

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// FileStorage is a mock implementation of the FileStorage interface.
type FileStorage struct {
	mock.Mock
}

// Save mocks the Save method.
func (m *FileStorage) Save(file io.Reader, fileName string) (string, error) {
	args := m.Called(file, fileName)
	return args.String(0), args.Error(1)
}

// Delete mocks the Delete method.
func (m *FileStorage) Delete(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

// Read mocks the Read method.
func (m *FileStorage) Read(filePath string) ([]byte, error) {
	args := m.Called(filePath)
	return args.Get(0).([]byte), args.Error(1)
}
