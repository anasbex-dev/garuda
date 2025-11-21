#!/bin/bash

echo "Building Garuda Minecraft Server..."

# Clean previous build
rm -rf bin/
mkdir -p bin/

# Build for current platform
go build -o bin/garuda cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "Build successful! Binary: bin/garuda"
    echo "File size: $(du -h bin/garuda | cut -f1)"
else
    echo "Build failed!"
    exit 1
fi