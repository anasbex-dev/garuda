.PHONY: build run clean test setup

build:
    @echo "Building Garuda Minecraft Server..."
    @go build -o bin/garuda ./cmd/server

run: build
    @echo "Starting Garuda Server..."
    @./bin/garuda

clean:
    @echo "Cleaning build files..."
    @rm -rf bin/

test:
    @echo "Running tests..."
    @go test ./...

setup:
    @echo "Setting up Garuda..."
    @go mod tidy
    @chmod +x scripts/*.sh

dev: setup build run