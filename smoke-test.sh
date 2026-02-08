#!/bin/bash
set -euo pipefail

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
  log_info "Test $TESTS_RUN: $1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

RUN_ID="$(date +%s)-$$"
DB_PATH="${TMPDIR:-/tmp}/test-vm-system-daemon-${RUN_ID}.db"
WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/test-workspace-smoke-${RUN_ID}-XXXX")"
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
SERVER_URL="http://$SERVER_ADDR"
CLI="./vm-system --server-url $SERVER_URL --db $DB_PATH"
DAEMON_PID=""

cleanup() {
  if [[ -n "${DAEMON_PID}" ]]; then
    kill "$DAEMON_PID" >/dev/null 2>&1 || true
    wait "$DAEMON_PID" >/dev/null 2>&1 || true
  fi
  rm -f "$DB_PATH" >/dev/null 2>&1 || true
  rm -rf "$WORKTREE" >/dev/null 2>&1 || true
}
trap cleanup EXIT

log_info "Starting daemon-first VM System smoke tests"
log_info "Preparing workspace and database"
mkdir -p "$WORKTREE"

cat > "$WORKTREE/startup.js" <<'JS'
console.log("Startup script loaded");
globalThis.SMOKE_VALUE = 40;
JS

cat > "$WORKTREE/app.js" <<'JS'
console.log("app.js running");
SMOKE_VALUE + 2;
JS

run_test "Build vm-system binary"
if GOWORK=off go build -o vm-system ./cmd/vm-system; then
  log_success "Binary built"
else
  log_error "Build failed"
  exit 1
fi

run_test "Start daemon"
./vm-system serve --db "$DB_PATH" --listen "$SERVER_ADDR" >/tmp/vm-system-smoke.log 2>&1 &
DAEMON_PID=$!
sleep 1
if curl -sS "$SERVER_URL/api/v1/health" | grep -q '"status":"ok"'; then
  log_success "Daemon health endpoint is ready"
else
  log_error "Daemon did not become healthy"
  exit 1
fi

run_test "List available modules"
OUTPUT=$($CLI modules list-available)
if echo "$OUTPUT" | grep -q "console"; then
  log_success "Module list contains console"
else
  log_error "Module list missing console"
fi

run_test "Create template"
CREATE_OUTPUT=$($CLI template create --name "SmokeTemplate" --engine goja)
TEMPLATE_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*(ID: \(.*\)).*/\1/p')
if [[ -n "$TEMPLATE_ID" ]]; then
  log_success "Template created ($TEMPLATE_ID)"
else
  log_error "Template create output missing ID"
  exit 1
fi

run_test "Add template capability"
if $CLI template add-capability "$TEMPLATE_ID" --kind module --name console --enabled; then
  log_success "Capability added"
else
  log_error "Capability add failed"
fi

run_test "Add template startup file"
REL_STARTUP="$(realpath --relative-to "$WORKTREE" "$WORKTREE/startup.js")"
if $CLI template add-startup "$TEMPLATE_ID" --path "$REL_STARTUP" --order 10 --mode eval; then
  log_success "Startup file added"
else
  log_error "Startup file add failed"
fi

run_test "Create session via daemon"
SESSION_OUTPUT=$($CLI session create \
  --template-id "$TEMPLATE_ID" \
  --workspace-id ws-smoke \
  --base-commit deadbeef \
  --worktree-path "$WORKTREE")
SESSION_ID=$(echo "$SESSION_OUTPUT" | awk '/Created session:/ {print $3}')
if [[ -n "$SESSION_ID" ]]; then
  log_success "Session created ($SESSION_ID)"
else
  log_error "Session create output missing ID"
  exit 1
fi

run_test "Execute REPL over daemon API"
OUTPUT=$($CLI exec repl "$SESSION_ID" 'SMOKE_VALUE + 2')
if echo "$OUTPUT" | grep -q '"preview":"42"'; then
  log_success "REPL execution returned expected result"
else
  log_error "REPL execution output did not include expected result"
fi

run_test "Execute run-file over daemon API"
OUTPUT=$($CLI exec run-file "$SESSION_ID" "app.js")
if echo "$OUTPUT" | grep -q "Execution ID:"; then
  log_success "run-file execution succeeded"
else
  log_error "run-file execution failed"
fi

run_test "Check runtime summary"
OUTPUT=$(curl -sS "$SERVER_URL/api/v1/runtime/summary")
if echo "$OUTPUT" | grep -q '"active_sessions":1'; then
  log_success "Runtime summary reports one active session"
else
  log_error "Unexpected runtime summary: $OUTPUT"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Smoke test summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Tests run:    $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $TESTS_FAILED"

if [[ "$TESTS_FAILED" -gt 0 ]]; then
  exit 1
fi

log_success "Daemon-first smoke test completed"
