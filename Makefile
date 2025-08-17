# DNS Bench Makefile

# Variables
BINARY_NAME=dnsbench
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} main.go

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-arm64.exe main.go

# Create release archives
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@mkdir -p dist/releases
	cd dist && tar -czf releases/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64 ../README.md
	cd dist && tar -czf releases/${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64 ../README.md
	cd dist && tar -czf releases/${BINARY_NAME}-${VERSION}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64 ../README.md
	cd dist && tar -czf releases/${BINARY_NAME}-${VERSION}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64 ../README.md
	cd dist && zip releases/${BINARY_NAME}-${VERSION}-windows-amd64.zip ${BINARY_NAME}-windows-amd64.exe ../README.md
	cd dist && zip releases/${BINARY_NAME}-${VERSION}-windows-arm64.zip ${BINARY_NAME}-windows-arm64.exe ../README.md

# Run tests
.PHONY: test
test:
	go test -v -race ./...

# Run linting
.PHONY: lint
lint:
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, run: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Run the tool with example parameters
.PHONY: run
run: build
	./${BINARY_NAME} -domain example.com -count 5

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all supported platforms"
	@echo "  package    - Create release packages"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linting tools"
	@echo "  fmt        - Format code"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download and verify dependencies"
	@echo "  run        - Build and run with example parameters"
	@echo "  help       - Show this help message"
