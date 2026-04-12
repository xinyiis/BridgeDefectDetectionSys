#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=========================================="
echo "  start_backend.sh 已切换为统一入口"
echo "  即将转发到 start_dev.sh"
echo "=========================================="

exec bash "${ROOT_DIR}/start_dev.sh"
