#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${ROOT_DIR}/src/backend"
ENV_FILE="${ROOT_DIR}/.env"
CONFIG_FILE="${BACKEND_DIR}/config.yaml"

RUNTIME_DIR="${ROOT_DIR}/.runtime"
LOG_DIR="${ROOT_DIR}/logs"
BACKEND_PID_FILE="${RUNTIME_DIR}/backend.pid"
SSH_PID_FILE="${RUNTIME_DIR}/ssh_tunnel.pid"
BACKEND_LOG_FILE="${LOG_DIR}/backend.log"
SSH_LOG_FILE="${LOG_DIR}/ssh_tunnel.log"

BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://127.0.0.1:8080/api/v1/health}"
BACKEND_STARTUP_RETRIES="${BACKEND_STARTUP_RETRIES:-30}"
ENABLE_SSH_TUNNEL="${ENABLE_SSH_TUNNEL:-1}"
TUNNEL_CHECK_HOST="${TUNNEL_CHECK_HOST:-127.0.0.1}"

mkdir -p "${RUNTIME_DIR}" "${LOG_DIR}"

usage() {
    cat <<'EOF'
用法:
  bash start_root_backend.sh start
  bash start_root_backend.sh stop
  bash start_root_backend.sh restart
  bash start_root_backend.sh status
  bash start_root_backend.sh logs

环境变量:
  ENABLE_SSH_TUNNEL=1|0        是否启动 SSH 隧道（默认 1）
  BACKEND_HEALTH_URL=...       后端健康检查地址
  BACKEND_STARTUP_RETRIES=30   健康检查重试次数（每次间隔 1 秒）

说明:
  1. 默认读取项目根目录 .env。
  2. start 会后台启动，不会阻塞当前终端。
  3. logs 会持续追踪后端日志。
EOF
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

load_env_file() {
    if [ -f "${ENV_FILE}" ]; then
        set -a
        # shellcheck disable=SC1090
        source "${ENV_FILE}"
        set +a
    fi
}

is_pid_running() {
    local pid_file="$1"
    if [ ! -f "${pid_file}" ]; then
        return 1
    fi

    local pid
    pid="$(cat "${pid_file}" 2>/dev/null || true)"
    if [ -z "${pid}" ]; then
        return 1
    fi

    if kill -0 "${pid}" 2>/dev/null; then
        return 0
    fi

    return 1
}

cleanup_pid_file_if_stale() {
    local pid_file="$1"
    if [ -f "${pid_file}" ] && ! is_pid_running "${pid_file}"; then
        rm -f "${pid_file}"
    fi
}

wait_for_port() {
    local host="$1"
    local port="$2"
    local retries="$3"

    for ((i = 1; i <= retries; i++)); do
        if bash -lc "exec 3<>/dev/tcp/${host}/${port}" 2>/dev/null; then
            return 0
        fi
        sleep 1
    done

    return 1
}

wait_for_http() {
    local url="$1"
    local retries="$2"

    for ((i = 1; i <= retries; i++)); do
        if curl -fsS "${url}" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done

    return 1
}

start_ssh_tunnel() {
    if [ "${ENABLE_SSH_TUNNEL}" != "1" ]; then
        if [ -z "${PYTHON_SERVICE_URL:-}" ]; then
            export PYTHON_SERVICE_URL="http://127.0.0.1:18080"
        fi
        echo "✓ 跳过 SSH 隧道，使用算法服务地址: ${PYTHON_SERVICE_URL}"
        return 0
    fi

    if [ -z "${SSH_HOST:-}" ] || [ -z "${SSH_PORT:-}" ] || [ -z "${SSH_USER:-}" ] || [ -z "${SSH_LOCAL_PORT:-}" ] || [ -z "${SSH_REMOTE_PORT:-}" ]; then
        echo "❌ 错误：ENABLE_SSH_TUNNEL=1 时，必须提供 SSH_HOST/SSH_PORT/SSH_USER/SSH_LOCAL_PORT/SSH_REMOTE_PORT"
        return 1
    fi

    if ! command_exists ssh; then
        echo "❌ 错误：未安装 ssh 命令"
        return 1
    fi

    cleanup_pid_file_if_stale "${SSH_PID_FILE}"
    if is_pid_running "${SSH_PID_FILE}"; then
        echo "✓ SSH 隧道已在运行（PID: $(cat "${SSH_PID_FILE}")）"
    else
        local remote_host
        remote_host="${SSH_REMOTE_HOST:-127.0.0.1}"

        local -a ssh_opts=(
            -CNg
            -L "${SSH_LOCAL_PORT}:${remote_host}:${SSH_REMOTE_PORT}"
            -p "${SSH_PORT}"
            -o StrictHostKeyChecking=no
            -o UserKnownHostsFile=/dev/null
            -o GlobalKnownHostsFile=/dev/null
            -o ServerAliveInterval=30
            -o ServerAliveCountMax=3
        )

        echo "🔄 启动 SSH 隧道..."
        if [ -n "${SSH_PASSWORD:-}" ] && [ "${SSH_PASSWORD}" != "你的密码写在这里" ]; then
            if ! command_exists sshpass; then
                echo "❌ 错误：检测到 SSH_PASSWORD，但未安装 sshpass"
                return 1
            fi
            nohup sshpass -p "${SSH_PASSWORD}" ssh "${ssh_opts[@]}" "${SSH_USER}@${SSH_HOST}" >>"${SSH_LOG_FILE}" 2>&1 &
        else
            nohup ssh "${ssh_opts[@]}" "${SSH_USER}@${SSH_HOST}" >>"${SSH_LOG_FILE}" 2>&1 &
        fi
        echo $! >"${SSH_PID_FILE}"
    fi

    if ! wait_for_port "${TUNNEL_CHECK_HOST}" "${SSH_LOCAL_PORT}" 12; then
        echo "❌ 错误：SSH 隧道未就绪，本地端口 ${SSH_LOCAL_PORT} 不可用"
        return 1
    fi

    if [ -z "${PYTHON_SERVICE_URL:-}" ]; then
        export PYTHON_SERVICE_URL="http://127.0.0.1:${SSH_LOCAL_PORT}"
    fi

    echo "✓ SSH 隧道已就绪: ${TUNNEL_CHECK_HOST}:${SSH_LOCAL_PORT}"
    echo "✓ 算法服务地址: ${PYTHON_SERVICE_URL}"
}

start_backend() {
    if ! command_exists go; then
        echo "❌ 错误：未安装 Go，请先安装 Go 1.25+"
        return 1
    fi

    if ! command_exists curl; then
        echo "❌ 错误：未安装 curl"
        return 1
    fi

    if [ ! -f "${CONFIG_FILE}" ]; then
        echo "❌ 错误：找不到配置文件 ${CONFIG_FILE}"
        return 1
    fi

    cleanup_pid_file_if_stale "${BACKEND_PID_FILE}"
    if is_pid_running "${BACKEND_PID_FILE}"; then
        echo "✓ 后端已在运行（PID: $(cat "${BACKEND_PID_FILE}")）"
        return 0
    fi

    local goflags
    goflags="${GOFLAGS:-}"
    if [[ "${goflags}" != *"-p="* ]] && [[ "${goflags}" != *"-p "* ]]; then
        goflags="${goflags} -p=1"
    fi
    export GOFLAGS="${goflags# }"
    export GOCACHE="${GOCACHE:-/tmp/gocache}"

    echo "🚀 启动后端服务..."
    echo "日志文件: ${BACKEND_LOG_FILE}"
    (
        cd "${BACKEND_DIR}"
        nohup go run cmd/server/main.go >>"${BACKEND_LOG_FILE}" 2>&1 &
        echo $! >"${BACKEND_PID_FILE}"
    )

    if ! wait_for_http "${BACKEND_HEALTH_URL}" "${BACKEND_STARTUP_RETRIES}"; then
        echo "❌ 错误：后端健康检查失败: ${BACKEND_HEALTH_URL}"
        echo "最近日志:"
        tail -n 40 "${BACKEND_LOG_FILE}" || true
        return 1
    fi

    echo "✓ 后端已启动（PID: $(cat "${BACKEND_PID_FILE}")）"
    echo "✓ 健康检查通过: ${BACKEND_HEALTH_URL}"
}

stop_process_by_pid_file() {
    local name="$1"
    local pid_file="$2"

    cleanup_pid_file_if_stale "${pid_file}"

    if ! is_pid_running "${pid_file}"; then
        echo "ℹ️ ${name} 未运行"
        rm -f "${pid_file}"
        return 0
    fi

    local pid
    pid="$(cat "${pid_file}")"
    echo "🛑 停止 ${name}（PID: ${pid}）..."
    kill "${pid}" 2>/dev/null || true

    for _ in {1..10}; do
        if ! kill -0 "${pid}" 2>/dev/null; then
            rm -f "${pid_file}"
            echo "✓ ${name} 已停止"
            return 0
        fi
        sleep 1
    done

    echo "⚠️ ${name} 未在 10 秒内退出，发送 SIGKILL"
    kill -9 "${pid}" 2>/dev/null || true
    rm -f "${pid_file}"
}

status_all() {
    cleanup_pid_file_if_stale "${BACKEND_PID_FILE}"
    cleanup_pid_file_if_stale "${SSH_PID_FILE}"

    if is_pid_running "${BACKEND_PID_FILE}"; then
        echo "BACKEND: running (PID: $(cat "${BACKEND_PID_FILE}"))"
    else
        echo "BACKEND: stopped"
    fi

    if is_pid_running "${SSH_PID_FILE}"; then
        echo "SSH_TUNNEL: running (PID: $(cat "${SSH_PID_FILE}"))"
    else
        echo "SSH_TUNNEL: stopped"
    fi
}

start_all() {
    load_env_file
    start_ssh_tunnel
    start_backend
    status_all
}

stop_all() {
    stop_process_by_pid_file "BACKEND" "${BACKEND_PID_FILE}"
    stop_process_by_pid_file "SSH_TUNNEL" "${SSH_PID_FILE}"
}

restart_all() {
    stop_all
    start_all
}

show_logs() {
    touch "${BACKEND_LOG_FILE}"
    tail -n 80 -f "${BACKEND_LOG_FILE}"
}

main() {
    local action="${1:-start}"

    case "${action}" in
        start)
            start_all
            ;;
        stop)
            stop_all
            ;;
        restart)
            restart_all
            ;;
        status)
            status_all
            ;;
        logs)
            show_logs
            ;;
        *)
            usage
            exit 1
            ;;
    esac
}

main "$@"
