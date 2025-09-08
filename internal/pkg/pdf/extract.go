package pdf

import (
	"bytes"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

const (
	minTextLength = 100 // Minimum characters to consider PDF as text-based
)

func ExtractText(pdfPath string) (string, bool) {
	f, err := os.Open(pdfPath)
	if err != nil {
		return "", false
	}
	defer f.Close()

	var buf bytes.Buffer
	fileInfo, err := f.Stat()
	if err != nil {
		return "", false
	}

	reader, err := pdf.NewReader(f, fileInfo.Size())
	if err != nil {
		return "", false
	}

	// Extract text from first 3 pages as a sample
	for i := 1; i <= 3 && i <= reader.NumPage(); i++ {
		p := reader.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
	}

	text := strings.TrimSpace(buf.String())
	return text, len(text) >= minTextLength
}
