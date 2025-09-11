# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

Smart-DocParser is a Go microservice for intelligent contract analysis using Large Language Models (LLMs). It processes PDF contracts through OCR, extracts structured data including payment milestones, performs risk assessments, and generates Smart Cheque configurations for milestone-based payments on the Ripple network.

## Essential Commands

### Build and Run
```bash
# Build the application
make build
# or
go build -o bin/server main.go

# Run from project root (uses config.yaml)
go run main.go

# Generate Swagger documentation
make swagger
# or
swag init -g main.go -o ./docs
```

### Testing
```bash
# Run all unit tests with coverage (recommended)
make test

# Run specific test suites
go test ./internal/services/document/... -v
go test ./internal/handlers/... -v

# Run repository integration tests (requires Docker)
go test -tags repo_integration ./internal/repository/...

# Run with detailed coverage
go test -coverprofile=coverage.out ./...
```

### Code Quality
```bash
# Format code
make format

# Run linters (includes go vet and staticcheck)
make lint

# Install development dependencies
make dev-deps
```

## Architecture

### Core Architecture Pattern
The project follows **Clean Architecture/Hexagonal Architecture** with strict dependency inversion:
- **Handlers** (HTTP layer) → **Services** (business logic) → **Repositories** (data access)
- **Container** pattern for dependency injection and lifecycle management
- **Repository** pattern abstracts database operations
- All external dependencies are injected via interfaces

### Key Components
- **LLM Service**: OpenRouter/OpenAI integration for contract analysis with circuit breakers
- **OCR Service**: Cached text extraction from images and PDFs using Qwen Vision API
- **Document Service**: File upload, validation, and processing orchestration  
- **Analysis Service**: Comprehensive contract analysis and risk assessment
- **Prompt Engine**: Centralized prompt management for all LLM interactions

### Data Flow Architecture
1. **Document Upload** → File validation → PDF rasterization (if needed)
2. **OCR Processing** → Text extraction → Redis caching
3. **LLM Analysis** → Structured extraction via prompt engine
4. **Data Persistence** → SQLite/PostgreSQL via GORM repositories
5. **Response** → Standardized JSON APIs with Swagger documentation

## Critical Development Rules

### Prompt Engineering (MANDATORY)
All LLM prompts MUST use the centralized prompt engine in `internal/services/llm/prompt_engine.go`:

```go
// ✅ CORRECT - Use prompt engine
instruction := s.promptEngine.BuildComprehensiveContractAnalysisPrompt()

// ❌ WRONG - Never embed prompts directly in services
instruction := `You are a legal expert...`
```

### Available Prompt Methods
- `BuildContractAnalysisPrompt(contractText)` - Basic analysis
- `BuildComprehensiveContractAnalysisPrompt()` - Advanced financial analysis
- `BuildTextBasedContractAnalysisPrompt(contractText)` - Text-based analysis  
- `BuildMilestoneSequencingPrompt(milestones)` - Payment sequencing
- `BuildRiskAssessmentPrompt(contractText, standards)` - Risk evaluation

### Configuration Management
- Primary config: `config.yaml` (server settings, database, timeouts)
- Environment variables: `OPENROUTER_API_KEY`, `QWEN_API_KEY`
- Use Viper for configuration loading in `configs/config.go`

### Database Patterns
- **Primary**: PostgreSQL with GORM for production
- **Development**: SQLite (`./local.db`) for local development
- **Testing**: TestContainers for isolated integration tests
- Repository pattern with interface abstractions in `internal/repositories/`

## Service Dependencies

### Required External Services
- **Redis**: Required for OCR caching and readiness checks (`localhost:6379`)
- **OpenRouter API**: LLM services for contract analysis
- **Qwen Vision API**: OCR fallback service

### Dependency Startup Order
1. Start Redis: `docker run --name redis -p 6379:6379 -d redis:7-alpine`
2. Set environment variables in `.env`
3. Run application: `go run main.go`

## API Architecture

### Core Endpoints
- `POST /contracts/upload` - File upload with validation
- `POST /contracts/upload-analyze` - Upload and immediate analysis
- `POST /contracts/analyze-comprehensive` - Advanced contract analysis
- `GET /contracts/:id` - Retrieve contract details
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe (validates Redis, DB)

### Response Patterns
- All endpoints return standardized JSON with proper HTTP status codes
- Error responses include structured error messages
- Swagger documentation available at `/swagger/index.html`

## Financial Calculations

The system performs sophisticated financial analysis:
- **Flexible JSON Parsing**: Handles both string and numeric amounts via `FlexibleString` type
- **Tax Calculations**: Applies country-specific tax rates (e.g., India GST 18%)
- **Percentage Calculations**: Converts milestone percentages to absolute amounts
- **Discrepancy Detection**: Compares contract-stated vs. calculated amounts

## Testing Strategy

### Test Coverage Requirements
- Maintain 95%+ test coverage
- Unit tests with mocked dependencies using testify/mock
- Integration tests with TestContainers for database operations
- HTTP handler tests using Gin test mode

### Key Testing Patterns
```bash
# Unit tests only (fast)
go test ./internal/services/... -short

# Integration tests (requires Docker)
go test ./internal/repository/... -v

# Full test suite with coverage
make test
```

## Production Features

### Resilience Patterns
- **Circuit Breakers**: All external API calls (Sony/gobreaker)
- **Retry Logic**: Configurable backoff for LLM and OCR services
- **Graceful Shutdown**: 30-second timeout for request completion
- **Caching**: Redis caching for OCR results (1-hour TTL)

### Monitoring & Observability
- **Structured Logging**: Zap with correlation IDs
- **Metrics**: Prometheus metrics at `/metrics` endpoint
- **Tracing**: Optional OpenTelemetry/Jaeger integration
- **Health Checks**: Kubernetes-ready liveness/readiness probes

## Module Information

- **Go Module**: `contract-analysis-service`
- **Go Version**: 1.25+
- **Primary Framework**: Gin (HTTP), GORM (ORM), Viper (config)
- **Key Dependencies**: OpenRouter/OpenAI clients, Redis, TestContainers

## Development Workflow

1. **Prerequisites**: Go 1.25+, Redis, Docker (for tests)
2. **Setup**: Run `go mod tidy` and create `.env` with API keys
3. **Development**: Use `make test` and `make lint` before commits
4. **Documentation**: Regenerate Swagger docs after handler changes
5. **Testing**: Run integration tests via `make test` (includes TestContainers)
