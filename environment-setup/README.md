# 比赛项目 - 环境配置脚本集合

这是比赛项目的统一环境配置脚本集合，包含后端环境和基础工具的自动安装脚本。

## 📋 目录

1. [基础工具配置](#基础工具配置) - Linux虚拟机必备工具
2. [后端环境配置](#后端环境配置) - Go + MySQL环境

## 📁 目录结构

```text
environment-setup/
├── setup_basic_tools.sh       # 基础工具安装脚本
├── backend-env/               # 后端环境相关
│   ├── setup_backend_env.sh   # 后端环境安装脚本
│   └── import_database.sh     # 数据库导入工具
└── README.md                  # 本文档
```

---

## 🔧 基础工具配置

**适用场景：** 全新的Ubuntu虚拟机，需要安装开发必备的基础工具。

### 包含工具

- **编辑器：** vim, nano
- **开发工具：** git, gcc, g++, make, cmake
- **网络工具：** curl, wget, net-tools, openssh-server
- **系统工具：** htop, tree, tmux, zip, unzip, jq

### 快速安装

```bash
# 1. 进入脚本目录
cd environment-setup

# 2. 运行安装脚本
sudo ./setup_basic_tools.sh

# 3. 检查安装状态
./setup_basic_tools.sh --check
```

### 特性

- ✅ 智能跳过已安装工具
- ✅ 自动配置SSH服务
- ✅ 可选配置Git用户信息
- ✅ 显示工具版本和使用提示

### 帮助命令

```bash
./setup_basic_tools.sh --help
```

---

## 🚀 后端环境配置

**适用场景：** 安装项目后端所需的Go语言和MySQL数据库。

### 快速开始（2步）

### 第1步：运行安装脚本

```bash
sudo ./backend-env/setup_backend_env.sh
```

### 第2步：重新加载环境变量

```bash
source ~/.bashrc
```

完成！现在可以使用 `go version` 和 `mysql -uroot -p123456` 测试。

---

## 📦 脚本会安装什么？

- **Go 1.21.6** - Go编程语言
- **MySQL 8.0** - 数据库（用户名：root，密码：123456）
- **Go依赖包** - Gin、GORM等7个包

**注意：** 脚本只安装MySQL，不会创建具体的数据库。

---

## 🎯 如果已经安装过MySQL或Go

运行脚本时会提示：

```
⚠️  检测到已安装的组件，可能导致版本冲突！

选择操作：
  1) 清理所有旧版本并重新安装（推荐）
  2) 跳过已安装的组件，仅安装缺失部分
  3) 退出脚本
```

**建议：**
- **开发环境，没有重要数据** → 选择 `1`
- **有重要数据库，不能删** → 选择 `2`

⚠️ **警告：** 选择1会删除MySQL所有数据库！

---

## 📝 安装后的操作

### 1. 验证安装

```bash
# 检查Go
go version

# 检查MySQL
mysql -uroot -p123456
```

### 2. 导入数据库

如果你有 `.sql` 文件：

```bash
./backend-env/import_database.sh your_database.sql
```

或手动创建数据库：

```bash
mysql -uroot -p123456 -e "CREATE DATABASE mydb;"
```

### 3. 开始开发

```bash
go run main.go
```

---

## 💾 MySQL配置信息

安装完成后，配置信息保存在 `mysql_config.txt`：

- **用户名：** root
- **密码：** 123456
- **端口：** 3306

**Go连接字符串：**
```go
dsn := "root:123456@tcp(localhost:3306)/数据库名?charset=utf8mb4&parseTime=True"
```

---

## 🆘 遇到问题？

### 权限不足

```bash
# 确保使用 sudo
sudo ./backend-env/setup_backend_env.sh
```

### 环境变量未生效

```bash
# 重新加载
source ~/.bashrc

# 或者退出终端重新登录
```

### MySQL连接失败

```bash
# 重启MySQL服务
sudo systemctl restart mysql

# 检查状态
sudo systemctl status mysql
```

---

## 📚 其他命令

```bash
# 查看帮助
./backend-env/setup_backend_env.sh --help

# 完全卸载（会删除所有数据）
sudo ./backend-env/setup_backend_env.sh --uninstall
```

---

## 📂 相关文档

- **后端技术栈方案.md** - 技术架构和详细设计
- **需求文档.md** - 项目需求说明

---

## ⏱️ 安装需要多久？

- **基础工具：** 约3-5分钟
- **后端环境：** 约5-10分钟

取决于网络速度和系统配置。

---

## ✅ 系统要求

### 所有脚本通用要求

- Ubuntu 22.04 或 24.04
- 需要sudo权限
- 稳定的网络连接

### 后端环境额外要求

- 至少2GB内存
- 至少5GB磁盘空间

---

## 🔒 安全提示

生产环境请修改默认密码：

```bash
mysql -uroot -p123456
ALTER USER 'root'@'localhost' IDENTIFIED BY '你的强密码';
FLUSH PRIVILEGES;
```

---

**有问题？** 查看脚本帮助：`./backend-env/setup_backend_env.sh --help`
