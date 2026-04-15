#!/bin/bash
# 视频检测端到端联调测试脚本
# 用途：验证后端 + 算法端视频异步检测链路是否正常工作
# 使用：./test_video_detection_e2e.sh [video_file_path]

set -e

# ============ 配置区 ============
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
ALGO_URL="${ALGO_URL:-http://localhost:18080}"
TEST_VIDEO="${1:-}"
# 使用数据库中实际存在的值
BRIDGE_ID="${BRIDGE_ID:-2}"      # 数据库中存在 ID: 2-11
DRONE_ID="${DRONE_ID:-3}"        # 数据库中存在 ID: 3, 5, 7, 9, 11
USERNAME="${USERNAME:-admin}"    # 使用 admin 账户
PASSWORD="${PASSWORD:-admin123}" # admin 密码

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============ 工具函数 ============
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "命令 '$1' 未安装，请先安装"
        exit 1
    fi
}

# ============ 步骤 1: 环境检查 ============
log_info "=========================================="
log_info "步骤 1: 环境检查"
log_info "=========================================="

check_command curl
check_command jq

# ============ 步骤 2: 服务健康检查 ============
log_info "=========================================="
log_info "步骤 2: 服务健康检查"
log_info "=========================================="

# 检查后端服务
log_info "检查后端服务: $BACKEND_URL"
if curl -sf "$BACKEND_URL/api/v1/health" > /dev/null 2>&1; then
    log_success "后端服务运行正常"
else
    log_error "后端服务不可用，请检查服务是否启动"
    log_info "提示: 确保后端服务运行在 $BACKEND_URL"
    exit 1
fi

# 检查算法服务
log_info "检查算法服务: $ALGO_URL"
if curl -sf "$ALGO_URL/docs" > /dev/null 2>&1; then
    log_success "算法服务运行正常"
else
    log_error "算法服务不可用，请检查服务是否启动"
    log_info "提示: 确保算法服务运行在 $ALGO_URL"
    exit 1
fi

# ============ 步骤 3: 用户登录 ============
log_info "=========================================="
log_info "步骤 3: 用户登录"
log_info "=========================================="

log_info "使用用户 '$USERNAME' 登录..."
LOGIN_RESPONSE=$(curl -sf -X POST "$BACKEND_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}" \
    -c /tmp/session_cookie.txt \
    2>&1)

if [ $? -ne 0 ]; then
    log_error "登录失败"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.id // .data.user_id // .user_id // .id // empty')
if [ -z "$USER_ID" ]; then
    log_error "无法从登录响应中提取 user_id"
    echo "响应内容: $LOGIN_RESPONSE"
    exit 1
fi

log_success "登录成功，用户ID: $USER_ID"

# ============ 步骤 4: 准备测试视频 ============
log_info "=========================================="
log_info "步骤 4: 准备测试视频"
log_info "=========================================="

if [ -z "$TEST_VIDEO" ]; then
    log_error "请提供测试视频文件路径"
    log_info "用法: $0 <video_file_path>"
    log_info "示例: $0 /path/to/test_video.mp4"
    exit 1
fi

if [ ! -f "$TEST_VIDEO" ]; then
    log_error "视频文件不存在: $TEST_VIDEO"
    exit 1
fi

VIDEO_SIZE=$(du -h "$TEST_VIDEO" | cut -f1)
VIDEO_SIZE_BYTES=$(stat -c%s "$TEST_VIDEO" 2>/dev/null || stat -f%z "$TEST_VIDEO" 2>/dev/null)
MAX_SIZE_BYTES=$((200 * 1024 * 1024))  # 200MB

if [ "$VIDEO_SIZE_BYTES" -gt "$MAX_SIZE_BYTES" ]; then
    log_error "视频文件过大: $VIDEO_SIZE ($(($VIDEO_SIZE_BYTES / 1024 / 1024))MB)"
    log_error "后端限制最大文件大小为 200MB"
    log_info "请使用更小的视频文件，或使用以下命令压缩视频："
    log_info "  ffmpeg -i $TEST_VIDEO -vcodec libx264 -crf 28 output.mp4"
    exit 1
fi

log_success "找到测试视频: $TEST_VIDEO (大小: $VIDEO_SIZE)"

# ============ 步骤 5: 上传视频 ============
log_info "=========================================="
log_info "步骤 5: 上传视频并创建任务"
log_info "=========================================="

log_info "正在上传视频..."
# 使用临时文件分离进度条和响应内容
UPLOAD_RESPONSE=$(curl -X POST "$BACKEND_URL/api/v1/detection/video/upload" \
    --max-time 300 \
    -b /tmp/session_cookie.txt \
    -F "video=@$TEST_VIDEO" \
    -F "bridge_id=$BRIDGE_ID" \
    -F "drone_id=$DRONE_ID" \
    -F "user_id=$USER_ID" \
    -F "pixel_ratio=0.5" \
    2>/tmp/upload_progress.log)

if [ $? -ne 0 ]; then
    log_error "视频上传失败"
    echo "$UPLOAD_RESPONSE"
    exit 1
fi

TASK_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.task_id // .task_id // empty')
if [ -z "$TASK_ID" ]; then
    log_error "无法从响应中提取 task_id"
    echo "响应内容: $UPLOAD_RESPONSE"
    exit 1
fi

log_success "视频上传成功，任务ID: $TASK_ID"

# ============ 步骤 6: 启动检测任务 ============
log_info "=========================================="
log_info "步骤 6: 启动检测任务"
log_info "=========================================="

SESSION_ID="test_session_$(date +%s)"
log_info "启动任务 (task_id=$TASK_ID, session_id=$SESSION_ID)..."

START_RESPONSE=$(curl -sf -X POST "$BACKEND_URL/api/v1/detection/video/$TASK_ID/start" \
    -b /tmp/session_cookie.txt \
    -H "Content-Type: application/json" \
    -d "{\"session_id\": \"$SESSION_ID\"}" \
    2>&1)

if [ $? -ne 0 ]; then
    log_error "启动任务失败"
    echo "$START_RESPONSE"
    exit 1
fi

log_success "任务启动成功"

# ============ 步骤 7: 轮询任务状态 ============
log_info "=========================================="
log_info "步骤 7: 轮询任务状态"
log_info "=========================================="

MAX_WAIT=300  # 最多等待5分钟
POLL_INTERVAL=3
ELAPSED=0
START_TIME=$(date +%s)

log_info "开始轮询任务状态 (最多等待 ${MAX_WAIT}s)..."

while [ $ELAPSED -lt $MAX_WAIT ]; do
    STATUS_RESPONSE=$(curl -sf "$BACKEND_URL/api/v1/detection/video/$TASK_ID" \
        -b /tmp/session_cookie.txt \
        2>&1)

    if [ $? -ne 0 ]; then
        log_warn "查询任务状态失败，继续重试..."
        sleep $POLL_INTERVAL
        ELAPSED=$(($(date +%s) - START_TIME))
        continue
    fi

    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.data.status // .status // empty')
    PROCESSED=$(echo "$STATUS_RESPONSE" | jq -r '.data.processed_frames // .processed_frames // 0')
    TOTAL=$(echo "$STATUS_RESPONSE" | jq -r '.data.total_frames // .total_frames // 0')
    CONFIRMED=$(echo "$STATUS_RESPONSE" | jq -r '.data.confirmed_defects // .confirmed_defects // 0')

    if [ -z "$STATUS" ]; then
        log_warn "无法解析任务状态，继续重试..."
        sleep $POLL_INTERVAL
        ELAPSED=$(($(date +%s) - START_TIME))
        continue
    fi

    # 显示进度
    if [ "$TOTAL" -gt 0 ]; then
        PROGRESS=$((PROCESSED * 100 / TOTAL))
        echo -ne "\r${BLUE}[INFO]${NC} 状态: $STATUS | 进度: $PROCESSED/$TOTAL ($PROGRESS%) | 确认缺陷: $CONFIRMED"
    else
        echo -ne "\r${BLUE}[INFO]${NC} 状态: $STATUS | 已处理: $PROCESSED 帧 | 确认缺陷: $CONFIRMED"
    fi

    # 检查终态
    if [ "$STATUS" = "completed" ]; then
        echo ""
        log_success "任务完成！"
        break
    elif [ "$STATUS" = "failed" ]; then
        echo ""
        ERROR_MSG=$(echo "$STATUS_RESPONSE" | jq -r '.data.error_message // .error_message // "未知错误"')
        log_error "任务失败: $ERROR_MSG"
        exit 1
    elif [ "$STATUS" = "canceled" ]; then
        echo ""
        log_error "任务已取消"
        exit 1
    fi

    sleep $POLL_INTERVAL
    ELAPSED=$(($(date +%s) - START_TIME))
done

echo ""

if [ $ELAPSED -ge $MAX_WAIT ]; then
    log_error "任务超时 (等待超过 ${MAX_WAIT}s)"
    exit 1
fi

# ============ 步骤 8: 获取最终结果 ============
log_info "=========================================="
log_info "步骤 8: 获取最终结果"
log_info "=========================================="

FINAL_RESPONSE=$(curl -sf "$BACKEND_URL/api/v1/detection/video/$TASK_ID" \
    -b /tmp/session_cookie.txt \
    2>&1)

if [ $? -ne 0 ]; then
    log_error "获取最终结果失败"
    exit 1
fi

# 解析结果
TOTAL_FRAMES=$(echo "$FINAL_RESPONSE" | jq -r '.data.total_frames // .total_frames // 0')
PROCESSED_FRAMES=$(echo "$FINAL_RESPONSE" | jq -r '.data.processed_frames // .processed_frames // 0')
CONFIRMED_DEFECTS=$(echo "$FINAL_RESPONSE" | jq -r '.data.confirmed_defects // .confirmed_defects // 0')
CREATED_AT=$(echo "$FINAL_RESPONSE" | jq -r '.data.created_at // .created_at // "N/A"')
COMPLETED_AT=$(echo "$FINAL_RESPONSE" | jq -r '.data.completed_at // .completed_at // "N/A"')

# 计算总耗时
END_TIME=$(date +%s)
TOTAL_TIME=$((END_TIME - START_TIME))

# ============ 步骤 9: 输出测试报告 ============
log_info "=========================================="
log_info "步骤 9: 测试报告"
log_info "=========================================="

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│                      测试报告                                │"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 任务ID:           $TASK_ID"
echo "│ 视频文件:         $TEST_VIDEO"
echo "│ 视频大小:         $VIDEO_SIZE"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 总帧数:           $TOTAL_FRAMES"
echo "│ 已处理帧数:       $PROCESSED_FRAMES"
echo "│ 确认缺陷数:       $CONFIRMED_DEFECTS"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 创建时间:         $CREATED_AT"
echo "│ 完成时间:         $COMPLETED_AT"
echo "│ 总耗时:           ${TOTAL_TIME}s"
echo "└─────────────────────────────────────────────────────────────┘"
echo ""

# 性能指标
if [ "$TOTAL_FRAMES" -gt 0 ] && [ "$TOTAL_TIME" -gt 0 ]; then
    AVG_TIME_PER_FRAME=$((TOTAL_TIME * 1000 / TOTAL_FRAMES))
    THROUGHPUT=$(echo "scale=2; $TOTAL_FRAMES / $TOTAL_TIME" | bc)

    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                      性能指标                                │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ 平均每帧耗时:     ${AVG_TIME_PER_FRAME}ms"
    echo "│ 吞吐量:           ${THROUGHPUT} 帧/秒"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo ""
fi

# ============ 步骤 10: 验证结果 ============
log_info "=========================================="
log_info "步骤 10: 验证结果"
log_info "=========================================="

VALIDATION_PASSED=true

# 验证1: 所有帧都已处理
if [ "$PROCESSED_FRAMES" -eq "$TOTAL_FRAMES" ]; then
    log_success "✓ 所有帧已处理 ($PROCESSED_FRAMES/$TOTAL_FRAMES)"
else
    log_error "✗ 帧处理不完整 ($PROCESSED_FRAMES/$TOTAL_FRAMES)"
    VALIDATION_PASSED=false
fi

# 验证2: 至少有一些帧被处理
if [ "$PROCESSED_FRAMES" -gt 0 ]; then
    log_success "✓ 至少处理了一些帧 ($PROCESSED_FRAMES)"
else
    log_error "✗ 没有处理任何帧"
    VALIDATION_PASSED=false
fi

# 验证3: 任务状态为完成
FINAL_STATUS=$(echo "$FINAL_RESPONSE" | jq -r '.data.status // .status // empty')
if [ "$FINAL_STATUS" = "completed" ]; then
    log_success "✓ 任务状态正确 (completed)"
else
    log_error "✗ 任务状态异常 ($FINAL_STATUS)"
    VALIDATION_PASSED=false
fi

echo ""

if [ "$VALIDATION_PASSED" = true ]; then
    log_success "=========================================="
    log_success "所有验证通过！视频检测链路工作正常 ✓"
    log_success "=========================================="
    exit 0
else
    log_error "=========================================="
    log_error "部分验证失败，请检查日志"
    log_error "=========================================="
    exit 1
fi
