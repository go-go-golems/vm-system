#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

TEST_SCRIPTS=(
  "smoke-test.sh"
  "test-e2e.sh"
  "test-library-matrix.sh"
)

PASSED=0
FAILED=0

for test_script in "${TEST_SCRIPTS[@]}"; do
  echo ""
  echo "============================================================"
  echo "Running ${test_script}"
  echo "============================================================"

  start_time=$(date +%s)
  if "${SCRIPT_DIR}/${test_script}"; then
    end_time=$(date +%s)
    echo "[PASS] ${test_script} ($((end_time - start_time))s)"
    PASSED=$((PASSED + 1))
  else
    end_time=$(date +%s)
    echo "[FAIL] ${test_script} ($((end_time - start_time))s)"
    FAILED=$((FAILED + 1))
  fi
done

echo ""
echo "============================================================"
echo "Integration script suite summary"
echo "============================================================"
echo "Passed: ${PASSED}"
echo "Failed: ${FAILED}"

if [[ "${FAILED}" -gt 0 ]]; then
  exit 1
fi

echo "All integration scripts passed"
