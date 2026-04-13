#!/usr/bin/env bash

set -euo pipefail

HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/api/v1/health}"
PORT="${PORT:-8080}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-123456}"
MYSQL_DB="${MYSQL_DB:-bridge_detection}"
HEALTH_RETRIES="${HEALTH_RETRIES:-20}"
TABLES=(
    users
    bridges
    drones
    defects
    reports
    video_analysis_tasks
    defect_observations
)

ok() {
    echo "PASS: $1"
}

fail() {
    echo "FAIL: $1"
    exit 1
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

wait_for_health() {
    local retries="$1"
    local response=""

    for ((i = 1; i <= retries; i++)); do
        if response="$(curl -fsS "${HEALTH_URL}" 2>/dev/null)"; then
            echo "${response}"
            return 0
        fi
        sleep 1
    done

    return 1
}

check_port() {
    if command_exists ss; then
        if ss -lntp | grep -q ":${PORT} "; then
            ok "端口 ${PORT} 正在监听（ss）"
            return 0
        fi
        fail "端口 ${PORT} 未监听（ss）"
    fi

    if command_exists netstat; then
        if netstat -lnt | grep -q "[.:]${PORT}[[:space:]]"; then
            ok "端口 ${PORT} 正在监听（netstat）"
            return 0
        fi
        fail "端口 ${PORT} 未监听（netstat）"
    fi

    if command_exists lsof; then
        if lsof -nP -iTCP:"${PORT}" -sTCP:LISTEN >/dev/null 2>&1; then
            ok "端口 ${PORT} 正在监听（lsof）"
            return 0
        fi
        fail "端口 ${PORT} 未监听（lsof）"
    fi

    ok "跳过端口监听检查（未安装 ss/netstat/lsof）"
}

check_health() {
    local response=""

    if ! response="$(wait_for_health "${HEALTH_RETRIES}")"; then
        fail "健康检查失败: ${HEALTH_URL}"
    fi

    echo "${response}" | grep -q '"status":"ok"' || fail "健康检查返回异常: ${response}"
    ok "健康检查通过: ${HEALTH_URL}"
}

check_tables() {
    local output=""
    output="$(mysql -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -D"${MYSQL_DB}" -N -e "SHOW TABLES;" 2>/dev/null)" || {
        fail "MySQL 查询失败（请确认数据库账号和库名）"
    }

    for table in "${TABLES[@]}"; do
        echo "${output}" | grep -qx "${table}" || fail "缺少数据表: ${table}"
    done

    ok "数据库表检查通过（${MYSQL_DB}）"
}

main() {
    command -v curl >/dev/null 2>&1 || fail "未安装 curl"
    command -v mysql >/dev/null 2>&1 || fail "未安装 mysql 客户端"

    check_health
    check_port
    check_tables

    echo "PASS: 后端迁移验收通过"
}

main "$@"
