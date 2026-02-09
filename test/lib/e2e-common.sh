#!/bin/bash
# Shared helpers for daemon-first shell integration tests.

vm_common_alloc_port() {
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
}

vm_common_init_run() {
  local run_name="$1"
  local run_id
  run_id="$(date +%s)-$$"

  DB_PATH="${TMPDIR:-/tmp}/${run_name}-${run_id}.db"
  WORKTREE="$(mktemp -d "${TMPDIR:-/tmp}/${run_name}-workspace-${run_id}-XXXX")"
  SERVER_PORT="$(vm_common_alloc_port)"
  SERVER_ADDR="127.0.0.1:${SERVER_PORT}"
  SERVER_URL="http://${SERVER_ADDR}"
  CLI="./vm-system --server-url ${SERVER_URL} --db ${DB_PATH}"
  DAEMON_PID=""

  # Allow callers to override where daemon logs are written.
  DAEMON_LOG_PATH="${TMPDIR:-/tmp}/vm-system-${run_name}.log"
}

vm_common_cleanup() {
  if [[ -n "${DAEMON_PID:-}" ]]; then
    kill "${DAEMON_PID}" >/dev/null 2>&1 || true
    wait "${DAEMON_PID}" >/dev/null 2>&1 || true
  fi
  rm -f "${DB_PATH:-}" >/dev/null 2>&1 || true
  rm -rf "${WORKTREE:-}" >/dev/null 2>&1 || true
}

vm_common_trap_cleanup() {
  trap vm_common_cleanup EXIT
}

vm_common_build_binary() {
  GOWORK=off go build -o vm-system ./cmd/vm-system
}

vm_common_start_daemon() {
  local log_path="${1:-${DAEMON_LOG_PATH}}"
  ./vm-system serve --db "${DB_PATH}" --listen "${SERVER_ADDR}" >"${log_path}" 2>&1 &
  DAEMON_PID=$!
}

vm_common_wait_for_health() {
  local attempts="${1:-40}"
  local sleep_seconds="${2:-0.25}"
  local i

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "${SERVER_URL}/api/v1/health" | grep -q '"status":"ok"'; then
      return 0
    fi
    sleep "${sleep_seconds}"
  done

  return 1
}

vm_common_extract_template_id() {
  echo "$1" | sed -n 's/.*(ID: \(.*\)).*/\1/p'
}

vm_common_extract_session_id() {
  echo "$1" | awk '/Created session:/ {print $3}'
}
