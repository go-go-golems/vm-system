#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

run_test() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Test $TESTS_RUN: $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Setup
log_info "Starting VM System Smoke Tests"
log_info "Cleaning up test environment..."

# Clean up any previous test data
rm -f test-vm-system.db
rm -rf test-workspace
mkdir -p test-workspace

# Set test database
export DB_PATH="test-vm-system.db"
CLI="./vm-system --db $DB_PATH"

# ============================================================================
# TEST 1: Build the CLI
# ============================================================================
run_test "Build VM System CLI"
if go build -o vm-system ./cmd/vm-system; then
    log_success "CLI built successfully"
else
    log_error "Failed to build CLI"
    exit 1
fi

# ============================================================================
# TEST 2: List available modules
# ============================================================================
run_test "List available modules"
OUTPUT=$($CLI modules list-available)
if echo "$OUTPUT" | grep -q "console"; then
    log_success "Module list contains 'console'"
else
    log_error "Module list missing 'console'"
fi

if echo "$OUTPUT" | grep -q "math"; then
    log_success "Module list contains 'math'"
else
    log_error "Module list missing 'math'"
fi

# ============================================================================
# TEST 3: List available libraries
# ============================================================================
run_test "List available libraries"
OUTPUT=$($CLI modules list-libraries)
if echo "$OUTPUT" | grep -q "lodash"; then
    log_success "Library list contains 'lodash'"
else
    log_error "Library list missing 'lodash'"
fi

if echo "$OUTPUT" | grep -q "zustand"; then
    log_success "Library list contains 'zustand'"
else
    log_error "Library list missing 'zustand'"
fi

# ============================================================================
# TEST 4: Download libraries
# ============================================================================
run_test "Download all libraries"
if $CLI libs download; then
    log_success "Libraries downloaded successfully"
else
    log_error "Failed to download libraries"
fi

# ============================================================================
# TEST 5: Check library cache
# ============================================================================
run_test "Verify library cache"
OUTPUT=$($CLI libs cache-info)
if echo "$OUTPUT" | grep -q "lodash"; then
    log_success "Lodash cached"
else
    log_error "Lodash not cached"
fi

if echo "$OUTPUT" | grep -q "zustand"; then
    log_success "Zustand cached"
else
    log_error "Zustand not cached"
fi

# ============================================================================
# TEST 6: Create VM profile
# ============================================================================
run_test "Create VM profile"
OUTPUT=$($CLI vm create --name "SmokeTestVM" --engine goja)
if echo "$OUTPUT" | grep -q "Created VM"; then
    log_success "VM profile created"
    VM_ID=$(echo "$OUTPUT" | grep -oP 'ID: \K[a-zA-Z0-9_-]+')
    log_info "VM ID: $VM_ID"
else
    log_error "Failed to create VM profile"
    exit 1
fi

# ============================================================================
# TEST 7: List VMs
# ============================================================================
run_test "List VM profiles"
OUTPUT=$($CLI vm list)
if echo "$OUTPUT" | grep -q "SmokeTestVM"; then
    log_success "VM appears in list"
else
    log_error "VM not found in list"
fi

# ============================================================================
# TEST 8: Get VM details
# ============================================================================
run_test "Get VM details"
OUTPUT=$($CLI vm get "$VM_ID")
if echo "$OUTPUT" | grep -q "SmokeTestVM"; then
    log_success "VM details retrieved"
else
    log_error "Failed to get VM details"
fi

# ============================================================================
# TEST 9: Add modules to VM
# ============================================================================
run_test "Add console module to VM"
if $CLI modules add-module --vm-id "$VM_ID" --module-id console; then
    log_success "Console module added"
else
    log_error "Failed to add console module"
fi

run_test "Add math module to VM"
if $CLI modules add-module --vm-id "$VM_ID" --module-id math; then
    log_success "Math module added"
else
    log_error "Failed to add math module"
fi

run_test "Add json module to VM"
if $CLI modules add-module --vm-id "$VM_ID" --module-id json; then
    log_success "JSON module added"
else
    log_error "Failed to add JSON module"
fi

# ============================================================================
# TEST 10: Add libraries to VM
# ============================================================================
run_test "Add lodash library to VM"
if $CLI modules add-library --vm-id "$VM_ID" --library-id lodash; then
    log_success "Lodash library added"
else
    log_error "Failed to add lodash library"
fi

run_test "Add zustand library to VM"
if $CLI modules add-library --vm-id "$VM_ID" --library-id zustand; then
    log_success "Zustand library added"
else
    log_error "Failed to add zustand library"
fi

# ============================================================================
# TEST 11: Verify VM configuration
# ============================================================================
run_test "Verify VM has modules and libraries"
OUTPUT=$($CLI vm get "$VM_ID")
if echo "$OUTPUT" | grep -q "console"; then
    log_success "Console module in VM config"
else
    log_error "Console module not in VM config"
fi

if echo "$OUTPUT" | grep -q "lodash"; then
    log_success "Lodash library in VM config"
else
    log_error "Lodash library not in VM config"
fi

# ============================================================================
# TEST 12: Add capabilities
# ============================================================================
run_test "Add console capability"
if $CLI vm add-capability "$VM_ID" --kind module --name console; then
    log_success "Console capability added"
else
    log_error "Failed to add console capability"
fi

# ============================================================================
# TEST 13: List capabilities
# ============================================================================
run_test "List VM capabilities"
OUTPUT=$($CLI vm list-capabilities "$VM_ID")
if echo "$OUTPUT" | grep -q "console"; then
    log_success "Capabilities listed correctly"
else
    log_error "Failed to list capabilities"
fi

# ============================================================================
# TEST 14: Create startup file
# ============================================================================
run_test "Create startup file"
cat > test-workspace/startup.js << 'EOF'
console.log("Startup script loaded");
const VERSION = "1.0.0";
EOF

if $CLI vm add-startup-file "$VM_ID" --path test-workspace/startup.js --order 1 --mode eval; then
    log_success "Startup file added"
else
    log_error "Failed to add startup file"
fi

# ============================================================================
# TEST 15: List startup files
# ============================================================================
run_test "List startup files"
OUTPUT=$($CLI vm list-startup-files "$VM_ID")
if echo "$OUTPUT" | grep -q "startup.js"; then
    log_success "Startup files listed correctly"
else
    log_error "Failed to list startup files"
fi

# ============================================================================
# TEST 16: Create session
# ============================================================================
run_test "Create VM session"
OUTPUT=$($CLI session create "$VM_ID" --workspace-id test-workspace)
if echo "$OUTPUT" | grep -q "Created session"; then
    log_success "Session created"
    SESSION_ID=$(echo "$OUTPUT" | grep -oP 'ID: \K[a-zA-Z0-9_-]+')
    log_info "Session ID: $SESSION_ID"
else
    log_error "Failed to create session"
fi

# ============================================================================
# TEST 17: List sessions
# ============================================================================
run_test "List sessions"
OUTPUT=$($CLI session list)
if echo "$OUTPUT" | grep -q "$SESSION_ID"; then
    log_success "Session appears in list"
else
    log_error "Session not found in list"
fi

# ============================================================================
# TEST 18: Test module configuration persistence
# ============================================================================
run_test "Verify module persistence after restart"
# Simulate restart by creating new CLI instance
OUTPUT=$($CLI vm get "$VM_ID")
if echo "$OUTPUT" | grep -q "console" && echo "$OUTPUT" | grep -q "math"; then
    log_success "Modules persisted correctly"
else
    log_error "Modules not persisted"
fi

# ============================================================================
# TEST 19: Test library configuration persistence
# ============================================================================
run_test "Verify library persistence after restart"
OUTPUT=$($CLI vm get "$VM_ID")
if echo "$OUTPUT" | grep -q "lodash" && echo "$OUTPUT" | grep -q "zustand"; then
    log_success "Libraries persisted correctly"
else
    log_error "Libraries not persisted"
fi

# ============================================================================
# TEST 20: Verify library cache files exist
# ============================================================================
run_test "Verify library cache files"
if [ -f ".vm-cache/libraries/lodash-4.17.21.js" ]; then
    log_success "Lodash cache file exists"
else
    log_error "Lodash cache file missing"
fi

if [ -f ".vm-cache/libraries/zustand-4.4.7.js" ]; then
    log_success "Zustand cache file exists"
else
    log_error "Zustand cache file missing"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                      TEST SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Tests Run:    $TESTS_RUN"
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
else
    echo "Tests Failed: $TESTS_FAILED"
fi
echo ""

# Calculate success rate
SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
echo "Success Rate: $SUCCESS_RATE%"
echo ""

# Cleanup
log_info "Cleaning up test environment..."
rm -f test-vm-system.db
rm -rf test-workspace

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}           ALL TESTS PASSED! ✓${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}           SOME TESTS FAILED ✗${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
