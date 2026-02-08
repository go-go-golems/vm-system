#!/bin/bash
set -euo pipefail

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

extract_template_id() {
  echo "$1" | sed -n 's/.*(ID: \(.*\)).*/\1/p'
}

RUN_ID="$(date +%s)-$$"
DB_PATH="${TMPDIR:-/tmp}/test-library-req-${RUN_ID}.db"
WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/test-workspace-lib-${RUN_ID}-XXXX")"
SERVER_PORT="$(
python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"
SERVER_ADDR="127.0.0.1:${SERVER_PORT}"
SERVER_URL="http://${SERVER_ADDR}"
CLI="./vm-system --db ${DB_PATH} --server-url ${SERVER_URL}"
DAEMON_PID=""

cleanup() {
  if [[ -n "${DAEMON_PID}" ]]; then
    kill "${DAEMON_PID}" >/dev/null 2>&1 || true
    wait "${DAEMON_PID}" >/dev/null 2>&1 || true
  fi
  rm -f "${DB_PATH}" >/dev/null 2>&1 || true
  rm -rf "${WORKTREE}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

log_info "Testing Library Configuration Requirements (Daemon-First)"
log_info "Starting daemon..."
./vm-system serve --db "${DB_PATH}" --listen "${SERVER_ADDR}" >/tmp/vm-system-library-requirements.log 2>&1 &
DAEMON_PID=$!
sleep 1
curl -sS "${SERVER_URL}/api/v1/health" >/dev/null

# ============================================================================
# TEST 1: Download libraries first
# ============================================================================
run_test "Download all libraries"
if ${CLI} libs download; then
  LODASH_LIB_ID="$(basename "$(ls .vm-cache/libraries/lodash*.js | head -n 1)" .js)"
  RAMDA_LIB_ID="$(basename "$(ls .vm-cache/libraries/ramda*.js | head -n 1)" .js)"
  ZUSTAND_LIB_ID="$(basename "$(ls .vm-cache/libraries/zustand*.js | head -n 1)" .js)"
  if [[ -z "${LODASH_LIB_ID}" || -z "${RAMDA_LIB_ID}" || -z "${ZUSTAND_LIB_ID}" ]]; then
    log_error "Failed to resolve downloaded library IDs from cache"
    exit 1
  fi
  log_success "Libraries downloaded"
else
  log_error "Failed to download libraries"
  exit 1
fi

# ============================================================================
# TEST 2: Create template WITHOUT libraries configured
# ============================================================================
run_test "Create template without library configuration"
OUTPUT=$(${CLI} http template create --name "NoLibsTemplate" --engine goja)
TEMPLATE_NO_LIBS=$(extract_template_id "${OUTPUT}")
if [[ -n "${TEMPLATE_NO_LIBS}" ]]; then
  log_success "Template created without libraries"
  log_info "Template ID: ${TEMPLATE_NO_LIBS}"
else
  log_error "Failed to create template"
  exit 1
fi

${CLI} http template add-module "${TEMPLATE_NO_LIBS}" --name console

# ============================================================================
# TEST 3: Create template WITH Lodash library configured
# ============================================================================
run_test "Create template with Lodash library"
OUTPUT=$(${CLI} http template create --name "LodashTemplate" --engine goja)
TEMPLATE_LODASH=$(extract_template_id "${OUTPUT}")
if [[ -n "${TEMPLATE_LODASH}" ]]; then
  log_success "Template created for Lodash"
  log_info "Template ID: ${TEMPLATE_LODASH}"
else
  log_error "Failed to create template"
  exit 1
fi

${CLI} http template add-module "${TEMPLATE_LODASH}" --name console
${CLI} http template add-library "${TEMPLATE_LODASH}" --name "${LODASH_LIB_ID}"

# ============================================================================
# TEST 4: Verify library is in template configuration
# ============================================================================
run_test "Verify Lodash is in template configuration"
OUTPUT=$(${CLI} http template get "${TEMPLATE_LODASH}")
if echo "${OUTPUT}" | grep -A 5 "Loaded Libraries" | grep -q "lodash"; then
  log_success "Template reports Lodash in Loaded Libraries"
else
  log_error "Template missing Lodash in Loaded Libraries"
fi

# ============================================================================
# TEST 5: Create test code that requires Lodash
# ============================================================================
run_test "Create test code that requires Lodash"
cat > "${WORKTREE}/test-lodash.js" <<'JS'
if (typeof _ === 'undefined') {
  throw new Error('Lodash library not available');
}
const users = [
  { name: 'Alice', age: 30 },
  { name: 'Bob', age: 25 },
  { name: 'Charlie', age: 35 }
];
const names = _.map(users, 'name');
console.log("SUCCESS_LODASH", JSON.stringify(names));
JS
log_success "Lodash test code created"

# ============================================================================
# TEST 6: Verify code fails without library configured
# ============================================================================
run_test "Verify code fails on template without Lodash"
SESSION_NO_LIBS=$(${CLI} http session create \
  --template-id "${TEMPLATE_NO_LIBS}" \
  --workspace-id "ws-no-libs" \
  --base-commit "deadbeef" \
  --worktree-path "${WORKTREE}" | awk '/Created session:/ {print $3}')

if ${CLI} http exec run-file "${SESSION_NO_LIBS}" "test-lodash.js" >/tmp/test-no-libs.out 2>&1; then
  if grep -q "Error" /tmp/test-no-libs.out; then
    log_success "Execution failed as expected without Lodash"
  else
    log_error "Execution unexpectedly succeeded without Lodash"
  fi
else
  log_success "Execution command returned non-zero without Lodash as expected"
fi

# ============================================================================
# TEST 7: Verify code succeeds with library configured
# ============================================================================
run_test "Verify code succeeds on template with Lodash"
SESSION_LODASH=$(${CLI} http session create \
  --template-id "${TEMPLATE_LODASH}" \
  --workspace-id "ws-lodash" \
  --base-commit "deadbeef" \
  --worktree-path "${WORKTREE}" | awk '/Created session:/ {print $3}')

OUTPUT=$(${CLI} http exec run-file "${SESSION_LODASH}" "test-lodash.js")
if echo "${OUTPUT}" | grep -q "SUCCESS_LODASH"; then
  log_success "Execution succeeded with Lodash configured"
else
  log_error "Execution output missing Lodash success marker"
fi

# ============================================================================
# TEST 8: Add library to existing template (post-hoc)
# ============================================================================
run_test "Add library to existing template (post-hoc configuration)"
OUTPUT=$(${CLI} http template create --name "PostHocTemplate" --engine goja)
TEMPLATE_POSTHOC=$(extract_template_id "${OUTPUT}")
${CLI} http template add-module "${TEMPLATE_POSTHOC}" --name console

if ${CLI} http template add-library "${TEMPLATE_POSTHOC}" --name "${LODASH_LIB_ID}"; then
  log_success "Library added post-hoc successfully"
else
  log_error "Failed to add library post-hoc"
fi

OUTPUT=$(${CLI} http template get "${TEMPLATE_POSTHOC}")
if echo "${OUTPUT}" | grep -A 5 "Loaded Libraries" | grep -q "lodash"; then
  log_success "Post-hoc library configuration persisted"
else
  log_error "Post-hoc library configuration not found"
fi

# ============================================================================
# TEST 9: Add multiple libraries to one template
# ============================================================================
run_test "Add multiple libraries to template"
OUTPUT=$(${CLI} http template create --name "MultiLibTemplate" --engine goja)
TEMPLATE_MULTI=$(extract_template_id "${OUTPUT}")

${CLI} http template add-module "${TEMPLATE_MULTI}" --name console
${CLI} http template add-library "${TEMPLATE_MULTI}" --name "${LODASH_LIB_ID}"
${CLI} http template add-library "${TEMPLATE_MULTI}" --name "${RAMDA_LIB_ID}"
${CLI} http template add-library "${TEMPLATE_MULTI}" --name "${ZUSTAND_LIB_ID}"

OUTPUT=$(${CLI} http template get "${TEMPLATE_MULTI}")
if echo "${OUTPUT}" | grep -q "lodash" && echo "${OUTPUT}" | grep -q "ramda" && echo "${OUTPUT}" | grep -q "zustand"; then
  log_success "Multiple libraries configured successfully"
else
  log_error "Not all libraries found in configuration"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                      TEST SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Tests Run:    ${TESTS_RUN}"
echo -e "${GREEN}Tests Passed: ${TESTS_PASSED}${NC}"
if [ "${TESTS_FAILED}" -gt 0 ]; then
  echo -e "${RED}Tests Failed: ${TESTS_FAILED}${NC}"
else
  echo "Tests Failed: ${TESTS_FAILED}"
fi
echo ""

EFFECTIVE_PASSED=${TESTS_PASSED}
if [ "${EFFECTIVE_PASSED}" -gt "${TESTS_RUN}" ]; then
  EFFECTIVE_PASSED=${TESTS_RUN}
fi
SUCCESS_RATE=$((EFFECTIVE_PASSED * 100 / TESTS_RUN))
echo "Success Rate: ${SUCCESS_RATE}%"
echo ""

echo "KEY FINDINGS:"
echo "1. Libraries can be configured post-hoc (after template creation)"
echo "2. Multiple libraries can be added to the same template"
echo "3. Library configuration persists across daemon/API operations"
echo "4. Runtime execution behavior now validated for with/without library configuration"
echo ""

if [ "${TESTS_FAILED}" -eq 0 ]; then
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
