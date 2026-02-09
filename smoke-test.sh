#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test/lib/e2e-common.sh
source "${SCRIPT_DIR}/test/lib/e2e-common.sh"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

log_info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; TESTS_PASSED=$((TESTS_PASSED + 1)); }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; TESTS_FAILED=$((TESTS_FAILED + 1)); }
run_test() {
  TESTS_RUN=$((TESTS_RUN + 1))
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  log_info "Test ${TESTS_RUN}: $1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

vm_common_init_run "vm-system-smoke"
vm_common_trap_cleanup

log_info "Starting daemon-first VM System smoke tests"
log_info "Preparing workspace and database"
mkdir -p "${WORKTREE}"

cat > "${WORKTREE}/startup.js" <<'JS'
console.log("Startup script loaded");
globalThis.SMOKE_VALUE = 40;
JS

cat > "${WORKTREE}/app.js" <<'JS'
console.log("app.js running");
SMOKE_VALUE + 2;
JS

run_test "Build vm-system binary"
if vm_common_build_binary; then
  log_success "Binary built"
else
  log_error "Build failed"
  exit 1
fi

run_test "Start daemon"
vm_common_start_daemon "${TMPDIR:-/tmp}/vm-system-smoke.log"
if vm_common_wait_for_health; then
  log_success "Daemon health endpoint is ready"
else
  log_error "Daemon did not become healthy"
  exit 1
fi

run_test "List available configurable modules"
OUTPUT=$(${CLI} template list-available-modules)
if [[ -n "${OUTPUT}" ]]; then
  log_success "Available module catalog command returned output"
else
  log_error "Available module catalog command returned empty output"
fi

run_test "Create template"
CREATE_OUTPUT=$(${CLI} template create --name "SmokeTemplate" --engine goja)
TEMPLATE_ID=$(vm_common_extract_template_id "${CREATE_OUTPUT}")
if [[ -n "${TEMPLATE_ID}" ]]; then
  log_success "Template created (${TEMPLATE_ID})"
else
  log_error "Template create output missing ID"
  exit 1
fi

run_test "Add template module"
if ${CLI} template add-module "${TEMPLATE_ID}" --name fs; then
  log_success "Module added"
else
  log_error "Module add failed"
fi

run_test "Add template startup file"
if ${CLI} template add-startup "${TEMPLATE_ID}" --path "startup.js" --order 10 --mode eval; then
  log_success "Startup file added"
else
  log_error "Startup file add failed"
fi

run_test "Create session via daemon"
SESSION_OUTPUT=$(${CLI} session create \
  --template-id "${TEMPLATE_ID}" \
  --workspace-id ws-smoke \
  --base-commit deadbeef \
  --worktree-path "${WORKTREE}")
SESSION_ID=$(vm_common_extract_session_id "${SESSION_OUTPUT}")
if [[ -n "${SESSION_ID}" ]]; then
  log_success "Session created (${SESSION_ID})"
else
  log_error "Session create output missing ID"
  exit 1
fi

run_test "Execute REPL over daemon API"
OUTPUT=$(${CLI} exec repl "${SESSION_ID}" 'SMOKE_VALUE + 2')
if echo "${OUTPUT}" | grep -q '"preview":"42"'; then
  log_success "REPL execution returned expected result"
else
  log_error "REPL execution output did not include expected result"
fi

run_test "Execute run-file over daemon API"
OUTPUT=$(${CLI} exec run-file "${SESSION_ID}" "app.js")
if echo "${OUTPUT}" | grep -q "Execution ID:"; then
  log_success "run-file execution succeeded"
else
  log_error "run-file execution failed"
fi

run_test "Check runtime summary"
OUTPUT=$(curl -sS "${SERVER_URL}/api/v1/runtime/summary")
if echo "${OUTPUT}" | grep -q '"active_sessions":1'; then
  log_success "Runtime summary reports one active session"
else
  log_error "Unexpected runtime summary: ${OUTPUT}"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Smoke test summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Tests run:    ${TESTS_RUN}"
echo "Tests passed: ${TESTS_PASSED}"
echo "Tests failed: ${TESTS_FAILED}"

if [[ "${TESTS_FAILED}" -gt 0 ]]; then
  exit 1
fi

log_success "Daemon-first smoke test completed"
