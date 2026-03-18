# 桥梁缺陷检测系统 - 后端服务

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go                      # 应用启动入口
├── internal/
│   ├── domain/
│   │   └── model/
│   │       └── user.go                  # 数据模型定义
│   ├── infrastructure/
│   │   └── persistence/
│   │       └── database.go              # 数据库连接和迁移
│   └── interfaces/
│       └── http/
│           ├── middleware/
│           │   └── auth.go              # 认证中间件
│           └── router/
│               └── router.go            # 路由管理
├── pkg/
│   ├── config/
│   │   └── config.go                    # 配置管理
│   └── response/
│       └── response.go                  # 统一响应格式
├── config.yaml                           # 配置文件
├── go.mod                                # Go 模块定义
└── README.md                             # 本文件
```

## 快速开始

### 1. 安装依赖

```bash
cd src/backend
go mod tidy
```

首次运行会自动下载以下依赖：
- `github.com/gin-gonic/gin` - Web 框架
- `gorm.io/gorm` - ORM 框架
- `gorm.io/driver/mysql` - MySQL 驱动
- `github.com/gin-contrib/cors` - CORS 中间件
- `github.com/gin-contrib/sessions` - Session 管理
- `gopkg.in/yaml.v3` - YAML 配置解析
- `golang.org/x/crypto` - 密码加密（bcrypt）

### 2. 配置数据库

确保 MySQL 服务运行，并创建数据库：

```sql
CREATE DATABASE bridge_detection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

修改 `config.yaml` 中的数据库配置：

```yaml
database:
  dsn: "root:你的密码@tcp(localhost:3306)/bridge_detection?charset=utf8mb4&parseTime=True&loc=Local"
```

### 3. 启动服务

```bash
go run cmd/server/main.go
```

看到以下输出表示启动成功：

```
✓ 配置加载成功
✓ 数据库连接成功
✓ 数据库表结构迁移完成
✓ 已创建默认管理员账户 (用户名: admin, 密码: admin123)
✓ 路由注册完成
✓ 服务器启动成功
✓ 监听地址: http://localhost:8080
```

### 4. 测试接口

访问健康检查接口：

```bash
curl http://localhost:8080/api/health
```

返回：

```json
{
  "status": "ok",
  "message": "Bridge Detection System API is running"
}
```

## 配置说明

### config.yaml 配置项

```yaml
# 服务器配置
server:
  port: 8080              # HTTP 端口
  mode: debug             # 运行模式: debug/release

# 数据库配置
database:
  dsn: "..."              # MySQL 连接串
  max_idle_conns: 10      # 最大空闲连接
  max_open_conns: 100     # 最大打开连接

# Python 算法服务
python_service:
  url: "http://localhost:8000"
  timeout: 30

# 文件上传
upload:
  image_dir: "./uploads/images"
  result_dir: "./uploads/results"

# Session 配置
session:
  secret: "your-secret-key"
  max_age: 86400

# CORS 配置
cors:
  allow_origins:
    - "http://localhost:5173"
```

## API 接口文档

### 公开接口（无需登录）

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| GET | `/api/health` | 健康检查 | ✅ |
| POST | `/api/register` | 用户注册 | 🚧 待实现 |
| POST | `/api/login` | 用户登录 | 🚧 待实现 |

### 认证接口（需要登录）

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| POST | `/api/logout` | 退出登录 | 🚧 待实现 |
| GET | `/api/user/info` | 获取用户信息 | 🚧 待实现 |
| GET | `/api/bridges` | 获取桥梁列表 | 🚧 待实现 |
| POST | `/api/bridges` | 创建桥梁 | 🚧 待实现 |
| GET | `/api/bridges/:id` | 获取桥梁详情 | 🚧 待实现 |
| POST | `/api/detect/image` | 图片检测 | 🚧 待实现 |

### 管理员接口（需要管理员权限）

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| GET | `/api/admin/users` | 用户管理 | 🚧 待实现 |
| GET | `/api/admin/stats` | 全局统计 | 🚧 待实现 |

## 数据库表结构

### users（用户表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | varchar(50) | 用户名（唯一） |
| password | varchar(255) | 密码（bcrypt加密） |
| email | varchar(100) | 邮箱 |
| role | varchar(20) | 角色（user/admin） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### bridges（桥梁表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | varchar(100) | 桥梁名称 |
| location | varchar(255) | 地理位置 |
| description | text | 描述 |
| user_id | uint | 所属用户ID |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### drones（无人机表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | varchar(100) | 名称 |
| model | varchar(100) | 型号 |
| stream_url | varchar(255) | 视频流地址 |
| user_id | uint | 所属用户ID |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### defects（缺陷表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| bridge_id | uint | 所属桥梁ID |
| defect_type | varchar(50) | 缺陷类型 |
| image_path | varchar(255) | 原始图片路径 |
| result_path | varchar(255) | 结果图片路径 |
| bbox | text | 边界框（JSON） |
| length | decimal(10,4) | 长度（米） |
| width | decimal(10,4) | 宽度（米） |
| area | decimal(10,4) | 面积（平方米） |
| confidence | decimal(5,4) | 置信度 |
| detected_at | datetime | 检测时间 |
| created_at | datetime | 创建时间 |

## 开发指南

### 添加新接口

1. 在 `internal/interfaces/http/handler/` 创建 handler
2. 在 `internal/application/usecase/` 创建业务逻辑
3. 在 `internal/interfaces/http/router/router.go` 注册路由

示例：

```go
// handler/user_handler.go
func Register(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 实现注册逻辑
    }
}

// router/router.go
r.POST("/register", handler.Register(db))
```

### 使用中间件

```go
// 需要登录
auth := r.Group("/api")
auth.Use(middleware.AuthRequired(db))
auth.GET("/user/info", handler.GetUserInfo)

// 需要管理员权限
admin := r.Group("/api/admin")
admin.Use(middleware.AuthRequired(db))
admin.Use(middleware.AdminRequired())
admin.GET("/users", handler.GetAllUsers(db))
```

### 获取当前用户

```go
func GetUserInfo(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    response.Success(c, user)
}
```

## 生产部署

### 1. 编译

```bash
go build -o bridge-server cmd/server/main.go
```

### 2. 修改配置

```yaml
server:
  mode: release  # 切换到生产模式

session:
  secret: "生产环境密钥"  # 修改为随机密钥

cors:
  allow_origins:
    - "https://yourdomain.com"  # 修改为生产域名
```

### 3. 运行

```bash
./bridge-server
```

## 故障排查

### 数据库连接失败

```
❌ 数据库连接失败: Error 1045: Access denied
```

解决方法：
1. 检查 MySQL 是否运行：`systemctl status mysql`
2. 检查用户名密码是否正确
3. 检查数据库是否创建：`SHOW DATABASES;`

### 端口被占用

```
❌ 服务器启动失败: bind: address already in use
```

解决方法：
1. 修改 `config.yaml` 中的端口
2. 或者杀死占用进程：`lsof -ti:8080 | xargs kill -9`

### 依赖下载失败

```bash
# 使用国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
```

## 技术栈

- **语言**: Go 1.25+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **Session**: Cookie-based
- **配置**: YAML

## 下一步开发

- [ ] 实现用户注册登录
- [ ] 实现桥梁 CRUD
- [ ] 实现图片检测接口
- [ ] 对接 Python 算法服务
- [ ] 添加单元测试
- [ ] 添加 API 文档（Swagger）

## 许可证

Copyright © 2026 Bridge Detection Team
