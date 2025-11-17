.PHONY: build clean install test help

# Default log file path
LOGFILE ?= ../../erupe.log

# Build the binary
build:
	@echo "Building loganalyzer..."
	@go build -o loganalyzer

# Install the binary to GOPATH/bin
install:
	@echo "Installing loganalyzer..."
	@go install

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f loganalyzer

# Run quick tests
test: build
	@echo "Running basic tests..."
	@./loganalyzer stats -f $(LOGFILE) > /dev/null && echo "✓ stats command works"
	@./loganalyzer errors -f $(LOGFILE) -summary > /dev/null && echo "✓ errors command works"
	@./loganalyzer connections -f $(LOGFILE) > /dev/null && echo "✓ connections command works"
	@./loganalyzer filter -f $(LOGFILE) -count > /dev/null && echo "✓ filter command works"
	@echo "All tests passed!"

# Quick stats
stats: build
	@./loganalyzer stats -f $(LOGFILE)

# Quick error summary
errors: build
	@./loganalyzer errors -f $(LOGFILE) -summary

# Show help
help:
	@echo "Erupe Log Analyzer - Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build       - Build the binary (default)"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run basic functionality tests"
	@echo "  stats       - Quick stats of log file"
	@echo "  errors      - Quick error summary"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  LOGFILE     - Path to log file (default: ../../erupe.log)"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make stats"
	@echo "  make test LOGFILE=/path/to/custom.log"

# Default target
.DEFAULT_GOAL := build
