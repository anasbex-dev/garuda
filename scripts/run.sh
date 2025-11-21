#!/bin/bash

# Garuda Framework Run Script
# Copyright © 2025 - AnasBex

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
BINARY_NAME="garuda"
BUILD_DIR="bin"
LOG_DIR="logs"
CONFIG_FILE="config.json"
MAX_MEMORY="2G"
DEBUG_MODE=false
AUTO_BUILD=true

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Print run banner
print_run_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                   GARUDA SERVER LAUNCHER                    ║"
    echo "║                  Copyright © 2025 - AnasBex                 ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Check if binary exists
check_binary() {
    if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
        log_warn "Binary not found: $BUILD_DIR/$BINARY_NAME"
        
        if [ "$AUTO_BUILD" = "true" ]; then
            log_info "Attempting auto-build..."
            if ./scripts/build.sh; then
                log_success "Auto-build completed successfully"
                return 0
            else
                log_error "Auto-build failed"
                return 1
            fi
        else
            log_error "Please build the server first: ./scripts/build.sh"
            return 1
        fi
    fi
    return 0
}

# Check configuration
check_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        log_info "Run ./scripts/setup.sh to create configuration"
        return 1
    fi
    
    # Validate JSON syntax if jq is available
    if command -v jq &> /dev/null; then
        if jq empty "$CONFIG_FILE" &> /dev/null; then
            log_success "Config validation: PASSED"
        else
            log_error "Config validation: FAILED - Invalid JSON"
            return 1
        fi
    fi
    
    return 0
}

# Setup environment
setup_environment() {
    log_info "Setting up environment..."
    
    # Create log directory
    mkdir -p "$LOG_DIR"
    
    # Set log file
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    export GARUDA_LOG_FILE="$LOG_DIR/server_$timestamp.log"
    
    # Load environment variables
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
        log_success "Environment variables loaded"
    fi
    
    # Set Go environment
    export GOGC=50  # More aggressive garbage collection for servers
    export GOMAXPROCS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2)
    
    log_info "CPU cores: $GOMAXPROCS"
    log_info "Log file: $GARUDA_LOG_FILE"
}

# Check ports
check_ports() {
    local port=$(grep -o '"address":[[:space:]]*"[^"]*' config.json | grep -o '[0-9]*$')
    
    if [ -z "$port" ]; then
        port="19132"  # Default port
    fi
    
    log_info "Checking port $port..."
    
    if command -v netstat &> /dev/null; then
        if netstat -tuln | grep ":$port " &> /dev/null; then
            log_warn "Port $port is already in use"
            return 1
        fi
    elif command -v ss &> /dev/null; then
        if ss -tuln | grep ":$port " &> /dev/null; then
            log_warn "Port $port is already in use"
            return 1
        fi
    fi
    
    log_success "Port $port is available"
    return 0
}

# Setup signal handlers
setup_handlers() {
    trap cleanup EXIT INT TERM
}

# Cleanup function
cleanup() {
    log_info "Shutting down Garuda server..."
    if [ ! -z "$SERVER_PID" ]; then
        kill -TERM "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    log_success "Server shutdown complete"
}

# Run the server
run_server() {
    local binary_path="$BUILD_DIR/$BINARY_NAME"
    local log_file="$GARUDA_LOG_FILE"
    
    log_info "Starting Garuda Minecraft Server..."
    log_info "Binary: $binary_path"
    log_info "Config: $CONFIG_FILE"
    log_info "Log: $log_file"
    
    # Display server info
    echo "┌──────────────────────────────────────────────────────┐"
    echo "│                SERVER STARTUP INFORMATION            │"
    echo "├──────────────────────────────────────────────────────┤"
    echo "│ Start Time:  $(date)"
    echo "│ Log File:    $log_file"
    echo "│ Config File: $CONFIG_FILE"
    echo "│ Binary:      $binary_path"
    echo "└──────────────────────────────────────────────────────┘"
    
    # Run the server
    if [ "$DEBUG_MODE" = "true" ]; then
        log_info "Debug mode enabled"
        "$binary_path" 2>&1 | tee "$log_file" &
    else
        "$binary_path" >> "$log_file" 2>&1 &
    fi
    
    SERVER_PID=$!
    log_success "Server started with PID: $SERVER_PID"
    
    # Wait for server to start
    log_info "Waiting for server to initialize..."
    sleep 3
    
    # Check if server is still running
    if kill -0 "$SERVER_PID" 2>/dev/null; then
        log_success "Server is running successfully"
        log_info "Connect with Minecraft Bedrock Edition to:"
        log_info "  Address: localhost:19132"
        log_info "  Version: 1.21.10 - 1.21.123"
        
        # Tail logs
        log_info "Showing server logs (Ctrl+C to stop):"
        echo "────────────────────────────────────────────────────────"
        tail -f "$log_file" &
        TAIL_PID=$!
        
        # Wait for server process
        wait "$SERVER_PID"
        kill "$TAIL_PID" 2>/dev/null || true
    else
        log_error "Server failed to start"
        log_info "Check log file for details: $log_file"
        exit 1
    fi
}

# Main run function
main() {
    print_run_banner
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --debug)
                DEBUG_MODE=true
                shift
                ;;
            --no-build)
                AUTO_BUILD=false
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --debug     Enable debug output and tee to console"
                echo "  --no-build  Skip auto-build if binary missing"
                echo "  --help      Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Run checks
    check_config || exit 1
    check_binary || exit 1
    check_ports || log_warn "Proceeding anyway..."
    
    # Setup and run
    setup_environment
    setup_handlers
    run_server
}

main "$@"