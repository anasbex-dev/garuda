#!/bin/bash

# Garuda Framework Setup Script
# Copyright © 2025 - AnasBex

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "$DEBUG" = "true" ]; then
        echo -e "${CYAN}[DEBUG]${NC} $1"
    fi
}

# Banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ██████╗  █████╗ ██████╗ ██╗   ██╗██████╗  █████╗     ███████╗██████╗     ║
║    ██╔════╝ ██╔══██╗██╔══██╗██║   ██║██╔══██╗██╔══██╗    ██╔════╝██╔══██╗    ║
║    ██║  ███╗███████║██████╔╝██║   ██║██████╔╝███████║    █████╗  ██████╔╝    ║
║    ██║   ██║██╔══██║██╔══██╗██║   ██║██╔══██╗██╔══██║    ██╔══╝  ██╔══██╗    ║
║    ╚██████╔╝██║  ██║██║  ██║╚██████╔╝██║  ██║██║  ██║    ███████╗██║  ██║    ║
║     ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝    ╚══════╝╚═╝  ╚═╝    ║
║                                                                              ║
║                    G A R U D A   F R A M E W O R K   M C                     ║
║                                                                              ║
║                   Copyright © 2025 - AnasBex - v1.0.0                       ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        log_info "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        log_error "Go version $GO_VERSION is not supported. Required: $REQUIRED_VERSION or later"
        exit 1
    fi
    
    log_success "Go version: $GO_VERSION"
    
    # Check Git
    if ! command -v git &> /dev/null; then
        log_warn "Git is not installed. Some features may not work properly."
    else
        log_success "Git is available"
    fi
    
    # Check platform
    detect_platform
}

# Detect platform and set variables
detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case $OS in
        Linux)
            if [ -f "/etc/termux-version" ] || [ -d "/data/data/com.termux" ]; then
                PLATFORM="Termux"
                BIN_DIR="$HOME/.local/bin"
            elif [ -f "/.dockerenv" ]; then
                PLATFORM="Docker"
                BIN_DIR="/usr/local/bin"
            else
                PLATFORM="Linux"
                BIN_DIR="$HOME/.local/bin"
            fi
            ;;
        Darwin)
            PLATFORM="macOS"
            BIN_DIR="/usr/local/bin"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            PLATFORM="Windows"
            BIN_DIR="/usr/local/bin"
            ;;
        *)
            PLATFORM="Unknown"
            BIN_DIR="./bin"
            ;;
    esac
    
    log_info "Platform: $PLATFORM ($OS $ARCH)"
    export GARUDA_PLATFORM="$PLATFORM"
    export GARUDA_BIN_DIR="$BIN_DIR"
}

# Setup directory structure
setup_directories() {
    log_info "Setting up directory structure..."
    
    local dirs=(
        "bin"
        "logs"
        "plugins"
        "worlds"
        "backups"
        "cache"
    )
    
    for dir in "${dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_debug "Created directory: $dir"
        fi
    done
    
    log_success "Directory structure created"
}

# Download dependencies
download_dependencies() {
    log_info "Downloading dependencies..."
    
    # Download Go modules
    if go mod download; then
        log_success "Dependencies downloaded successfully"
    else
        log_error "Failed to download dependencies"
        exit 1
    fi
    
    # Verify critical dependencies
    log_info "Verifying critical dependencies..."
    if go list -m github.com/google/uuid &> /dev/null; then
        log_success "UUID library: OK"
    else
        log_error "Missing critical dependency: github.com/google/uuid"
        exit 1
    fi
}

# Create configuration files
create_configs() {
    log_info "Creating configuration files..."
    
    # Main config
    if [ ! -f "config.json" ]; then
        if [ -f "config.example.json" ]; then
            cp config.example.json config.json
            log_success "Created config.json from example"
        else
            log_warn "config.example.json not found, creating basic config.json"
            cat > config.json << EOF
{
  "server": {
    "address": "0.0.0.0:19132",
    "max_players": 20,
    "motd": "§bGaruda§f Minecraft Server",
    "version": "1.21.50"
  },
  "world": {
    "name": "world",
    "seed": "garuda",
    "gamemode": "survival",
    "difficulty": 2,
    "view_distance": 8
  },
  "protocol": {
    "version": "1.21.50",
    "auto_negotiate": true,
    "strict_version_check": false
  },
  "performance": {
    "max_entities": 100,
    "max_chunks": 100,
    "enable_physics": true,
    "enable_redstone": false,
    "enable_mobs": true,
    "enable_weather": true,
    "compression_level": 1
  },
  "debug": true
}
EOF
            log_success "Created basic config.json"
        fi
    else
        log_info "config.json already exists"
    fi
    
    # Environment file
    if [ ! -f ".env" ]; then
        cat > .env << EOF
# Garuda Server Environment Configuration
GARUDA_DEBUG=true
GARUDA_DATA_DIR=./data
GARUDA_LOG_LEVEL=info
GARUDA_MAX_MEMORY=2G
EOF
        log_success "Created .env file"
    fi
}

# Setup permissions
setup_permissions() {
    log_info "Setting up permissions..."
    
    # Make scripts executable
    chmod +x scripts/*.sh
    
    # Make bin directory writable
    if [ -d "bin" ]; then
        chmod 755 bin
    fi
    
    # Setup log permissions
    if [ -d "logs" ]; then
        chmod 755 logs
        touch logs/server.log
        chmod 644 logs/server.log
    fi
    
    log_success "Permissions configured"
}

# Platform-specific optimizations
platform_optimizations() {
    log_info "Applying platform-specific optimizations..."
    
    case $PLATFORM in
        Termux)
            log_info "Applying Termux optimizations..."
            # Increase file limits
            ulimit -n 8192 2>/dev/null || true
            # Set TMPDIR if not set
            export TMPDIR=$HOME/tmp
            mkdir -p $TMPDIR
            ;;
        Docker)
            log_info "Applying Docker optimizations..."
            # Ensure proper signal handling
            export GOTRACEBACK=all
            ;;
        Linux)
            log_info "Applying Linux optimizations..."
            # Increase file limits if possible
            ulimit -n 8192 2>/dev/null || true
            ;;
    esac
    
    log_success "Platform optimizations applied"
}

# Generate version info
generate_version_info() {
    log_info "Generating version information..."
    
    cat > version.txt << EOF
Garuda Minecraft Server Framework
Version: 1.0.0
Build Date: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
Platform: $PLATFORM
Go Version: $GO_VERSION
Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
Supported MC: 1.21.10 - 1.21.123
EOF

    log_success "Version information generated"
}

# Run health check
health_check() {
    log_info "Running health check..."
    
    local errors=0
    
    # Check if we can build
    if go build -o bin/garuda-test ./cmd/server 2>/dev/null; then
        log_success "Build test: PASSED"
        rm -f bin/garuda-test
    else
        log_error "Build test: FAILED"
        ((errors++))
    fi
    
    # Check config
    if [ -f "config.json" ]; then
        if python3 -m json.tool config.json >/dev/null 2>&1 || jq . config.json >/dev/null 2>&1; then
            log_success "Config validation: PASSED"
        else
            log_error "Config validation: FAILED - config.json may be invalid"
            ((errors++))
        fi
    fi
    
    # Check directory structure
    local required_dirs=("bin" "logs")
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_success "Directory $dir: EXISTS"
        else
            log_error "Directory $dir: MISSING"
            ((errors++))
        fi
    done
    
    if [ $errors -eq 0 ]; then
        log_success "Health check: ALL TESTS PASSED"
    else
        log_error "Health check: $errors TEST(S) FAILED"
        return 1
    fi
}

# Main setup function
main() {
    print_banner
    
    log_info "Starting Garuda Framework setup..."
    log_info "Platform: $(uname -s) $(uname -m)"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --debug)
                DEBUG="true"
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --debug    Enable debug output"
                echo "  --help     Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Run setup steps
    check_prerequisites
    setup_directories
    download_dependencies
    create_configs
    setup_permissions
    platform_optimizations
    generate_version_info
    health_check
    
    log_success "Garuda Framework setup completed successfully!"
    log_info ""
    log_info "Next steps:"
    log_info "  ./scripts/build.sh    - Build the server"
    log_info "  ./scripts/run.sh      - Start the server"
    log_info "  ./scripts/dev.sh      - Start in development mode"
    log_info ""
    log_info "Configuration file: config.json"
    log_info "Logs directory: logs/"
    log_info "Plugins directory: plugins/"
}

# Run main function
main "$@"