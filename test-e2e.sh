#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test/lib/e2e-common.sh
source "${SCRIPT_DIR}/test/lib/e2e-common.sh"

echo "=== VM System End-to-End Test (Daemon-First) ==="
echo ""

vm_common_init_run "vm-system-e2e"
vm_common_trap_cleanup

echo "1. Creating test workspace..."
mkdir -p "${WORKTREE}/runtime"
cat > "${WORKTREE}/runtime/init.js" <<'JS'
console.log("Startup: init.js loaded")
globalThis.testGlobal = "initialized"
JS

cat > "${WORKTREE}/runtime/bootstrap.js" <<'JS'
console.log("Startup: bootstrap.js loaded")
JS

cat > "${WORKTREE}/test.js" <<'JS'
console.log("Hello from test.js")
const result = testGlobal + "-ok"
console.log("Result:", result)
result
JS

echo "✓ Workspace ready"
echo ""

echo "2. Building vm-system..."
vm_common_build_binary
echo "✓ Build successful"
echo ""

echo "3. Starting daemon..."
vm_common_start_daemon "${TMPDIR:-/tmp}/vm-system-e2e.log"
vm_common_wait_for_health
echo "✓ Daemon started"
echo ""

echo "4. Creating template..."
CREATE_OUTPUT=$(${CLI} template create --name "test-template" --engine goja)
TEMPLATE_ID=$(vm_common_extract_template_id "${CREATE_OUTPUT}")
echo "${CREATE_OUTPUT}"
echo "Template ID: ${TEMPLATE_ID}"
echo ""

echo "5. Adding startup files..."
${CLI} template add-startup "${TEMPLATE_ID}" --path "runtime/init.js" --order 10 --mode eval
${CLI} template add-startup "${TEMPLATE_ID}" --path "runtime/bootstrap.js" --order 20 --mode eval
echo "✓ Startup files added"
echo ""

echo "6. Creating session..."
SESSION_OUTPUT=$(${CLI} session create \
  --template-id "${TEMPLATE_ID}" \
  --workspace-id ws-test-1 \
  --base-commit abc123 \
  --worktree-path "${WORKTREE}")
SESSION_ID=$(vm_common_extract_session_id "${SESSION_OUTPUT}")
echo "${SESSION_OUTPUT}"
echo "Session ID: ${SESSION_ID}"
echo ""

echo "7. Executing REPL snippet..."
${CLI} exec repl "${SESSION_ID}" '1 + 2'
echo ""

echo "8. Executing run-file..."
${CLI} exec run-file "${SESSION_ID}" 'test.js'
echo ""

echo "9. Listing sessions and executions..."
${CLI} session list
${CLI} exec list "${SESSION_ID}"
echo ""

echo "10. Runtime summary..."
curl -sS "${SERVER_URL}/api/v1/runtime/summary"
echo ""

echo "=== All E2E Steps Passed ==="
echo ""
echo "Summary:"
echo "- Template created over daemon API"
echo "- Startup files executed during session creation"
echo "- Session persisted in daemon"
echo "- REPL and run-file executions succeeded via API client mode"
