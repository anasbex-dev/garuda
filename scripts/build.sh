#!/bin/bash

# Garuda Framework Build Script
# Copyright © 2025 - AnasBex

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BUILD_DIR="bin"
BINARY_NAME="garuda"
MAIN_PACKAGE="./cmd/server"
VERSION="1.0.0"
BUILD_DATE=$(date -u +"%Y-%m-%d_%H:%M:%S_UTC")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Build banner
print_build_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                     GARUDA BUILD SYSTEM                     ║"
    echo "║                  Copyright © 2025 - AnasBex                 ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Check if setup was run
check_setup() {
    if [ ! -f "config.json" ] && [ ! -f "go.mod" ]; then
        log_error "Setup not completed. Run ./scripts/setup.sh first"
        exit 1
    fi
}

# Clean build directory
clean_build() {
    log_info "Cleaning build directory..."
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"/*
        log_success "Build directory cleaned"
    else
        mkdir -p "$BUILD_DIR"
        log_success "Build directory created"
    fi
}

# Run tests
run_tests() {
    log_info "Running tests..."
    if go test ./... -v; then
        log_success "All tests passed"
    else
        log_error "Tests failed"
        exit 1
    fi
}

# Build for current platform
build_native() {
    local platform=$1
    local arch=$2
    local output_name="$BINARY_NAME"
    
    if [ "$platform" != "native" ]; then
        output_name="${BINARY_NAME}-${platform}-${arch}"
    fi
    
    log_info "Building for $platform/$arch..."
    
    # Set build flags
    local ldflags="-X main.version=$VERSION \
                  -X main.buildDate=$BUILD_DATE \
                  -X main.gitCommit=$GIT_COMMIT \
                  -X main.buildPlatform=$platform \
                  -w -s"
    
    if go build -ldflags "$ldflags" -o "$BUILD_DIR/$output_name" "$MAIN_PACKAGE"; then
        log_success "Build successful: $BUILD_DIR/$output_name"
        
        # Generate checksum
        cd "$BUILD_DIR"
        sha256sum "$output_name" > "$output_name.sha256"
        cd ..
        log_success "Checksum generated: $BUILD_DIR/$output_name.sha256"
    else
        log_error "Build failed for $platform/$arch"
        return 1
    fi
}

# Build for multiple platforms
build_cross_platform() {
    log_info "Starting cross-platform build..."
    
    local platforms=(
        "linux/amd64"
        "linux/arm64" 
        "windows/amd64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        local os=$(echo $platform | cut -d'/' -f1)
        local arch=$(echo $platform | cut -d'/' -f2)
        local output_name="${BINARY_NAME}-${os}-${arch}"
        
        if [ "$os" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        log_info "Building for $os/$arch..."
        
        if GOOS=$os GOARCH=$arch go build \
           -ldflags "-X main.version=$VERSION -X main.buildDate=$BUILD_DATE -X main.gitCommit=$GIT_COMMIT -w -s" \
           -o "$BUILD_DIR/$output_name" "$MAIN_PACKAGE"; then
            log_success "Built: $output_name"
            
            # Generate checksum
            cd "$BUILD_DIR"
            sha256sum "$output_name" > "$output_name.sha256"
            cd ..
        else
            log_error "Failed to build for $os/$arch"
        fi
    done
}

# Build optimized for current platform
build_optimized() {
    local platform=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    log_info "Building optimized binary for $platform/$arch..."
    
    local optimization_flags=""
    case $platform in
        linux)
            optimization_flags="-tags=netgo -ldflags=-extldflags=-static"
            ;;
        darwin)
            optimization_flags=""
            ;;
    esac
    
    local ldflags="-X main.version=$VERSION \
                  -X main.buildDate=$BUILD_DATE \
                  -X main.gitCommit=$GIT_COMMIT \
                  -w -s"
    
    if go build $optimization_flags -ldflags "$ldflags" -o "$BUILD_DIR/$BINARY_NAME" "$MAIN_PACKAGE"; then
        log_success "Optimized build completed: $BUILD_DIR/$BINARY_NAME"
        
        # Strip symbols to reduce size
        if command -v strip &> /dev/null; then
            strip "$BUILD_DIR/$BINARY_NAME"
            log_success "Binary stripped"
        fi
        
        # Show binary info
        log_info "Binary information:"
        file "$BUILD_DIR/$BINARY_NAME"
        du -h "$BUILD_DIR/$BINARY_NAME"
    else
        log_error "Optimized build failed"
        return 1
    fi
}

# Build for development (fast build)
build_dev() {
    log_info "Building development version..."
    
    if go build -race -o "$BUILD_DIR/${BINARY_NAME}-dev" "$MAIN_PACKAGE"; then
        log_success "Development build completed: $BUILD_DIR/${BINARY_NAME}-dev"
    else
        log_error "Development build failed"
        return 1
    fi
}

# Show build summary
show_summary() {
    log_info "Build Summary:"
    echo "┌──────────────────────────────────────────────────────┐"
    echo "│ Version:    $VERSION"
    echo "│ Build Date: $BUILD_DATE"  
    echo "│ Git Commit: $GIT_COMMIT"
    echo "│ Platform:   $(uname -s) $(uname -m)"
    echo "│ Output:     $BUILD_DIR/"
    echo "└──────────────────────────────────────────────────────┘"
    
    if [ -d "$BUILD_DIR" ]; then
        log_info "Built files:"
        ls -la "$BUILD_DIR/" | grep -E "(garuda|.sha256)" | while read line; do
            echo "  $line"
        done
    fi
}

# Main build function
main() {
    print_build_banner
    
    local build_type="optimized"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --clean)
                clean_build
                shift
                ;;
            --test)
                run_tests
                shift
                ;;
            --dev)
                build_type="dev"
                shift
                ;;
            --cross-platform)
                build_type="cross"
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --clean          Clean build directory first"
                echo "  --test           Run tests before building"
                echo "  --dev            Build development version with race detector"
                echo "  --cross-platform Build for multiple platforms"
                echo "  --help           Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    check_setup
    
    log_info "Starting build process..."
    log_info "Build type: $build_type"
    
    case $build_type in
        dev)
            build_dev
            ;;
        cross)
            build_cross_platform
            ;;
        optimized)
            build_optimized
            ;;
    esac
    
    show_summary
    log_success "Build process completed successfully!"
}

main "$@"