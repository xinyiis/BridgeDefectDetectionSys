<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">智能分析</h1>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>病害信息</span>
          </template>
          <div v-if="defect" class="defect-info">
            <img v-if="defect.defect_image" :src="defect.defect_image" alt="病害图片" class="defect-image" />
            <div class="info-grid">
              <div class="info-item">
                <span class="label">病害类型</span>
                <el-tag>{{ defect.defect_type }}</el-tag>
              </div>
              <div class="info-item">
                <span class="label">位置</span>
                <span>{{ defect.defect_position }}</span>
              </div>
              <div class="info-item">
                <span class="label">等级</span>
                <el-tag :type="getLevelType(defect.level)">{{ defect.level }}</el-tag>
              </div>
              <div class="info-item">
                <span class="label">尺寸</span>
                <span>{{ defect.length }}×{{ defect.width }}mm</span>
              </div>
            </div>
          </div>
          <el-empty v-else description="请选择要分析的病害" />
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>AI 智能分析</span>
          </template>
          <div class="analysis-section">
            <el-form :model="form">
              <el-form-item label="分析提示词">
                <el-input
                    v-model="form.prompt"
                    type="textarea"
                    rows="4"
                    placeholder="请输入您想要分析的内容，例如：分析这个裂缝的成因和危害程度"
                />
              </el-form-item>
              <el-form-item>
                <el-button
                    type="primary"
                    :loading="analyzing"
                    :disabled="!defect"
                    @click="handleAnalyze"
                >
                  <el-icon><Cpu /></el-icon>开始分析
                </el-button>
              </el-form-item>
            </el-form>

            <div v-if="analysisResult" class="analysis-result">
              <div class="result-header">
                <el-icon><ChatDotRound /></el-icon>
                <span>分析结果</span>
              </div>
              <div class="result-content">
                {{ analysisResult.analysis_content }}
              </div>
              <div class="result-footer">
                <span>分析耗时: {{ analysisResult.cost_time?.toFixed(2) }}秒</span>
                <span>模型: {{ analysisResult.model_version || 'GPT-4' }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="mt-4" v-if="analysisHistory.length > 0">
      <template #header>
        <span>分析历史</span>
      </template>
      <el-timeline>
        <el-timeline-item
            v-for="item in analysisHistory"
            :key="item.analysis_id"
            :timestamp="item.analysis_time"
            type="primary"
        >
          <h4>{{ item.prompt }}</h4>
          <p class="history-content">{{ item.analysis_content }}</p>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Cpu, ChatDotRound } from '@element-plus/icons-vue'
import { createAnalysis } from '@/api/analysis'

const route = useRoute()

const defect = ref<any>(null)
const analyzing = ref(false)
const analysisResult = ref<any>(null)
const analysisHistory = ref<any[]>([])

const form = reactive({
  prompt: '请分析这个桥梁病害的成因、危害程度以及建议的修复措施。'
})

const getLevelType = (level: string) => {
  const map: Record<string, string> = {
    '轻微': 'info',
    '一般': 'warning',
    '严重': 'danger'
  }
  return map[level] || 'info'
}

const handleAnalyze = async () => {
  if (!defect.value) {
    ElMessage.warning('请先选择病害')
    return
  }

  analyzing.value = true
  try {
    const result = await createAnalysis({
      defect_id: defect.value.defect_id,
      prompt: form.prompt
    })

    analysisResult.value = result
    analysisHistory.value.unshift({
      analysis_id: result.analysis_id,
      prompt: form.prompt,
      analysis_content: result.analysis_content,
      analysis_time: result.analysis_time,
      cost_time: result.cost_time,
      model_version: result.model_version
    })

    ElMessage.success('分析完成')
  } catch (error: any) {
    ElMessage.error(error.message || '分析失败')
  } finally {
    analyzing.value = false
  }
}

onMounted(() => {
  // 从路由参数获取病害信息
  const defectId = route.query.defect_id
  if (defectId) {
    // 模拟加载病害数据
    defect.value = {
      defect_id: defectId,
      defect_type: '裂缝',
      defect_position: '桥墩底部',
      defect_image: '/results/defect1.jpg',
      level: '一般',
      length: 50.5,
      width: 2.0,
      area: 101.0
    }
  }
})
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

.defect-info {
  text-align: center;
}

.defect-image {
  max-width: 100%;
  max-height: 300px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  text-align: left;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.info-item .label {
  font-size: 12px;
  color: #999;
}

.analysis-section {
  min-height: 400px;
}

.analysis-result {
  margin-top: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #409EFF;
  margin-bottom: 15px;
}

.result-content {
  line-height: 1.8;
  color: #333;
  white-space: pre-wrap;
}

.result-footer {
  display: flex;
  gap: 20px;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #999;
}

.history-content {
  color: #666;
  margin-top: 8px;
  line-height: 1.6;
}
</style>
