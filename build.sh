#!/bin/bash
# Build script for production installer
# Usage: ./build.sh

echo "Building undangan-digital for Linux amd64..."

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go

if [ $? -eq 0 ]; then
    echo "Build complete: undangan-digital (Linux amd64)"
    echo "Upload to VPS and run: chmod +x undangan-digital && ./undangan-digital"
else
    echo "Build failed!"
    exit 1
fi
