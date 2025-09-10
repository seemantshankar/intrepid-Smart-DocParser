# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Smart-DocParser is a Go-based contract analysis microservice that uses Large Language Models (LLMs) for contract parsing and extraction of obligations, payments, and milestones. The system leverages OpenAI/Claude APIs for analysis and generates Smart Cheque configurations for milestone-based payments.

## Development Commands

### Build and Run
```bash
# Build the application
make build
# or
go build -o bin/server main.go

# Run the server (from project root, uses config.yaml)
go run main.go

# Run the server with specific command
go run cmd/server/main.go
```

### Testing
```bash
# Run all unit tests with coverage
make test
# or
go test -v -coverprofile=coverage.out ./...

# Run specific test suites
go test ./internal/services/document/... -v
go test ./internal/handlers/... -v

# Run repository integration tests (requires Docker for testcontainers)
go test -tags repo_integration ./internal/repository/...
```

### Code Quality
```bash
# Format code
make format
# or 
go fmt ./...

# Run linters
make lint
# or
go vet ./...
staticcheck ./...
```

### Dependencies and Documentation
```bash
# Install development dependencies
make dev-deps

# Generate Swagger documentation
make swagger
# or
swag init -g main.go -o ./docs
```

### Prerequisites
- Go 1.25+
- Redis (for caching and readiness checks)
- Docker (for integration tests)

## Architecture Overview

### Clean Architecture Structure
The project follows hexagonal architecture with clear separation:

```
├── cmd/                  # Application entry points
├── configs/              # Configuration management (Viper-based)
├── internal/
│   ├── handlers/         # HTTP request handlers (Gin-based)
│   ├── middleware/       # HTTP middleware (auth, logging, rate limiting)
│   ├── models/           # Domain entities and DTOs
│   ├── pkg/              # Shared packages (DB, container, logging)
│   ├── repositories/     # Data access layer (GORM)
│   └── services/         # Business logic services
├── docs/                 # API documentation and requirements
└── api/                  # OpenAPI/Swagger specifications
```

### Key Components

**Main Entry Points:**
- `main.go`: Production server with full middleware stack
- `cmd/server/main.go`: Alternative entry point
- `cmd/ocr-demo/main.go`: OCR demonstration tool

**Core Services:**
- **Document Service**: File upload, validation, OCR processing
- **Analysis Service**: LLM-powered contract analysis and milestone extraction
- **LLM Service**: Integration with OpenRouter/OpenAI APIs (uses prompt engine)
- **Prompt Engine**: Centralized prompt management for all LLM interactions
- **OCR Service**: Text extraction from images and PDFs using Qwen API
- **Risk Assessment**: Industry knowledge database and vulnerability analysis

**Data Layer:**
- PostgreSQL/SQLite with GORM
- Redis for caching and performance
- Repository pattern for data access

### Configuration

Configuration is managed via `config.yaml` and environment variables:

```yaml
server:
  port: "9091"                    # HTTP server port
database:
  dialect: "sqlite3"              # or "postgres"
  name: "./local.db"              # database file/connection
redis:
  address: "localhost:6379"       # Redis connection
llm:
  openrouter:
    base_url: "https://openrouter.ai/api/v1"
    api_key: ""                   # Set via OPENROUTER_API_KEY env var
```

**Environment Variables:**
- `OPENROUTER_API_KEY`: OpenRouter API key for LLM services
- `QWEN_API_KEY`: Qwen API key for OCR services

### API Structure

**Core Endpoints:**
- `POST /contracts/upload`: Upload and validate documents
- `POST /contracts/upload-analyze`: Upload and immediately analyze
- `POST /contracts/analyze`: Analyze existing contract
- `GET /contracts/:id`: Retrieve contract details
- `DELETE /contracts/:id`: Remove contract

**Health Checks:**
- `GET /health`: Liveness check (returns 200 with status)
- `GET /ready`: Readiness check (validates DB, Redis dependencies)

**Documentation:**
- `GET /swagger/*any`: Interactive API documentation

### External Dependencies

**LLM Integration:**
- OpenRouter API for GPT-4o, Claude models
- Circuit breakers and retry logic for resilience
- Configurable timeouts and fallback strategies

**OCR Integration:**
- Qwen Vision API for image and scanned PDF text extraction
- PDF text extraction fallback using ledongthuc/pdf

**Database:**
- Primary: PostgreSQL with GORM
- Development: SQLite (./local.db)
- Integration tests: TestContainers for isolated testing

### Testing Strategy

**Test Coverage Requirements:** 95%+ coverage maintained

**Test Types:**
- Unit tests with mocked dependencies (testify/mock)
- Repository integration tests with TestContainers
- HTTP handler tests with Gin test mode
- End-to-end contract analysis flows

**Key Test Commands:**
```bash
# Unit tests only
go test ./internal/services/... -short

# Integration tests (requires Docker)  
go test ./internal/repository/... -v

# All tests with coverage
go test -coverprofile=coverage.out ./...
```

### Production Considerations

**Middleware Stack:**
- CORS, request ID propagation, Zap logging
- JWT authentication and authorization
- Rate limiting and circuit breakers
- Graceful shutdown handling

**Monitoring:**
- Prometheus metrics at `/metrics`
- Structured logging with correlation IDs
- Health and readiness endpoints for K8s

**Security:**
- File upload size limits (10MB)
- Input validation and sanitization
- API key management via environment variables
- No sensitive data in logs or commits

### Prompt Engineering Guidelines

**IMPORTANT:** All LLM prompts must be managed through the centralized `prompt_engine.go` file:

```go
// ✅ CORRECT: Use prompt engine
instruction := s.promptEngine.BuildComprehensiveContractAnalysisPrompt()

// ❌ INCORRECT: Embedded prompts in service files
instruction := `You are a legal expert...` // Don't do this!
```

**Prompt Engine Location:** `internal/services/llm/prompt_engine.go`

**Available Prompt Methods:**
- `BuildContractAnalysisPrompt(contractText)` - Basic contract analysis
- `BuildComprehensiveContractAnalysisPrompt()` - Advanced financial analysis for images
- `BuildTextBasedContractAnalysisPrompt(contractText)` - Advanced analysis for text
- `BuildMilestoneSequencingPrompt(milestones)` - Milestone ordering
- `BuildRiskAssessmentPrompt(contractText, standards)` - Risk analysis

**When adding new LLM functionality:**
1. Add new prompt method to `prompt_engine.go`
2. Use the prompt engine in your service
3. Never embed prompts directly in service files
4. Follow existing naming conventions: `Build{Purpose}Prompt(...)`

### Development Workflow

1. **Start dependencies**: Ensure Redis is running (`docker run -p 6379:6379 redis:7-alpine`)
2. **Environment setup**: Create `.env` with required API keys
3. **Generate docs**: Run `make swagger` after handler changes
4. **Run tests**: Use `make test` to ensure quality gates pass
5. **Check build**: Use `make build` to verify compilation
6. **Linting**: Run `make lint` to catch issues early

### Module Information

**Go Module:** `contract-analysis-service`
**Go Version:** 1.25+

**Key Dependencies:**
- `github.com/gin-gonic/gin`: HTTP framework
- `gorm.io/gorm`: ORM and database toolkit
- `github.com/spf13/viper`: Configuration management
- `go.uber.org/zap`: Structured logging
- `github.com/swaggo/gin-swagger`: API documentation
- `github.com/shopspring/decimal`: Precise financial calculations
- `github.com/testcontainers/testcontainers-go`: Integration testing