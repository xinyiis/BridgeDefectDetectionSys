# 后端环境配置

本目录包含后端开发环境（Go + MySQL）的安装和配置脚本。

## 📦 包含文件

- `setup_backend_env.sh` - 后端环境一键安装脚本
- `import_database.sh` - 数据库导入工具

## 🚀 快速开始

### 安装后端环境

```bash
# 在 environment-setup 目录下运行
sudo ./backend-env/setup_backend_env.sh
```

安装完成后重新加载环境变量：

```bash
source ~/.bashrc
```

### 导入数据库

```bash
# 导入 SQL 文件到数据库
./backend-env/import_database.sh your_database.sql
```

## 📋 安装内容

- **Go 1.25.0** - Go 编程语言环境
- **MySQL 8.0** - 数据库服务器
  - 用户名: `root`
  - 密码: `123456`
  - 端口: `3306`
- **Go 依赖包** - Gin、GORM 等常用包

## 📚 详细文档

详细的安装说明和故障排除，请查看上级目录的 [README.md](../README.md)

## 💡 提示

如果需要查看脚本帮助信息：

```bash
./backend-env/setup_backend_env.sh --help
```
