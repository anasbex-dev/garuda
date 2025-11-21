#!/bin/bash

# Garuda Development Server Script
# Copyright © 2025 - AnasBex

set -e

# Colors
CYAN='\033[0;36m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[DEV]${NC} $1"; }
log_success() { echo -e "${GREEN}[DEV]${NC} $1"; }

print_dev_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                  GARUDA DEVELOPMENT MODE                    ║"
    echo "║                  Copyright © 2025 - AnasBex                 ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

main() {
    print_dev_banner
    
    log_info "Starting development environment..."
    
    # Kill any existing server
    pkill -f "garuda" || true
    sleep 1
    
    # Build development version
    log_info "Building development server..."
    if ! ./scripts/build.sh --dev; then
        echo -e "${RED}[DEV] Build failed${NC}"
        exit 1
    fi
    
    # Run with development settings
    log_info "Starting development server..."
    export GARUDA_DEBUG=true
    export GARUDA_LOG_LEVEL=debug
    export GORACE="halt_on_error=1"
    
    ./bin/garuda-dev
}

main "$@"