# Makefile
.PHONY: all build lint clean

# Default target
all: lint build

.PHONY: ci
ci: inst tidy all vulncheck

.PHONY: lint
# Run linters
lint:
    golangci-lint run

# Install dependencies (if needed in the future)
deps:
    go mod tidy

# Run vulnerability check
vulncheck:
		gosec ./...

# Build the plugin
build:
    go build -buildmode=plugin -o traefik-plugin-block-useragents.so block_useragents.go

# Run tests
test:
		go test -v ./...

# Clean up generated files
clean:
    rm -f traefik-plugin-block-useragents.so
