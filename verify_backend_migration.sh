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
COLUMN_CHECKS=(
    users:id
    users:username
    users:email
    users:role
    users:created_at
    bridges:id
    bridges:bridge_name
    bridges:bridge_code
    bridges:user_id
    drones:id
    drones:name
    drones:user_id
    defects:id
    defects:bridge_id
    defects:defect_type
    defects:image_path
    defects:detected_at
    reports:id
    reports:report_name
    reports:report_type
    reports:user_id
    reports:bridge_id
    reports:status
    video_analysis_tasks:id
    video_analysis_tasks:task_id
    video_analysis_tasks:user_id
    video_analysis_tasks:bridge_id
    video_analysis_tasks:status
    defect_observations:id
    defect_observations:task_id
    defect_observations:bridge_id
    defect_observations:frame_no
    defect_observations:defect_type
)
INDEX_CHECKS=(
    users:username:UNIQUE
    users:email:UNIQUE
    bridges:bridge_code:UNIQUE
    reports:user_id:ANY
    reports:bridge_id:ANY
    video_analysis_tasks:task_id:UNIQUE
    defect_observations:task_id:ANY
    defect_observations:frame_no:ANY
)
FK_CHECKS=(
    reports:user_id:users:id
    reports:bridge_id:bridges:id
)

mysql_query() {
    local query="$1"
    MYSQL_PWD="${MYSQL_PASSWORD}" mysql -u"${MYSQL_USER}" -D"${MYSQL_DB}" -N -B -e "${query}" 2>/dev/null
}

mysql_exec() {
    local sql="$1"
    MYSQL_PWD="${MYSQL_PASSWORD}" mysql -u"${MYSQL_USER}" -D"${MYSQL_DB}" -e "${sql}" >/dev/null 2>&1
}

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
    output="$(mysql_query "SHOW TABLES;")" || {
        fail "MySQL 查询失败（请确认数据库账号和库名）"
    }

    for table in "${TABLES[@]}"; do
        echo "${output}" | grep -qx "${table}" || fail "缺少数据表: ${table}"
    done

    ok "数据库表检查通过（${MYSQL_DB}）"
}

check_columns() {
    local item=""
    local table=""
    local column=""
    local count=""

    for item in "${COLUMN_CHECKS[@]}"; do
        IFS=: read -r table column <<<"${item}"
        count="$(mysql_query "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='${MYSQL_DB}' AND table_name='${table}' AND column_name='${column}';")" || {
            fail "字段检查查询失败: ${table}.${column}"
        }
        [ "${count}" -ge 1 ] || fail "缺少字段: ${table}.${column}"
    done

    ok "核心字段检查通过"
}

check_indexes() {
    local item=""
    local table=""
    local column=""
    local index_type=""
    local count=""
    local query=""

    for item in "${INDEX_CHECKS[@]}"; do
        IFS=: read -r table column index_type <<<"${item}"
        query="SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics WHERE table_schema='${MYSQL_DB}' AND table_name='${table}' AND column_name='${column}'"
        if [ "${index_type}" = "UNIQUE" ]; then
            query="${query} AND non_unique = 0"
        fi
        query="${query};"

        count="$(mysql_query "${query}")" || {
            fail "索引检查查询失败: ${table}.${column}"
        }
        [ "${count}" -ge 1 ] || fail "缺少索引: ${table}.${column} (${index_type})"
    done

    ok "核心索引检查通过"
}

check_foreign_keys() {
    local item=""
    local table=""
    local column=""
    local ref_table=""
    local ref_column=""
    local count=""

    for item in "${FK_CHECKS[@]}"; do
        IFS=: read -r table column ref_table ref_column <<<"${item}"
        count="$(mysql_query "SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema='${MYSQL_DB}' AND table_name='${table}' AND column_name='${column}' AND referenced_table_name='${ref_table}' AND referenced_column_name='${ref_column}';")" || {
            fail "外键检查查询失败: ${table}.${column}"
        }
        [ "${count}" -ge 1 ] || fail "缺少外键: ${table}.${column} -> ${ref_table}.${ref_column}"
    done

    ok "核心外键检查通过"
}

check_default_admin() {
    local count=""
    count="$(mysql_query "SELECT COUNT(*) FROM users WHERE role='admin';")" || {
        fail "默认管理员检查查询失败"
    }
    [ "${count}" -ge 1 ] || fail "未检测到管理员账户（role='admin'）"
    ok "管理员账户检查通过"
}

check_transaction_smoke() {
    local stamp=""
    local sql=""

    stamp="$(date +%s)"
    sql="
START TRANSACTION;
INSERT INTO users (username, password, real_name, email, role, created_at, updated_at)
VALUES ('verify_u_${stamp}', 'verify_password_hash', 'verify-user', 'verify_${stamp}@example.com', 'user', NOW(), NOW());
SET @uid = LAST_INSERT_ID();

INSERT INTO bridges (bridge_name, bridge_code, user_id, created_at, updated_at)
VALUES ('verify-bridge-${stamp}', 'verify-code-${stamp}', @uid, NOW(), NOW());
SET @bid = LAST_INSERT_ID();

INSERT INTO drones (name, model, stream_url, user_id, created_at, updated_at)
VALUES ('verify-drone-${stamp}', 'verify-model', 'rtsp://example.com/live', @uid, NOW(), NOW());

INSERT INTO defects (bridge_id, defect_type, image_path, result_path, bbox, length, width, area, confidence, detected_at, created_at, updated_at)
VALUES (@bid, 'crack', '/tmp/source.jpg', '/tmp/result.jpg', '{}', 0.1, 0.1, 0.01, 0.9, NOW(), NOW(), NOW());

INSERT INTO reports (report_name, report_type, user_id, bridge_id, start_time, end_time, status, created_at, updated_at)
VALUES ('verify-report-${stamp}', 'bridge_inspection', @uid, @bid, NOW(), NOW(), 'generating', NOW(), NOW());

INSERT INTO video_analysis_tasks (task_id, session_id, user_id, bridge_id, video_path, fps, model_name, pixel_ratio, enable_segment, status, total_frames, processed_frames, confirmed_defects, created_at, updated_at)
VALUES ('verify-task-${stamp}', 'verify-session-${stamp}', @uid, @bid, '/tmp/mock.mp4', 1, 'baseline', 1, 0, 'uploaded', 0, 0, 0, NOW(), NOW());

INSERT INTO defect_observations (task_id, bridge_id, frame_no, timestamp_ms, defect_type, bbox, confidence, frame_path, result_path, track_id, created_at)
VALUES ('verify-task-${stamp}', @bid, 1, 0, 'crack', '{}', 0.9, '/tmp/frame.jpg', '/tmp/result.jpg', 'verify-track-${stamp}', NOW());

ROLLBACK;
"

    mysql_exec "${sql}" || fail "最小读写事务冒烟失败（检查字段约束、外键或默认值）"
    ok "最小读写事务冒烟通过（已回滚）"
}

main() {
    command -v curl >/dev/null 2>&1 || fail "未安装 curl"
    command -v mysql >/dev/null 2>&1 || fail "未安装 mysql 客户端"

    check_health
    check_port
    check_tables
    check_columns
    check_indexes
    check_foreign_keys
    check_default_admin
    check_transaction_smoke

    echo "PASS: 后端迁移验收通过"
}

main "$@"
