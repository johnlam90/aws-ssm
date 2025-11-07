.PHONY: build install clean test run help

# Binary name
BINARY_NAME=aws-ssm

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	go install .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application (example: make run ARGS="list")
run: build
	./$(BINARY_NAME) $(ARGS)

# Display help
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  install  - Install the binary to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  run      - Build and run (use ARGS='...' to pass arguments)"
	@echo "  help     - Display this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run ARGS='list'"
	@echo "  make run ARGS='session web-server'"

