#!/bin/bash
set -euo pipefail

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "           LIBRARY LOADING TEST (Daemon-First)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

RUN_ID="$(date +%s)-$$"
DB_PATH="${TMPDIR:-/tmp}/test-lib-loading-${RUN_ID}.db"
WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/test-lib-worktree-${RUN_ID}-XXXX")"
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
./vm-system serve --db "${DB_PATH}" --listen "${SERVER_ADDR}" >/tmp/vm-system-lib-loading.log 2>&1 &
DAEMON_PID=$!
sleep 1
curl -sS "${SERVER_URL}/api/v1/health" >/dev/null

echo "[INFO] Creating test script..."
cat > "${WORKTREE}/test-lodash.js" <<'JS'
if (typeof _ === 'undefined') {
  throw new Error('Lodash library not available');
}
const doubled = _.map([1, 2, 3], n => n * 2);
console.log("LODASH_OK", JSON.stringify(doubled));
"SCRIPT_OK";
JS

echo ""
echo "[TEST 1] Creating template..."
CREATE_OUTPUT=$(${CLI} template create --name "test-template-lib-loading" --engine goja)
TEMPLATE_ID=$(echo "${CREATE_OUTPUT}" | sed -n 's/.*(ID: \(.*\)).*/\1/p')
echo "${CREATE_OUTPUT}"
[[ -n "${TEMPLATE_ID}" ]]
echo "✓ Created template: ${TEMPLATE_ID}"

echo ""
echo "[TEST 2] Downloading libraries to cache..."
${CLI} libs download
LODASH_LIB_ID="$(basename "$(ls .vm-cache/libraries/lodash*.js | head -n 1)" .js)"
if [[ -z "${LODASH_LIB_ID}" ]]; then
  echo "[FAIL] Could not resolve downloaded Lodash library ID"
  exit 1
fi
echo "✓ Libraries downloaded"

echo ""
echo "[TEST 3] Adding console module + Lodash library to template..."
${CLI} template add-module "${TEMPLATE_ID}" --name console
${CLI} template add-library "${TEMPLATE_ID}" --name "${LODASH_LIB_ID}"
echo "✓ Template configured"

echo ""
echo "[TEST 4] Verifying template library configuration..."
${CLI} template get "${TEMPLATE_ID}" | grep -A 5 "Loaded Libraries" | grep -q "lodash"
echo "✓ Lodash configuration verified"

echo ""
echo "[TEST 5] Creating session..."
SESSION_OUTPUT=$(${CLI} session create \
  --template-id "${TEMPLATE_ID}" \
  --workspace-id "ws-lib-loading" \
  --base-commit "deadbeef" \
  --worktree-path "${WORKTREE}")
SESSION_ID=$(echo "${SESSION_OUTPUT}" | awk '/Created session:/ {print $3}')
echo "${SESSION_OUTPUT}"
[[ -n "${SESSION_ID}" ]]
echo "✓ Session created: ${SESSION_ID}"

echo ""
echo "[TEST 6] Executing Lodash test run-file..."
EXEC_OUTPUT=$(${CLI} exec run-file "${SESSION_ID}" "test-lodash.js")
echo "${EXEC_OUTPUT}"
echo "${EXEC_OUTPUT}" | grep -q "LODASH_OK"
echo "✓ Lodash executed successfully in goja runtime"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                      TEST SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✓ Template created via daemon API"
echo "✓ Libraries downloaded and template configured"
echo "✓ Session created via daemon API"
echo "✓ run-file executed with Lodash loaded"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "           ALL TESTS PASSED! ✓"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
