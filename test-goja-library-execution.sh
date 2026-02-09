#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
echo "[INFO] test-goja-library-execution.sh is deprecated; delegating to test-library-matrix.sh"
exec "${SCRIPT_DIR}/test-library-matrix.sh"
