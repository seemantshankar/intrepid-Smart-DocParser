package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// RasterizeToJPEGs uses the external `pdftoppm` tool to rasterize up to maxPages
// of the input PDF into JPEG images. It returns the list of generated image paths.
// Requires poppler-utils to be installed (e.g., `brew install poppler`).
func RasterizeToJPEGs(pdfPath string, maxPages int) ([]string, error) {
	if maxPages <= 0 {
		maxPages = 10
	}

	// Create a temporary directory for generated images
	tmpDir, err := os.MkdirTemp("", "docparser-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Output prefix; pdftoppm will append -1.jpg, -2.jpg, ...
	outPrefix := filepath.Join(tmpDir, "page")

	cmd := exec.Command(
		"pdftoppm",
		"-jpeg",
		"-f", "1",
		"-l", strconv.Itoa(maxPages),
		pdfPath,
		outPrefix,
	)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %w", err)
	}

	// Collect generated files; pdftoppm names them like prefix-1.jpg, prefix-2.jpg
	var images []string
	for i := 1; i <= maxPages; i++ {
		img := fmt.Sprintf("%s-%d.jpg", outPrefix, i)
		if _, statErr := os.Stat(img); statErr == nil {
			images = append(images, img)
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images were generated from PDF")
	}

	return images, nil
}
