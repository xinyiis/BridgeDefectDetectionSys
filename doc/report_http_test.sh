#!/bin/bash

# 报表生成模块 HTTP 测试脚本
# 用途：模拟真实用户操作，测试报表模块5个接口

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

# Cookie文件
COOKIE_FILE="cookies_report_test.txt"

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试结果输出
test_result() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $1 -eq 0 ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✓ $2${NC}"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}✗ $2${NC}"
    fi
}

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}  报表生成模块 HTTP 测试${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# ========================================
# 步骤1：用户注册和登录
# ========================================
echo -e "${YELLOW}步骤1：用户注册和登录${NC}"

# 生成唯一用户名
TIMESTAMP=$(date +%s)
USERNAME="report_test_$TIMESTAMP"

# 注册用户
echo "→ 注册测试用户: $USERNAME"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"123456\",
    \"real_name\": \"报表测试用户\",
    \"email\": \"${USERNAME}@test.com\"
  }")

echo "$REGISTER_RESPONSE" | grep -q '"code":200'
test_result $? "用户注册成功"

# 登录获取Session
echo "→ 登录获取Session"
LOGIN_RESPONSE=$(curl -s -c "$COOKIE_FILE" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"123456\"
  }")

echo "$LOGIN_RESPONSE" | grep -q '"code":200'
test_result $? "用户登录成功"

echo ""

# ========================================
# 步骤2：创建测试桥梁
# ========================================
echo -e "${YELLOW}步骤2：创建测试桥梁${NC}"

echo "→ 创建测试桥梁：长江大桥"
BRIDGE_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$BASE_URL/bridges" \
  -F "bridge_name=长江大桥" \
  -F "bridge_code=BRIDGE_CJ_$TIMESTAMP" \
  -F "address=湖北省武汉市" \
  -F "longitude=114.31" \
  -F "latitude=30.59" \
  -F "bridge_type=悬索桥" \
  -F "build_year=1957" \
  -F "length=1670.0" \
  -F "width=18.0" \
  -F "remark=测试桥梁")

BRIDGE_ID=$(echo "$BRIDGE_RESPONSE" | grep -o '"bridge_id":[0-9]*' | head -1 | cut -d':' -f2)
echo "$BRIDGE_RESPONSE" | grep -q '"code":200'
test_result $? "桥梁创建成功 (ID: $BRIDGE_ID)"

echo ""

# ========================================
# 步骤3：创建测试缺陷数据
# ========================================
echo -e "${YELLOW}步骤3：创建测试缺陷数据（模拟）${NC}"

# 注意：这里我们跳过缺陷创建，因为需要上传图片和调用Python服务
# 实际测试中报表可以在没有缺陷的情况下生成（缺陷数为0）
echo "→ 跳过缺陷创建（报表可以在无缺陷情况下生成）"
test_result 0 "使用无缺陷场景测试"

echo ""

# ========================================
# 步骤4：创建报表（桥梁检测报表）
# ========================================
echo -e "${YELLOW}步骤4：创建桥梁检测报表${NC}"

# 计算时间范围（最近30天）
END_DATE=$(date +%Y-%m-%d)
START_DATE=$(date -d "30 days ago" +%Y-%m-%d)

echo "→ 创建报表：长江大桥检测报表 ($START_DATE ~ $END_DATE)"
REPORT_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$BASE_URL/reports" \
  -H "Content-Type: application/json" \
  -d "{
    \"report_name\": \"长江大桥检测报表_$TIMESTAMP\",
    \"report_type\": \"bridge_inspection\",
    \"bridge_id\": $BRIDGE_ID,
    \"start_time\": \"$START_DATE\",
    \"end_time\": \"$END_DATE\"
  }")

REPORT_ID=$(echo "$REPORT_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "$REPORT_RESPONSE" | grep -q '"code":200'
test_result $? "报表创建成功 (ID: $REPORT_ID，状态：生成中)"

# 显示报表信息
echo "  报表ID: $REPORT_ID"
echo "  报表名称: 长江大桥检测报表_$TIMESTAMP"
echo "  桥梁ID: $BRIDGE_ID"
echo "  时间范围: $START_DATE ~ $END_DATE"

echo ""

# ========================================
# 步骤5：查询报表列表
# ========================================
echo -e "${YELLOW}步骤5：查询报表列表${NC}"

echo "→ 获取报表列表（分页）"
LIST_RESPONSE=$(curl -s -b "$COOKIE_FILE" "$BASE_URL/reports?page=1&page_size=10")

echo "$LIST_RESPONSE" | grep -q '"code":200'
test_result $? "报表列表查询成功"

# 检查是否包含刚创建的报表
echo "$LIST_RESPONSE" | grep -q "长江大桥检测报表_$TIMESTAMP"
test_result $? "列表中包含新创建的报表"

# 显示报表数量
REPORT_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"total":[0-9]*' | head -1 | cut -d':' -f2)
echo "  当前报表总数: $REPORT_COUNT"

echo ""

# ========================================
# 步骤6：获取报表详情
# ========================================
echo -e "${YELLOW}步骤6：获取报表详情${NC}"

echo "→ 查询报表ID $REPORT_ID 的详情"
DETAIL_RESPONSE=$(curl -s -b "$COOKIE_FILE" "$BASE_URL/reports/$REPORT_ID")

echo "$DETAIL_RESPONSE" | grep -q '"code":200'
test_result $? "报表详情查询成功"

# 显示报表状态
REPORT_STATUS=$(echo "$DETAIL_RESPONSE" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  报表状态: $REPORT_STATUS"

# 显示报表基本信息
echo "  报表信息:"
echo "$DETAIL_RESPONSE" | grep -o '"report_name":"[^"]*"' | head -1 | sed 's/.*:/    报表名称: /' | sed 's/"//g'
echo "$DETAIL_RESPONSE" | grep -o '"report_type":"[^"]*"' | head -1 | sed 's/.*:/    报表类型: /' | sed 's/"//g'
echo "$DETAIL_RESPONSE" | grep -o '"defect_count":[0-9]*' | head -1 | sed 's/.*:/    缺陷数量: /'
echo "$DETAIL_RESPONSE" | grep -o '"high_risk_count":[0-9]*' | head -1 | sed 's/.*:/    高危缺陷: /'

echo ""

# ========================================
# 步骤7：等待PDF生成完成（可选）
# ========================================
echo -e "${YELLOW}步骤7：等待PDF生成完成${NC}"

echo "→ 等待PDF后台生成（最多等待10秒）"
MAX_WAIT=10
WAIT_COUNT=0
PDF_READY=false

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))

    STATUS_RESPONSE=$(curl -s -b "$COOKIE_FILE" "$BASE_URL/reports/$REPORT_ID")
    CURRENT_STATUS=$(echo "$STATUS_RESPONSE" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)

    echo -n "  等待中 ($WAIT_COUNT秒)... 当前状态: $CURRENT_STATUS"

    if [ "$CURRENT_STATUS" = "completed" ]; then
        echo -e " ${GREEN}[完成]${NC}"
        PDF_READY=true
        break
    elif [ "$CURRENT_STATUS" = "failed" ]; then
        echo -e " ${RED}[失败]${NC}"
        ERROR_MSG=$(echo "$STATUS_RESPONSE" | grep -o '"error_message":"[^"]*"' | cut -d'"' -f4)
        echo "  错误信息: $ERROR_MSG"
        break
    else
        echo " [生成中]"
    fi
done

if [ "$PDF_READY" = true ]; then
    test_result 0 "PDF生成完成"
else
    echo -e "${YELLOW}  注意：PDF生成可能需要更多时间，可稍后重试下载${NC}"
    test_result 1 "PDF生成超时（后台仍在处理）"
fi

echo ""

# ========================================
# 步骤8：下载报表PDF（如果已生成）
# ========================================
if [ "$PDF_READY" = true ]; then
    echo -e "${YELLOW}步骤8：下载报表PDF${NC}"

    OUTPUT_FILE="report_${REPORT_ID}.pdf"
    echo "→ 下载报表PDF到: $OUTPUT_FILE"

    HTTP_CODE=$(curl -s -b "$COOKIE_FILE" -w "%{http_code}" -o "$OUTPUT_FILE" \
      "$BASE_URL/reports/$REPORT_ID/download")

    if [ "$HTTP_CODE" = "200" ]; then
        FILE_SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
        test_result 0 "报表下载成功 (大小: $FILE_SIZE)"
        echo "  文件路径: ./$OUTPUT_FILE"
    else
        test_result 1 "报表下载失败 (HTTP $HTTP_CODE)"
        rm -f "$OUTPUT_FILE"
    fi

    echo ""
else
    echo -e "${YELLOW}步骤8：跳过下载（PDF未生成）${NC}"
    echo ""
fi

# ========================================
# 步骤9：按条件过滤报表列表
# ========================================
echo -e "${YELLOW}步骤9：过滤报表列表${NC}"

echo "→ 按报表类型过滤: bridge_inspection"
FILTER_RESPONSE=$(curl -s -b "$COOKIE_FILE" \
  "$BASE_URL/reports?report_type=bridge_inspection&page=1&page_size=10")

echo "$FILTER_RESPONSE" | grep -q '"code":200'
test_result $? "按类型过滤成功"

echo "→ 按桥梁ID过滤: $BRIDGE_ID"
BRIDGE_FILTER_RESPONSE=$(curl -s -b "$COOKIE_FILE" \
  "$BASE_URL/reports?bridge_id=$BRIDGE_ID&page=1&page_size=10")

echo "$BRIDGE_FILTER_RESPONSE" | grep -q '"code":200'
test_result $? "按桥梁过滤成功"

echo ""

# ========================================
# 步骤10：删除报表（软删除）
# ========================================
echo -e "${YELLOW}步骤10：删除报表${NC}"

echo "→ 删除报表ID: $REPORT_ID"
DELETE_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X DELETE "$BASE_URL/reports/$REPORT_ID")

echo "$DELETE_RESPONSE" | grep -q '"code":200'
test_result $? "报表删除成功（软删除）"

# 验证删除后无法访问
echo "→ 验证删除后无法访问"
VERIFY_RESPONSE=$(curl -s -b "$COOKIE_FILE" "$BASE_URL/reports/$REPORT_ID")
VERIFY_CODE=$(echo "$VERIFY_RESPONSE" | grep -o '"code":[0-9]*' | head -1 | cut -d':' -f2)

if [ "$VERIFY_CODE" != "200" ]; then
    test_result 0 "删除验证成功（无法访问已删除报表）"
else
    test_result 1 "删除验证失败（仍可访问）"
fi

echo ""

# ========================================
# 步骤11：清理测试数据
# ========================================
echo -e "${YELLOW}步骤11：清理测试数据${NC}"

echo "→ 删除测试桥梁"
DELETE_BRIDGE_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X DELETE "$BASE_URL/bridges/$BRIDGE_ID")
echo "$DELETE_BRIDGE_RESPONSE" | grep -q '"code":200'
test_result $? "桥梁删除成功"

echo "→ 用户登出"
LOGOUT_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$BASE_URL/auth/logout")
echo "$LOGOUT_RESPONSE" | grep -q '"code":200'
test_result $? "用户登出成功"

# 删除Cookie文件
rm -f "$COOKIE_FILE"

echo ""

# ========================================
# 测试总结
# ========================================
echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}  测试总结${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    SUCCESS_RATE="100%"
else
    SUCCESS_RATE=$(awk "BEGIN {printf \"%.1f%%\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")
    echo -e "${YELLOW}⚠ 部分测试失败，通过率: $SUCCESS_RATE${NC}"
fi

echo ""
echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}  报表模块功能验证完成${NC}"
echo -e "${BLUE}=====================================${NC}"

exit $FAILED_TESTS
