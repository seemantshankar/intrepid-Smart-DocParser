# Smart-DocParser

Smart-DocParser is a Go-based document analysis system that leverages Large Language Models (LLMs) for contract parsing and extraction of key information such as obligations, payments, and milestones.

## Features
- **Document Upload and Analysis**: Upload PDFs or other files via REST API, rasterize PDFs to images for multimodal LLM analysis, with fallback to text-based extraction.
- **LLM Integration**: Supports OpenAI (GPT-4o) and other providers for contract analysis, including prompt engineering for structured JSON responses.
- **Analysis Endpoint**: POST `/contracts/analyze` handles file processing, OCR if needed, and returns structured results with contract details, risks, and milestones.
- **Document Management**: CRUD operations for documents with analysis retrieval.
- **Security**: JWT-based authorization for endpoints.

## Architecture
- **Handlers**: `analysis_handler.go` for analysis, `document_handler.go` for document operations.
- **Services**: LLM service in `internal/services/llm/` for contract analysis, milestone sequencing.
- **Models**: DTOs for contract analysis, risks, milestones in `internal/models/`.
- **Database**: PostgreSQL/SQLite repositories for persistence.

## Setup
1. Install dependencies: `go mod tidy`
2. Configure `.env` or `config.yaml` with LLM API keys.
3. Run: `go run main.go`

## API Endpoints
- `POST /contracts/analyze`: Analyze uploaded contract file.
- `POST /documents/upload`: Upload document.
- `GET /documents/{id}/analysis`: Retrieve analysis results.

## Requirements
See `docs/Requirements/requirements.md` for detailed specs.

## Testing

### Current Test Status
- ✅ **Unit Tests**: All unit tests pass
- ✅ **Build**: Application builds successfully
- ✅ **Linter**: All linting issues resolved
- ⚠️ **Integration Tests**: Some integration tests are currently skipped (require environment setup)
- ⚠️ **Repository Tests**: Some repository integration tests need fixes

### Running Tests
```bash
# Run all unit tests
make test

# Run specific test suites
go test ./internal/services/document/... -v
go test ./internal/handlers/... -v

# Run with coverage
go test -coverprofile=coverage.out ./...
```

## ✨ Features

### Document Upload Service (✅ COMPLETED)
- ✅ **Document Upload Handler**: Supports PDF, DOCX, TXT, JPG, PNG, TIFF formats
- ✅ **File Validation**: 10MB size limit and format verification
- ✅ **Secure Storage**: Document storage with metadata tracking
- ✅ **Authorization**: Proper user authorization for all document operations
- ✅ **Lifecycle Management**: Document retention and cleanup functionality
- ✅ **Unit Tests**: Comprehensive unit tests for upload validation and storage
- ⚠️ **Integration Tests**: Available but currently skipped (require environment setup)

### LLM-Powered Analysis
- **Contract Validation**: Determines if a document is a valid contract.
- **Element Detection**: Extracts key elements like parties, obligations, and terms.
- **Contract Classification**: Identifies the type of contract (e.g., Sale of Goods, Service Agreement).

### Infrastructure & Quality
- **OCR Integration**: Extracts text from images and scanned PDFs using cloud vision APIs.
- **Resilient External Clients**: Built-in retry logic and circuit breakers for all external API calls.
- **Caching Layer**: Uses Redis to cache OCR results, improving performance and reducing costs.
- **RESTful API**: A robust API built with the Gin framework (switched from Gorilla Mux for better performance and middleware support).
- **Swagger/OpenAPI**: Automatically generated, interactive API documentation.
- **Monitoring**: Exposes Prometheus metrics for performance monitoring.
- **Database Integration**: Uses GORM with PostgreSQL and SQLite for data persistence.
- **Production-Ready**: Features rate limiting, structured logging (Zap), request tracing, and graceful shutdown.

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (for PostgreSQL and Redis)
- Make (optional)

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/yourusername/contract-analysis-service.git
    cd contract-analysis-service
    ```

2.  **Set up environment variables:**
    Create a `.env` file in the project root and add your API keys:
    ```env
    OPENROUTER_API_KEY="your-openrouter-api-key"
    ```

3.  **Install dependencies:**
    ```bash
    go mod tidy
    go mod vendor
    ```

### Running the Service

1.  **Start Redis (required for readiness and OCR caching):**
    ```bash
    # Option A: Docker
    docker run --name redis -p 6379:6379 -d redis:7-alpine

    # Option B: macOS Homebrew
    # brew install redis
    # brew services start redis
    ```

2.  **Run the application:**
    ```bash
    # From the repository root (uses config.yaml)
    go run main.go
    ```

The service will be available at `http://localhost:9091`.

## 📚 API Documentation

Interactive API documentation is available at:
- **Swagger UI**: `http://localhost:9091/swagger/index.html`
- **Raw OpenAPI JSON**: `http://localhost:9091/swagger/doc.json`

If you update handler annotations and need to regenerate the spec:

```bash
# Install if needed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs targeting the current entry point
swag init -g main.go -o ./api/swagger
```

## 🩺 Health and Readiness

- Liveness: `GET /health` → returns `{"status":"healthy"}` with HTTP 200.
- Readiness: `GET /ready` → validates dependencies and returns cumulative status.

Example successful readiness response:

```json
{
  "status": "ready",
  "details": {
    "database": "ok",
    "redis": "ok"
  }
}
```

If a dependency is unavailable, `/ready` returns HTTP 503 with details (e.g., `redis: unhealthy: <error>`).

## 🛠 Development

### Testing

```bash
# Run all unit tests
make test

# Run repository integration tests (requires Docker; Postgres via testcontainers)
go test -tags repo_integration ./internal/repository/...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
make lint
```

### Build

```bash
# Build the server binary
make build
# or
go build -o bin/server main.go
```

## 🏗 Project Structure

```
.
├── api/              # OpenAPI/Swagger specifications
├── configs/          # Configuration structs and loading logic
├── docs/             # Project documentation
├── internal/         # Private application code
│   ├── handlers/     # HTTP request handlers
│   ├── middleware/   # Gin middleware
│   ├── models/       # Core data models
│   ├── pkg/          # Internal shared packages (db, logging, etc.)
│   ├── repositories/ # Data access layer (database interactions)
│   └── services/     # Business logic and service implementations
├── vendor/           # Go module dependencies
└── ...
```

## ⚙️ Configuration

Configuration is loaded from `config.yaml` and environment variables.

- `server.port`: HTTP port (default configured: `9091`)
- `database`: DB dialect/name; defaults to SQLite `./local.db` for local dev
- `redis.address`: Redis endpoint (default `localhost:6379`)
- `llm.openrouter`: External LLM settings; `OPENROUTER_API_KEY` can be provided via env

Example: `configs/config.go` and `config.yaml` define the supported fields.

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
