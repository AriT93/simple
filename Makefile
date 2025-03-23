# Makefile for AI Joke Agent

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=ai-agent

# Ginkgo parameters
GINKGO=ginkgo
GINKGO_OPTS=-v -r

# Colors
GREEN=\033[0;32m
NC=\033[0m

# Default target
all: test build

# Build the application
build:
	@echo "${GREEN}Building application...${NC}"
	$(GOBUILD) -o $(BINARY_NAME) .

# Run tests
test:
	@echo "${GREEN}Running tests...${NC}"
	$(GINKGO) $(GINKGO_OPTS)

# Run specific test suite
test-suite:
	@echo "${GREEN}Running specific test suite...${NC}"
	$(GINKGO) -v $(suite)

# Run the application
run: build
	@echo "${GREEN}Running application...${NC}"
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "${GREEN}Cleaning build artifacts...${NC}"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Install dependencies
deps:
	@echo "${GREEN}Installing dependencies...${NC}"
	$(GOCMD) mod tidy

# Help target
help:
	@echo "Available targets:"
	@echo "  all      - Run tests and build the application"
	@echo "  build    - Build the application"
	@echo "  test     - Run all tests"
	@echo "  test-suite suite=<package> - Run tests for a specific package"
	@echo "  run      - Build and run the application"
	@echo "  clean    - Remove build artifacts"
	@echo "  deps     - Install dependencies"
	@echo "  help     - Show this help message"

.PHONY: all build test test-suite run clean deps help
