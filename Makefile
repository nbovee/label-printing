# Label Printer Makefile

# Default target
all: build

# Build the GUI executable (no console window)
build:
	@echo "Building Label Printer (GUI mode)..."
	go build -o label-printer.exe main.go
	@echo "Build successful! Run label-printer.exe to start the application."
	@echo "Note: Console window may appear briefly. For true GUI mode, use build-gui target."

# Build true GUI executable (alternative method)
build-gui:
	@echo "Building Label Printer (true GUI mode)..."
	go build -ldflags='-H windowsgui -s -w' -o label-printer-gui.exe main.go
	@echo "Build successful! Run label-printer-gui.exe to start the application."

# Build with console window (for debugging)
build-debug:
	@echo "Building Label Printer (debug mode)..."
	go build -o label-printer-debug.exe main.go
	@echo "Build successful! Run label-printer-debug.exe to start with console."

# Run the application directly
run:
	@echo "Running Label Printer..."
	go run main.go

# Install/update dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	@echo "Dependencies updated."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	del /Q *.exe 2>nul || echo "No executables to clean."
	del /Q *.pdf 2>nul || echo "No PDFs to clean."
	@echo "Clean complete."

# Build and run in one command
dev: build run

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build GUI executable (no console window)"
	@echo "  build-gui  - Build true GUI executable (alternative method)"
	@echo "  build-debug- Build executable with console window"
	@echo "  run        - Run application directly with go run"
	@echo "  deps       - Install/update dependencies"
	@echo "  clean      - Remove build artifacts and generated PDFs"
	@echo "  dev        - Build and run in one command"
	@echo "  help       - Show this help message"

.PHONY: all build build-gui build-debug run deps clean dev help 