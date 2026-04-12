#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${ROOT_DIR}/src/backend"
ENV_FILE="${ROOT_DIR}/.env"
CONFIG_FILE="${BACKEND_DIR}/config.yaml"
BACKEND_HEALTH_URL="http://localhost:8080/api/v1/health"
TUNNEL_CHECK_HOST="localhost"

SSH_PID=""
BACKEND_PID=""

cleanup() {
    local exit_code=$?

    if [ -n "${BACKEND_PID}" ] && kill -0 "${BACKEND_PID}" 2>/dev/null; then
        kill "${BACKEND_PID}" 2>/dev/null || true
        wait "${BACKEND_PID}" 2>/dev/null || true
    fi

    if [ -n "${SSH_PID}" ] && kill -0 "${SSH_PID}" 2>/dev/null; then
        kill "${SSH_PID}" 2>/dev/null || true
        wait "${SSH_PID}" 2>/dev/null || true
    fi

    exit "${exit_code}"
}

wait_for_port() {
    local host=$1
    local port=$2
    local retries=$3

    for ((i = 1; i <= retries; i++)); do
        if bash -lc "exec 3<>/dev/tcp/${host}/${port}" 2>/dev/null; then
            return 0
        fi
        sleep 1
    done

    return 1
}

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

trap cleanup INT TERM EXIT

echo "=========================================="
echo "  启动桥梁缺陷检测系统开发环境"
echo "=========================================="

if [ ! -f "${ENV_FILE}" ]; then
    echo "⚠️  警告：找不到 ${ENV_FILE}，将仅使用当前 shell 环境变量"
else
    set -a
    # shellcheck disable=SC1090
    source "${ENV_FILE}"
    set +a
fi

if [ ! -f "${CONFIG_FILE}" ]; then
    echo "❌ 错误：找不到 ${CONFIG_FILE}"
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "❌ 错误：未安装 Go，请先安装 Go 1.25+"
    exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
    echo "❌ 错误：未安装 curl"
    exit 1
fi

if [ -z "${SSH_HOST:-}" ] || [ -z "${SSH_PORT:-}" ] || [ -z "${SSH_USER:-}" ] || [ -z "${SSH_LOCAL_PORT:-}" ] || [ -z "${SSH_REMOTE_PORT:-}" ]; then
    echo "❌ 错误：缺少 SSH 隧道必要配置（SSH_HOST/SSH_PORT/SSH_USER/SSH_LOCAL_PORT/SSH_REMOTE_PORT）"
    exit 1
fi

SSH_REMOTE_HOST="${SSH_REMOTE_HOST:-127.0.0.1}"
SSH_COMMON_OPTS=(
    -o StrictHostKeyChecking=no
    -o UserKnownHostsFile=/dev/null
    -o GlobalKnownHostsFile=/dev/null
    -o ServerAliveInterval=30
    -o ServerAliveCountMax=3
)

if [ -z "${PYTHON_SERVICE_URL:-}" ]; then
    export PYTHON_SERVICE_URL="http://localhost:${SSH_LOCAL_PORT}"
fi

echo "🔄 正在启动 SSH 隧道连接算法服务器..."
pkill -f "ssh -CNg -L ${SSH_LOCAL_PORT}:${SSH_REMOTE_HOST}:${SSH_REMOTE_PORT}" 2>/dev/null || true

if [ -n "${SSH_PASSWORD:-}" ] && [ "${SSH_PASSWORD}" != "你的密码写在这里" ]; then
    if ! command -v sshpass >/dev/null 2>&1; then
        echo "❌ 错误：检测到 SSH_PASSWORD，但未安装 sshpass。请安装后重试，或改用 SSH key 登录"
        exit 1
    fi
    sshpass -p "${SSH_PASSWORD}" ssh \
        -CNg \
        -L "${SSH_LOCAL_PORT}:${SSH_REMOTE_HOST}:${SSH_REMOTE_PORT}" \
        -p "${SSH_PORT}" \
        "${SSH_COMMON_OPTS[@]}" \
        "${SSH_USER}@${SSH_HOST}" &
else
    ssh \
        -CNg \
        -L "${SSH_LOCAL_PORT}:${SSH_REMOTE_HOST}:${SSH_REMOTE_PORT}" \
        -p "${SSH_PORT}" \
        "${SSH_COMMON_OPTS[@]}" \
        "${SSH_USER}@${SSH_HOST}" &
fi
SSH_PID=$!

if ! wait_for_port "${TUNNEL_CHECK_HOST}" "${SSH_LOCAL_PORT}" 10; then
    echo "❌ 错误：SSH 隧道启动失败，本地端口 ${SSH_LOCAL_PORT} 未就绪"
    exit 1
fi

echo "✓ SSH 隧道已建立: localhost:${SSH_LOCAL_PORT} -> ${SSH_HOST}:${SSH_REMOTE_HOST}:${SSH_REMOTE_PORT}"
echo "✓ 算法服务地址: ${PYTHON_SERVICE_URL}"

echo ""
echo "🚀 正在启动 Go 后端服务..."
cd "${BACKEND_DIR}"
go run cmd/server/main.go &
BACKEND_PID=$!

if ! wait_for_http "${BACKEND_HEALTH_URL}" 20; then
    echo "❌ 错误：后端启动失败，健康检查未通过"
    exit 1
fi

echo "✓ 后端服务已启动: http://localhost:8080"
echo "✓ 健康检查通过: ${BACKEND_HEALTH_URL}"
echo ""
echo "按 Ctrl+C 停止后端和 SSH 隧道"
echo "=========================================="

wait "${BACKEND_PID}"
