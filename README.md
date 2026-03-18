# 桥梁病害检测系统

> 基于低空无人机视觉的桥梁表观病害精细化智能检测算法系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.8+-3776AB?logo=python)](https://www.python.org/)

## 项目简介

本项目是为**服务外包大赛**（杭州师范大学赛题）开发的桥梁病害智能检测系统。系统通过深度学习算法，自动识别和分析无人机采集的桥梁图像中的各类病害（裂缝、剥落、锈蚀、泛碱等），实现像素级精准分割和量化计算。

### 核心特性

- **多类别病害检测**：精准识别混凝土裂缝、剥落/掉块、钢筋裸露、钢结构锈蚀、泛碱等常见病害
- **像素级分割**：基于YOLO检测 + SAM分割的两阶段算法，实现高精度病害分割
- **实时视频流处理**：支持无人机视频流实时监测和缺陷检测
- **智能分析报告**：集成大模型API，自动生成桥梁健康状况分析报告
- **3D可视化展示**：基于Three.js的桥梁3D模型展示和缺陷标注
- **权限管理**：区分普通用户和管理员，保证数据安全

### 技术亮点

- **YOLO目标检测**：快速定位病害区域，支持小目标和极端长宽比检测
- **SAM实例分割**：精细化分割病害边缘，支持后续几何参数计算
- **图像增强**：解决桥下光照不均、运动模糊等复杂环境干扰
- **量化计算**：自动计算病害物理尺寸（长度、宽度、面积）

---

## 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                   前端（Vue.js）                        │
│  • 用户界面  • Three.js 3D可视化  • 数据统计图表        │
└───────────────────┬─────────────────────────────────────┘
                    │ HTTP / WebSocket
┌───────────────────▼─────────────────────────────────────┐
│               Go后端服务（Gin框架）                      │
│  • 用户管理  • 桥梁管理  • 无人机管理  • 数据统计       │
│  • 文件存储  • 权限控制  • API接口                      │
└───────┬─────────────────────────┬───────────────────────┘
        │                         │ HTTP调用
        ▼                         ▼
┌───────────────┐         ┌──────────────────────────────┐
│  MySQL 8.0    │         │   Python算法服务（FastAPI）   │
│  数据持久化    │         │  • YOLO目标检测              │
│               │         │  • SAM实例分割               │
│               │         │  • 图像增强预处理             │
│               │         │  • 尺寸量化计算               │
└───────────────┘         └──────────────────────────────┘
```

---

## 目录结构

```
BridgeDefectDetectionSys/
│
├── README.md                   # 项目说明文档（本文件）
│
├── doc/                        # 📄 项目文档
│   ├── 需求文档.md              # 比赛需求和功能说明
│   └── 后端技术栈方案.md        # 后端详细技术设计方案
│
├── environment-setup/          # 🔧 环境配置脚本
│   ├── README.md               # 环境安装使用指南
│   ├── setup_basic_tools.sh    # 基础开发工具安装脚本
│   ├── setup_backend_env.sh    # Go + MySQL环境一键安装
│   ├── import_database.sh      # 数据库导入工具
│   ├── go.mod                  # Go项目依赖配置
│   ├── go.sum                  # Go依赖版本锁定
│   └── mysql_config.txt        # MySQL连接配置信息
│
├── backend/                    # 🚀 Go后端服务（待开发）
│   ├── main.go                 # 程序入口
│   ├── config.yaml             # 配置文件
│   ├── models/                 # 数据模型
│   ├── handlers/               # API控制器
│   ├── middleware/             # 中间件
│   ├── utils/                  # 工具函数
│   └── uploads/                # 文件存储目录
│
├── algorithm/                  # 🧠 Python算法服务（待开发）
│   ├── main.py                 # FastAPI入口
│   ├── models/                 # 模型文件
│   ├── detect.py               # YOLO检测模块
│   ├── segment.py              # SAM分割模块
│   ├── enhance.py              # 图像增强模块
│   └── requirements.txt        # Python依赖
│
└── frontend/                   # 🎨 Vue前端（待开发）
    ├── src/
    ├── public/
    └── package.json
```

> **当前状态**：项目处于规划阶段，已完成技术方案设计和环境配置脚本。

---

## 快速上手

### 环境要求

- **操作系统**：Ubuntu 22.04 / 24.04（推荐）或其他Linux发行版
- **硬件要求**：
  - CPU：4核及以上
  - 内存：至少8GB（算法训练推荐16GB+）
  - GPU：NVIDIA GPU（推荐RTX 4090，支持CUDA）
  - 磁盘：至少20GB可用空间

### 第一步：安装基础工具

如果你使用的是全新的Ubuntu虚拟机，首先安装开发必备工具：

```bash
cd environment-setup
sudo ./setup_basic_tools.sh
```

**包含工具**：vim, git, gcc, g++, make, cmake, curl, wget, htop, tree 等

### 第二步：安装后端环境

一键安装Go语言和MySQL数据库：

```bash
# 运行安装脚本
sudo ./setup_backend_env.sh

# 重新加载环境变量
source ~/.bashrc
```

**安装内容**：
- Go 1.25.0
- MySQL 8.0（用户名：root，密码：123456）
- Go依赖包：Gin, GORM, Sessions, CORS, bcrypt, WebSocket

### 第三步：验证安装

```bash
# 检查Go版本
go version

# 检查MySQL连接
mysql -uroot -p123456

# 检查Go依赖
cd environment-setup
go mod verify
```

### 第四步：导入数据库（可选）

如果有现成的数据库文件：

```bash
./import_database.sh your_database.sql
```

或手动创建数据库：

```bash
mysql -uroot -p123456 -e "CREATE DATABASE bridge_detection CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 第五步：安装前端环境

**1. 安装 Node.js v20运行环境**

如果系统中尚未安装 Node.js，请执行以下命令：

```bash
# 更新系统包列表
sudo apt update

# 安装 Node.js 和 npm
# 1. 下载并添加 Node.js 20 的官方源
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -

# 2. 安装 Node.js
sudo apt-get install -y nodejs
# 3. 安装 npm
sudo apt install npm -y

# 验证安装（建议 Node 版本 v20）
node -v
npm -v
```



## 开发指南

### 后端开发

```bash
# 1. 进入后端目录（待创建）
cd backend

# 2. 安装依赖
go mod tidy

# 3. 配置数据库连接
# 编辑 config.yaml 文件

# 4. 启动服务
go run main.go
```

**默认端口**：http://localhost:8080

### 算法服务开发

```bash
# 1. 进入算法目录（待创建）
cd algorithm

# 2. 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 3. 安装依赖
pip install -r requirements.txt

# 4. 启动算法服务
python main.py
```

**默认端口**：http://localhost:8000

### 前端开发

```bash
# 1. 进入前端目录（待创建）
cd frontend

# 2. 安装依赖
npm install

# 3. 启动开发服务器
npm run dev
```

**默认端口**：http://localhost:5173

---

## 核心功能模块

### 1. 用户管理
- 用户注册/登录/登出
- 角色权限控制（普通用户/管理员）
- 个人信息管理

### 2. 桥梁管理
- 桥梁信息CRUD
- 3D模型可视化展示
- 历史缺陷记录查询

### 3. 无人机管理
- 无人机信息管理
- 视频流源配置

### 4. 图片上传识别
- 本地图片上传
- YOLO + SAM检测分割
- 缺陷尺寸量化计算
- 结果可视化展示

### 5. 视频流识别
- 实时视频流接入
- 逐帧检测分析
- WebSocket实时推送结果

### 6. 智能分析
- 集成大模型API（Claude/GPT）
- 自动生成缺陷分析报告
- 维护建议生成

### 7. 数据统计
- 缺陷数量统计
- 类型分布可视化
- 时间趋势分析

### 8. 报表生成
- 检测报告导出
- PDF格式输出（可选）

---

## API接口文档

### 用户相关

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/register` | POST | 用户注册 |
| `/api/login` | POST | 用户登录 |
| `/api/logout` | POST | 用户登出 |
| `/api/user/info` | GET | 获取个人信息 |

### 桥梁管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/bridges` | GET | 获取桥梁列表 |
| `/api/bridges` | POST | 添加桥梁 |
| `/api/bridges/:id` | GET | 获取桥梁详情 |
| `/api/bridges/:id` | PUT | 更新桥梁 |
| `/api/bridges/:id` | DELETE | 删除桥梁 |
| `/api/bridges/:id/defects` | GET | 获取桥梁历史缺陷 |

### 检测功能

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/detect/image` | POST | 上传图片检测 |
| `/api/detect/video/start` | POST | 开始视频流检测 |
| `/api/detect/video/stop` | POST | 停止视频流检测 |
| `/api/detect/analyze` | POST | 智能分析（调用大模型） |

### 数据统计

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/stats/overview` | GET | 主页统计数据 |

---

## 数据库设计

### 核心数据表

```sql
-- 用户表
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    real_name VARCHAR(50) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100) UNIQUE NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 桥梁表
CREATE TABLE bridges (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(255),
    description TEXT,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 无人机表
CREATE TABLE drones (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    model VARCHAR(100),
    stream_url VARCHAR(255),
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 缺陷记录表
CREATE TABLE defects (
    id INT PRIMARY KEY AUTO_INCREMENT,
    bridge_id INT NOT NULL,
    defect_type VARCHAR(50) NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    result_path VARCHAR(255),
    bbox TEXT,
    length DECIMAL(10,4),
    width DECIMAL(10,4),
    area DECIMAL(10,4),
    confidence DECIMAL(5,4),
    detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bridge_id) REFERENCES bridges(id)
);
```

---

## 算法技术方案

### 检测算法：YOLO系列

- **模型选择**：YOLOv8 / YOLOv9
- **优化方向**：
  - 小目标检测优化
  - 极端长宽比（细长裂缝）适配
  - 多尺度特征融合

### 分割算法：SAM

- **模型**：Segment Anything Model
- **输入**：YOLO检测框 + 提示点
- **输出**：精确的病害掩膜（Mask）

### 图像增强

- 光照均衡化（CLAHE）
- 去模糊（Wiener滤波）
- 噪声抑制（双边滤波）

### 量化计算

```python
# 像素尺寸 → 物理尺寸
physical_length = pixel_length * pixel_to_mm_ratio
physical_width = pixel_width * pixel_to_mm_ratio
physical_area = pixel_area * (pixel_to_mm_ratio ** 2)
```

---

## 常见问题

### 1. Go环境变量未生效？

```bash
source ~/.bashrc
# 或退出终端重新登录
```

### 2. MySQL连接失败？

```bash
# 重启MySQL服务
sudo systemctl restart mysql

# 检查状态
sudo systemctl status mysql
```

### 3. 权限不足？

所有安装脚本都需要使用 `sudo` 权限运行。

### 4. 如何修改MySQL密码？

```bash
mysql -uroot -p123456
ALTER USER 'root'@'localhost' IDENTIFIED BY '你的新密码';
FLUSH PRIVILEGES;
```

---

## 技术选型说明

### 为什么选择Go？

- 高性能、高并发
- 简单易学，开发效率高
- 编译型语言，部署方便

### 为什么选择Gin框架？

- 轻量级、高性能
- 社区活跃，文档完善
- 中间件丰富

### 为什么使用Session而不是JWT？

- 比赛项目，简单够用
- 无需分布式部署
- 减少前端存储复杂度

### 为什么用本地文件系统而不是MinIO？

- 避免过度设计
- 降低部署复杂度
- 满足项目需求

---

## 项目进度

- [x] 项目规划和技术方案设计
- [x] 环境配置脚本开发
- [ ] 数据库表结构设计
- [ ] Go后端框架搭建
- [ ] 用户认证模块开发
- [ ] 桥梁管理模块开发
- [ ] Python算法服务开发
- [ ] YOLO模型训练和优化
- [ ] SAM分割集成
- [ ] 前端界面开发
- [ ] 系统联调测试
- [ ] 项目文档整理

---

## 贡献指南

本项目为服务外包大赛参赛项目，暂不接受外部贡献。

---

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

---

## 联系方式

如有问题，请通过以下方式联系：

- **项目仓库**：[GitHub Repository]
- **技术文档**：查看 `doc/` 目录
- **环境配置**：查看 `environment-setup/README.md`

---

## 致谢

感谢杭州师范大学提供的赛题支持和技术指导。

---

**最后更新时间**：2026-03-12
