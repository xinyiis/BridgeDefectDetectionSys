import os
os.environ["OMP_NUM_THREADS"] = "1"
os.environ["MKL_NUM_THREADS"] = "1"
os.environ["YOLO_VERBOSE"] = "False"
import sys
import time
import cv2
import numpy as np
import base64
import json
import torch
import asyncio
import httpx
import io
import importlib.util  # <--- 修正 1：必须显式导入这个用于动态加载
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from typing import List, Optional, Dict, Any
from fastapi import FastAPI, File, UploadFile, Form
from pydantic import BaseModel
from PIL import Image
import supervision as sv
import uvicorn

sys.path.insert(0, "/root/miniconda3/lib/python3.12/site-packages/ultralytics")

from ultralytics import YOLO

# ==========================================
# 0. 核心配置与路径定义
# ==========================================
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
BPE_PATH = "assets/bpe_simple_vocab_16e6.txt.gz"
DISEASE_NAMES = ["Crack", "Breakage", "Comb", "Hole", "Reinforcement", "Seepage"]
# ==========================================
# 1. 路径定义 (确保指向 sam3 文件夹的上一层)
# ==========================================
TEACHER_PARENT = "/root/autodl-tmp/NLP/sam"
STUDENT_PARENT = "/root/autodl-tmp/efficientsam3/sam3"

def safe_load_student():
    print("正在加载蒸馏版 (Student) 模型环境...")
    # 强制清理
    if "sam3" in sys.modules:
        del sys.modules["sam3"]
    for mod in list(sys.modules.keys()):
        if mod.startswith("sam3."):
            del sys.modules[mod]
    
    # 暂时切换路径进行模块抓取
    sys.path.insert(0, STUDENT_PARENT)
    try:
        from sam3.model_builder import build_efficientsam3_image_model
        from sam3.model.sam3_image_processor import Sam3Processor
        return build_efficientsam3_image_model, Sam3Processor
    finally:
        sys.path.pop(0)

def safe_load_teacher():
    print("正在加载原版 (Teacher) 模型环境...")
    if "sam3" in sys.modules:
        del sys.modules["sam3"]
    for mod in list(sys.modules.keys()):
        if mod.startswith("sam3."):
            del sys.modules[mod]
            
    sys.path.insert(0, TEACHER_PARENT)
    try:
        from sam3.model_builder import build_sam3_image_model
        from sam3.model.sam3_image_processor import Sam3Processor
        return build_sam3_image_model, Sam3Processor
    finally:
        sys.path.pop(0)

# ==========================================
# 2. 执行初始化 (关键点：初始化时必须在路径上下文内)
# ==========================================

# --- A. 初始化 Student ---
sys.path.insert(0, STUDENT_PARENT) # <--- 必须在这里手动插入
if "sam3" in sys.modules: del sys.modules["sam3"] # 确保 clean
build_fn_student, proc_cls_student = safe_load_student()

print("正在初始化 Student 权重...")
sam3_model_student = build_fn_student(
    checkpoint_path="/root/autodl-tmp/efficientsam3/output/efficient_sam3_repvit_m1_1_mobileclip_s1.pth",
    bpe_path=BPE_PATH,
    device=DEVICE,
    backbone_type="repvit", 
    model_name="m1_1",
    enable_inst_interactivity=True
)
sam3_processor_student = proc_cls_student(sam3_model_student, confidence_threshold=0.3, device=DEVICE)
sys.path.pop(0) # 初始化完了再弹出

# --- B. 初始化 Teacher ---
sys.path.insert(0, TEACHER_PARENT) # <--- 切换到 Teacher 路径
if "sam3" in sys.modules: del sys.modules["sam3"]
build_fn_teacher, proc_cls_teacher = safe_load_teacher()

print("正在初始化 Teacher 权重...")
sam3_model_teacher = build_fn_teacher(
    checkpoint_path="weights/sam3.pt",
    bpe_path=BPE_PATH,
    device=DEVICE,
    enable_inst_interactivity=True
)
sam3_processor_teacher = proc_cls_teacher(sam3_model_teacher, confidence_threshold=0.3, device=DEVICE)
sys.path.pop(0)
# 加載 YOLO（新版优化：启动时预加载到 GPU）
MODELS = {
    "baseline": YOLO("xe/best.pt").to(DEVICE),
    "my_trained": YOLO("weights/last.pt").to(DEVICE)
}





# ==========================================
# 2. 異步隊列模型定義
# ==========================================
task_queue = asyncio.Queue()

class AsyncTask(BaseModel):
    task_type: str  # "preprocess", "detect", "segment", "video_frame_detect"
    request_id: str
    callback_url: str = ""
    payload: Dict[str, Any]
    enter_time: float = 0.0

class VideoFrameDetectRequest(BaseModel):
    request_id: str
    task_id: str
    bridge_id: int
    frame_no: int
    timestamp_ms: int
    frame_ref: str
    model_name: str = "baseline"
    conf: float = 0.25
    callback_base_url: str

def normalize_frame_ref(frame_ref: str) -> str:
    if frame_ref.startswith("file://"):
        return frame_ref[len("file://"):]
    return frame_ref

def build_callback_url(callback_base_url: str, event: str) -> str:
    return callback_base_url.rstrip("/") + "/" + event

async def post_callback(client: httpx.AsyncClient, url: str, payload: Dict[str, Any]) -> None:
    try:
        await client.post(url, json=payload)
    except Exception as callback_error:
        print(f"⚠️ 回調失敗 {url}: {callback_error}")

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 啟動後台 Worker
    worker_task = asyncio.create_task(algo_worker())
    yield
    worker_task.cancel()

app = FastAPI(title="橋梁算法中台 - 異步隊列全功能版（合并版）", lifespan=lifespan)

# ==========================================
# 3. 核心：後台消費 Worker (含詳細審計打印)
# ==========================================
async def algo_worker():
    async with httpx.AsyncClient() as client:
        while True:
            task: AsyncTask = await task_queue.get()
            t_start = time.time()
            q_latency = (t_start - task.enter_time) * 1000

            try:
                # 執行業務邏輯並獲取詳細耗時
                if task.task_type == "preprocess":
                    result, detail_times = await handle_preprocess(task.payload)
                elif task.task_type == "detect":
                    result, detail_times = await handle_detect(task.payload)
                elif task.task_type == "segment":
                    result, detail_times = await handle_segment(task.payload)
                elif task.task_type == "video_frame_detect":
                    # 视频帧检测：先回调 frame-started
                    await post_callback(client, build_callback_url(task.payload["callback_base_url"], "frame-started"), {
                        "request_id": task.request_id,
                        "task_id": task.payload["task_id"],
                        "frame_no": task.payload["frame_no"],
                        "timestamp_ms": task.payload["timestamp_ms"],
                        "status": "processing",
                        "started_at": datetime.now(timezone.utc).isoformat()
                    })
                    result, detail_times = await handle_detect(task.payload)
                else:
                    raise ValueError(f"不支持的任務類型: {task.task_type}")

                t_done = time.time()
                total_process_ms = (t_done - t_start) * 1000

                # --- 核心：格式化打印報表 ---
                print("\n" + "—"*45)
                print(f"🕵️  算法性能審計 | 任務: {task.task_type.upper()}")
                print(f"  ID: {task.request_id}")
                print(f"  1. 排隊等待 (Queue):    {q_latency:.2f} ms")
                print(f"  2. 圖像解碼 (Decode):   {detail_times.get('decode_ms', 0):.2f} ms")

                # 根據任務類型顯示特定模型耗時
                if task.task_type in ("detect", "video_frame_detect"):
                    print(f"  3. YOLO 详细耗时:")
                    print(f"     └─ 图像预处理: {detail_times.get('decode_ms', 0):.2f} ms")
                    print(f"     └─ 模型准备:   {detail_times.get('model_prepare_ms', 0):.2f} ms")
                    print(f"     └─ 纯显卡推理: {detail_times.get('pure_infer_ms', 0):.2f} ms")
                    print(f"     └─ 后处理提取: {detail_times.get('post_process_ms', 0):.2f} ms")

                if task.task_type == "segment":
                    sam_total = detail_times.get('sam_feature_ms', 0) + detail_times.get('sam_infer_ms', 0)
                    print(f"  3. SAM3 分割 (Segment): {sam_total:.2f} ms")
                    print(f"     └─ 特徵提取: {detail_times.get('sam_feature_ms', 0):.2f} ms")
                    print(f"     └─ 掩碼推理: {detail_times.get('sam_infer_ms', 0):.2f} ms")

                print(f"  4. 結果渲染 (Render):   {detail_times.get('render_ms', 0):.2f} ms")
                print(f"  🚀 單幀總計 (Total):    {total_process_ms:.2f} ms")
                print("—"*45 + "\n")

                # 视频帧检测：回调 frame-result
                if task.task_type == "video_frame_detect":
                    callback_payload = {
                        "request_id": task.request_id,
                        "task_id": task.payload["task_id"],
                        "frame_no": task.payload["frame_no"],
                        "timestamp_ms": task.payload["timestamp_ms"],
                        "status": "success",
                        "queue_latency_ms": int(round(q_latency)),
                        "detect_total_ms": int(round(total_process_ms)),
                        "decode_ms": int(round(detail_times.get("decode_ms", 0))),
                        "infer_ms": int(round(detail_times.get("infer_ms", 0))),
                        "postprocess_ms": int(round(detail_times.get("postprocess_ms", detail_times.get("render_ms", 0)))),
                        "frame_ref": task.payload["frame_ref"],
                        "yolo_bboxes": result.get("yolo_bboxes", [])
                    }
                    await post_callback(
                        client,
                        build_callback_url(task.payload["callback_base_url"], "frame-result"),
                        callback_payload
                    )
                else:
                    # 普通任务：回调原有格式
                    callback_payload = {
                        "request_id": task.request_id,
                        "status": "success",
                        "task_type": task.task_type,
                        "metrics": {
                            "queue_latency_ms": round(q_latency, 2),
                            "total_process_ms": round(total_process_ms, 2),
                            **detail_times
                        },
                        "results": result
                    }
                    await client.post(task.callback_url, json=callback_payload)

            except Exception as e:
                print(f"❌ 任務失敗: {str(e)}")
                if task.task_type == "video_frame_detect":
                    await post_callback(client, build_callback_url(task.payload["callback_base_url"], "frame-result"), {
                        "request_id": task.request_id,
                        "task_id": task.payload.get("task_id", ""),
                        "frame_no": task.payload.get("frame_no", 0),
                        "timestamp_ms": task.payload.get("timestamp_ms", 0),
                        "status": "failed",
                        "frame_ref": task.payload.get("frame_ref", ""),
                        "error_message": str(e)
                    })
            finally:
                task_queue.task_done()

# ==========================================
# 4. 業務邏輯實現 (封裝耗時統計)
# ==========================================

async def handle_preprocess(p):
    t0 = time.time()
    img_bytes = base64.b64decode(p['img_b64'])
    img = cv2.imdecode(np.frombuffer(img_bytes, np.uint8), cv2.IMREAD_COLOR)
    t1 = time.time()

    if p['mode'] == "blur":
        size = 4
        kernel = np.zeros((size, size))
        kernel[int((size - 1) / 2), :] = np.ones(size) / size
        img = cv2.filter2D(img, -1, kernel)
        img = cv2.fastNlMeansDenoisingColored(img, None, 5, 5, 7, 21)
    elif p['mode'] == "light":
        lab = cv2.cvtColor(img, cv2.COLOR_BGR2LAB)
        l, a, b = cv2.split(lab)
        clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
        img = cv2.cvtColor(cv2.merge((clahe.apply(l), a, b)), cv2.COLOR_LAB2BGR)

    t2 = time.time()
    res_b64 = cv2_to_base64(img)
    t3 = time.time()

    return {"image_base64": res_b64}, {
        "decode_ms": (t1 - t0) * 1000,
        "algo_ms": (t2 - t1) * 1000,
        "encode_ms": (t3 - t2) * 1000
    }

# 在 handle_detect 中加入强制性能优化
async def handle_detect(p):
    return _sync_detect_logic(p)

def _sync_detect_logic(p):
    t_start = time.time()

    # --- 1. 解码 ---
    img_bytes = base64.b64decode(p['img_b64'])
    img = cv2.imdecode(np.frombuffer(img_bytes, np.uint8), cv2.IMREAD_COLOR)
    h, w = img.shape[:2]
    if max(h, w) > 1280:
        scale = 1280 / max(h, w)
        img = cv2.resize(img, (int(w * scale), int(h * scale)))

    t_decode = time.time()

    # --- 2. 模型准备 ---
    model = MODELS.get(p['model_name'], MODELS["baseline"])
    t_model_ready = time.time()

    # --- 3. 核心推理（新版优化：CUDA 同步确保准确计时）---
    if torch.cuda.is_available():
        torch.cuda.synchronize()

    t_infer_start = time.time()
    results = model.predict(img, conf=p['conf'], device=DEVICE, verbose=False)[0]

    if torch.cuda.is_available():
        torch.cuda.synchronize()
    t_infer_end = time.time()

    # --- 4. 后处理 ---
    bboxes = []
    for i, box in enumerate(results.boxes):
        idx = int(box.cls[0])
        bboxes.append({
            "box_id": i, "class_idx": idx, "class_name": DISEASE_NAMES[idx],
            "confidence": round(float(box.conf[0]), 4), "yolo_coords": box.xywhn[0].tolist()
        })
    t_post_process = time.time()

    # # --- 5. 渲染 ---
    # plot_img = results.plot()
    # _, buffer = cv2.imencode('.jpg', plot_img, [int(cv2.IMWRITE_JPEG_QUALITY), 85])
    # plot_b64 = base64.b64encode(buffer).decode('utf-8')
    # t_render = time.time()

    return {
        "model_used": p['model_name'],
        "yolo_bboxes": bboxes,
        #"image_results": plot_b64
    }, {
        "decode_ms": (t_decode - t_start) * 1000,
        "model_prepare_ms": (t_model_ready - t_decode) * 1000,
        "pure_infer_ms": (t_infer_end - t_infer_start) * 1000,
        "post_process_ms": (t_post_process - t_infer_end) * 1000,
        #"render_ms": (t_render - t_post_process) * 1000,
        "infer_ms": (t_infer_end - t_decode) * 1000
    }

async def handle_segment(p):
    t0 = time.time()
    img_bytes = base64.b64decode(p['img_b64'])
    img_cv2 = cv2.imdecode(np.frombuffer(img_bytes, np.uint8), cv2.IMREAD_COLOR)
    h, w = img_cv2.shape[:2]

    # 根据请求选择模型
    if p.get('model_type') == "teacher":
        model = sam3_model_teacher
        processor = sam3_processor_teacher
    else:
        model = sam3_model_student
        processor = sam3_processor_student

    # SAM3 特徵提取 (Feature Encoding)
    img_pil = Image.open(io.BytesIO(img_bytes)).convert('RGB')
    inference_state = processor.set_image(img_pil)
    t1 = time.time()

    # 解析 BBOX (从 YOLO 归一化坐标转像素坐标)
    bboxes_data = json.loads(p['bboxes_json'])
    input_boxes = []
    input_class_ids = []
    for item in bboxes_data:
        xc, yc, bw, bh = item['yolo_coords']
        # 转为 [x1, y1, x2, y2]
        input_boxes.append([(xc-bw/2)*w, (yc-bh/2)*h, (xc+bw/2)*w, (yc+bh/2)*h])
        input_class_ids.append(item['class_idx'])

    if not input_boxes:
        return {"fusion_image": cv2_to_base64(img_cv2), "individual_masks": []}, {"total_ms": (time.time()-t0)*1000}

    # 多 Box 同时推理 (Batch Inference)
    masks, scores, _ = model.predict_inst(
        inference_state, 
        box=np.array(input_boxes, dtype=np.float32), 
        multimask_output=False
    )
    t2 = time.time()

    # 结果渲染准备
    if masks.ndim == 4:
        masks = masks.squeeze(1)
    
    masks_np = masks.cpu().numpy() if hasattr(masks, "cpu") else masks
    
    # 使用 Supervision 进行融合渲染
    detections = sv.Detections(
        xyxy=np.array(input_boxes, dtype=np.float32), 
        mask=masks_np.astype(bool), 
        class_id=np.array(input_class_ids)
    )
    
    mask_ann = sv.MaskAnnotator(opacity=p['alpha'])
    # 将 PIL 转 OpenCV 格式进行标注
    annotated = mask_ann.annotate(scene=cv2.cvtColor(img_cv2, cv2.COLOR_BGR2RGB), detections=detections)
    
    # 生成个体 Mask (黑白图片)
    individual_masks = []
    for i in range(len(masks_np)):
        m255 = (masks_np[i] * 255).astype(np.uint8)
        # 获取类别名称
        cls_name = DISEASE_NAMES[input_class_ids[i]] if input_class_ids[i] < len(DISEASE_NAMES) else "Unknown"
        
        individual_masks.append({
            "box_index": bboxes_data[i].get('box_id', i),
            "class_name": cls_name,
            "mask_base64": cv2_to_base64(m255)
        })

    t3 = time.time()
    return {
        "fusion_image": cv2_to_base64(cv2.cvtColor(annotated, cv2.COLOR_RGB2BGR)),
        "individual_masks": individual_masks
    }, {
        "sam_feature_ms": (t1 - t0) * 1000,
        "sam_infer_ms": (t2 - t1) * 1000,
        "render_ms": (t3 - t2) * 1000
    }

# ==========================================
# 5. API 路由 (接收請求入隊)
# ==========================================

@app.post("/algo/preprocess", tags=["1.預處理"])
async def api_preprocess(callback_url: str = Form(...), request_id: str = Form(...), mode: str = Form(...), file: UploadFile = File(...)):
    img_b64 = base64.b64encode(await file.read()).decode()
    await task_queue.put(AsyncTask(
        task_type="preprocess", request_id=request_id, callback_url=callback_url,
        payload={"mode": mode, "img_b64": img_b64}, enter_time=time.time()
    ))
    return {"status": "accepted", "request_id": request_id}

@app.post("/algo/detect", tags=["2.檢測"])
async def api_detect(callback_url: str = Form(...), request_id: str = Form(...), model_name: str = Form("my_trained"), conf: float = Form(0.25), file: UploadFile = File(...)):
    img_b64 = base64.b64encode(await file.read()).decode()
    await task_queue.put(AsyncTask(
        task_type="detect", request_id=request_id, callback_url=callback_url,
        payload={"model_name": model_name, "conf": conf, "img_b64": img_b64}, enter_time=time.time()
    ))
    return {"status": "accepted", "request_id": request_id}

@app.post("/algo/segment", tags=["3.分割"])
async def api_segment(
    bboxes_json: str = Form(...),
    alpha: float = Form(0.5),
    model_type: str = Form("student"),
    file: UploadFile = File(...),
):
    img_b64 = base64.b64encode(await file.read()).decode()
    result, _ = await handle_segment({
        "bboxes_json": bboxes_json,
        "alpha": alpha,
        "img_b64": img_b64,
        "model_type": model_type,
    })
    return {"status": "success", **result}

@app.post("/algo/video/frames/detect", tags=["4.視頻幀異步檢測"])
async def api_video_frame_detect(req: VideoFrameDetectRequest):
    frame_path = normalize_frame_ref(req.frame_ref)
    if not os.path.isfile(frame_path):
        return {
            "status": "failed",
            "request_id": req.request_id,
            "task_id": req.task_id,
            "error_message": "frame_ref not accessible"
        }

    def read_frame_bytes() -> bytes:
        with open(frame_path, "rb") as f:
            return f.read()

    try:
        frame_bytes = await asyncio.to_thread(read_frame_bytes)
    except Exception as read_error:
        return {
            "status": "failed",
            "request_id": req.request_id,
            "task_id": req.task_id,
            "error_message": f"read frame failed: {read_error}"
        }

    img_b64 = base64.b64encode(frame_bytes).decode()
    queued_at = datetime.now(timezone.utc).isoformat()
    await task_queue.put(AsyncTask(
        task_type="video_frame_detect",
        request_id=req.request_id,
        payload={
            "task_id": req.task_id,
            "bridge_id": req.bridge_id,
            "frame_no": req.frame_no,
            "timestamp_ms": req.timestamp_ms,
            "frame_ref": req.frame_ref,
            "model_name": req.model_name,
            "conf": req.conf,
            "callback_base_url": req.callback_base_url,
            "img_b64": img_b64
        },
        enter_time=time.time()
    ))

    async with httpx.AsyncClient() as client:
        await post_callback(client, build_callback_url(req.callback_base_url, "frame-queued"), {
            "request_id": req.request_id,
            "task_id": req.task_id,
            "frame_no": req.frame_no,
            "timestamp_ms": req.timestamp_ms,
            "status": "queued",
            "queued_at": queued_at
        })

    return {
        "status": "accepted",
        "request_id": req.request_id,
        "task_id": req.task_id,
        "queued_at": queued_at
    }

# 工具函數
def cv2_to_base64(img):
    _, buffer = cv2.imencode('.jpg', img)
    return base64.b64encode(buffer).decode('utf-8')


# 新版优化：启动诊断
print(f"\n{'='*50}")
print(f"🚀 橋梁算法中台啟動診斷")
print(f"{'='*50}")
print(f"CUDA 可用: {torch.cuda.is_available()}")
if torch.cuda.is_available():
    print(f"當前設備: {torch.cuda.get_device_name(0)}")
    print(f"CUDA 版本: {torch.version.cuda}")
else:
    print(f"當前設備: CPU")
print(f"PyTorch 版本: {torch.__version__}")
print(f"YOLO baseline 設備: {MODELS['baseline'].device}")
print(f"YOLO my_trained 設備: {MODELS['my_trained'].device}")
print(f"SAM3 設備: {DEVICE}")
print(f"{'='*50}\n")



if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=18080)
