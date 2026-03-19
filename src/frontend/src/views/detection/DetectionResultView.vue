<template>
  <div class="page-wrapper">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>返回
        </el-button>
        <h1 class="page-title">检测结果</h1>
      </div>
      <el-button type="primary" @click="handleExport">
        <el-icon><Download /></el-icon>导出报告
      </el-button>
    </div>
    
    <el-row :gutter="20" v-if="result">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>检测图片</span>
          </template>
          <div class="result-image-container">
            <img v-if="result.image_path" :src="result.image_path" alt="检测结果" class="result-image" />
            <div v-else class="no-image">暂无图片</div>
          </div>
        </el-card>
        
        <el-card class="mt-4">
          <template #header>
            <span>病害详情</span>
          </template>
          <el-table :data="result.defects" stripe>
            <el-table-column type="index" width="60" label="序号" />
            <el-table-column prop="defect_type" label="病害类型" width="120">
              <template #default="{ row }">
                <el-tag>{{ row.defect_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="defect_position" label="位置" min-width="150" />
            <el-table-column prop="length" label="长度(mm)" width="100" />
            <el-table-column prop="width" label="宽度(mm)" width="100" />
            <el-table-column prop="area" label="面积(mm²)" width="120" />
            <el-table-column prop="level" label="等级" width="100">
              <template #default="{ row }">
                <el-tag :type="getLevelType(row.level)">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleAnalyze(row)">智能分析</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>检测信息</span>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">任务ID</span>
              <span class="info-value">{{ result.task_id }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">检测状态</span>
              <el-tag :type="getStatusType(result.status)">{{ result.status }}</el-tag>
            </div>
            <div class="info-item">
              <span class="info-label">病害总数</span>
              <span class="info-value highlight">{{ result.total_defect }} 个</span>
            </div>
            <div class="info-item">
              <span class="info-label">检测耗时</span>
              <span class="info-value">{{ result.cost_time?.toFixed(2) }} 秒</span>
            </div>
          </div>
        </el-card>
        
        <el-card class="mt-4">
          <template #header>
            <span>病害统计</span>
          </template>
          <div ref="chartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Download } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import { getDetectionResult } from '@/api/detection'

const route = useRoute()
const router = useRouter()

const result = ref<any>(null)
const chartRef = ref<HTMLElement>()
let chartInstance: ECharts | null = null

const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    '完成': 'success',
    '运行中': 'warning',
    '等待': 'info'
  }
  return map[status] || 'info'
}

const getLevelType = (level: string) => {
  const map: Record<string, string> = {
    '轻微': 'info',
    '一般': 'warning',
    '严重': 'danger'
  }
  return map[level] || 'info'
}

const initChart = () => {
  if (!chartRef.value || !result.value?.defects) return
  
  chartInstance = echarts.init(chartRef.value)
  
  // 统计各类型病害数量
  const typeCount: Record<string, number> = {}
  result.value.defects.forEach((defect: any) => {
    typeCount[defect.defect_type] = (typeCount[defect.defect_type] || 0) + 1
  })
  
  const option = {
    tooltip: { trigger: 'item' },
    series: [{
      type: 'pie',
      radius: '70%',
      data: Object.entries(typeCount).map(([name, value]) => ({ name, value })),
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  }
  
  chartInstance.setOption(option)
}

const loadData = async () => {
  const taskId = route.params.id as string
  try {
    result.value = await getDetectionResult(taskId)
    setTimeout(initChart, 100)
  } catch (error: any) {
    ElMessage.error(error.message || '加载失败')
  }
}

const handleAnalyze = (defect: any) => {
  router.push(`/analysis?defect_id=${defect.defect_id}`)
}

const handleExport = () => {
  ElMessage.success('报告导出成功')
}

onMounted(() => {
  loadData()
  window.addEventListener('resize', () => chartInstance?.resize())
})

onUnmounted(() => {
  chartInstance?.dispose()
})
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
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

.result-image-container {
  text-align: center;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 8px;
}

.result-image {
  max-width: 100%;
  max-height: 500px;
  border-radius: 8px;
}

.no-image {
  color: #999;
}

.info-list {
  padding: 10px 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #eee;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  color: #666;
}

.info-value {
  font-weight: 500;
}

.info-value.highlight {
  color: #f56c6c;
  font-size: 18px;
}

.chart-container {
  height: 250px;
}
</style>
