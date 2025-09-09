package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	Save(file io.Reader, fileName string) (string, error)
	Delete(filePath string) error
	Read(filePath string) ([]byte, error)
}

// LocalStorage implements FileStorage for the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath},
		nil
}

// Save saves a file to the local filesystem and returns the path.
func (s *LocalStorage) Save(file io.Reader, fileName string) (string, error) {
	// Generate a unique filename to prevent collisions
	ext := filepath.Ext(fileName)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(s.basePath, newFileName)

	// Create the destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return dstPath, nil
}

// Delete removes a file from the local filesystem.
func (s *LocalStorage) Delete(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Read reads a file from the local filesystem.
func (s *LocalStorage) Read(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
