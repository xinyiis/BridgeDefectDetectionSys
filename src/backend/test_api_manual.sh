#!/bin/bash

# 真实HTTP API测试脚本
# 测试运行中的服务器（需要先启动：go run cmd/server/main.go）

BASE_URL="http://localhost:8080"
COOKIE_FILE="cookies.txt"

echo "========================================="
echo "   真实HTTP API测试"
echo "========================================="
echo ""

# 检查服务器是否运行
echo "检查服务器状态..."
if ! curl -s "$BASE_URL/api/health" > /dev/null; then
    echo "❌ 服务器未运行"
    echo "请先启动服务器："
    echo "  cd src/backend"
    echo "  go run cmd/server/main.go"
    exit 1
fi
echo "✅ 服务器正在运行"
echo ""

# 清理旧Cookie
rm -f $COOKIE_FILE

echo "========================================="
echo "   开始测试用户接口"
echo "========================================="
echo ""

# 1. 健康检查
echo "1️⃣ 测试健康检查接口"
curl -s "$BASE_URL/api/health" | jq '.'
echo ""

# 2. 用户注册
echo "2️⃣ 测试用户注册"
curl -s -X POST "$BASE_URL/api/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "real_name": "测试用户",
    "email": "test@example.com",
    "phone": "13800138000"
  }' | jq '.'
echo ""

# 3. 用户登录
echo "3️⃣ 测试用户登录"
curl -s -c $COOKIE_FILE -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }' | jq '.'
echo ""

# 4. 获取当前用户信息
echo "4️⃣ 测试获取当前用户信息"
curl -s -b $COOKIE_FILE "$BASE_URL/api/user/info" | jq '.'
echo ""

# 5. 更新用户信息
echo "5️⃣ 测试更新用户信息"
curl -s -b $COOKIE_FILE -X PUT "$BASE_URL/api/user/info" \
  -H "Content-Type: application/json" \
  -d '{
    "real_name": "更新后的名字",
    "phone": "13900139000"
  }' | jq '.'
echo ""

# 6. 用户登出
echo "6️⃣ 测试用户登出"
curl -s -b $COOKIE_FILE -X POST "$BASE_URL/api/logout" | jq '.'
echo ""

# 7. 管理员登录
echo "7️⃣ 测试管理员登录"
curl -s -c $COOKIE_FILE -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }' | jq '.'
echo ""

# 8. 管理员获取用户列表
echo "8️⃣ 测试管理员获取用户列表"
curl -s -b $COOKIE_FILE "$BASE_URL/api/admin/users?page=1&page_size=10" | jq '.'
echo ""

# 9. 管理员获取用户详情
echo "9️⃣ 测试管理员获取用户详情"
curl -s -b $COOKIE_FILE "$BASE_URL/api/admin/users/2" | jq '.'
echo ""

# 清理
rm -f $COOKIE_FILE

echo "========================================="
echo "   测试完成"
echo "========================================="
