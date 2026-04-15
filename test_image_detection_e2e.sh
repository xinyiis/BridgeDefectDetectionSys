#!/bin/bash
# 图片检测端到端联调测试脚本
# 用途：验证后端 + 算法端图片同步检测链路是否正常工作
# 使用：./test_image_detection_e2e.sh [image_file_path]

set -e

# ============ 配置区 ============
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
ALGO_URL="${ALGO_URL:-http://localhost:18080}"
TEST_IMAGE="${1:-}"
# 使用数据库中实际存在的值
BRIDGE_ID="${BRIDGE_ID:-2}"      # 数据库中存在 ID: 2-11
MODEL_NAME="${MODEL_NAME:-yolov8}" # 模型名称
PIXEL_RATIO="${PIXEL_RATIO:-0.5}" # 像素比例（米/像素）
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

# ============ 步骤 4: 准备测试图片 ============
log_info "=========================================="
log_info "步骤 4: 准备测试图片"
log_info "=========================================="

if [ -z "$TEST_IMAGE" ]; then
    log_error "请提供测试图片文件路径"
    log_info "用法: $0 <image_file_path>"
    log_info "示例: $0 /path/to/test_image.jpg"
    exit 1
fi

if [ ! -f "$TEST_IMAGE" ]; then
    log_error "图片文件不存在: $TEST_IMAGE"
    exit 1
fi

IMAGE_SIZE=$(du -h "$TEST_IMAGE" | cut -f1)
IMAGE_SIZE_BYTES=$(stat -c%s "$TEST_IMAGE" 2>/dev/null || stat -f%z "$TEST_IMAGE" 2>/dev/null)
MAX_SIZE_BYTES=$((10 * 1024 * 1024))  # 10MB

if [ "$IMAGE_SIZE_BYTES" -gt "$MAX_SIZE_BYTES" ]; then
    log_error "图片文件过大: $IMAGE_SIZE ($(($IMAGE_SIZE_BYTES / 1024 / 1024))MB)"
    log_error "后端限制最大文件大小为 10MB"
    exit 1
fi

# 验证图片格式
IMAGE_EXT="${TEST_IMAGE##*.}"
IMAGE_EXT_LOWER=$(echo "$IMAGE_EXT" | tr '[:upper:]' '[:lower:]')
if [[ ! "$IMAGE_EXT_LOWER" =~ ^(jpg|jpeg|png|bmp)$ ]]; then
    log_error "不支持的图片格式: .$IMAGE_EXT_LOWER"
    log_error "仅支持: jpg, jpeg, png, bmp"
    exit 1
fi

log_success "找到测试图片: $TEST_IMAGE (大小: $IMAGE_SIZE, 格式: $IMAGE_EXT_LOWER)"

# ============ 步骤 5: 上传图片并检测 ============
log_info "=========================================="
log_info "步骤 5: 上传图片并进行检测"
log_info "=========================================="

log_info "正在上传图片并检测..."
START_TIME=$(date +%s%3N)  # 毫秒级时间戳

DETECT_RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/detection/upload" \
    --max-time 120 \
    -b /tmp/session_cookie.txt \
    -F "image=@$TEST_IMAGE" \
    -F "bridge_id=$BRIDGE_ID" \
    -F "model_name=$MODEL_NAME" \
    -F "pixel_ratio=$PIXEL_RATIO")

END_TIME=$(date +%s%3N)
CURL_EXIT_CODE=$?

if [ $CURL_EXIT_CODE -ne 0 ]; then
    log_error "图片上传失败 (curl exit code: $CURL_EXIT_CODE)"
    echo "$DETECT_RESPONSE"
    exit 1
fi

# 检查响应是否为有效 JSON
if ! echo "$DETECT_RESPONSE" | jq empty 2>/dev/null; then
    log_error "响应不是有效的 JSON"
    echo "响应内容: $DETECT_RESPONSE"
    exit 1
fi

# 检查响应码（兼容 code=0 与 code=200 两种成功语义）
RESPONSE_CODE=$(echo "$DETECT_RESPONSE" | jq -r '.code // empty')
if [ "$RESPONSE_CODE" != "0" ] && [ "$RESPONSE_CODE" != "200" ]; then
    ERROR_MSG=$(echo "$DETECT_RESPONSE" | jq -r '.message // "未知错误"')
    log_error "检测失败: $ERROR_MSG"
    echo "完整响应: $DETECT_RESPONSE"
    exit 1
fi

REQUEST_TIME=$((END_TIME - START_TIME))
log_success "检测完成，耗时: ${REQUEST_TIME}ms"

# ============ 步骤 6: 解析检测结果 ============
log_info "=========================================="
log_info "步骤 6: 解析检测结果"
log_info "=========================================="

# 提取关键字段
TOTAL_DEFECTS=$(echo "$DETECT_RESPONSE" | jq -r '.data.total_defects // 0')
IMAGE_PATH=$(echo "$DETECT_RESPONSE" | jq -r '.data.image_path // "N/A"')
RESULT_PATH=$(echo "$DETECT_RESPONSE" | jq -r '.data.result_path // "N/A"')
PROCESSING_TIME=$(echo "$DETECT_RESPONSE" | jq -r '.data.processing_time // 0')
PERSISTENCE_STATUS=$(echo "$DETECT_RESPONSE" | jq -r '.data.persistence_status // "unknown"')
PERSISTENCE_TASK_ID=$(echo "$DETECT_RESPONSE" | jq -r '.data.persistence_task_id // "N/A"')

log_info "检测到缺陷数量: $TOTAL_DEFECTS"
log_info "原始图片路径: $IMAGE_PATH"
log_info "结果图片路径: $RESULT_PATH"
log_info "算法处理时间: ${PROCESSING_TIME}s"
log_info "持久化状态: $PERSISTENCE_STATUS"
log_info "持久化任务ID: $PERSISTENCE_TASK_ID"

# 提取缺陷列表
DEFECTS=$(echo "$DETECT_RESPONSE" | jq -c '.data.defects // []')
DEFECT_COUNT=$(echo "$DEFECTS" | jq 'length')

if [ "$DEFECT_COUNT" -gt 0 ]; then
    echo ""
    log_info "缺陷详情:"
    echo "$DEFECTS" | jq -r '.[] | "  - 类型: \(.defect_type), 置信度: \(.confidence), 位置: (\(.bbox.x), \(.bbox.y)), 尺寸: \(.bbox.width)x\(.bbox.height)"'
fi

# 提取缺陷统计
DEFECT_SUMMARY=$(echo "$DETECT_RESPONSE" | jq -c '.data.defect_summary // {}')
if [ "$DEFECT_SUMMARY" != "{}" ]; then
    echo ""
    log_info "缺陷统计:"
    echo "$DEFECT_SUMMARY" | jq -r 'to_entries[] | "  - \(.key): \(.value) 个"'
fi

# ============ 步骤 7: 验证持久化状态（如果有） ============
if [ "$PERSISTENCE_TASK_ID" != "N/A" ] && [ "$PERSISTENCE_TASK_ID" != "null" ]; then
    log_info "=========================================="
    log_info "步骤 7: 验证持久化状态"
    log_info "=========================================="

    log_info "等待持久化完成..."
    MAX_WAIT=30
    POLL_INTERVAL=2
    ELAPSED=0

    while [ $ELAPSED -lt $MAX_WAIT ]; do
        PERSISTENCE_RESPONSE=$(curl -sf "$BACKEND_URL/api/v1/detection/persistence/$PERSISTENCE_TASK_ID" \
            -b /tmp/session_cookie.txt \
            2>&1)

        if [ $? -eq 0 ]; then
            PERSIST_STATUS=$(echo "$PERSISTENCE_RESPONSE" | jq -r '.data.status // "unknown"')

            if [ "$PERSIST_STATUS" = "completed" ]; then
                SAVED_DEFECT_IDS=$(echo "$PERSISTENCE_RESPONSE" | jq -r '.data.saved_defect_ids // [] | length')
                log_success "持久化完成，已保存 $SAVED_DEFECT_IDS 个缺陷到数据库"
                break
            elif [ "$PERSIST_STATUS" = "failed" ]; then
                ERROR_MSG=$(echo "$PERSISTENCE_RESPONSE" | jq -r '.data.error_message // "未知错误"')
                log_error "持久化失败: $ERROR_MSG"
                break
            else
                echo -ne "\r${BLUE}[INFO]${NC} 持久化状态: $PERSIST_STATUS"
            fi
        fi

        sleep $POLL_INTERVAL
        ELAPSED=$((ELAPSED + POLL_INTERVAL))
    done

    echo ""

    if [ $ELAPSED -ge $MAX_WAIT ]; then
        log_warn "持久化状态查询超时"
    fi
fi

# ============ 步骤 8: 输出测试报告 ============
log_info "=========================================="
log_info "步骤 8: 测试报告"
log_info "=========================================="

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│                      测试报告                                │"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 测试图片:         $TEST_IMAGE"
echo "│ 图片大小:         $IMAGE_SIZE"
echo "│ 图片格式:         $IMAGE_EXT_LOWER"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 桥梁ID:           $BRIDGE_ID"
echo "│ 模型名称:         $MODEL_NAME"
echo "│ 像素比例:         $PIXEL_RATIO 米/像素"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 检测到缺陷数:     $TOTAL_DEFECTS"
echo "│ 算法处理时间:     ${PROCESSING_TIME}s"
echo "│ 请求总耗时:       ${REQUEST_TIME}ms"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│ 原始图片路径:     $IMAGE_PATH"
echo "│ 结果图片路径:     $RESULT_PATH"
echo "│ 持久化状态:       $PERSISTENCE_STATUS"
echo "└─────────────────────────────────────────────────────────────┘"
echo ""

# ============ 步骤 9: 验证结果 ============
log_info "=========================================="
log_info "步骤 9: 验证结果"
log_info "=========================================="

VALIDATION_PASSED=true

# 验证1: 响应码正确
if [ "$RESPONSE_CODE" = "0" ]; then
    log_success "✓ 响应码正确 (code=0)"
else
    log_error "✗ 响应码错误 (code=$RESPONSE_CODE)"
    VALIDATION_PASSED=false
fi

# 验证2: 返回了缺陷数量字段
if [ -n "$TOTAL_DEFECTS" ]; then
    log_success "✓ 返回了缺陷数量字段 (total_defects=$TOTAL_DEFECTS)"
else
    log_error "✗ 缺少缺陷数量字段"
    VALIDATION_PASSED=false
fi

# 验证3: 返回了图片路径
if [ "$IMAGE_PATH" != "N/A" ] && [ "$IMAGE_PATH" != "null" ]; then
    log_success "✓ 返回了原始图片路径"
else
    log_error "✗ 缺少原始图片路径"
    VALIDATION_PASSED=false
fi

# 验证4: 返回了结果图片路径
if [ "$RESULT_PATH" != "N/A" ] && [ "$RESULT_PATH" != "null" ]; then
    log_success "✓ 返回了结果图片路径"
else
    log_error "✗ 缺少结果图片路径"
    VALIDATION_PASSED=false
fi

# 验证5: 处理时间合理（< 10秒）
if [ "$PROCESSING_TIME" != "0" ]; then
    PROCESSING_TIME_INT=$(echo "$PROCESSING_TIME" | awk '{print int($1)}')
    if [ "$PROCESSING_TIME_INT" -lt 10 ]; then
        log_success "✓ 处理时间合理 (${PROCESSING_TIME}s < 10s)"
    else
        log_warn "⚠ 处理时间较长 (${PROCESSING_TIME}s)"
    fi
else
    log_warn "⚠ 处理时间为0，可能未正确记录"
fi

# 验证6: 如果检测到缺陷，验证缺陷数据结构
if [ "$TOTAL_DEFECTS" -gt 0 ]; then
    FIRST_DEFECT=$(echo "$DEFECTS" | jq -r '.[0] // empty')
    if [ -n "$FIRST_DEFECT" ]; then
        HAS_TYPE=$(echo "$FIRST_DEFECT" | jq -r '.defect_type // empty')
        HAS_BBOX=$(echo "$FIRST_DEFECT" | jq -r '.bbox // empty')
        HAS_CONFIDENCE=$(echo "$FIRST_DEFECT" | jq -r '.confidence // empty')

        if [ -n "$HAS_TYPE" ] && [ -n "$HAS_BBOX" ] && [ -n "$HAS_CONFIDENCE" ]; then
            log_success "✓ 缺陷数据结构完整 (包含 type, bbox, confidence)"
        else
            log_error "✗ 缺陷数据结构不完整"
            VALIDATION_PASSED=false
        fi
    fi
fi

echo ""

if [ "$VALIDATION_PASSED" = true ]; then
    log_success "=========================================="
    log_success "所有验证通过！图片检测链路工作正常 ✓"
    log_success "=========================================="
    exit 0
else
    log_error "=========================================="
    log_error "部分验证失败，请检查日志"
    log_error "=========================================="
    exit 1
fi
