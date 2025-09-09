# OCR Demo for Scanned PDF Contract Processing

This demo application demonstrates the complete workflow for processing scanned PDF contracts using OCR (Optical Character Recognition) and contract validation.

## Overview

The OCR demo performs the following steps:

1. **PDF Rasterization**: Converts PDF pages to JPEG images using `pdftoppm`
2. **OCR Text Extraction**: Extracts text from images using OpenRouter's vision models
3. **Contract Validation**: Validates the extracted text as a legal contract using LLM
4. **Results Display**: Shows validation results including contract type, confidence, and detected elements

## Prerequisites

### System Dependencies

```bash
# Install poppler-utils for PDF rasterization
brew install poppler  # macOS
# or
sudo apt-get install poppler-utils  # Ubuntu/Debian
```

### Environment Variables

Create a `.env` file in the project root with:

```env
OPENROUTER_API_KEY=your_openrouter_api_key_here
```

## Usage

### Build the Demo

```bash
go build -o ocr-demo cmd/ocr-demo/main.go
```

### Run the Demo

```bash
./ocr-demo <path-to-pdf>
```

**Example:**
```bash
./ocr-demo uploads/sample_contract.pdf
```

### File Size Limitations

- Maximum PDF size: 10MB (to avoid excessive API usage)
- Maximum pages processed: 3 (configurable in code)
- Minimum OCR confidence: 0.1 (configurable)

## Sample Output

```
Processing PDF: uploads/contract.pdf (3705086 bytes)
Step 1: Converting PDF to images...
Generated 2 images
Step 2: Setting up OCR service...
Step 3: Extracting text from images...
Processing image 1/2: page-1.jpg
OCR confidence for image 1: 0.85
Extracted text length: 1247 characters
Processing image 2/2: page-2.jpg
OCR confidence for image 2: 0.92
Extracted text length: 1156 characters
Step 4: Combined extracted text length: 2403 characters
First 300 characters: This Agreement is entered into on [DATE] between...
Step 5: Setting up validation service...
Step 6: Validating extracted text as contract...

=== VALIDATION RESULTS ===
Is Valid Contract: true
Contract Type: Service Agreement
Confidence: 0.87
Reason: Document contains all essential contract elements

Detected Elements:
  1. Parties identification
  2. Consideration/payment terms
  3. Scope of work
  4. Termination clauses

=== OCR DEMO COMPLETED ===
Successfully processed scanned PDF through OCR pipeline
```

## Architecture

### Components Used

1. **PDF Rasterizer** (`internal/pkg/pdf`): Converts PDF to images
2. **OCR Service** (`internal/services/ocr`): Extracts text from images
3. **Validation Service** (`internal/services/validation`): Validates contract content
4. **LLM Service** (`internal/services/llm`): Provides AI-powered analysis

### OCR Models

- Primary: `qwen/qwen2.5-vl-32b-instruct:free` (OpenRouter)
- Fallback models configurable in code
- Vision-language models for document understanding

## Error Handling

The demo includes comprehensive error handling for:

- Missing PDF files
- File size limitations
- PDF rasterization failures
- OCR processing errors
- Low confidence results
- API failures with retry logic

## Troubleshooting

### Common Issues

1. **"pdftoppm failed"**: Install poppler-utils
2. **"PDF file too large"**: Use smaller PDF files (<500KB)
3. **"OCR confidence too low"**: Try higher quality scanned documents
4. **"OPENROUTER_API_KEY required"**: Set the environment variable

### Debug Mode

The demo includes detailed logging for each step. Check the output for:
- Image generation status
- OCR confidence scores
- Text extraction lengths
- Validation results

## Integration

This demo can be integrated into larger applications by:

1. Using the individual services directly
2. Implementing as a REST API endpoint
3. Adding to existing document processing pipelines
4. Extending with additional OCR providers

## Performance Considerations

- OCR processing time depends on image size and complexity
- API rate limits may apply for OpenRouter
- Memory usage scales with PDF size and page count
- Consider caching OCR results for repeated processing