#!/bin/bash
# Build script for Red Team C2 Implant
# Cross-compiles for multiple platforms

set -e

echo "================================"
echo "Red Team C2 Implant Build Script"
echo "================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build directory
BUILD_DIR="build"
mkdir -p $BUILD_DIR

# Version info
VERSION="1.0.0"
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${YELLOW}Version:${NC} $VERSION"
echo -e "${YELLOW}Build Time:${NC} $BUILD_TIME"
echo -e "${YELLOW}Git Commit:${NC} $GIT_COMMIT"
echo ""

# Build flags
LDFLAGS="-s -w -X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT"

cd implant

# Linux AMD64
echo -e "${YELLOW}Building for Linux AMD64...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o ../$BUILD_DIR/implant-linux-amd64 main.go
echo -e "${GREEN}✓ Built: implant-linux-amd64${NC}"

# Linux ARM64
echo -e "${YELLOW}Building for Linux ARM64...${NC}"
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o ../$BUILD_DIR/implant-linux-arm64 main.go
echo -e "${GREEN}✓ Built: implant-linux-arm64${NC}"

# Windows AMD64
echo -e "${YELLOW}Building for Windows AMD64...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o ../$BUILD_DIR/implant-windows-amd64.exe main.go
echo -e "${GREEN}✓ Built: implant-windows-amd64.exe${NC}"

# macOS AMD64
echo -e "${YELLOW}Building for macOS AMD64...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o ../$BUILD_DIR/implant-darwin-amd64 main.go
echo -e "${GREEN}✓ Built: implant-darwin-amd64${NC}"

# macOS ARM64 (M1/M2)
echo -e "${YELLOW}Building for macOS ARM64...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o ../$BUILD_DIR/implant-darwin-arm64 main.go
echo -e "${GREEN}✓ Built: implant-darwin-arm64${NC}"

cd ..

# Generate checksums
echo ""
echo -e "${YELLOW}Generating checksums...${NC}"
cd $BUILD_DIR
sha256sum * > checksums.txt
echo -e "${GREEN}✓ Checksums saved to checksums.txt${NC}"

# Display build summary
echo ""
echo "================================"
echo "Build Summary"
echo "================================"
ls -lh
echo ""
echo -e "${GREEN}All builds completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Test implant: ./$BUILD_DIR/implant-linux-amd64 --help"
echo "  2. Deploy to target (authorized testing only!)"
echo "  3. Monitor with teamserver"
