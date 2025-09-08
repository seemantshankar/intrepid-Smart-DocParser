package pdf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// findSamplePDF tries to find a PDF to use for tests.
// It first looks at the environment variable RASTER_TEST_PDF,
// otherwise it tries to pick the first PDF under the repository uploads/ folder.
func findSamplePDF(t *testing.T) (string, bool) {
	t.Helper()
	if p := os.Getenv("RASTER_TEST_PDF"); p != "" {
		return p, true
	}
	// try to locate a PDF under uploads/
	cwd, _ := os.Getwd()
	// Walk up to repo root by searching for go.mod
	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", false
		}
		root = parent
	}
	matches, _ := filepath.Glob(filepath.Join(root, "uploads", "*.pdf"))
	if len(matches) == 0 {
		return "", false
	}
	return matches[0], true
}

func TestRasterizeToJPEGs_Smoke(t *testing.T) {
	// Skip if pdftoppm is not available
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm not found in PATH; brew install poppler to run this test")
	}

	pdfPath, ok := findSamplePDF(t)
	if !ok {
		t.Skip("no sample PDF found; set RASTER_TEST_PDF to a valid PDF path")
	}

	images, err := RasterizeToJPEGs(pdfPath, 2)
	if err != nil {
		t.Fatalf("RasterizeToJPEGs failed: %v", err)
	}
	if len(images) == 0 {
		t.Fatalf("expected at least one generated image, got 0")
	}
	// Verify files exist
	for _, img := range images {
		if _, err := os.Stat(img); err != nil {
			t.Fatalf("generated image not found: %s (%v)", img, err)
		}
	}
}
