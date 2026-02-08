#!/bin/bash
set -euo pipefail

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "      GOJA LIBRARY EXECUTION TEST (Daemon-First End-to-End)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

RUN_ID="$(date +%s)-$$"
DB_PATH="${TMPDIR:-/tmp}/test-goja-exec-${RUN_ID}.db"
WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/test-goja-workspace-${RUN_ID}-XXXX")"
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
CLI="./vm-system --server-url ${SERVER_URL} --db ${DB_PATH}"
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

echo "[INFO] Starting daemon..."
./vm-system serve --db "${DB_PATH}" --listen "${SERVER_ADDR}" >/tmp/vm-system-goja-library.log 2>&1 &
DAEMON_PID=$!
sleep 1
curl -sS "${SERVER_URL}/api/v1/health" >/dev/null

echo "[INFO] Creating test workspace with git repo..."
(
  cd "${WORKTREE}"
  git init
  git config user.email "test@example.com"
  git config user.name "Test User"

  cat > test-lodash.js <<'JS'
if (typeof _ === 'undefined') {
  console.log("FAIL: Lodash is not loaded!");
  throw new Error("Lodash not available");
}

console.log("SUCCESS: Lodash is loaded!");

const numbers = [1, 2, 3, 4, 5];
const doubled = _.map(numbers, n => n * 2);
console.log("Original numbers:", JSON.stringify(numbers));
console.log("Doubled numbers:", JSON.stringify(doubled));

const chunked = _.chunk([1, 2, 3, 4, 5, 6], 2);
console.log("Chunked array:", JSON.stringify(chunked));

const users = [
  { name: 'Alice', age: 30 },
  { name: 'Bob', age: 25 },
  { name: 'Charlie', age: 35 }
];
const sorted = _.sortBy(users, 'age');
console.log("Sorted users:", JSON.stringify(sorted));

console.log("All Lodash tests passed!");
JS

  git add test-lodash.js
  git commit -m "Add Lodash test"
)

BASE_COMMIT=$(git -C "${WORKTREE}" rev-parse HEAD)
echo "✓ Test workspace created with commit: ${BASE_COMMIT}"

# Test 1: Download libraries
echo ""
echo "[TEST 1] Downloading libraries..."
${CLI} libs download
LODASH_LIB_ID="$(basename "$(ls .vm-cache/libraries/lodash*.js | head -n 1)" .js)"
if [[ -z "${LODASH_LIB_ID}" ]]; then
  echo "[FAIL] Could not resolve downloaded Lodash library ID"
  exit 1
fi
echo "✓ Libraries downloaded"

# Test 2: Create template with Lodash
echo ""
echo "[TEST 2] Creating template with Lodash library..."
TIMESTAMP=$(date +%s)
CREATE_OUTPUT=$(${CLI} template create --name "goja-test-${TIMESTAMP}" --engine goja)
TEMPLATE_ID=$(echo "${CREATE_OUTPUT}" | sed -n 's/.*(ID: \(.*\)).*/\1/p')
echo "${CREATE_OUTPUT}"
echo "✓ Created template: ${TEMPLATE_ID}"

# Test 3: Add Lodash to template
echo ""
echo "[TEST 3] Adding Lodash library to template..."
${CLI} modules add-library --vm-id "${TEMPLATE_ID}" --library-id "${LODASH_LIB_ID}"
${CLI} modules add-module --vm-id "${TEMPLATE_ID}" --module-id console
echo "✓ Lodash configured"

# Test 4: Verify configuration
echo ""
echo "[TEST 4] Verifying template configuration..."
${CLI} template get "${TEMPLATE_ID}" | grep -A 5 "Loaded Libraries" | grep -q "lodash"
echo "✓ Configuration verified"

# Test 5: Create session (loads libraries into runtime)
echo ""
echo "[TEST 5] Creating session (loading libraries into goja runtime)..."
SESSION_OUTPUT=$(${CLI} session create \
  --template-id "${TEMPLATE_ID}" \
  --workspace-id "test-workspace" \
  --base-commit "${BASE_COMMIT}" \
  --worktree-path "${WORKTREE}")
SESSION_ID=$(echo "${SESSION_OUTPUT}" | awk '/Created session:/ {print $3}')
echo "${SESSION_OUTPUT}"

echo ""
echo "[TEST 6] Executing JavaScript code that uses Lodash..."
if [[ -n "${SESSION_ID}" ]]; then
  EXEC_OUTPUT=$(${CLI} exec run-file "${SESSION_ID}" "test-lodash.js")
  echo "Execution output:"
  echo "${EXEC_OUTPUT}"
  echo ""

  if echo "${EXEC_OUTPUT}" | grep -q "All Lodash tests passed!"; then
    echo "✓✓✓ SUCCESS! Lodash is working in goja runtime! ✓✓✓"
  else
    echo "⚠ Execution completed but didn't find success message"
    exit 1
  fi
else
  echo "[FAIL] No session ID available, skipping execution test"
  exit 1
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                      TEST SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "KEY FINDINGS:"
echo "1. ✓ Libraries downloaded to cache"
echo "2. ✓ Template created and configured with Lodash"
echo "3. ✓ Session created via daemon API"
echo "4. ✓ JavaScript code executed with Lodash functions"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "           END-TO-END TEST COMPLETE! ✓"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
