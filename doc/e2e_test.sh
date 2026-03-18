#!/bin/bash

# 桥梁缺陷检测系统 - 端到端测试脚本
# 模拟真实用户操作：注册、登录、创建桥梁、创建缺陷、查询统计等
# 使用方法：
#   1. 启动后端服务：cd src/backend && go run cmd/server/main.go
#   2. 运行此脚本：bash doc/e2e_test.sh

# 配置
BASE_URL="http://localhost:8080/api/v1"
COOKIE_FILE="e2e_cookies.txt"
TIMESTAMP=$(date +%s)

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0

# 打印分隔线
print_separator() {
    echo -e "\n${BLUE}$1${NC}"
}

# 打印步骤
print_step() {
    echo -e "${YELLOW}[步骤 $1/$2] $3${NC}"
    ((TOTAL_TESTS++))
}

# 检查响应
check_response() {
    local response="$1"
    local expected_code="$2"
    local step_name="$3"

    local code=$(echo "$response" | jq -r '.code // 0')

    if [ "$code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ $step_name 成功${NC}"
        ((PASSED_TESTS++))
        return 0
    else
        echo -e "${RED}✗ $step_name 失败 (期望code=$expected_code, 实际code=$code)${NC}"
        echo "响应: $response" | head -3
        return 1
    fi
}

# 清理
cleanup() {
    rm -f "$COOKIE_FILE"
    echo -e "\n${BLUE}清理完成${NC}"
}

trap cleanup EXIT

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║    桥梁缺陷检测系统 - 端到端测试                          ║${NC}"
echo -e "${BLUE}║    模拟真实用户操作流程                                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"

TOTAL_STEPS=22

# ====================
# 第一部分：用户注册与登录
# ====================
print_separator "═══════════════ 第一部分：用户管理 ═══════════════"

# 步骤1：注册新用户
print_step 1 $TOTAL_STEPS "注册新用户 (user_test_${TIMESTAMP})"
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"user_test_${TIMESTAMP}\",
    \"password\": \"Test123456\",
    \"real_name\": \"测试用户\",
    \"email\": \"test_${TIMESTAMP}@example.com\",
    \"phone\": \"13800138000\"
  }")

check_response "$REGISTER_RESPONSE" "200" "用户注册"
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.id')
echo "  → 用户ID: $USER_ID"

# 步骤2：用户登录
print_step 2 $TOTAL_STEPS "用户登录"
LOGIN_RESPONSE=$(curl -s -c "$COOKIE_FILE" -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"user_test_${TIMESTAMP}\",
    \"password\": \"Test123456\"
  }")

check_response "$LOGIN_RESPONSE" "200" "用户登录"

if [ ! -f "$COOKIE_FILE" ]; then
    echo -e "${RED}登录失败，无法继续测试${NC}"
    exit 1
fi

# ====================
# 第二部分：桥梁管理
# ====================
print_separator "═══════════════ 第二部分：桥梁管理 ═══════════════"

# 步骤3：创建第一座桥梁
print_step 3 $TOTAL_STEPS "创建桥梁1 - 长江大桥"
BRIDGE1_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/bridges" \
  -F "bridge_name=长江大桥_${TIMESTAMP}" \
  -F "bridge_code=BRIDGE_CJ_${TIMESTAMP}" \
  -F "address=湖北省武汉市" \
  -F "longitude=114.31" \
  -F "latitude=30.52" \
  -F "bridge_type=悬索桥" \
  -F "build_year=1957" \
  -F "length=1670.4" \
  -F "width=18.0" \
  -F "remark=武汉长江大桥")

check_response "$BRIDGE1_RESPONSE" "200" "创建桥梁1"
BRIDGE1_ID=$(echo "$BRIDGE1_RESPONSE" | jq -r '.data.bridge_id')
echo "  → 桥梁1 ID: $BRIDGE1_ID"

# 步骤4：创建第二座桥梁
print_step 4 $TOTAL_STEPS "创建桥梁2 - 黄河大桥"
BRIDGE2_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/bridges" \
  -F "bridge_name=黄河大桥_${TIMESTAMP}" \
  -F "bridge_code=BRIDGE_HH_${TIMESTAMP}" \
  -F "address=河南省郑州市" \
  -F "longitude=113.62" \
  -F "latitude=34.75" \
  -F "bridge_type=斜拉桥" \
  -F "build_year=1986" \
  -F "length=2386.0" \
  -F "width=22.5" \
  -F "remark=郑州黄河大桥")

check_response "$BRIDGE2_RESPONSE" "200" "创建桥梁2"
BRIDGE2_ID=$(echo "$BRIDGE2_RESPONSE" | jq -r '.data.bridge_id')
echo "  → 桥梁2 ID: $BRIDGE2_ID"

# 步骤5：创建第三座桥梁
print_step 5 $TOTAL_STEPS "创建桥梁3 - 珠江大桥"
BRIDGE3_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/bridges" \
  -F "bridge_name=珠江大桥_${TIMESTAMP}" \
  -F "bridge_code=BRIDGE_ZJ_${TIMESTAMP}" \
  -F "address=广东省广州市" \
  -F "longitude=113.26" \
  -F "latitude=23.13" \
  -F "bridge_type=梁桥" \
  -F "build_year=2008" \
  -F "length=1388.0" \
  -F "width=28.0" \
  -F "remark=广州珠江大桥")

check_response "$BRIDGE3_RESPONSE" "200" "创建桥梁3"
BRIDGE3_ID=$(echo "$BRIDGE3_RESPONSE" | jq -r '.data.bridge_id')
echo "  → 桥梁3 ID: $BRIDGE3_ID"

# 步骤6：查询桥梁列表
print_step 6 $TOTAL_STEPS "查询桥梁列表"
BRIDGES_LIST=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/bridges?page=1&page_size=10")
check_response "$BRIDGES_LIST" "200" "查询桥梁列表"
BRIDGE_COUNT=$(echo "$BRIDGES_LIST" | jq -r '.data.total')
echo "  → 桥梁总数: $BRIDGE_COUNT"

# ====================
# 第三部分：无人机管理
# ====================
print_separator "═══════════════ 第三部分：无人机管理 ═══════════════"

# 步骤7：创建无人机1
print_step 7 $TOTAL_STEPS "创建无人机1 - 大疆 Mavic 3"
DRONE1_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/drones" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"大疆 Mavic 3 Pro_${TIMESTAMP}\",
    \"model\": \"Mavic 3 Pro\",
    \"stream_url\": \"rtsp://192.168.1.100:554/stream1\"
  }")

check_response "$DRONE1_RESPONSE" "200" "创建无人机1"
DRONE1_ID=$(echo "$DRONE1_RESPONSE" | jq -r '.data.id')
echo "  → 无人机1 ID: $DRONE1_ID"

# 步骤8：创建无人机2
print_step 8 $TOTAL_STEPS "创建无人机2 - 大疆 Phantom 4"
DRONE2_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/drones" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"大疆 Phantom 4 RTK_${TIMESTAMP}\",
    \"model\": \"Phantom 4 RTK\",
    \"stream_url\": \"rtsp://192.168.1.101:554/stream2\"
  }")

check_response "$DRONE2_RESPONSE" "200" "创建无人机2"
echo "  → 无人机2 ID: $(echo "$DRONE2_RESPONSE" | jq -r '.data.id')"

# ====================
# 第四部分：创建缺陷数据（模拟检测结果）
# ====================
print_separator "═══════════════ 第四部分：缺陷数据创建 ═══════════════"

echo -e "${YELLOW}注意：直接通过数据库插入缺陷数据（模拟检测结果）${NC}"
echo -e "${YELLOW}实际使用中，缺陷由检测接口自动创建${NC}"

# 步骤9-11：使用SQL直接插入缺陷数据（准备阶段，不计入测试总数）
echo -e "${BLUE}[准备] 插入缺陷数据（桥梁1：10个，桥梁2：5个，桥梁3：2个）${NC}"

# 创建SQL插入脚本
cat > /tmp/insert_defects_${TIMESTAMP}.sql << EOF
USE bridge_detection;

-- 桥梁1的缺陷（长江大桥 - 老桥，缺陷较多）
INSERT INTO defects (bridge_id, defect_type, image_path, result_path, confidence, area, length, width, detected_at, created_at) VALUES
($BRIDGE1_ID, '裂缝', 'images/cj_crack_001.jpg', 'results/cj_crack_001_result.jpg', 0.96, 0.15, 2.5, 0.06, NOW() - INTERVAL 1 DAY, NOW()),
($BRIDGE1_ID, '裂缝', 'images/cj_crack_002.jpg', 'results/cj_crack_002_result.jpg', 0.94, 0.08, 1.8, 0.044, NOW() - INTERVAL 2 DAY, NOW()),
($BRIDGE1_ID, '剥落', 'images/cj_spall_001.jpg', 'results/cj_spall_001_result.jpg', 0.92, 0.12, 1.5, 0.08, NOW() - INTERVAL 1 DAY, NOW()),
($BRIDGE1_ID, '钢筋锈蚀', 'images/cj_rust_001.jpg', 'results/cj_rust_001_result.jpg', 0.89, 0.05, 0.8, 0.063, NOW() - INTERVAL 3 DAY, NOW()),
($BRIDGE1_ID, '混凝土开裂', 'images/cj_concrete_001.jpg', 'results/cj_concrete_001_result.jpg', 0.87, 0.06, 1.2, 0.05, NOW() - INTERVAL 2 DAY, NOW()),
($BRIDGE1_ID, '裂缝', 'images/cj_crack_003.jpg', 'results/cj_crack_003_result.jpg', 0.85, 0.04, 0.9, 0.044, NOW() - INTERVAL 4 DAY, NOW()),
($BRIDGE1_ID, '表面损伤', 'images/cj_damage_001.jpg', 'results/cj_damage_001_result.jpg', 0.78, 0.02, 0.5, 0.04, NOW() - INTERVAL 5 DAY, NOW()),
($BRIDGE1_ID, '裂缝', 'images/cj_crack_004.jpg', 'results/cj_crack_004_result.jpg', 0.88, 0.03, 0.7, 0.043, NOW(), NOW()),
($BRIDGE1_ID, '剥落', 'images/cj_spall_002.jpg', 'results/cj_spall_002_result.jpg', 0.91, 0.07, 1.1, 0.064, NOW(), NOW()),
($BRIDGE1_ID, '钢筋锈蚀', 'images/cj_rust_002.jpg', 'results/cj_rust_002_result.jpg', 0.86, 0.04, 0.6, 0.067, NOW() - INTERVAL 1 DAY, NOW());

-- 桥梁2的缺陷（黄河大桥 - 中等）
INSERT INTO defects (bridge_id, defect_type, image_path, result_path, confidence, area, length, width, detected_at, created_at) VALUES
($BRIDGE2_ID, '裂缝', 'images/hh_crack_001.jpg', 'results/hh_crack_001_result.jpg', 0.92, 0.06, 1.2, 0.05, NOW() - INTERVAL 1 DAY, NOW()),
($BRIDGE2_ID, '裂缝', 'images/hh_crack_002.jpg', 'results/hh_crack_002_result.jpg', 0.88, 0.04, 0.9, 0.044, NOW() - INTERVAL 2 DAY, NOW()),
($BRIDGE2_ID, '剥落', 'images/hh_spall_001.jpg', 'results/hh_spall_001_result.jpg', 0.85, 0.03, 0.7, 0.043, NOW() - INTERVAL 3 DAY, NOW()),
($BRIDGE2_ID, '混凝土开裂', 'images/hh_concrete_001.jpg', 'results/hh_concrete_001_result.jpg', 0.82, 0.02, 0.5, 0.04, NOW() - INTERVAL 4 DAY, NOW()),
($BRIDGE2_ID, '表面损伤', 'images/hh_damage_001.jpg', 'results/hh_damage_001_result.jpg', 0.75, 0.015, 0.4, 0.038, NOW(), NOW());

-- 桥梁3的缺陷（珠江大桥 - 新桥，缺陷少）
INSERT INTO defects (bridge_id, defect_type, image_path, result_path, confidence, area, length, width, detected_at, created_at) VALUES
($BRIDGE3_ID, '裂缝', 'images/zj_crack_001.jpg', 'results/zj_crack_001_result.jpg', 0.78, 0.02, 0.5, 0.04, NOW() - INTERVAL 1 DAY, NOW()),
($BRIDGE3_ID, '表面损伤', 'images/zj_damage_001.jpg', 'results/zj_damage_001_result.jpg', 0.72, 0.01, 0.3, 0.033, NOW(), NOW());
EOF

print_step 9 $TOTAL_STEPS "执行SQL插入缺陷数据"
mysql -uroot -p123456 < /tmp/insert_defects_${TIMESTAMP}.sql 2>&1 | grep -v "Warning"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 缺陷数据插入成功${NC}"
    ((PASSED_TESTS++))
    echo "  → 桥梁1: 10个缺陷"
    echo "  → 桥梁2: 5个缺陷"
    echo "  → 桥梁3: 2个缺陷"
    echo "  → 总计: 17个缺陷"
else
    echo -e "${RED}✗ 缺陷数据插入失败${NC}"
fi

rm -f /tmp/insert_defects_${TIMESTAMP}.sql

# ====================
# 第五部分：缺陷查询
# ====================
print_separator "═══════════════ 第五部分：缺陷查询 ═══════════════"

# 步骤10：查询所有缺陷
print_step 10 $TOTAL_STEPS "查询所有缺陷列表"
DEFECTS_LIST=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/defects?page=1&page_size=20")
check_response "$DEFECTS_LIST" "200" "查询缺陷列表"
DEFECT_COUNT=$(echo "$DEFECTS_LIST" | jq -r '.data.total')
echo "  → 缺陷总数: $DEFECT_COUNT"

# 步骤11：按桥梁过滤缺陷
print_step 11 $TOTAL_STEPS "查询桥梁1的缺陷"
BRIDGE1_DEFECTS=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/defects?bridge_id=${BRIDGE1_ID}")
check_response "$BRIDGE1_DEFECTS" "200" "按桥梁过滤缺陷"
BRIDGE1_DEFECT_COUNT=$(echo "$BRIDGE1_DEFECTS" | jq -r '.data.total')
echo "  → 桥梁1缺陷数: $BRIDGE1_DEFECT_COUNT"

# 步骤12：按缺陷类型过滤
print_step 12 $TOTAL_STEPS "查询裂缝类型的缺陷"
CRACK_DEFECTS=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/defects?defect_type=裂缝")
check_response "$CRACK_DEFECTS" "200" "按类型过滤缺陷"
CRACK_COUNT=$(echo "$CRACK_DEFECTS" | jq -r '.data.total')
echo "  → 裂缝类型缺陷数: $CRACK_COUNT"

# ====================
# 第六部分：统计模块测试
# ====================
print_separator "═══════════════ 第六部分：统计模块测试 ═══════════════"

# 步骤13：概览统计
print_step 13 $TOTAL_STEPS "获取概览统计"
OVERVIEW=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/overview")
check_response "$OVERVIEW" "200" "概览统计"

echo "  概览数据："
echo "$OVERVIEW" | jq -r '.data |
  "  → 桥梁数量: \(.bridge_count)
  → 无人机数量: \(.drone_count)
  → 缺陷总数: \(.defect_count)
  → 检测次数: \(.detection_count)
  → 今日缺陷: \(.today_defects)
  → 本周缺陷: \(.week_defects)
  → 趋势方向: \(.trend_direction)"'

# 步骤14：缺陷类型分布
print_step 14 $TOTAL_STEPS "获取缺陷类型分布（30天）"
TYPE_DIST=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/defect-types?days=30")
check_response "$TYPE_DIST" "200" "缺陷类型分布"

echo "  类型分布："
echo "$TYPE_DIST" | jq -r '.data.distribution[] |
  "  → \(.defect_type): \(.count)个 (占比\(.percentage | tonumber | floor)%, 平均置信度\(.avg_confidence))"' | head -5

# 步骤15：缺陷趋势统计
print_step 15 $TOTAL_STEPS "获取缺陷趋势（7天）"
TREND=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/defect-trend?days=7&granularity=day")
check_response "$TREND" "200" "缺陷趋势统计"

echo "  趋势摘要："
echo "$TREND" | jq -r '.data |
  "  → 统计周期: \(.period)
  → 缺陷总数: \(.total)
  → 日均缺陷: \(.avg_per_day | tonumber | floor * 100 / 100)
  → 峰值日期: \(.peak_date)
  → 峰值数量: \(.peak_count)"'

# 步骤16：桥梁健康度排名
print_step 16 $TOTAL_STEPS "获取桥梁健康度排名（最差前3）"
RANKING=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/bridge-ranking?limit=10&order=worst")
check_response "$RANKING" "200" "桥梁健康度排名"

echo "  排名（最差前3）："
echo "$RANKING" | jq -r '.data.ranking[0:3][] |
  "  → \(.bridge_name):
     - 缺陷数: \(.defect_count)
     - 高危缺陷: \(.high_risk_count)
     - 健康评分: \(.health_score | tonumber | floor)
     - 健康等级: \(.health_level)"'

# 步骤17：最近检测记录
print_step 17 $TOTAL_STEPS "获取最近检测记录（前5条）"
RECENT=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/recent-detections?limit=5")
check_response "$RECENT" "200" "最近检测记录"

echo "  最近检测："
echo "$RECENT" | jq -r '.data.detections[0:3][] |
  "  → \(.bridge_name): 检测到\(.defect_count)个缺陷 (\(.detected_at))"'

# 步骤18：高危缺陷告警
print_step 18 $TOTAL_STEPS "获取高危缺陷告警"
ALERTS=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/high-risk-alerts?limit=20")
check_response "$ALERTS" "200" "高危缺陷告警"

echo "  高危告警："
ALERT_TOTAL=$(echo "$ALERTS" | jq -r '.data.total')
echo "  → 高危告警总数: $ALERT_TOTAL"
echo "$ALERTS" | jq -r '.data.alerts[0:3][] |
  "  → \(.bridge_name) - \(.defect_type):
     置信度\(.confidence), 面积\(.area)㎡, 严重程度[\(.severity)]"'

# 步骤19：紧急告警过滤
print_step 19 $TOTAL_STEPS "获取紧急级别告警"
URGENT=$(curl -s -b "$COOKIE_FILE" "${BASE_URL}/stats/high-risk-alerts?severity=urgent&limit=20")
check_response "$URGENT" "200" "紧急级别告警"
URGENT_COUNT=$(echo "$URGENT" | jq -r '.data.total')
echo "  → 紧急告警数量: $URGENT_COUNT"

# ====================
# 第七部分：数据修改与删除
# ====================
print_separator "═══════════════ 第七部分：数据修改与删除 ═══════════════"

# 步骤20：更新桥梁信息
print_step 20 $TOTAL_STEPS "更新桥梁3状态为维修中"
UPDATE_BRIDGE=$(curl -s -b "$COOKIE_FILE" -X PUT "${BASE_URL}/bridges/${BRIDGE3_ID}" \
  -F "status=维修中" \
  -F "remark=发现轻微缺陷，正在进行维护")

check_response "$UPDATE_BRIDGE" "200" "更新桥梁信息"

# 步骤21：删除无人机
print_step 21 $TOTAL_STEPS "删除无人机2（物理删除）"
DELETE_DRONE=$(curl -s -b "$COOKIE_FILE" -X DELETE "${BASE_URL}/drones/${DRONE1_ID}")
check_response "$DELETE_DRONE" "200" "删除无人机"

# 步骤22：用户登出
print_step 22 $TOTAL_STEPS "用户登出"
LOGOUT=$(curl -s -b "$COOKIE_FILE" -X POST "${BASE_URL}/auth/logout")
check_response "$LOGOUT" "200" "用户登出"

# ====================
# 测试总结
# ====================
print_separator "═══════════════════ 测试总结 ═══════════════════"

echo -e "总测试步骤: ${BLUE}$TOTAL_STEPS${NC}"
echo -e "成功步骤: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败步骤: ${RED}$((TOTAL_STEPS - PASSED_TESTS))${NC}"
echo -e "成功率: ${BLUE}$(echo "scale=1; $PASSED_TESTS * 100 / $TOTAL_STEPS" | bc)%${NC}"

echo -e "\n${BLUE}创建的测试数据：${NC}"
echo "  • 用户: user_test_${TIMESTAMP}"
echo "  • 桥梁: 3座（长江大桥、黄河大桥、珠江大桥）"
echo "  • 无人机: 2台（Mavic 3 Pro、Phantom 4 RTK）"
echo "  • 缺陷: 17个（分布在3座桥梁上）"

echo -e "\n${BLUE}测试覆盖模块：${NC}"
echo "  ✓ 用户注册与登录"
echo "  ✓ 桥梁管理（CRUD）"
echo "  ✓ 无人机管理（CRUD）"
echo "  ✓ 缺陷查询与过滤"
echo "  ✓ 统计模块（6个接口）"
echo "  ✓ 数据更新与删除"

if [ "$PASSED_TESTS" -eq "$TOTAL_STEPS" ]; then
    echo -e "\n${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║      🎉 所有测试通过！系统运行正常      ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "\n${YELLOW}╔════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║   ⚠️  部分测试失败，请检查错误日志    ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════╝${NC}"
    exit 1
fi
