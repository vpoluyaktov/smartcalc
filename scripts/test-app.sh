#!/bin/bash

# SmartCalc Test Runner
# Runs all tests: Go backend, and Playwright E2E tests
# Usage: ./scripts/test-app.sh [--skip-e2e] [--verbose]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Options
SKIP_E2E=false
VERBOSE=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --skip-e2e)
            SKIP_E2E=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--skip-e2e] [--verbose|-v]"
            echo ""
            echo "Options:"
            echo "  --skip-e2e    Skip Playwright E2E tests"
            echo "  --verbose,-v  Show verbose output"
            echo "  --help,-h     Show this help message"
            exit 0
            ;;
    esac
done

# Results tracking
GO_TESTS_PASSED=0
GO_TESTS_FAILED=0
E2E_TESTS_PASSED=0
E2E_TESTS_FAILED=0
E2E_TESTS_SKIPPED=0

# Print section header
print_header() {
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Print success message
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Print error message
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Print warning message
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Print info message
print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Run Go backend tests
run_go_tests() {
    print_header "Running Go Backend Tests"
    
    cd "$PROJECT_ROOT"
    
    if $VERBOSE; then
        if go test ./... -v -count=1; then
            GO_TESTS_PASSED=1
            print_success "All Go tests passed"
        else
            GO_TESTS_FAILED=1
            print_error "Some Go tests failed"
        fi
    else
        # Capture output and show summary
        local output
        if output=$(go test ./... -count=1 2>&1); then
            GO_TESTS_PASSED=1
            # Count packages tested
            local pkg_count=$(echo "$output" | grep -c "^ok" || true)
            print_success "All Go tests passed ($pkg_count packages)"
            
            # Show package summary
            echo "$output" | grep "^ok\|^---\|^FAIL" | head -20
        else
            GO_TESTS_FAILED=1
            print_error "Some Go tests failed"
            echo "$output"
        fi
    fi
}

# Run Playwright E2E tests
run_e2e_tests() {
    if $SKIP_E2E; then
        print_header "Playwright E2E Tests (Skipped)"
        print_warning "E2E tests skipped via --skip-e2e flag"
        E2E_TESTS_SKIPPED=1
        return
    fi
    
    print_header "Running Playwright E2E Tests"
    
    cd "$PROJECT_ROOT"
    
    # Check if node_modules exists in frontend
    if [ ! -d "frontend/node_modules" ]; then
        print_info "Installing frontend dependencies..."
        cd frontend && npm install && cd ..
    fi
    
    # Check if Playwright is installed
    if [ ! -d "frontend/node_modules/@playwright" ]; then
        print_info "Installing Playwright..."
        cd frontend && npm install -D @playwright/test && npx playwright install chromium && cd ..
    fi
    
    # Check if wails is available
    if ! command -v wails &> /dev/null; then
        # Try to find wails in go/bin
        export PATH="$PATH:$HOME/go/bin"
        if ! command -v wails &> /dev/null; then
            print_error "Wails not found. Please install wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
            E2E_TESTS_FAILED=1
            return
        fi
    fi
    
    # Check if we have a display, if not use xvfb
    if [ -z "$DISPLAY" ]; then
        if ! command -v xvfb-run &> /dev/null; then
            print_error "No display available and xvfb not installed"
            print_info "Install xvfb: sudo apt install xvfb"
            print_info "Or run with a display (X11/Wayland)"
            E2E_TESTS_FAILED=1
            return
        fi
        print_info "No display detected, using xvfb for headless mode..."
        WAILS_CMD="xvfb-run -a wails dev"
    else
        WAILS_CMD="wails dev"
    fi
    
    # Start Wails dev server in background
    print_info "Starting Wails dev server..."
    $WAILS_CMD > /tmp/wails-dev.log 2>&1 &
    WAILS_PID=$!
    
    # Wait for Wails dev server to be ready
    print_info "Waiting for Wails dev server to start (port 34115)..."
    local max_wait=60
    local waited=0
    while ! curl -s http://localhost:34115 > /dev/null 2>&1; do
        sleep 1
        waited=$((waited + 1))
        
        # Check for webkit2gtk error early
        if grep -q "webkit2gtk-4.0" /tmp/wails-dev.log 2>/dev/null; then
            print_error "Missing webkit2gtk-4.0 library"
            echo ""
            print_info "On Ubuntu 24.04, webkit2gtk-4.0 is not available."
            print_info "Options:"
            print_info "  1. Use Ubuntu 22.04 or earlier"
            print_info "  2. Install webkit2gtk-4.0 from a PPA"
            print_info "  3. Skip E2E tests with --skip-e2e"
            echo ""
            kill $WAILS_PID 2>/dev/null || true
            E2E_TESTS_FAILED=1
            return
        fi
        
        if [ $waited -ge $max_wait ]; then
            print_error "Wails dev server failed to start within ${max_wait}s"
            print_info "Check /tmp/wails-dev.log for details"
            kill $WAILS_PID 2>/dev/null || true
            E2E_TESTS_FAILED=1
            return
        fi
        # Check if wails process is still running
        if ! kill -0 $WAILS_PID 2>/dev/null; then
            print_error "Wails dev server process died"
            print_info "Log output:"
            tail -20 /tmp/wails-dev.log
            E2E_TESTS_FAILED=1
            return
        fi
    done
    print_success "Wails dev server started (PID: $WAILS_PID)"
    
    # Run Playwright tests with timeout
    cd frontend
    echo ""
    print_info "Running Playwright tests..."
    
    # Use timeout to prevent hanging (2 minutes max for 37 tests)
    local playwright_timeout=120
    
    # Use both list (for console output) and html (for report) reporters
    if timeout $playwright_timeout npx playwright test --reporter=list,html 2>&1; then
        E2E_TESTS_PASSED=1
        print_success "All E2E tests passed"
        print_info "View report: cd frontend && npx playwright show-report"
    else
        local exit_code=$?
        E2E_TESTS_FAILED=1
        if [ $exit_code -eq 124 ]; then
            print_error "Playwright tests timed out after ${playwright_timeout}s"
        else
            print_error "Some E2E tests failed"
            print_info "View detailed report: cd frontend && npx playwright show-report"
        fi
    fi
    
    # Stop Wails dev server and all related processes
    print_info "Stopping Wails dev server..."
    # Kill the main process
    kill $WAILS_PID 2>/dev/null || true
    # Kill child processes
    pkill -P $WAILS_PID 2>/dev/null || true
    # Kill any remaining wails/xvfb processes from this session
    pkill -f "wails dev" 2>/dev/null || true
    pkill -f "Xvfb" 2>/dev/null || true
    # Wait briefly for cleanup
    sleep 1
    
    cd "$PROJECT_ROOT"
}

# Print final summary
print_summary() {
    print_header "Test Summary"
    
    local total_passed=0
    local total_failed=0
    local total_skipped=0
    
    # Go tests
    if [ $GO_TESTS_PASSED -eq 1 ]; then
        print_success "Go Backend Tests: PASSED"
        total_passed=$((total_passed + 1))
    elif [ $GO_TESTS_FAILED -eq 1 ]; then
        print_error "Go Backend Tests: FAILED"
        total_failed=$((total_failed + 1))
    fi
    
    # E2E tests
    if [ $E2E_TESTS_PASSED -eq 1 ]; then
        print_success "Playwright E2E Tests: PASSED"
        total_passed=$((total_passed + 1))
    elif [ $E2E_TESTS_FAILED -eq 1 ]; then
        print_error "Playwright E2E Tests: FAILED"
        total_failed=$((total_failed + 1))
    elif [ $E2E_TESTS_SKIPPED -eq 1 ]; then
        print_warning "Playwright E2E Tests: SKIPPED"
        total_skipped=$((total_skipped + 1))
    fi
    
    echo ""
    echo -e "${BLUE}────────────────────────────────────────────────────────────${NC}"
    
    if [ $total_failed -eq 0 ] && [ $total_passed -gt 0 ]; then
        echo -e "${GREEN}All test suites passed! ✓${NC}"
        if [ $total_skipped -gt 0 ]; then
            echo -e "${YELLOW}($total_skipped suite(s) skipped)${NC}"
        fi
        exit 0
    elif [ $total_failed -gt 0 ]; then
        echo -e "${RED}$total_failed test suite(s) failed ✗${NC}"
        exit 1
    else
        echo -e "${YELLOW}No tests were run${NC}"
        exit 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║           SmartCalc Test Runner                           ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Run all test suites
    run_go_tests
    run_e2e_tests
    
    # Print summary and exit with appropriate code
    print_summary
}

# Run main
main
