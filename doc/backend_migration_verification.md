# 后端迁移验证清单（服务器）

本文档对应脚本：`verify_backend_migration.sh`。

目标：验证后端迁移是否有漏项，并确认服务可正常运行。
默认策略：非破坏性检查（不删库、不改表），事务冒烟会自动回滚。

## 1. 服务可用性

- 检查项：
  - 健康接口是否可访问（默认 `GET /api/v1/health`）
  - 后端端口是否处于监听状态
- 方法：
  - `curl` 重试健康检查 URL
  - 使用 `ss` / `netstat` / `lsof` 之一检查端口监听

## 2. 核心数据表完整性

- 检查项：
  - `users`
  - `bridges`
  - `drones`
  - `defects`
  - `reports`
  - `video_analysis_tasks`
  - `defect_observations`
- 方法：
  - `SHOW TABLES` 后逐表断言存在

## 3. 核心字段完整性

- 检查项（示例）：
  - `users`: `id`, `username`, `email`, `role`, `created_at`
  - `bridges`: `id`, `bridge_name`, `bridge_code`, `user_id`
  - `defects`: `id`, `bridge_id`, `defect_type`, `image_path`, `detected_at`
  - `reports`: `id`, `report_name`, `report_type`, `user_id`, `bridge_id`, `status`
  - `video_analysis_tasks`: `id`, `task_id`, `user_id`, `bridge_id`, `status`
  - `defect_observations`: `id`, `task_id`, `bridge_id`, `frame_no`, `defect_type`
- 方法：
  - 查询 `information_schema.columns`，逐字段判断存在性

## 4. 核心索引完整性

- 检查项：
  - `users.username`（唯一）
  - `users.email`（唯一）
  - `bridges.bridge_code`（唯一）
  - `reports.user_id`
  - `reports.bridge_id`
  - `video_analysis_tasks.task_id`（唯一）
  - `defect_observations.task_id`
  - `defect_observations.frame_no`
- 方法：
  - 查询 `information_schema.statistics`
  - 唯一索引检查会同时验证 `non_unique = 0`

## 5. 关键外键完整性

- 检查项：
  - `reports.user_id -> users.id`
  - `reports.bridge_id -> bridges.id`
- 方法：
  - 查询 `information_schema.key_column_usage` 验证引用关系

## 6. 默认管理员检查

- 检查项：
  - `users` 表中是否存在 `role='admin'` 账户
- 方法：
  - 直接统计 `SELECT COUNT(*) FROM users WHERE role='admin'`

## 7. 最小读写事务冒烟（回滚）

- 检查项：
  - 关键表是否可执行一次完整最小写入链路
  - 写入是否满足约束（字段、索引、外键）
  - 事务能否成功回滚
- 方法：
  - 事务内依次插入：`users -> bridges -> drones -> defects -> reports -> video_analysis_tasks -> defect_observations`
  - 末尾执行 `ROLLBACK`，不污染数据库

## 使用方式

```bash
cd /path/to/BridgeDefectDetectionSys
bash verify_backend_migration.sh
```

可用环境变量：

- `HEALTH_URL`：默认 `http://127.0.0.1:8080/api/v1/health`
- `PORT`：默认 `8080`
- `MYSQL_USER`：默认 `root`
- `MYSQL_PASSWORD`：默认 `123456`
- `MYSQL_DB`：默认 `bridge_detection`
- `HEALTH_RETRIES`：默认 `20`

## 判定标准

脚本输出 `PASS: 后端迁移验收通过` 且退出码为 `0`，表示验收通过；任一检查失败会立即输出 `FAIL: ...` 并以非 0 退出。
