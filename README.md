# Contract Analysis and Milestone Extraction Service

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/contract-analysis-service)](https://goreportcard.com/report/github.com/yourusername/contract-analysis-service)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/contract-analysis-service.svg)](https://pkg.go.dev/github.com/yourusername/contract-analysis-service)

A high-performance, production-ready microservice built with Go that automates the analysis of uploaded contracts. It leverages external LLMs to extract key information, validate contract integrity, and identify payment milestones. The service is designed with a clean architecture, ensuring maintainability and scalability.

## ✨ Features

- **Document Upload**: Supports various formats (PDF, DOCX, TXT, JPG, PNG, TIFF) with validation for file size and type.
- **LLM-Powered Analysis**: Integrates with external LLM providers (e.g., OpenRouter) for:
  - **Contract Validation**: Determines if a document is a valid contract.
  - **Element Detection**: Extracts key elements like parties, obligations, and terms.
  - **Contract Classification**: Identifies the type of contract (e.g., Sale of Goods, Service Agreement).
- **OCR Integration**: Extracts text from images and scanned PDFs using cloud vision APIs.
- **Resilient External Clients**: Built-in retry logic and circuit breakers for all external API calls.
- **Caching Layer**: Uses Redis to cache OCR results, improving performance and reducing costs.
- **RESTful API**: A robust API built with the Gin framework.
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

1.  **Start the infrastructure (PostgreSQL & Redis):**
    ```bash
    docker-compose up -d
    ```

2.  **Run the application:**
    ```bash
    make run
    ```

The service will be available at `http://localhost:9091`.

## 📚 API Documentation

Interactive API documentation is available at:
- **Swagger UI**: `http://localhost:9091/swagger/index.html`

## 🛠 Development

### Testing

```bash
# Run all unit tests
make test

# Run integration tests (requires Docker and .env file)
make test-integration
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
make lint
```

## 🏗 Project Structure

```
.
├── api/              # OpenAPI/Swagger specifications
├── cmd/              # Main application entry points
│   └── server/       # The main server application
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

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
