#!/bin/bash
set -euo pipefail

echo "=== VM System End-to-End Test (Daemon-First) ==="
echo ""

RUN_ID="$(date +%s)-$$"
DB_PATH="${TMPDIR:-/tmp}/vm-system-test-${RUN_ID}.db"
WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/test-workspace-e2e-${RUN_ID}-XXXX")"
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

echo "1. Cleaning previous artifacts..."
:


echo "2. Creating test workspace..."
mkdir -p "$WORKTREE/runtime"
cat > "$WORKTREE/runtime/init.js" <<'JS'
console.log("Startup: init.js loaded")
globalThis.testGlobal = "initialized"
JS

cat > "$WORKTREE/runtime/bootstrap.js" <<'JS'
console.log("Startup: bootstrap.js loaded")
JS

cat > "$WORKTREE/test.js" <<'JS'
console.log("Hello from test.js")
const result = testGlobal + "-ok"
console.log("Result:", result)
result
JS

echo "✓ Workspace ready"
echo ""


echo "3. Building vm-system..."
GOWORK=off go build -o vm-system ./cmd/vm-system
echo "✓ Build successful"
echo ""


echo "4. Starting daemon..."
./vm-system serve --db "$DB_PATH" --listen "$SERVER_ADDR" >/tmp/vm-system-e2e.log 2>&1 &
DAEMON_PID=$!
sleep 1
curl -sS "$SERVER_URL/api/v1/health" >/dev/null
echo "✓ Daemon started"
echo ""


echo "5. Creating template..."
CREATE_OUTPUT=$($CLI template create --name "test-template" --engine goja)
TEMPLATE_ID=$(echo "$CREATE_OUTPUT" | sed -n 's/.*(ID: \(.*\)).*/\1/p')
echo "$CREATE_OUTPUT"
echo "Template ID: $TEMPLATE_ID"
echo ""


echo "6. Adding startup files..."
$CLI template add-startup "$TEMPLATE_ID" --path "runtime/init.js" --order 10 --mode eval
$CLI template add-startup "$TEMPLATE_ID" --path "runtime/bootstrap.js" --order 20 --mode eval
echo "✓ Startup files added"
echo ""


echo "7. Creating session..."
WORKTREE_PATH="$WORKTREE"
SESSION_OUTPUT=$($CLI session create \
  --template-id "$TEMPLATE_ID" \
  --workspace-id ws-test-1 \
  --base-commit abc123 \
  --worktree-path "$WORKTREE_PATH")
SESSION_ID=$(echo "$SESSION_OUTPUT" | awk '/Created session:/ {print $3}')
echo "$SESSION_OUTPUT"
echo "Session ID: $SESSION_ID"
echo ""


echo "8. Executing REPL snippet..."
$CLI exec repl "$SESSION_ID" '1 + 2'
echo ""


echo "9. Executing run-file..."
$CLI exec run-file "$SESSION_ID" 'test.js'
echo ""


echo "10. Listing sessions and executions..."
$CLI session list
$CLI exec list "$SESSION_ID"
echo ""


echo "11. Runtime summary..."
curl -sS "$SERVER_URL/api/v1/runtime/summary"
echo ""


echo "=== All E2E Steps Passed ==="
echo ""
echo "Summary:"
echo "- Template created over daemon API"
echo "- Startup files executed during session creation"
echo "- Session persisted in daemon"
echo "- REPL and run-file executions succeeded via API client mode"
