# 算法模块改造说明（视频异步 Detect 版）

## 1. 目的

本文档面向算法模块，说明为了配合新版视频检测链路，算法侧需要完成哪些改造。

对应的总方案文档为：

- `BridgeDefectDetectionSys/doc/设计变更_视频算法异步对接方案.md`

本说明只保留算法侧必须知道的内容，便于快速实施。

---

## 2. 最终边界

请算法模块按以下边界实现：

- **后端负责抽帧**
- **后端按帧派发 detect 请求**
- **算法端只处理单帧 detect 请求**
- **算法端异步返回结构化结果**
- **算法端不负责视频抽帧**
- **算法端不返回 Base64 图片，不返回渲染结果图**
- **算法端不做 segment，只做 detect**

一句话概括：

> 算法模块从“处理整段视频”改为“异步消费后端派发的单帧 detect 任务”。

---

## 3. 你们需要新增/调整的能力

## 3.1 新增帧级异步 detect 接口

算法端需要提供接口：

`POST /algo/video/frames/detect`

用途：

- 接收后端发来的单帧 detect 请求
- 只做快速校验和入队
- 不要同步阻塞到推理完成后再返回

### 请求体示例

```json
{
  "request_id": "frame_req_123_45",
  "task_id": "video_task_123",
  "bridge_id": 101,
  "frame_no": 45,
  "timestamp_ms": 45000,
  "frame_ref": "file:///data/bridge_detect/frames/video_task_123/frame_000045.jpg",
  "model_name": "baseline",
  "conf": 0.25,
  "callback_base_url": "http://backend.internal/api/v1/detection/video/callback"
}
```

### 成功响应示例

```json
{
  "status": "accepted",
  "request_id": "frame_req_123_45",
  "task_id": "video_task_123",
  "queued_at": "2026-04-12T14:00:00Z"
}
```

### 最低要求

- 校验 `frame_ref` 可访问
- 校验必要字段完整
- 生成内部队列任务
- 返回 `accepted`

---

## 3.2 新增算法侧内部队列/worker

虽然外部只看到一个 HTTP 接口，但算法端内部必须改成异步消费模型：

- 接口线程只负责接收任务并入队
- worker 线程/进程负责真正执行 detect
- worker 数量可配置
- 队列长度和并发要可控制

### 至少要支持的内部状态

建议内部至少有：

- `queued`
- `processing`
- `completed`
- `failed`

如果你们已有现成队列框架，可以复用，不要求一定新增数据库表。

---

## 3.3 新增回调后端的能力

算法端在处理过程中，至少需要回调后端以下事件：

### 1）单帧入队确认

`POST /api/v1/detection/video/callback/frame-queued`

### 2）单帧开始处理

`POST /api/v1/detection/video/callback/frame-started`

### 3）单帧结果回调（核心）

`POST /api/v1/detection/video/callback/frame-result`

### 4）任务级失败回调（建议）

`POST /api/v1/detection/video/callback/task-failed`

### 5）任务级完成回调（可选增强）

`POST /api/v1/detection/video/callback/task-completed`

> 第一版里，最核心必须先做通的是：`frame-result`

---

## 4. 单帧结果回调要求

## 4.1 成功回调

成功时，算法端必须回调：

```json
{
  "request_id": "frame_req_123_45",
  "task_id": "video_task_123",
  "frame_no": 45,
  "timestamp_ms": 45000,
  "status": "success",
  "queue_latency_ms": 180,
  "detect_total_ms": 1500,
  "decode_ms": 120,
  "infer_ms": 1080,
  "postprocess_ms": 300,
  "frame_ref": "file:///data/bridge_detect/frames/video_task_123/frame_000045.jpg",
  "yolo_bboxes": [
    {
      "box_id": 0,
      "class_idx": 0,
      "class_name": "Crack",
      "yolo_coords": [0.51, 0.48, 0.18, 0.07],
      "confidence": 0.92
    }
  ]
}
```

## 4.2 失败回调

失败时至少回调：

```json
{
  "request_id": "frame_req_123_45",
  "task_id": "video_task_123",
  "frame_no": 45,
  "timestamp_ms": 45000,
  "status": "failed",
  "error_message": "decode frame failed"
}
```

## 4.3 必须遵守的约束

- `request_id` 必须原样回传
- `frame_no`、`timestamp_ms` 必须原样回传
- `frame_ref` 必须与输入帧对应
- 结果中只返回结构化 bbox
- 不返回 Base64 图片
- 不返回渲染结果图
- 不返回 segment/mask 结果

---

## 5. 耗时字段必须补齐

为了后续定位瓶颈，算法端回调时必须带这些耗时：

- `queue_latency_ms`
- `detect_total_ms`
- `decode_ms`
- `infer_ms`
- `postprocess_ms`

说明：

- `queue_latency_ms`：请求进入算法侧后，排队等待开始处理的时间
- `detect_total_ms`：单帧 detect 总耗时
- `decode_ms`：读图/解码耗时
- `infer_ms`：模型推理耗时
- `postprocess_ms`：后处理耗时

这些字段不是可选优化项，而是联调和调参必须项。

---

## 6. 旧行为需要删除或停止依赖

请算法模块停止沿用以下旧行为：

### 6.1 不再接收完整视频做本地抽帧

新版视频链路中：

- 视频抽帧已经交给后端
- 算法端不要再假设自己会接收到完整视频并自行抽帧

### 6.2 不再返回 Base64 图片

视频模式下不要再返回：

- `image_base64`
- 带框渲染图
- 大图二进制内容

### 6.3 不再做 segment

视频模式下只做 detect，不做：

- mask 分割
- 长度/面积测量
- 高精度几何结果输出

---

## 7. 取消、重试、幂等要求

## 7.1 幂等

算法端要接受后端可能重复发送同一个 `request_id` 的情况。

建议：

- 以 `request_id` 作为算法侧幂等键
- 重复请求不要重复创建多份内部任务
- 如果该请求已完成，可直接返回已接受/已存在状态

## 7.2 回调重试

算法端回调后端失败时，建议在以下场景自动重试：

- HTTP 超时
- 后端返回 `5xx`
- 网络异常

不建议无限重试：

- 后端返回 `4xx`
- `request_id/task_id` 不存在
- 签名校验失败

## 7.3 取消能力

算法端建议支持：

`POST /algo/video/tasks/{task_id}/cancel`

语义：

- 停止继续处理该视频任务相关的未完成帧请求
- 已在运行中的任务尽量安全停止，或跑完后不再继续后续消费

第一版如果细粒度取消单帧太复杂，可以先只支持整任务取消。

---

## 8. 算法模块最少改造清单

如果按“先跑通，再优化”的原则，算法侧至少要完成以下 6 件事：

### 必做 1：新增帧级 detect 入队接口
- `POST /algo/video/frames/detect`

### 必做 2：实现内部异步队列/worker
- 接口快速返回
- worker 异步处理

### 必做 3：实现 `frame-result` 回调
- 成功/失败都能回调

### 必做 4：结果只返回结构化 bbox
- 不返图
- 不返 segment

### 必做 5：补齐耗时统计
- queue/decode/infer/postprocess/total

### 必做 6：支持基本幂等
- 按 `request_id` 去重

---

## 9. 推荐实施顺序

### 第一阶段：打通最小闭环

- 新增 `POST /algo/video/frames/detect`
- 请求入队
- worker 异步消费
- 回调 `frame-result`

### 第二阶段：补全状态事件

- 回调 `frame-queued`
- 回调 `frame-started`
- 支持 `task-failed`
- 可选支持 `task-completed`

### 第三阶段：补全稳定性

- 幂等去重
- 回调失败重试
- 任务取消
- 队列并发与积压控制

---

## 10. 最终一句话要求

算法模块这次改造的核心不是“提升模型能力”，而是：

> **把视频检测场景改造成一个稳定的单帧异步 detect worker。**

只要做到：

- 能稳定接收帧任务
- 能异步 detect
- 能回调结构化结果
- 能提供耗时数据

就满足第一阶段联调要求。
