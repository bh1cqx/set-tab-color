.PHONY: build compile release clean test

# Default target - compile then run tests
build: compile test

# Compile only (no tests)
compile:
	mkdir -p build
	go build -o build/set-tab-color .

# Cross-compile for all target platforms
release: build
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/set-tab-color-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o build/set-tab-color-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o build/set-tab-color-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o build/set-tab-color-darwin-arm64 .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf build/

# Default target when just running 'make'
.DEFAULT_GOAL := build
