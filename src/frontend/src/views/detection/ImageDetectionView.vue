<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">图片识别</h1>
    </div>
    
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>上传图片</span>
          </template>
          <el-upload
            class="upload-area"
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            :show-file-list="false"
            accept="image/*"
          >
            <el-icon class="upload-icon"><Upload /></el-icon>
            <div class="upload-text">
              <span>拖拽图片到此处或</span>
              <em>点击上传</em>
            </div>
            <template #tip>
              <div class="upload-tip">
                支持 JPG、PNG、BMP 格式，单张图片不超过 20MB
              </div>
            </template>
          </el-upload>
          
          <div v-if="previewUrl" class="preview-section">
            <img :src="previewUrl" alt="预览" class="preview-image" />
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>检测参数</span>
          </template>
          <el-form :model="form" label-width="120px">
            <el-form-item label="关联桥梁">
              <el-select v-model="form.bridge_id" placeholder="请选择桥梁（可选）" clearable style="width: 100%">
                <el-option
                  v-for="bridge in bridgeList"
                  :key="bridge.bridge_id"
                  :label="bridge.bridge_name"
                  :value="bridge.bridge_id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="检测模型">
              <el-select v-model="form.model_name" placeholder="请选择模型" style="width: 100%">
                <el-option label="YOLOv5s" value="yolov5s" />
                <el-option label="YOLOv5m" value="yolov5m" />
                <el-option label="YOLOv5l" value="yolov5l" />
              </el-select>
            </el-form-item>
            <el-form-item label="像素比例">
              <el-input-number v-model="form.pixel_ratio" :precision="4" :step="0.0001" :min="0.0001" style="width: 100%" />
              <div class="form-tip">像素与实际尺寸的比例系数</div>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" :loading="uploading" :disabled="!selectedFile" @click="handleUpload">
                <el-icon><VideoPlay /></el-icon>开始检测
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
        
        <el-card class="mt-4" v-if="taskStatus">
          <template #header>
            <span>检测状态</span>
          </template>
          <div class="status-content">
            <el-steps :active="activeStep" finish-status="success">
              <el-step title="上传图片" />
              <el-step title="开始检测" />
              <el-step title="检测完成" />
            </el-steps>
            <div class="status-info" v-if="taskStatus">
              <p>任务ID: {{ taskId }}</p>
              <p>当前状态: <el-tag>{{ taskStatus }}</el-tag></p>
              <el-button v-if="taskStatus === '完成'" type="primary" @click="viewResult">查看结果</el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { Upload, VideoPlay } from '@element-plus/icons-vue'
import { uploadImage, startDetection } from '@/api/detection'
import { getBridgeList } from '@/api/bridge'
import type { Bridge } from '@/types'

const router = useRouter()

const selectedFile = ref<File | null>(null)
const previewUrl = ref('')
const uploading = ref(false)
const bridgeList = ref<Bridge[]>([])
const taskId = ref('')
const taskStatus = ref('')
const activeStep = ref(0)

const form = reactive({
  bridge_id: '',
  model_name: 'yolov5s',
  pixel_ratio: 0.1
})

const handleFileChange = (uploadFile: UploadFile) => {
  if (uploadFile.raw) {
    selectedFile.value = uploadFile.raw
    previewUrl.value = URL.createObjectURL(uploadFile.raw)
  }
}

const handleUpload = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择图片')
    return
  }
  
  uploading.value = true
  activeStep.value = 0
  
  try {
    const result = await uploadImage({
      image: selectedFile.value,
      bridge_id: form.bridge_id,
      model_name: form.model_name,
      pixel_ratio: form.pixel_ratio
    })
    
    taskId.value = result.task_id
    taskStatus.value = result.status
    activeStep.value = 1
    
    ElMessage.success('上传成功，开始检测')
    
    // 开始检测
    const startResult = await startDetection(result.task_id)
    taskStatus.value = startResult.status
    activeStep.value = 2
    
    ElMessage.success('检测完成')
    taskStatus.value = '完成'
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

const viewResult = () => {
  router.push(`/detection/result/${taskId.value}`)
}

const loadBridges = async () => {
  try {
    const data = await getBridgeList({ page: 1, page_size: 100 })
    bridgeList.value = data.list
  } catch {
    // 忽略错误
  }
}

onMounted(loadBridges)
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

.upload-area {
  width: 100%;
}

.upload-area :deep(.el-upload-dragger) {
  width: 100%;
  height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-icon {
  font-size: 48px;
  color: #409EFF;
  margin-bottom: 15px;
}

.upload-text {
  font-size: 14px;
  color: #666;
}

.upload-text em {
  color: #409EFF;
  font-style: normal;
  margin-left: 5px;
}

.upload-tip {
  font-size: 12px;
  color: #999;
  margin-top: 10px;
  text-align: center;
}

.preview-section {
  margin-top: 20px;
  text-align: center;
}

.preview-image {
  max-width: 100%;
  max-height: 300px;
  border-radius: 8px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.mt-4 {
  margin-top: 16px;
}

.status-content {
  padding: 20px 0;
}

.status-info {
  margin-top: 30px;
  text-align: center;
}

.status-info p {
  margin: 10px 0;
}
</style>
