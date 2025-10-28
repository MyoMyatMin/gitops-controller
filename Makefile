BINARY_NAME=gitops-controller

.PHONY: all build clean test

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/main.go

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@go clean
	@rm -f $(BINARY_NAME)

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)