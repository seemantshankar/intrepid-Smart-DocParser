.PHONY: build test lint format swagger

# Build the application
build:
	go build -o bin/server main.go

# Run tests
test:
	go test -v -coverprofile=coverage.out ./...

# Run linters
lint:
	go vet ./...
	staticcheck ./...

# Format code
format:
	gofmt -w .

# Install development dependencies
dev-deps:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o ./api/swagger

# Run the application
run:
	go run cmd/server/main.go
