#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_SETUP_DIR="${ROOT_DIR}/environment-setup"
BACKEND_ENV_DIR="${ENV_SETUP_DIR}/backend-env"
BACKEND_DIR="${ROOT_DIR}/src/backend"
MYSQL_DATABASE="${MYSQL_DATABASE:-bridge_detection}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-123456}"

echo "=========================================="
echo "  BridgeDefectDetectionSys 同机后端初始化"
echo "=========================================="
echo "项目目录: ${ROOT_DIR}"
echo "数据库名: ${MYSQL_DATABASE}"
echo ""

cd "${ENV_SETUP_DIR}"
./setup_basic_tools.sh

cd "${BACKEND_ENV_DIR}"
./setup_backend_env.sh

if [ -f "${HOME}/.bashrc" ]; then
    # shellcheck disable=SC1090
    source "${HOME}/.bashrc"
fi

echo ""
echo "========== 环境验证 =========="
go version
mysql --version

echo ""
echo "========== 初始化数据库 =========="
mysql -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" -e "CREATE DATABASE IF NOT EXISTS ${MYSQL_DATABASE} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo ""
echo "========== 后端测试 =========="
cd "${BACKEND_DIR}"
GOCACHE=/tmp/gocache go test ./...

echo ""
echo "=========================================="
echo "同机后端环境初始化完成"
echo "下一步："
echo "1. 启动算法服务，并确认监听在 127.0.0.1:18080"
echo "2. 在项目根目录执行: bash start_same_host_backend.sh"
echo "=========================================="
