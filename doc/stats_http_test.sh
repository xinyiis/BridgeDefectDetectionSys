#!/bin/bash

# 统计模块 HTTP 请求测试脚本
# 使用方法：
#   1. 启动后端服务：cd src/backend && go run cmd/server/main.go
#   2. 运行此脚本：bash doc/stats_http_test.sh

# 配置
BASE_URL="http://localhost:8080/api/v1"
COOKIE_FILE="cookies.txt"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0

# 打印测试标题
print_test() {
    echo -e "\n${BLUE}========== $1 ==========${NC}"
    ((TOTAL_TESTS++))
}

# 检查响应
check_response() {
    local response="$1"
    local expected_code="$2"
    local test_name="$3"

    local code=$(echo "$response" | jq -r '.code')

    if [ "$code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ $test_name PASSED${NC}"
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}✗ $test_name FAILED (expected code $expected_code, got $code)${NC}"
        echo "Response: $response"
        return 1
    fi
}

# 清理
cleanup() {
    rm -f "$COOKIE_FILE"
}

trap cleanup EXIT

echo -e "${BLUE}统计模块 HTTP 请求测试${NC}"
echo "========================================"

# 1. 登录获取 Session
print_test "用户登录"
LOGIN_RESPONSE=$(curl -s -c "$COOKIE_FILE" -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "password": "123456"
  }')

check_response "$LOGIN_RESPONSE" "200" "用户登录"

if [ ! -f "$COOKIE_FILE" ]; then
    echo -e "${RED}登录失败，无法获取 Session${NC}"
    exit 1
fi

# 2. 测试概览统计
print_test "获取概览统计 (GET /stats/overview)"
OVERVIEW_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/overview")
check_response "$OVERVIEW_RESPONSE" "200" "概览统计"

echo "概览数据："
echo "$OVERVIEW_RESPONSE" | jq '.data'

# 3. 测试缺陷类型分布
print_test "获取缺陷类型分布 (GET /stats/defect-types)"
TYPE_DIST_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/defect-types?days=30")
check_response "$TYPE_DIST_RESPONSE" "200" "缺陷类型分布"

echo "类型分布（前3种）："
echo "$TYPE_DIST_RESPONSE" | jq '.data.distribution | .[0:3]'

# 4. 测试缺陷趋势统计
print_test "获取缺陷趋势统计 (GET /stats/defect-trend)"
TREND_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/defect-trend?days=7&granularity=day")
check_response "$TREND_RESPONSE" "200" "缺陷趋势统计"

echo "趋势统计摘要："
echo "$TREND_RESPONSE" | jq '{period: .data.period, total: .data.total, avg_per_day: .data.avg_per_day}'

# 5. 测试桥梁健康度排名
print_test "获取桥梁健康度排名 (GET /stats/bridge-ranking)"
RANKING_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/bridge-ranking?limit=10&order=worst")
check_response "$RANKING_RESPONSE" "200" "桥梁健康度排名"

echo "排名（前3座桥梁）："
echo "$RANKING_RESPONSE" | jq '.data.ranking | .[0:3] | .[] | {bridge_name, defect_count, health_score, health_level}'

# 6. 测试最近检测记录
print_test "获取最近检测记录 (GET /stats/recent-detections)"
RECENT_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/recent-detections?limit=10")
check_response "$RECENT_RESPONSE" "200" "最近检测记录"

echo "最近检测（前3条）："
echo "$RECENT_RESPONSE" | jq '.data.detections | .[0:3] | .[] | {bridge_name, defect_count, detected_at}'

# 7. 测试高危缺陷告警
print_test "获取高危缺陷告警 (GET /stats/high-risk-alerts)"
ALERTS_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/high-risk-alerts?limit=20")
check_response "$ALERTS_RESPONSE" "200" "高危缺陷告警"

echo "高危告警（前3条）："
echo "$ALERTS_RESPONSE" | jq '.data.alerts | .[0:3] | .[] | {bridge_name, defect_type, confidence, area, severity}'

# 8. 测试高危告警严重程度过滤
print_test "获取紧急级别告警 (GET /stats/high-risk-alerts?severity=urgent)"
URGENT_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/high-risk-alerts?severity=urgent&limit=20")
check_response "$URGENT_RESPONSE" "200" "紧急级别告警"

echo "紧急告警数量："
echo "$URGENT_RESPONSE" | jq '.data.total'

# 9. 测试参数验证（无效days参数）
print_test "参数验证 - 超大天数 (days=1000)"
INVALID_DAYS_RESPONSE=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/defect-types?days=1000")
# 应该被限制为最大365天，但仍然返回200
check_response "$INVALID_DAYS_RESPONSE" "200" "参数验证（大天数）"

# 10. 测试缓存（连续两次请求应该很快）
print_test "测试缓存机制（连续两次请求）"
START_TIME=$(date +%s%3N)
CACHE_TEST_1=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/overview")
END_TIME_1=$(date +%s%3N)
TIME_1=$((END_TIME_1 - START_TIME))

START_TIME_2=$(date +%s%3N)
CACHE_TEST_2=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/overview")
END_TIME_2=$(date +%s%3N)
TIME_2=$((END_TIME_2 - START_TIME_2))

echo "第一次请求耗时: ${TIME_1}ms"
echo "第二次请求耗时: ${TIME_2}ms (应该更快，因为缓存)"

if [ "$TIME_2" -lt "$TIME_1" ] || [ "$TIME_2" -lt 100 ]; then
    echo -e "${GREEN}✓ 缓存机制工作正常${NC}"
    ((PASSED_TESTS++))
else
    echo -e "${RED}✗ 缓存可能未生效${NC}"
fi

# 总结
echo -e "\n${BLUE}========== 测试总结 ==========${NC}"
echo "总测试数: $TOTAL_TESTS"
echo "通过数: $PASSED_TESTS"
echo "失败数: $((TOTAL_TESTS - PASSED_TESTS))"

if [ "$PASSED_TESTS" -eq "$TOTAL_TESTS" ]; then
    echo -e "${GREEN}所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}部分测试失败${NC}"
    exit 1
fi
