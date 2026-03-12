#!/bin/bash

##############################################
# Linux虚拟机开发环境 - 基础工具一键配置脚本
# 适用于: Ubuntu 22.04/24.04
# 功能: 安装开发必备的基础工具
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
    echo "Linux虚拟机开发环境 - 基础工具配置脚本"
    echo ""
    echo "用法:"
    echo "  ./setup_basic_tools.sh [选项]"
    echo ""
    echo "选项:"
    echo "  无参数          安装所有工具（智能跳过已安装）"
    echo "  -h, --help      显示此帮助信息"
    echo "  -c, --check     仅检查工具安装状态"
    echo ""
    echo "安装工具分类:"
    echo ""
    echo "📝 编辑器工具:"
    echo "  - vim          强大的文本编辑器"
    echo "  - nano         简单易用的编辑器"
    echo ""
    echo "🔧 开发工具:"
    echo "  - git          版本控制系统"
    echo "  - build-essential  编译工具链(gcc, g++, make)"
    echo "  - cmake        跨平台构建工具"
    echo ""
    echo "🌐 网络工具:"
    echo "  - curl         命令行HTTP客户端"
    echo "  - wget         文件下载工具"
    echo "  - net-tools    网络工具集(ifconfig等)"
    echo "  - openssh-server  SSH服务器"
    echo ""
    echo "💻 系统工具:"
    echo "  - htop         进程监视器"
    echo "  - tree         目录树显示"
    echo "  - tmux         终端复用器"
    echo "  - zip/unzip    压缩解压工具"
    echo "  - jq           JSON处理工具"
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

# 检查单个工具
check_tool() {
    local tool_name=$1
    if command -v "$tool_name" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# 检查包是否安装
check_package() {
    local package_name=$1
    if dpkg -l | grep -q "^ii  $package_name"; then
        return 0
    else
        return 1
    fi
}

# 检查所有工具状态
check_all_tools() {
    echo ""
    echo "=========================================="
    echo "  工具安装状态检查"
    echo "=========================================="
    echo ""

    local total=0
    local installed=0

    # 编辑器工具
    echo "📝 编辑器工具:"
    for tool in vim nano; do
        total=$((total + 1))
        if check_tool "$tool"; then
            echo -e "  ${GREEN}✓${NC} $tool"
            installed=$((installed + 1))
        else
            echo -e "  ${RED}✗${NC} $tool"
        fi
    done
    echo ""

    # 开发工具
    echo "🔧 开发工具:"
    for tool in git gcc g++ make cmake; do
        total=$((total + 1))
        if check_tool "$tool"; then
            echo -e "  ${GREEN}✓${NC} $tool"
            installed=$((installed + 1))
        else
            echo -e "  ${RED}✗${NC} $tool"
        fi
    done
    echo ""

    # 网络工具
    echo "🌐 网络工具:"
    for tool in curl wget ifconfig ssh; do
        total=$((total + 1))
        if check_tool "$tool"; then
            echo -e "  ${GREEN}✓${NC} $tool"
            installed=$((installed + 1))
        else
            echo -e "  ${RED}✗${NC} $tool"
        fi
    done
    echo ""

    # 系统工具
    echo "💻 系统工具:"
    for tool in htop tree tmux zip unzip jq; do
        total=$((total + 1))
        if check_tool "$tool"; then
            echo -e "  ${GREEN}✓${NC} $tool"
            installed=$((installed + 1))
        else
            echo -e "  ${RED}✗${NC} $tool"
        fi
    done
    echo ""

    echo "=========================================="
    echo "统计: $installed/$total 工具已安装"
    echo "=========================================="
    echo ""
}

# 更新包管理器
update_system() {
    print_info "更新包管理器..."
    sudo apt-get update -y
    print_success "包管理器更新完成"
}

# 安装编辑器工具
install_editors() {
    print_info "安装编辑器工具..."

    local packages=""

    if ! check_tool vim; then
        packages="$packages vim"
    fi

    if ! check_tool nano; then
        packages="$packages nano"
    fi

    if [ -n "$packages" ]; then
        sudo apt-get install -y $packages
        print_success "编辑器工具安装完成"
    else
        print_success "编辑器工具已全部安装，跳过"
    fi
}

# 安装开发工具
install_dev_tools() {
    print_info "安装开发工具..."

    local packages=""

    if ! check_tool git; then
        packages="$packages git"
    fi

    if ! check_package build-essential; then
        packages="$packages build-essential"
    fi

    if ! check_tool cmake; then
        packages="$packages cmake"
    fi

    if [ -n "$packages" ]; then
        sudo apt-get install -y $packages
        print_success "开发工具安装完成"
    else
        print_success "开发工具已全部安装，跳过"
    fi
}

# 安装网络工具
install_network_tools() {
    print_info "安装网络工具..."

    local packages=""

    if ! check_tool curl; then
        packages="$packages curl"
    fi

    if ! check_tool wget; then
        packages="$packages wget"
    fi

    if ! check_tool ifconfig; then
        packages="$packages net-tools"
    fi

    if ! check_tool ssh; then
        packages="$packages openssh-server"
    fi

    if [ -n "$packages" ]; then
        sudo apt-get install -y $packages

        # 启动SSH服务
        if echo "$packages" | grep -q "openssh-server"; then
            sudo systemctl start ssh
            sudo systemctl enable ssh
            print_info "SSH服务已启动并设置为开机自启"
        fi

        print_success "网络工具安装完成"
    else
        print_success "网络工具已全部安装，跳过"
    fi
}

# 安装系统工具
install_system_tools() {
    print_info "安装系统工具..."

    local packages=""

    if ! check_tool htop; then
        packages="$packages htop"
    fi

    if ! check_tool tree; then
        packages="$packages tree"
    fi

    if ! check_tool tmux; then
        packages="$packages tmux"
    fi

    if ! check_tool zip; then
        packages="$packages zip"
    fi

    if ! check_tool unzip; then
        packages="$packages unzip"
    fi

    if ! check_tool jq; then
        packages="$packages jq"
    fi

    if [ -n "$packages" ]; then
        sudo apt-get install -y $packages
        print_success "系统工具安装完成"
    else
        print_success "系统工具已全部安装，跳过"
    fi
}

# 配置Git（可选）
configure_git() {
    if check_tool git; then
        # 检查git是否已配置
        if [ -z "$(git config --global user.name)" ]; then
            print_warning "Git用户信息未配置"
            read -p "是否现在配置Git用户信息? (y/n): " config_git

            if [ "$config_git" = "y" ]; then
                read -p "输入你的名字: " git_name
                read -p "输入你的邮箱: " git_email

                git config --global user.name "$git_name"
                git config --global user.email "$git_email"

                print_success "Git配置完成"
                echo "  用户名: $(git config --global user.name)"
                echo "  邮箱: $(git config --global user.email)"
            fi
        else
            print_success "Git已配置: $(git config --global user.name) <$(git config --global user.email)>"
        fi
    fi
}

# 显示工具版本
show_versions() {
    echo ""
    echo "=========================================="
    echo "📦 已安装工具版本信息"
    echo "=========================================="
    echo ""

    # 编辑器
    if check_tool vim; then
        echo "vim:  $(vim --version | head -n1 | awk '{print $5}')"
    fi

    # 开发工具
    if check_tool git; then
        echo "git:  $(git --version | awk '{print $3}')"
    fi

    if check_tool gcc; then
        echo "gcc:  $(gcc --version | head -n1 | awk '{print $4}')"
    fi

    if check_tool cmake; then
        echo "cmake: $(cmake --version | head -n1 | awk '{print $3}')"
    fi

    # 网络工具
    if check_tool curl; then
        echo "curl: $(curl --version | head -n1 | awk '{print $2}')"
    fi

    # 系统工具
    if check_tool tmux; then
        echo "tmux: $(tmux -V | awk '{print $2}')"
    fi

    echo ""
}

# 显示常用命令提示
show_tips() {
    echo "=========================================="
    echo "💡 常用工具快速入门"
    echo "=========================================="
    echo ""
    echo "📝 文本编辑:"
    echo "  vim 文件名          使用vim编辑"
    echo "  nano 文件名         使用nano编辑"
    echo ""
    echo "🔧 版本控制:"
    echo "  git clone URL       克隆仓库"
    echo "  git status          查看状态"
    echo "  git add .           添加所有改动"
    echo "  git commit -m \"msg\" 提交改动"
    echo ""
    echo "🌐 网络工具:"
    echo "  curl URL            访问URL"
    echo "  wget URL            下载文件"
    echo "  ifconfig            查看网络配置"
    echo ""
    echo "💻 系统工具:"
    echo "  htop                查看进程"
    echo "  tree                显示目录树"
    echo "  tmux                启动终端复用器"
    echo ""
    echo "=========================================="
}

# 显示总结
show_summary() {
    echo ""
    echo "=========================================="
    echo "✅ 基础工具配置完成！"
    echo "=========================================="
    echo ""

    show_versions
    show_tips

    echo ""
    echo "🚀 下一步："
    echo "  1. 使用 --check 查看工具状态:"
    echo "     ./setup_basic_tools.sh --check"
    echo ""
    echo "  2. 配置Git用户信息（如未配置）:"
    echo "     git config --global user.name \"你的名字\""
    echo "     git config --global user.email \"你的邮箱\""
    echo ""
    echo "  3. 开始开发！"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    # 处理命令行参数
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        show_help
        exit 0
    fi

    if [ "$1" = "--check" ] || [ "$1" = "-c" ]; then
        check_system
        check_all_tools
        exit 0
    fi

    echo ""
    echo "=========================================="
    echo "  Linux开发环境 - 基础工具配置脚本  "
    echo "=========================================="
    echo ""

    # 检测系统
    check_system

    # 显示当前状态
    check_all_tools

    # 询问是否继续
    read -p "是否开始安装缺失的工具? (y/n): " install_confirm
    if [ "$install_confirm" != "y" ]; then
        print_info "已取消安装"
        exit 0
    fi

    # 更新系统
    update_system

    # 安装各类工具
    install_editors
    install_dev_tools
    install_network_tools
    install_system_tools

    # 配置Git
    configure_git

    # 显示总结
    show_summary
}

# 执行主函数
main "$@"
