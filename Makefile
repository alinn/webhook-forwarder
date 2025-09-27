.PHONY: build-server build-client build-all clean proto run-server run-client

# Build the server binary
build-server:
	@echo "Building server..."
	@go build -o bin/webhook-server cmd/server/main.go

# Build the client binary
build-client:
	@echo "Building client..."
	@go build -o bin/webhook-client cmd/client/main.go

# Build both binaries
build-all: build-server build-client

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/webhook.proto

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
