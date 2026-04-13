#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${ROOT_DIR}/src/backend"
BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://127.0.0.1:8080/api/v1/health}"

wait_for_http() {
    local url=$1
    local retries=$2

    for ((i = 1; i <= retries; i++)); do
        if curl -fsS "${url}" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done

    return 1
}

echo "=========================================="
echo "  启动桥梁缺陷检测系统后端（同机算法版）"
echo "=========================================="

if ! command -v go >/dev/null 2>&1; then
    echo "❌ 错误：未安装 Go，请先安装 Go 1.25+"
    exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
    echo "❌ 错误：未安装 curl"
    exit 1
fi

export PYTHON_SERVICE_URL="${PYTHON_SERVICE_URL:-http://127.0.0.1:18080}"

echo "✓ 使用算法服务地址: ${PYTHON_SERVICE_URL}"
echo "🚀 正在启动 Go 后端服务..."

cd "${BACKEND_DIR}"
go run cmd/server/main.go &
BACKEND_PID=$!

cleanup() {
    local exit_code=$?
    if [ -n "${BACKEND_PID:-}" ] && kill -0 "${BACKEND_PID}" 2>/dev/null; then
        kill "${BACKEND_PID}" 2>/dev/null || true
        wait "${BACKEND_PID}" 2>/dev/null || true
    fi
    exit "${exit_code}"
}

trap cleanup INT TERM EXIT

if ! wait_for_http "${BACKEND_HEALTH_URL}" 20; then
    echo "❌ 错误：后端启动失败，健康检查未通过"
    exit 1
fi

echo "✓ 后端服务已启动: http://127.0.0.1:8080"
echo "✓ 健康检查通过: ${BACKEND_HEALTH_URL}"
echo "按 Ctrl+C 停止后端"
echo "=========================================="

wait "${BACKEND_PID}"
