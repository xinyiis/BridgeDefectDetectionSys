<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">视频流识别</h1>
    </div>
    
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>检测配置</span>
          </template>
          <el-form :model="form" label-width="100px">
            <el-form-item label="选择无人机">
              <el-select v-model="form.drone_id" placeholder="请选择无人机" style="width: 100%">
                <el-option
                  v-for="drone in droneList"
                  :key="drone.drone_id"
                  :label="drone.drone_name"
                  :value="drone.drone_id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="关联桥梁">
              <el-select v-model="form.bridge_id" placeholder="请选择桥梁" style="width: 100%">
                <el-option
                  v-for="bridge in bridgeList"
                  :key="bridge.bridge_id"
                  :label="bridge.bridge_name"
                  :value="bridge.bridge_id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="视频流地址">
              <el-input v-model="form.video_stream_url" placeholder="请输入RTSP地址" />
            </el-form-item>
            <el-form-item label="检测模型">
              <el-select v-model="form.model_name" placeholder="请选择模型" style="width: 100%">
                <el-option label="YOLOv5s" value="yolov5s" />
                <el-option label="YOLOv5m" value="yolov5m" />
              </el-select>
            </el-form-item>
            <el-form-item label="像素比例">
              <el-input-number v-model="form.pixel_ratio" :precision="4" :step="0.0001" :min="0.0001" style="width: 100%" />
            </el-form-item>
            <el-form-item>
              <el-button
                v-if="!isDetecting"
                type="primary"
                size="large"
                @click="startDetectionTask"
                :loading="starting"
              >
                <el-icon><VideoPlay /></el-icon>开始检测
              </el-button>
              <el-button
                v-else
                type="danger"
                size="large"
                @click="stopDetectionTask"
                :loading="stopping"
              >
                <el-icon><VideoPause /></el-icon>停止检测
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
        
        <el-card class="mt-4" v-if="taskInfo">
          <template #header>
            <span>检测信息</span>
          </template>
          <div class="task-info">
            <p>任务ID: {{ taskInfo.task_id }}</p>
            <p>状态: <el-tag :type="getStatusType(taskInfo.status)">{{ taskInfo.status }}</el-tag></p>
            <p v-if="taskInfo.total_defect !== undefined">检测病害: {{ taskInfo.total_defect }} 个</p>
            <p v-if="taskInfo.cost_time !== undefined">耗时: {{ taskInfo.cost_time.toFixed(2) }} 秒</p>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="video-header">
              <span>实时视频</span>
              <div class="status-indicator" :class="{ active: isDetecting }">
                <span class="dot"></span>
                {{ isDetecting ? '检测中' : '未开始' }}
              </div>
            </div>
          </template>
          <div class="video-container">
            <div v-if="!isDetecting" class="video-placeholder">
              <el-icon><VideoCamera /></el-icon>
              <span>视频流将在此处显示</span>
            </div>
            <div v-else class="video-stream">
              <div class="stream-placeholder">
                <el-icon><Loading /></el-icon>
                <span>正在接收视频流...</span>
              </div>
            </div>
          </div>
        </el-card>
        
        <el-card class="mt-4" v-if="detectedDefects.length > 0">
          <template #header>
            <span>实时检测结果</span>
          </template>
          <div class="defect-list">
            <div v-for="(defect, index) in detectedDefects" :key="index" class="defect-item">
              <el-tag :type="getDefectLevelType(defect.level)">{{ defect.defect_type }}</el-tag>
              <span class="defect-position">{{ defect.defect_position }}</span>
              <span class="defect-time">{{ formatTime(defect.detect_time) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay, VideoPause, VideoCamera, Loading } from '@element-plus/icons-vue'
import { createVideoTask, startVideoDetection, stopVideoDetection } from '@/api/detection'
import { getBridgeList } from '@/api/bridge'
import { getDroneList } from '@/api/drone'
import type { Bridge, Drone } from '@/types'
import dayjs from 'dayjs'

const bridgeList = ref<Bridge[]>([])
const droneList = ref<Drone[]>([])
const isDetecting = ref(false)
const starting = ref(false)
const stopping = ref(false)
const taskInfo = ref<any>(null)
const detectedDefects = ref<any[]>([])

const form = reactive({
  drone_id: '',
  bridge_id: '',
  video_stream_url: 'rtsp://192.168.1.100:554/stream1',
  model_name: 'yolov5s',
  pixel_ratio: 0.1
})

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    '等待': 'info',
    '运行中': 'success',
    '完成': 'success',
    '停止': 'warning'
  }
  return map[status] || 'info'
}

const getDefectLevelType = (level: string) => {
  const map: Record<string, string> = {
    '轻微': 'info',
    '一般': 'warning',
    '严重': 'danger'
  }
  return map[level] || 'info'
}

const formatTime = (time: string) => {
  return dayjs(time).format('HH:mm:ss')
}

const startDetectionTask = async () => {
  if (!form.drone_id || !form.bridge_id) {
    ElMessage.warning('请选择无人机和桥梁')
    return
  }
  
  starting.value = true
  try {
    const task = await createVideoTask(form)
    taskInfo.value = task
    
    const startResult = await startVideoDetection(task.task_id)
    taskInfo.value = { ...taskInfo.value, ...startResult }
    isDetecting.value = true
    
    ElMessage.success('开始视频流检测')
    
    // 模拟接收检测结果
    startSimulatingResults()
  } catch (error: any) {
    ElMessage.error(error.message || '启动失败')
  } finally {
    starting.value = false
  }
}

const stopDetectionTask = async () => {
  if (!taskInfo.value) return
  
  stopping.value = true
  try {
    const result = await stopVideoDetection(taskInfo.value.task_id)
    taskInfo.value = { ...taskInfo.value, ...result }
    isDetecting.value = false
    ElMessage.success('检测已停止')
  } catch (error: any) {
    ElMessage.error(error.message || '停止失败')
  } finally {
    stopping.value = false
  }
}

const startSimulatingResults = () => {
  // 模拟实时检测结果
  const defectTypes = ['裂缝', '剥落', '露筋', '变形']
  const positions = ['桥墩底部', '桥面', '护栏', '桥台']
  const levels = ['轻微', '一般', '严重']
  
  const interval = setInterval(() => {
    if (!isDetecting.value) {
      clearInterval(interval)
      return
    }
    
    if (Math.random() > 0.7) {
      detectedDefects.value.unshift({
        defect_type: defectTypes[Math.floor(Math.random() * defectTypes.length)],
        defect_position: positions[Math.floor(Math.random() * positions.length)],
        level: levels[Math.floor(Math.random() * levels.length)],
        detect_time: new Date().toISOString()
      })
      
      if (detectedDefects.value.length > 10) {
        detectedDefects.value = detectedDefects.value.slice(0, 10)
      }
    }
  }, 2000)
}

const loadData = async () => {
  try {
    const [bridges, drones] = await Promise.all([
      getBridgeList({ page: 1, page_size: 100 }),
      getDroneList({ page: 1, page_size: 100 })
    ])
    bridgeList.value = bridges.list
    droneList.value = drones.list
  } catch {
    // 忽略错误
  }
}

onMounted(loadData)
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0;
}

.mt-4 {
  margin-top: 16px;
}

.video-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #999;
}

.status-indicator.active {
  color: #67c23a;
}

.status-indicator .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #999;
}

.status-indicator.active .dot {
  background: #67c23a;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.video-container {
  height: 400px;
  background: #1a1a2e;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-placeholder, .stream-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  color: #666;
}

.video-placeholder .el-icon, .stream-placeholder .el-icon {
  font-size: 48px;
}

.task-info p {
  margin: 10px 0;
}

.defect-list {
  max-height: 300px;
  overflow-y: auto;
}

.defect-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 12px;
  border-bottom: 1px solid #eee;
}

.defect-item:last-child {
  border-bottom: none;
}

.defect-position {
  flex: 1;
  color: #666;
}

.defect-time {
  color: #999;
  font-size: 13px;
}
</style>
