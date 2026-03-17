#!/bin/bash

##############################################
# 桥梁病害检测系统 - 后端环境一键配置脚本
# 适用于: Ubuntu 22.04/24.04
# 功能: 安装Go、MySQL及相关依赖
##############################################

set -e  # 遇到错误立即退出

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助
show_help() {
    echo "桥梁病害检测系统 - 后端环境配置脚本"
    echo ""
    echo "用法:"
    echo "  ./setup_backend_env.sh [选项]"
    echo ""
    echo "选项:"
    echo "  无参数          安装所有组件（智能检测已安装组件）"
    echo "  -h, --help      显示此帮助信息"
    echo "  -u, --uninstall 完全卸载所有组件"
    echo ""
    echo "安装组件:"
    echo "  - Go 1.25.0"
    echo "  - MySQL 8.0 (root/123456)"
    echo "  - Go依赖包（Gin, GORM等）"
    echo ""
    echo "说明:"
    echo "  - MySQL仅安装并设置root密码，不创建数据库"
    echo "  - 使用 import_database.sh 导入.sql文件创建数据库"
    echo ""
    echo "适用系统:"
    echo "  - Ubuntu 22.04/24.04"
}

# 检测系统
check_system() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [ "$ID" != "ubuntu" ]; then
            print_error "此脚本仅支持Ubuntu系统"
            exit 1
        fi

        VERSION_NUM=$(echo $VERSION_ID | cut -d. -f1)
        if [ "$VERSION_NUM" != "22" ] && [ "$VERSION_NUM" != "24" ]; then
            print_warning "建议使用Ubuntu 22.04或24.04"
            read -p "是否继续? (y/n): " continue_choice
            if [ "$continue_choice" != "y" ]; then
                exit 1
            fi
        fi
        print_info "系统: Ubuntu $VERSION_ID"
    else
        print_error "无法检测操作系统"
        exit 1
    fi
}

# 检查并清理旧版本
check_and_clean() {
    print_info "检查已安装的组件..."

    local need_clean=false
    local go_installed=false
    local mysql_installed=false

    # 检查Go
    if command -v go &> /dev/null; then
        go_installed=true
        need_clean=true
        print_warning "检测到已安装的Go: $(go version)"
    fi

    # 检查MySQL
    if command -v mysql &> /dev/null; then
        mysql_installed=true
        need_clean=true
        print_warning "检测到已安装的MySQL: $(mysql --version | awk '{print $3}')"
    fi

    if [ "$need_clean" = true ]; then
        echo ""
        print_warning "⚠️  检测到已安装的组件，可能导致版本冲突！"
        echo ""
        echo "选择操作："
        echo "  1) 清理所有旧版本并重新安装（推荐）"
        echo "  2) 跳过已安装的组件，仅安装缺失部分"
        echo "  3) 退出脚本"
        echo ""
        read -p "请选择 [1-3]: " clean_choice

        case $clean_choice in
            1)
                print_info "开始清理旧版本..."
                clean_old_installations "$go_installed" "$mysql_installed"
                print_success "清理完成！"
                export FORCE_INSTALL="true"
                ;;
            2)
                print_warning "将跳过已安装的组件"
                export FORCE_INSTALL="false"
                ;;
            3)
                print_info "退出脚本"
                exit 0
                ;;
            *)
                print_error "无效的选择"
                exit 1
                ;;
        esac
    else
        print_success "未检测到已安装的组件，开始全新安装"
        export FORCE_INSTALL="true"
    fi
}

# 清理旧版本
clean_old_installations() {
    local go_installed=$1
    local mysql_installed=$2

    # 清理Go
    if [ "$go_installed" = true ]; then
        print_info "清理旧版本Go..."
        sudo rm -rf /usr/local/go

        if [ -f ~/.bashrc ]; then
            cp ~/.bashrc ~/.bashrc.backup.$(date +%Y%m%d%H%M%S)
            sed -i '/\/usr\/local\/go\/bin/d' ~/.bashrc
            sed -i '/export GOPATH/d' ~/.bashrc
            sed -i '/export GOPROXY/d' ~/.bashrc
        fi

        # GOPATH目录
        if [ -d "$HOME/go" ]; then
            read -p "是否删除GOPATH目录 $HOME/go ? (y/n): " delete_gopath
            if [ "$delete_gopath" = "y" ]; then
                rm -rf "$HOME/go"
                print_success "GOPATH目录已删除"
            fi
        fi

        print_success "Go清理完成"
    fi

    # 清理MySQL
    if [ "$mysql_installed" = true ]; then
        print_warning "⚠️  清理MySQL将删除所有数据库数据！"
        read -p "确认删除MySQL及所有数据? (yes/no): " confirm_mysql

        if [ "$confirm_mysql" = "yes" ]; then
            print_info "清理MySQL..."
            sudo systemctl stop mysql &> /dev/null || true
            sudo apt-get remove --purge -y mysql-server mysql-client mysql-common
            sudo apt-get autoremove -y
            sudo apt-get autoclean
            sudo rm -rf /etc/mysql /var/lib/mysql
            print_success "MySQL清理完成"
        else
            print_warning "跳过MySQL清理"
            export SKIP_MYSQL="true"
        fi
    fi
}

# 完全卸载
uninstall_all() {
    print_warning "=========================================="
    print_warning "      ⚠️  完全卸载模式 ⚠️"
    print_warning "=========================================="
    echo ""
    print_warning "此操作将删除："
    echo "  - Go及其环境变量"
    echo "  - MySQL及所有数据库"
    echo ""
    print_error "⚠️  此操作不可恢复！⚠️"
    echo ""
    read -p "确认完全卸载? (输入 YES 继续): " confirm_uninstall

    if [ "$confirm_uninstall" != "YES" ]; then
        print_info "已取消卸载"
        exit 0
    fi

    # 卸载Go
    print_info "卸载Go..."
    sudo rm -rf /usr/local/go
    rm -rf "$HOME/go"
    if [ -f ~/.bashrc ]; then
        cp ~/.bashrc ~/.bashrc.backup.$(date +%Y%m%d%H%M%S)
        sed -i '/\/usr\/local\/go\/bin/d' ~/.bashrc
        sed -i '/export GOPATH/d' ~/.bashrc
        sed -i '/export GOPROXY/d' ~/.bashrc
    fi
    print_success "Go已卸载"

    # 卸载MySQL
    print_info "卸载MySQL..."
    sudo systemctl stop mysql &> /dev/null || true
    sudo apt-get remove --purge -y mysql-server mysql-client mysql-common
    sudo apt-get autoremove -y
    sudo apt-get autoclean
    sudo rm -rf /etc/mysql /var/lib/mysql
    print_success "MySQL已卸载"

    print_success "=========================================="
    print_success "完全卸载完成！"
    print_success "=========================================="
    exit 0
}

# 更新包管理器
update_system() {
    print_info "更新包管理器..."
    sudo apt-get update -y
    print_success "包管理器更新完成"
}

# 安装基础依赖
install_dependencies() {
    print_info "安装基础依赖..."
    sudo apt-get install -y wget curl git build-essential
    print_success "基础依赖安装完成"
}

# 安装Go
install_go() {
    if command -v go &> /dev/null && [ "$FORCE_INSTALL" != "true" ]; then
        INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go已安装: $INSTALLED_VERSION"
        return
    fi

    GO_VERSION="1.25.0"
    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"

    print_info "开始安装Go ${GO_VERSION}..."

    # 下载Go
    cd /tmp
    if [ ! -f "$GO_TAR" ]; then
        print_info "下载Go安装包..."
        wget -q --show-progress "$GO_URL" || {
            print_error "Go下载失败，尝试使用国内镜像..."
            wget -q --show-progress "https://mirrors.aliyun.com/golang/${GO_TAR}"
        }
    fi

    # 删除旧版本
    sudo rm -rf /usr/local/go

    # 解压安装
    print_info "解压Go安装包..."
    sudo tar -C /usr/local -xzf "$GO_TAR"

    # 配置环境变量
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo '' >> ~/.bashrc
        echo '# Go environment' >> ~/.bashrc
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
        echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc
    fi

    # 立即生效
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    export GOPROXY=https://goproxy.cn,direct

    # 验证安装
    if go version &> /dev/null; then
        print_success "Go安装成功: $(go version)"
    else
        print_error "Go安装失败"
        exit 1
    fi

    # 清理
    rm -f "/tmp/$GO_TAR"
}

# 安装MySQL
install_mysql() {
    if command -v mysql &> /dev/null && [ "$FORCE_INSTALL" != "true" ] || [ "$SKIP_MYSQL" = "true" ]; then
        print_success "MySQL已安装，跳过"
        return
    fi

    print_info "开始安装MySQL 8.0..."

    # 设置非交互式安装
    export DEBIAN_FRONTEND=noninteractive

    # 注意：MySQL 8.0不支持通过debconf预设密码，将在安装后配置

    # 安装MySQL
    print_info "安装MySQL软件包..."
    sudo apt-get install -y mysql-server

    print_success "MySQL安装完成"
}

# 配置MySQL
configure_mysql() {
    # 如果跳过MySQL安装，也跳过配置
    if [ "$SKIP_MYSQL" = "true" ]; then
        print_warning "跳过MySQL配置"
        return
    fi

    print_info "配置MySQL..."

    # 启动MySQL服务
    sudo systemctl start mysql
    sudo systemctl enable mysql

    # 等待MySQL完全启动
    print_info "等待MySQL服务启动..."
    sleep 3

    # 检测MySQL当前状态
    print_info "检测MySQL认证状态..."

    # 测试1：尝试使用密码123456连接（可能已配置过）
    if mysql -uroot -p123456 -e "SELECT 1;" &> /dev/null; then
        print_success "MySQL已配置，密码为123456"
        MYSQL_CONFIGURED=true
    # 测试2：尝试sudo无密码连接（全新安装）
    elif sudo mysql -e "SELECT 1;" &> /dev/null; then
        print_info "检测到全新MySQL安装，开始配置密码..."

        # 修改root密码（Ubuntu的MySQL默认使用auth_socket插件）
        if sudo mysql <<EOF
ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '123456';
FLUSH PRIVILEGES;
EOF
        then
            print_success "MySQL密码设置成功"
            MYSQL_CONFIGURED=true
        else
            print_error "MySQL密码设置失败"
            exit 1
        fi
    else
        print_error "无法连接到MySQL，请检查MySQL服务状态"
        print_info "尝试运行: sudo systemctl status mysql"
        exit 1
    fi

    if [ "$MYSQL_CONFIGURED" = true ]; then
        # 测试连接
        print_info "测试数据库连接..."
        if mysql -uroot -p123456 -e "SELECT VERSION();" &> /dev/null; then
            MYSQL_VERSION=$(mysql -uroot -p123456 -e "SELECT VERSION();" -s -N)
            print_success "数据库连接测试成功，版本: $MYSQL_VERSION"
        else
            print_warning "数据库连接测试失败，请手动检查"
        fi

        # 保存配置信息
        cat > mysql_config.txt <<EOF
MySQL配置信息
================
用户名: root
密码: 123456
主机: localhost
端口: 3306
================

连接命令:
mysql -uroot -p123456

连接字符串示例（Go）:
DSN: root:123456@tcp(localhost:3306)/数据库名?charset=utf8mb4&parseTime=True
================

说明:
- 使用 import_database.sh 脚本导入.sql文件创建数据库
- 或手动创建: mysql -uroot -p123456 -e "CREATE DATABASE mydb;"
================
EOF
        print_success "配置信息已保存到: mysql_config.txt"
        print_success "MySQL配置完成"
    fi
}

# 安装Go依赖包
install_go_dependencies() {
    print_info "安装Go依赖包..."

    # 初始化go.mod（如果不存在）
    if [ ! -f "go.mod" ]; then
        print_info "初始化Go模块..."
        go mod init bridge-detection-backend
    fi

    print_info "安装依赖包（可能需要几分钟）..."

    go get github.com/gin-gonic/gin@v1.10.0
    go get gorm.io/gorm@v1.25.7
    go get gorm.io/driver/mysql@v1.5.2
    go get github.com/gin-contrib/sessions@v1.0.1
    go get github.com/gin-contrib/cors@v1.6.0
    go get golang.org/x/crypto@v0.19.0
    go get github.com/gorilla/websocket@v1.5.1

    go mod tidy

    print_success "Go依赖包安装完成"
}

# 显示总结
show_summary() {
    echo ""
    echo "=========================================="
    echo "✅ 后端环境配置完成！"
    echo "=========================================="
    echo ""
    echo "📦 已安装组件："
    echo "  - Go $(go version | awk '{print $3}')"
    echo "  - MySQL $(mysql -uroot -p123456 -e "SELECT VERSION();" -s -N 2>/dev/null || echo "8.0")"
    echo "  - Go依赖包（Gin, GORM等）"
    echo ""

    if [ -f "mysql_config.txt" ]; then
        echo "🗄️  MySQL配置："
        cat mysql_config.txt
        echo ""
    fi

    echo "🚀 下一步："
    echo "  1. 重新加载环境变量："
    echo "     source ~/.bashrc"
    echo ""
    echo "  2. 验证安装："
    echo "     go version"
    echo "     mysql -uroot -p123456"
    echo ""
    echo "  3. 导入数据库："
    echo "     ./import_database.sh your_database.sql"
    echo ""
    echo "  4. 或手动创建数据库："
    echo "     mysql -uroot -p123456 -e \"CREATE DATABASE mydb;\""
    echo ""
    echo "  5. 开始开发"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    # 处理命令行参数
    if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ]; then
        check_system
        uninstall_all
        exit 0
    fi

    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        show_help
        exit 0
    fi

    echo ""
    echo "=========================================="
    echo "  桥梁病害检测系统 - 后端环境配置脚本  "
    echo "=========================================="
    echo ""

    # 检测系统
    check_system

    # 检查并清理旧版本
    check_and_clean

    # 更新系统
    update_system

    # 安装基础依赖
    install_dependencies

    # 安装Go
    install_go

    # 安装MySQL
    install_mysql

    # 配置MySQL
    configure_mysql

    # 安装Go依赖
    install_go_dependencies

    # 显示总结
    show_summary
}

# 执行主函数
main "$@"
