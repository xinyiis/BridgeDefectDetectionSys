#!/bin/bash

# 启动后端服务脚本
# 使用方法：bash start_backend.sh

echo "=========================================="
echo "  启动桥梁缺陷检测系统后端服务"
echo "=========================================="

# 检查配置文件
if [ ! -f "src/backend/config.yaml" ]; then
    echo "❌ 错误：找不到配置文件 src/backend/config.yaml"
    exit 1
fi

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误：未安装 Go，请先安装 Go 1.25+"
    exit 1
fi

echo "✓ 检查通过"
echo ""

# 切换到后端目录
cd src/backend

echo "启动服务..."
echo "服务地址: http://localhost:8080"
echo "API 文档: http://localhost:8080/swagger/index.html"
echo ""
echo "按 Ctrl+C 停止服务"
echo "=========================================="
echo ""

# 运行服务
go run cmd/server/main.go
