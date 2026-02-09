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

vm_common_init_run "vm-system-library-matrix"
vm_common_trap_cleanup

log_info "Starting library/module capability matrix test"
mkdir -p "${WORKTREE}"

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
console.log('SUCCESS_LODASH', JSON.stringify(names));
JS

cat > "${WORKTREE}/test-json-builtin.js" <<'JS'
const payload = { a: 1, b: 2 };
const encoded = JSON.stringify(payload);
if (encoded !== '{"a":1,"b":2}') {
  throw new Error('JSON builtin behavior mismatch: ' + encoded);
}
console.log('SUCCESS_JSON_BUILTIN', encoded);
JS

run_test "Build vm-system binary"
if vm_common_build_binary; then
  log_success "Binary built"
else
  log_error "Build failed"
  exit 1
fi

run_test "Start daemon"
vm_common_start_daemon "${TMPDIR:-/tmp}/vm-system-library-matrix.log"
if vm_common_wait_for_health; then
  log_success "Daemon health endpoint is ready"
else
  log_error "Daemon did not become healthy"
  exit 1
fi

run_test "Download library catalog to local cache"
if ${CLI} libs download; then
  LODASH_LIB_ID="$(basename "$(ls .vm-cache/libraries/lodash*.js | head -n 1)" .js)"
  if [[ -n "${LODASH_LIB_ID}" ]]; then
    log_success "Resolved Lodash library ID: ${LODASH_LIB_ID}"
  else
    log_error "Could not resolve Lodash library ID"
    exit 1
  fi
else
  log_error "Library download failed"
  exit 1
fi

run_test "Create template without libraries"
OUTPUT=$(${CLI} template create --name "NoLibsTemplate" --engine goja)
TEMPLATE_NO_LIBS=$(vm_common_extract_template_id "${OUTPUT}")
if [[ -n "${TEMPLATE_NO_LIBS}" ]]; then
  log_success "Template created without libraries (${TEMPLATE_NO_LIBS})"
else
  log_error "Template creation failed"
  exit 1
fi

run_test "JSON is not template-configurable module"
if ${CLI} template add-module "${TEMPLATE_NO_LIBS}" --name json >/tmp/vm-system-json-module.out 2>&1; then
  log_error "JSON module unexpectedly accepted"
else
  if grep -q "MODULE_NOT_ALLOWED" /tmp/vm-system-json-module.out || grep -q "not allowed" /tmp/vm-system-json-module.out; then
    log_success "JSON module rejected as expected"
  else
    log_error "JSON module failed for unexpected reason"
    cat /tmp/vm-system-json-module.out
  fi
fi

run_test "JSON built-in works without any library"
SESSION_NO_LIBS=$(${CLI} session create \
  --template-id "${TEMPLATE_NO_LIBS}" \
  --workspace-id "ws-no-libs" \
  --base-commit "deadbeef" \
  --worktree-path "${WORKTREE}" | awk '/Created session:/ {print $3}')

OUTPUT=$(${CLI} exec run-file "${SESSION_NO_LIBS}" "test-json-builtin.js")
if echo "${OUTPUT}" | grep -q "SUCCESS_JSON_BUILTIN"; then
  log_success "JSON built-in remained available"
else
  log_error "JSON built-in execution failed"
fi

run_test "Lodash usage fails without lodash configured"
if ${CLI} exec run-file "${SESSION_NO_LIBS}" "test-lodash.js" >/tmp/vm-system-no-lodash.out 2>&1; then
  if grep -q "Lodash library not available" /tmp/vm-system-no-lodash.out || grep -q "Error" /tmp/vm-system-no-lodash.out; then
    log_success "Execution failed as expected without Lodash"
  else
    log_error "Execution unexpectedly succeeded without Lodash"
    cat /tmp/vm-system-no-lodash.out
  fi
else
  log_success "Execution command returned non-zero without Lodash as expected"
fi

run_test "Create template with Lodash library configured"
OUTPUT=$(${CLI} template create --name "LodashTemplate" --engine goja)
TEMPLATE_LODASH=$(vm_common_extract_template_id "${OUTPUT}")
if [[ -z "${TEMPLATE_LODASH}" ]]; then
  log_error "Failed to create Lodash template"
  exit 1
fi
if ${CLI} template add-library "${TEMPLATE_LODASH}" --name "${LODASH_LIB_ID}"; then
  log_success "Lodash configured on template (${TEMPLATE_LODASH})"
else
  log_error "Failed to add Lodash to template"
  exit 1
fi

run_test "Lodash usage succeeds when Lodash configured"
SESSION_LODASH=$(${CLI} session create \
  --template-id "${TEMPLATE_LODASH}" \
  --workspace-id "ws-lodash" \
  --base-commit "deadbeef" \
  --worktree-path "${WORKTREE}" | awk '/Created session:/ {print $3}')

OUTPUT=$(${CLI} exec run-file "${SESSION_LODASH}" "test-lodash.js")
if echo "${OUTPUT}" | grep -q "SUCCESS_LODASH"; then
  log_success "Execution succeeded with Lodash configured"
else
  log_error "Execution output missing Lodash success marker"
fi

run_test "Post-hoc library configuration works"
OUTPUT=$(${CLI} template create --name "PostHocTemplate" --engine goja)
TEMPLATE_POSTHOC=$(vm_common_extract_template_id "${OUTPUT}")
if [[ -z "${TEMPLATE_POSTHOC}" ]]; then
  log_error "Failed to create post-hoc template"
  exit 1
fi
if ${CLI} template add-library "${TEMPLATE_POSTHOC}" --name "${LODASH_LIB_ID}"; then
  log_success "Post-hoc library add succeeded"
else
  log_error "Post-hoc library add failed"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Library matrix summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Tests run:    ${TESTS_RUN}"
echo "Tests passed: ${TESTS_PASSED}"
echo "Tests failed: ${TESTS_FAILED}"

if [[ "${TESTS_FAILED}" -gt 0 ]]; then
  exit 1
fi

log_success "Library/module capability matrix test completed"
