# 真实HTTP测试指南

## 两种测试方式对比

### 1. 单元/集成测试（已完成）✅
- **方式**: `go test`
- **特点**: 使用内存SQLite数据库，测试速度快，自动化
- **适用**: 开发过程中快速验证代码逻辑
- **状态**: ✅ 21个测试全部通过

### 2. 真实HTTP测试（本指南）🌐
- **方式**: 启动真实服务器，发送HTTP请求
- **特点**: 使用真实MySQL数据库，模拟前端调用场景
- **适用**: 验证完整的请求-响应流程，测试Session、CORS等
- **状态**: 准备就绪，按以下步骤操作

---

## 🚀 真实HTTP测试步骤

### 第一步：准备数据库

```bash
# 确保MySQL服务运行
sudo systemctl status mysql

# 创建数据库（如果还没创建）
mysql -uroot -p123456 -e "CREATE DATABASE IF NOT EXISTS bridge_detection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 第二步：启动服务器

```bash
cd /home/cjy/workspace/服务外包/BridgeDefectDetectionSys/src/backend

# 启动服务器（终端1）
go run cmd/server/main.go
```

**预期输出**：
```
╔══════════════════════════════════════════════════════════════╗
║         桥梁缺陷检测系统 - Bridge Defect Detection           ║
║                    Version: 1.0.0                            ║
╚══════════════════════════════════════════════════════════════╝

========== 初始化配置 ==========
✓ 配置加载成功

========== 初始化数据库 ==========
✓ 数据库连接成功
开始数据库表结构迁移...
✓ 数据库表结构迁移完成
✓ 已创建默认管理员账户 (用户名: admin, 密码: admin123)
  ⚠️  警告: 请尽快修改默认密码！
✓ 数据库索引检查完成

========== 初始化路由 ==========
✓ 路由注册完成
  公开路由:
    GET  /api/health          - 健康检查
    POST /api/register        - 用户注册
    POST /api/login           - 用户登录
  ...

========== 启动服务器 ==========
✓ 服务器启动成功
✓ 监听地址: http://localhost:8080
✓ 健康检查: http://localhost:8080/api/health
✓ 按 Ctrl+C 优雅退出

========== 服务器运行中 ==========
```

### 第三步：测试API接口

打开**新终端**（终端2），选择以下任一方式测试：

#### 方式1：使用自动化测试脚本（推荐）

```bash
cd /home/cjy/workspace/服务外包/BridgeDefectDetectionSys/src/backend

# 运行测试脚本
./test_api_manual.sh
```

#### 方式2：使用curl手动测试

```bash
# 1. 健康检查
curl http://localhost:8080/api/health

# 2. 用户注册
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "real_name": "测试用户",
    "email": "test@example.com",
    "phone": "13800138000"
  }'

# 3. 用户登录（保存Cookie）
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'

# 4. 获取用户信息（需要登录）
curl -b cookies.txt http://localhost:8080/api/user/info

# 5. 更新用户信息
curl -b cookies.txt -X PUT http://localhost:8080/api/user/info \
  -H "Content-Type: application/json" \
  -d '{
    "real_name": "新名字",
    "phone": "13900139000"
  }'

# 6. 退出登录
curl -b cookies.txt -X POST http://localhost:8080/api/logout

# 7. 管理员登录
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 8. 获取用户列表（管理员）
curl -b cookies.txt "http://localhost:8080/api/admin/users?page=1&page_size=10"

# 9. 获取指定用户信息（管理员）
curl -b cookies.txt http://localhost:8080/api/admin/users/1

# 10. 提升用户为管理员
curl -b cookies.txt -X POST http://localhost:8080/api/admin/users/promote \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2
  }'
```

#### 方式3：使用Postman测试

1. **导入API集合**（可以根据以下接口创建）

**基础设置**：
- Base URL: `http://localhost:8080`
- Headers: `Content-Type: application/json`

**公开接口**（无需登录）：
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| POST | `/api/register` | 用户注册 |
| POST | `/api/login` | 用户登录 |

**认证接口**（需要登录）：
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/logout` | 退出登录 |
| GET | `/api/user/info` | 获取用户信息 |
| PUT | `/api/user/info` | 更新用户信息 |

**管理员接口**：
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/users` | 获取用户列表 |
| GET | `/api/admin/users/:id` | 获取用户详情 |
| DELETE | `/api/admin/users/:id` | 删除用户 |
| POST | `/api/admin/users/promote` | 提升为管理员 |

2. **Postman设置Session**：
   - 在Postman设置中启用 `Automatically follow redirects`
   - Cookie会自动管理，无需手动处理

---

## 🎯 测试要点

### 1. Session认证测试

**重点验证**：
- ✅ 登录后获得Cookie
- ✅ 携带Cookie可以访问认证接口
- ✅ 登出后Cookie失效
- ✅ 无Cookie返回401

**测试步骤**：
```bash
# 步骤1：未登录访问 → 应返回401
curl http://localhost:8080/api/user/info

# 步骤2：登录获取Cookie
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 步骤3：携带Cookie访问 → 应返回200
curl -b cookies.txt http://localhost:8080/api/user/info

# 步骤4：登出
curl -b cookies.txt -X POST http://localhost:8080/api/logout

# 步骤5：使用旧Cookie访问 → 应返回401
curl -b cookies.txt http://localhost:8080/api/user/info
```

### 2. 权限控制测试

**重点验证**：
- ✅ 普通用户无法访问管理员接口（403）
- ✅ 管理员可以访问所有接口

**测试步骤**：
```bash
# 步骤1：注册并登录普通用户
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"123456","real_name":"普通用户","email":"user1@example.com"}'

curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"123456"}'

# 步骤2：普通用户访问管理员接口 → 应返回403
curl -b cookies.txt http://localhost:8080/api/admin/users

# 步骤3：管理员登录
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 步骤4：管理员访问 → 应返回200
curl -b cookies.txt http://localhost:8080/api/admin/users
```

### 3. 数据验证测试

**重点验证**：
- ✅ 用户名唯一性
- ✅ 邮箱唯一性
- ✅ 参数格式验证

**测试步骤**：
```bash
# 测试1：重复用户名 → 应返回400
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456","real_name":"测试","email":"test@example.com"}'

# 测试2：邮箱格式错误 → 应返回400
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","real_name":"测试","email":"invalid-email"}'

# 测试3：密码过短 → 应返回400
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123","real_name":"测试","email":"test@example.com"}'
```

---

## 🔍 观察服务器日志

服务器终端会显示实时日志，观察：

1. **SQL查询日志**（debug模式）
```
[2026-03-15 18:34:38] SELECT * FROM `users` WHERE username = "testuser"
[2026-03-15 18:34:38] INSERT INTO `users` (...)
```

2. **HTTP请求日志**
```
[GIN] 2026/03/15 - 18:34:38 | 200 |   12.345ms |   127.0.0.1 | POST     "/api/register"
[GIN] 2026/03/15 - 18:34:39 | 200 |    5.678ms |   127.0.0.1 | POST     "/api/login"
```

3. **错误日志**（如果有）
```
[GIN] 2026/03/15 - 18:34:40 | 401 |    1.234ms |   127.0.0.1 | GET      "/api/user/info"
```

---

## 📊 验证数据库变化

在测试过程中，可以查看数据库的实时变化：

```bash
# 打开MySQL客户端（新终端）
mysql -uroot -p123456 bridge_detection

# 查看用户表
SELECT id, username, real_name, email, role, created_at FROM users;

# 查看最新注册的用户
SELECT * FROM users ORDER BY created_at DESC LIMIT 5;

# 统计用户数量
SELECT role, COUNT(*) FROM users GROUP BY role;
```

---

## 🎨 前端集成测试

如果你有Vue前端，可以这样集成：

### 前端配置

```javascript
// src/api/request.js
import axios from 'axios'

const request = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 5000,
  withCredentials: true  // 重要：允许携带Cookie
})

export default request
```

### 前端API调用

```javascript
// src/api/user.js
import request from './request'

// 用户注册
export const register = (data) => {
  return request.post('/register', data)
}

// 用户登录
export const login = (data) => {
  return request.post('/login', data)
}

// 获取用户信息
export const getUserInfo = () => {
  return request.get('/user/info')
}

// 更新用户信息
export const updateUserInfo = (data) => {
  return request.put('/user/info', data)
}
```

### 前端测试页面

```vue
<template>
  <div>
    <button @click="testRegister">测试注册</button>
    <button @click="testLogin">测试登录</button>
    <button @click="testGetInfo">获取用户信息</button>
  </div>
</template>

<script>
import { register, login, getUserInfo } from '@/api/user'

export default {
  methods: {
    async testRegister() {
      try {
        const res = await register({
          username: 'testuser',
          password: '123456',
          real_name: '测试用户',
          email: 'test@example.com'
        })
        console.log('注册成功', res.data)
      } catch (error) {
        console.error('注册失败', error)
      }
    },

    async testLogin() {
      try {
        const res = await login({
          username: 'testuser',
          password: '123456'
        })
        console.log('登录成功', res.data)
      } catch (error) {
        console.error('登录失败', error)
      }
    },

    async testGetInfo() {
      try {
        const res = await getUserInfo()
        console.log('用户信息', res.data)
      } catch (error) {
        console.error('获取失败', error)
      }
    }
  }
}
</script>
```

---

## 🐛 常见问题

### 1. 服务器无法启动

```bash
# 检查端口占用
lsof -i:8080

# 如果被占用，杀死进程
kill -9 <PID>
```

### 2. 数据库连接失败

```bash
# 检查MySQL状态
sudo systemctl status mysql

# 重启MySQL
sudo systemctl restart mysql

# 检查数据库是否存在
mysql -uroot -p123456 -e "SHOW DATABASES LIKE 'bridge_detection';"
```

### 3. Session不生效

- 确保前端配置了 `withCredentials: true`
- 确保CORS配置正确（`config.yaml`中的`allow_credentials: true`）
- 检查Cookie是否正确发送（浏览器开发者工具 → Network → Cookies）

### 4. 401/403错误

- 401：未登录或Session过期 → 重新登录
- 403：权限不足 → 使用管理员账户或检查角色

---

## ✅ 测试检查清单

完整测试应该验证：

- [ ] 服务器成功启动
- [ ] 健康检查接口返回200
- [ ] 用户注册成功
- [ ] 重复用户名注册返回400
- [ ] 重复邮箱注册返回400
- [ ] 参数验证生效（邮箱格式、密码长度等）
- [ ] 用户登录成功并获得Cookie
- [ ] 密码错误返回401
- [ ] 携带Cookie可以访问认证接口
- [ ] 无Cookie访问返回401
- [ ] 用户信息获取正确
- [ ] 用户信息更新成功
- [ ] 密码修改后可用新密码登录
- [ ] 用户登出后Cookie失效
- [ ] 普通用户访问管理员接口返回403
- [ ] 管理员可以访问所有接口
- [ ] 用户列表分页正确
- [ ] 删除用户功能正常
- [ ] 提升管理员功能正常

---

**测试完成后，记得停止服务器（Ctrl+C）并清理测试数据！** 🎉
