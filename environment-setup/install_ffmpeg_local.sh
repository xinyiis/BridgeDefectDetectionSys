#!/bin/bash

##############################################
# 桥梁病害检测系统 - FFmpeg 本地安装脚本
# 适用于: Ubuntu 22.04/24.04 x86_64
# 功能: 下载静态版 ffmpeg/ffprobe 到仓库本地工具目录
##############################################

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
INSTALL_ROOT="${REPO_ROOT}/.local-tools/ffmpeg"
BIN_DIR="${INSTALL_ROOT}/bin"
CACHE_DIR="${INSTALL_ROOT}/cache"
VERSION_FILE="${INSTALL_ROOT}/VERSION"
PRIMARY_ARCHIVE_URL="${FFMPEG_ARCHIVE_URL:-https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz}"
FALLBACK_ARCHIVE_URL="https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz"

resolve_owner() {
    if [ -n "${SUDO_USER:-}" ] && id -u "${SUDO_USER}" >/dev/null 2>&1; then
        echo "${SUDO_USER}"
        return
    fi

    echo "${USER}"
}

ensure_supported_platform() {
    local machine
    machine="$(uname -m)"

    if [ "${machine}" != "x86_64" ]; then
        print_error "当前脚本仅支持 x86_64，检测到架构: ${machine}"
        exit 1
    fi
}

download_archive() {
    local archive_path="$1"
    local source_url="$2"

    print_info "下载静态版 FFmpeg..."
    print_info "下载命令: curl -L --fail -C - --output \"${archive_path}\" \"${source_url}\""

    if command -v curl >/dev/null 2>&1; then
        curl -L --fail --retry 3 --retry-delay 2 -C - --output "${archive_path}" "${source_url}"
        return
    fi

    if command -v wget >/dev/null 2>&1; then
        wget -c -O "${archive_path}" "${source_url}"
        return
    fi

    print_error "未找到 curl 或 wget，无法下载 FFmpeg"
    exit 1
}

already_installed() {
    [ -x "${BIN_DIR}/ffmpeg" ] && [ -x "${BIN_DIR}/ffprobe" ]
}

install_ffmpeg_local() {
    local owner
    owner="$(resolve_owner)"

    ensure_supported_platform

    if already_installed; then
        print_success "本地 FFmpeg 已存在，跳过下载"
        "${BIN_DIR}/ffmpeg" -version | sed -n '1p'
        "${BIN_DIR}/ffprobe" -version | sed -n '1p'
        return
    fi

    local temp_dir
    temp_dir="$(mktemp -d)"
    trap "rm -rf '${temp_dir}'" EXIT

    mkdir -p "${CACHE_DIR}"
    local archive_path="${CACHE_DIR}/$(basename "${PRIMARY_ARCHIVE_URL}")"
    local source_url="${PRIMARY_ARCHIVE_URL}"

    if ! download_archive "${archive_path}" "${source_url}"; then
        print_info "主下载源失败，切换到备用源..."
        archive_path="${CACHE_DIR}/$(basename "${FALLBACK_ARCHIVE_URL}")"
        source_url="${FALLBACK_ARCHIVE_URL}"
        download_archive "${archive_path}" "${source_url}"
    fi

    print_info "解压 FFmpeg 压缩包..."
    tar -xJf "${archive_path}" -C "${temp_dir}"

    local extracted_dir
    extracted_dir="$(find "${temp_dir}" -mindepth 1 -maxdepth 1 -type d | head -n 1)"
    if [ -z "${extracted_dir}" ]; then
        print_error "未找到解压后的 FFmpeg 目录"
        exit 1
    fi

    local extracted_ffmpeg="${extracted_dir}/ffmpeg"
    local extracted_ffprobe="${extracted_dir}/ffprobe"
    if [ ! -x "${extracted_ffmpeg}" ] && [ -x "${extracted_dir}/bin/ffmpeg" ]; then
        extracted_ffmpeg="${extracted_dir}/bin/ffmpeg"
    fi
    if [ ! -x "${extracted_ffprobe}" ] && [ -x "${extracted_dir}/bin/ffprobe" ]; then
        extracted_ffprobe="${extracted_dir}/bin/ffprobe"
    fi
    if [ ! -x "${extracted_ffmpeg}" ] || [ ! -x "${extracted_ffprobe}" ]; then
        print_error "未在压缩包中找到可执行的 ffmpeg/ffprobe"
        exit 1
    fi

    mkdir -p "${BIN_DIR}"
    install -m 755 "${extracted_ffmpeg}" "${BIN_DIR}/ffmpeg"
    install -m 755 "${extracted_ffprobe}" "${BIN_DIR}/ffprobe"

    {
        echo "source_url=${source_url}"
        echo "installed_at=$(date '+%Y-%m-%d %H:%M:%S %z')"
        echo "ffmpeg_version=$("${BIN_DIR}/ffmpeg" -version | sed -n '1p')"
        echo "ffprobe_version=$("${BIN_DIR}/ffprobe" -version | sed -n '1p')"
    } > "${VERSION_FILE}"

    if [ -n "${owner}" ] && id -u "${owner}" >/dev/null 2>&1; then
        chown -R "${owner}:$(id -gn "${owner}")" "${INSTALL_ROOT}" 2>/dev/null || true
    fi

    print_success "FFmpeg 本地安装完成"
    print_info "安装目录: ${BIN_DIR}"
    "${BIN_DIR}/ffmpeg" -version | sed -n '1,2p'
    "${BIN_DIR}/ffprobe" -version | sed -n '1,2p'
}

install_ffmpeg_local "$@"
