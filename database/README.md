# 数据库迁移指南

## 📦 文件说明

- **database_export.sql** - 完整数据库导出文件（32KB）
  - 包含所有表结构
  - 包含所有数据（24用户、26桥梁、5无人机、51缺陷、10报表）
  - MySQL 8.0 格式
  - UTF-8MB4 编码

## 🚀 在新电脑上导入数据库

### 方法1：命令行导入（推荐）

```bash
# 1. 创建数据库
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS bridge_detection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 导入数据
mysql -uroot -p bridge_detection < database_export.sql

# 3. 验证导入
mysql -uroot -p -Dbridge_detection -e "SHOW TABLES; SELECT COUNT(*) FROM users;"
```

### 方法2：MySQL Workbench 导入

1. 打开 MySQL Workbench
2. 连接到数据库
3. 菜单：Server → Data Import
4. 选择 "Import from Self-Contained File"
5. 选择 `database_export.sql` 文件
6. 点击 "Start Import"

### 方法3：phpMyAdmin 导入

1. 登录 phpMyAdmin
2. 创建数据库 `bridge_detection`
3. 选择该数据库
4. 点击 "导入" 选项卡
5. 选择 `database_export.sql` 文件
6. 点击 "执行"

## 📊 数据库统计

| 表名 | 数据量 | 说明 |
|------|--------|------|
| users | 24条 | 用户账号 |
| bridges | 26条 | 桥梁信息 |
| drones | 5条 | 无人机设备 |
| defects | 51条 | 缺陷检测记录 |
| reports | 10条 | 生成的报表 |

## ⚙️ 配置更新

导入数据库后，需要更新后端配置文件：

### 更新 `src/backend/config.yaml`

```yaml
database:
  dsn: "root:你的密码@tcp(localhost:3306)/bridge_detection?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600
```

**注意**：将 `你的密码` 替换为新电脑的MySQL密码。

## 🔑 默认管理员账号

导入后可使用以下账号登录：

- **用户名**：根据实际导出的数据
- **查询管理员账号**：
  ```bash
  mysql -uroot -p -Dbridge_detection -e "SELECT username, real_name, email, role FROM users WHERE role='admin';"
  ```

## 🗂️ 文件迁移

除了数据库，还需要迁移以下文件：

### 1. 上传文件（如果有）
```bash
# 压缩上传文件
tar -czf uploads.tar.gz uploads/

# 在新电脑上解压
tar -xzf uploads.tar.gz -C src/backend/
```

### 2. 报表文件（如果有）
```bash
# 压缩报表文件
tar -czf reports.tar.gz reports/

# 在新电脑上解压
tar -xzf reports.tar.gz
```

### 3. 字体文件
```bash
# 新电脑上运行字体下载脚本
cd environment-setup/backend-env
sudo bash setup_backend_env.sh
```

## ✅ 验证清单

导入完成后，执行以下检查：

- [ ] 数据库连接成功
- [ ] 所有表都存在（5个表）
- [ ] 数据正确导入（查询几条记录验证）
- [ ] 后端服务启动成功
- [ ] 登录功能正常
- [ ] 文件上传目录存在

## 🔄 重新导出数据（当前电脑）

如果需要重新导出最新数据：

```bash
cd /path/to/BridgeDefectDetectionSys
mysqldump -uroot -p123456 --single-transaction --routines --triggers --events --default-character-set=utf8mb4 bridge_detection > database/database_export_new.sql
```

## ⚠️ 注意事项

1. **MySQL版本兼容性**
   - 导出：MySQL 8.0.45
   - 建议新电脑使用 MySQL 8.0+ 版本

2. **字符编码**
   - 统一使用 UTF-8MB4 编码
   - 确保支持中文和emoji

3. **外键约束**
   - 导入时会自动处理外键关系
   - 不要单独导入某张表

4. **文件大小**
   - 当前导出文件：32KB
   - 如果数据量增大，考虑分表导出

## 🆘 常见问题

### 问题1：导入时报错 "Unknown database"
**解决**：先创建数据库
```bash
mysql -uroot -p -e "CREATE DATABASE bridge_detection;"
```

### 问题2：导入时报错字符集问题
**解决**：指定字符集
```bash
mysql -uroot -p --default-character-set=utf8mb4 bridge_detection < database_export.sql
```

### 问题3：导入后无法登录
**解决**：检查用户表
```bash
mysql -uroot -p -Dbridge_detection -e "SELECT id, username, role FROM users LIMIT 5;"
```

## 📞 技术支持

如有问题，请查看项目文档：
- `doc/接口文档.md` - API接口说明
- `doc/开发进度报告.md` - 项目状态
- `TEST_README.md` - 测试指南
