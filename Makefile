.PHONY: build release clean

# Default target - build for current platform
build:
	mkdir -p build
	go build -o build/set-tab-color .

# Cross-compile for all target platforms
release:
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/set-tab-color-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o build/set-tab-color-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o build/set-tab-color-macos-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o build/set-tab-color-macos-arm64 .

# Clean build artifacts
clean:
	rm -rf build/

# Default target when just running 'make'
.DEFAULT_GOAL := build
