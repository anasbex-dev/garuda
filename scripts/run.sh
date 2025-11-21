#!/bin/bash

echo "Starting Garuda Minecraft Server..."

# Check if binary exists
if [ ! -f "bin/garuda" ]; then
    echo "Binary not found. Building first..."
    ./scripts/build.sh
fi

# Run the server
./bin/garuda