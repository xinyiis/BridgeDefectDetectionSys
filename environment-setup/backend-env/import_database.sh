#!/bin/bash

##############################################
# 数据库导入脚本
# 功能: 使用.sql文件创建数据库
##############################################

set -e

# 颜色输出
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

# 显示帮助
show_help() {
    echo "数据库导入脚本"
    echo ""
    echo "用法:"
    echo "  ./import_database.sh [SQL文件路径]"
    echo ""
    echo "示例:"
    echo "  ./import_database.sh database.sql"
    echo "  ./import_database.sh /path/to/backup.sql"
    echo ""
    echo "说明:"
    echo "  - 默认使用 root/123456 连接MySQL"
    echo "  - SQL文件应包含 CREATE DATABASE 语句"
}

# MySQL连接配置
MYSQL_USER="root"
MYSQL_PASSWORD="123456"
MYSQL_HOST="localhost"
MYSQL_PORT="3306"

# 检查MySQL连接
check_mysql_connection() {
    print_info "检查MySQL连接..."

    if ! mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -h"$MYSQL_HOST" -P"$MYSQL_PORT" -e "SELECT 1;" &> /dev/null; then
        print_error "无法连接到MySQL"
        print_error "请检查MySQL是否已安装并正在运行"
        print_error "用户名: $MYSQL_USER"
        print_error "密码: $MYSQL_PASSWORD"
        exit 1
    fi

    print_success "MySQL连接正常"
}

# 导入SQL文件
import_sql() {
    local sql_file=$1

    # 检查文件是否存在
    if [ ! -f "$sql_file" ]; then
        print_error "SQL文件不存在: $sql_file"
        exit 1
    fi

    print_info "开始导入SQL文件: $sql_file"
    print_info "这可能需要几分钟，请耐心等待..."

    # 导入SQL文件
    if mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -h"$MYSQL_HOST" -P"$MYSQL_PORT" < "$sql_file"; then
        print_success "SQL文件导入成功"

        # 显示已创建的数据库
        print_info "当前数据库列表："
        mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -h"$MYSQL_HOST" -P"$MYSQL_PORT" -e "SHOW DATABASES;" | grep -v "Database\|information_schema\|performance_schema\|mysql\|sys"

    else
        print_error "SQL文件导入失败"
        exit 1
    fi
}

# 主函数
main() {
    # 处理参数
    if [ "$1" = "--help" ] || [ "$1" = "-h" ] || [ -z "$1" ]; then
        show_help
        exit 0
    fi

    local sql_file=$1

    echo ""
    echo "=========================================="
    echo "      数据库导入脚本"
    echo "=========================================="
    echo ""

    # 检查MySQL连接
    check_mysql_connection

    # 导入SQL文件
    import_sql "$sql_file"

    echo ""
    echo "=========================================="
    echo "✅ 数据库导入完成！"
    echo "=========================================="
    echo ""
    echo "💡 提示："
    echo "  - 登录MySQL: mysql -uroot -p123456"
    echo "  - 查看数据库: SHOW DATABASES;"
    echo "  - 使用数据库: USE 数据库名;"
    echo ""
}

# 执行主函数
main "$@"
