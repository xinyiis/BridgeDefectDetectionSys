#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${ROOT_DIR}/src/backend"
CONFIG_FILE="${BACKEND_DIR}/config.yaml"
HEALTH_RETRIES="${HEALTH_RETRIES:-40}"
BACKEND_PID=""

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

run_with_optional_sudo() {
    if [ "$(id -u)" -eq 0 ]; then
        "$@"
        return $?
    fi

    if command_exists sudo; then
        sudo "$@"
        return $?
    fi

    return 1
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

extract_config_value() {
    local section="$1"
    local key="$2"
    awk -v section="${section}" -v key="${key}" '
        $0 ~ "^" section ":" { in_section=1; next }
        in_section && $0 ~ "^[^[:space:]]" { in_section=0 }
        in_section && $1 == key ":" {
            value=$2
            gsub(/"/, "", value)
            print value
            exit
        }
    ' "${CONFIG_FILE}"
}

is_local_host() {
    local host="$1"
    case "${host}" in
        localhost | 127.0.0.1 | 0.0.0.0)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

try_start_mysql_service() {
    local unit

    if command_exists systemctl; then
        for unit in mysql mysqld mariadb; do
            echo "🔄 尝试通过 systemctl 启动 ${unit}.service ..."
            if run_with_optional_sudo systemctl start "${unit}" >/dev/null 2>&1; then
                return 0
            fi
        done
    fi

    if command_exists service; then
        for unit in mysql mysqld mariadb; do
            echo "🔄 尝试通过 service 启动 ${unit} ..."
            if run_with_optional_sudo service "${unit}" start >/dev/null 2>&1; then
                return 0
            fi
        done
    fi

    return 1
}

ensure_mysql_available() {
    local db_host="$1"
    local db_port="$2"

    if wait_for_port "${db_host}" "${db_port}" 3; then
        echo "✓ MySQL 可达: ${db_host}:${db_port}"
        return 0
    fi

    if ! is_local_host "${db_host}"; then
        echo "❌ MySQL 不可达: ${db_host}:${db_port}"
        echo "   当前是远程数据库地址，脚本不会自动启动远程服务"
        return 1
    fi

    echo "⚠️  MySQL 不可达: ${db_host}:${db_port}"
    echo "🔄 尝试自动启动本机 MySQL 服务..."

    if ! try_start_mysql_service; then
        echo "❌ 自动启动 MySQL 失败（systemctl/service 均未成功）"
        echo "   请手动执行: systemctl start mysql 或 service mysql start"
        return 1
    fi

    if ! wait_for_port "${db_host}" "${db_port}" 15; then
        echo "❌ MySQL 已尝试启动，但端口仍不可达: ${db_host}:${db_port}"
        return 1
    fi

    echo "✓ MySQL 已自动启动并可达: ${db_host}:${db_port}"
}

cleanup() {
    local exit_code=$?
    if [ -n "${BACKEND_PID}" ] && kill -0 "${BACKEND_PID}" 2>/dev/null; then
        kill "${BACKEND_PID}" 2>/dev/null || true
        wait "${BACKEND_PID}" 2>/dev/null || true
    fi
    exit "${exit_code}"
}

check_prerequisites() {
    if ! command_exists go; then
        echo "❌ 未检测到 Go，请先安装 Go 1.25+"
        return 1
    fi

    if ! command_exists curl; then
        echo "❌ 未检测到 curl，请先安装 curl"
        return 1
    fi

    if [ ! -f "${CONFIG_FILE}" ]; then
        echo "❌ 未找到配置文件: ${CONFIG_FILE}"
        return 1
    fi
}

check_services() {
    local dsn
    local db_host
    local db_port
    local python_enabled
    local python_url
    local python_host
    local python_port

    dsn="$(extract_config_value "database" "dsn")"
    if [[ "${dsn}" =~ @tcp\(([^:]+):([0-9]+)\) ]]; then
        db_host="${BASH_REMATCH[1]}"
        db_port="${BASH_REMATCH[2]}"
    else
        db_host="127.0.0.1"
        db_port="3306"
    fi

    ensure_mysql_available "${db_host}" "${db_port}"

    python_enabled="$(extract_config_value "python_service" "enabled")"
    python_url="${PYTHON_SERVICE_URL:-$(extract_config_value "python_service" "url")}"
    if [ "${python_enabled}" = "true" ] && [[ "${python_url}" =~ ^http://([^:/]+):([0-9]+) ]]; then
        python_host="${BASH_REMATCH[1]}"
        python_port="${BASH_REMATCH[2]}"
        if wait_for_port "${python_host}" "${python_port}" 2; then
            echo "✓ Python 服务可达: ${python_host}:${python_port}"
        else
            echo "⚠️  Python 服务暂不可达: ${python_host}:${python_port}"
            echo "   后端仍会启动，但检测相关接口可能报错"
        fi
    fi
}

start_backend() {
    local server_port
    local health_url
    server_port="$(extract_config_value "server" "port")"
    if [ -z "${server_port}" ]; then
        server_port="8080"
    fi
    health_url="http://127.0.0.1:${server_port}/api/v1/health"

    if curl -fsS "${health_url}" >/dev/null 2>&1; then
        echo "✓ 后端已在运行: ${health_url}"
        return 0
    fi

    cd "${BACKEND_DIR}"
    export GOCACHE="${GOCACHE:-/tmp/gocache}"
    mkdir -p "${GOCACHE}"

    go run cmd/server/main.go &
    BACKEND_PID=$!

    for ((i = 1; i <= HEALTH_RETRIES; i++)); do
        if curl -fsS "${health_url}" >/dev/null 2>&1; then
            echo "✓ 后端启动成功: ${health_url}"
            wait "${BACKEND_PID}"
            return 0
        fi

        if ! kill -0 "${BACKEND_PID}" 2>/dev/null; then
            echo "❌ 后端进程启动失败（进程已退出）"
            return 1
        fi
        sleep 1
    done

    echo "❌ 后端健康检查超时: ${health_url}"
    return 1
}

trap cleanup INT TERM EXIT

check_prerequisites
check_services
start_backend
