# Build script for production installer
# Usage: ./build.sh (on Linux VPS) or build manually on Windows

# Cross-compile for Linux amd64
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -ldflags="-s -w" -o undangan-digital ./cmd/main.go

echo "Build complete: undangan-digital (Linux amd64)"
echo "Upload to VPS and run: ./undangan-digital"