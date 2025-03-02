# Makefile
.PHONY: all build lint clean

# Default target
all: lint build

# Build the plugin
build:
    GO111MODULE=on go build -buildmode=plugin -o traefik-plugin-block-useragents.so .

# Run linters
lint:
    golangci-lint run

# Clean up generated files
clean:
    rm -f traefik-plugin-block-useragents.so

# Install dependencies (if needed in the future)
deps:
    go mod tidy