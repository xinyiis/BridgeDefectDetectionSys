<template>
  <div class="home-page">
    <!-- Hero Section -->
    <section class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">桥梁表观病害智能检测系统</h1>
        <p class="hero-desc">
          基于低空无人机视觉技术和深度学习算法，实现桥梁表观病害的自动化、智能化检测与分析
        </p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="$router.push('/detection/image')">
            <el-icon><Upload /></el-icon>开始检测
          </el-button>
          <el-button size="large" @click="$router.push('/bridges')">
            <el-icon><OfficeBuilding /></el-icon>管理桥梁
          </el-button>
        </div>
      </div>
      <div class="hero-stats" v-if="userStore.isLoggedIn">
        <div class="stat-card">
          <el-icon><OfficeBuilding /></el-icon>
          <div class="stat-info">
            <span class="stat-value">{{ stats.bridge_count }}</span>
            <span class="stat-label">桥梁数量</span>
          </div>
        </div>
        <div class="stat-card">
          <el-icon><VideoCamera /></el-icon>
          <div class="stat-info">
            <span class="stat-value">{{ stats.drone_count }}</span>
            <span class="stat-label">无人机</span>
          </div>
        </div>
        <div class="stat-card">
          <el-icon><DocumentChecked /></el-icon>
          <div class="stat-info">
            <span class="stat-value">{{ stats.task_count }}</span>
            <span class="stat-label">检测任务</span>
          </div>
        </div>
        <div class="stat-card">
          <el-icon><Warning /></el-icon>
          <div class="stat-info">
            <span class="stat-value">{{ stats.defect_count }}</span>
            <span class="stat-label">病害记录</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="features-section">
      <h2 class="section-title">核心功能</h2>
      <div class="features-grid">
        <div class="feature-card" @click="$router.push('/detection/image')">
          <div class="feature-icon">
            <el-icon><Picture /></el-icon>
          </div>
          <h3>图片识别</h3>
          <p>上传桥梁图片，AI自动识别裂缝、剥落、露筋等表观病害</p>
        </div>
        <div class="feature-card" @click="$router.push('/detection/video')">
          <div class="feature-icon">
            <el-icon><VideoCamera /></el-icon>
          </div>
          <h3>视频流识别</h3>
          <p>实时接入无人机视频流，进行动态病害检测与跟踪</p>
        </div>
        <div class="feature-card" @click="$router.push('/analysis')">
          <div class="feature-icon">
            <el-icon><Cpu /></el-icon>
          </div>
          <h3>智能分析</h3>
          <p>基于大语言模型的病害智能分析，生成专业诊断报告</p>
        </div>
        <div class="feature-card" @click="$router.push('/reports')">
          <div class="feature-icon">
            <el-icon><Document /></el-icon>
          </div>
          <h3>报表生成</h3>
          <p>自动生成检测报告，支持PDF导出和数据可视化</p>
        </div>
      </div>
    </section>

    <!-- Charts Section -->
    <section class="charts-section" v-if="userStore.isLoggedIn">
      <div class="chart-row">
        <div class="chart-card">
          <h3 class="chart-title">病害类型分布</h3>
          <div ref="defectTypeChart" class="chart-container"></div>
        </div>
        <div class="chart-card">
          <h3 class="chart-title">检测趋势统计</h3>
          <div ref="trendChart" class="chart-container"></div>
        </div>
      </div>
    </section>

    <!-- Recent Activity -->
    <section class="activity-section" v-if="userStore.isLoggedIn">
      <div class="section-header">
        <h2 class="section-title">最近活动</h2>
        <el-button link type="primary" @click="$router.push('/detection/image')">查看更多</el-button>
      </div>
      <el-empty v-if="!recentTasks.length" description="暂无检测记录" />
      <div v-else class="activity-list">
        <div v-for="task in recentTasks" :key="task.task_id" class="activity-item">
          <div class="activity-icon" :class="task.task_type === '图片上传识别' ? 'image' : 'video'">
            <el-icon><component :is="task.task_type === '图片上传识别' ? Picture : VideoCamera" /></el-icon>
          </div>
          <div class="activity-content">
            <h4>{{ task.task_type }}</h4>
            <p>模型: {{ task.model_name }} | 状态: {{ task.status }}</p>
          </div>
          <div class="activity-time">{{ formatTime(task.start_time) }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useUserStore } from '@/stores/user'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'
import {
  Upload,
  OfficeBuilding,
  VideoCamera,
  DocumentChecked,
  Warning,
  Picture,
  Cpu,
  Document
} from '@element-plus/icons-vue'
import { getUserStats, getDefectStats, getTrendStats } from '@/api/stats'
import type { StatsData, DefectTypeStat, TrendData } from '@/types'
import dayjs from 'dayjs'

const userStore = useUserStore()

const stats = ref<StatsData>({
  bridge_count: 0,
  drone_count: 0,
  task_count: 0,
  defect_count: 0
})

const defectTypeData = ref<DefectTypeStat[]>([])
const trendData = ref<TrendData>({
  dates: [],
  task_counts: [],
  defect_counts: []
})

const recentTasks = ref<any[]>([])

const defectTypeChart = ref<HTMLElement>()
const trendChart = ref<HTMLElement>()
let defectChartInstance: ECharts | null = null
let trendChartInstance: ECharts | null = null

const formatTime = (time?: string) => {
  if (!time) return '-'
  return dayjs(time).format('MM-DD HH:mm')
}

const initDefectTypeChart = () => {
  if (!defectTypeChart.value) return
  defectChartInstance = echarts.init(defectTypeChart.value)
  const option = {
    tooltip: { trigger: 'item' },
    legend: { bottom: '5%', left: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
      data: defectTypeData.value.map(item => ({ value: item.count, name: item.type }))
    }]
  }
  defectChartInstance.setOption(option)
}

const initTrendChart = () => {
  if (!trendChart.value) return
  trendChartInstance = echarts.init(trendChart.value)
  const option = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['检测任务', '病害数量'], bottom: '5%' },
    grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: trendData.value.dates },
    yAxis: { type: 'value' },
    series: [
      { name: '检测任务', type: 'line', smooth: true, data: trendData.value.task_counts, itemStyle: { color: '#409EFF' } },
      { name: '病害数量', type: 'line', smooth: true, data: trendData.value.defect_counts, itemStyle: { color: '#F56C6C' } }
    ]
  }
  trendChartInstance.setOption(option)
}

const loadData = async () => {
  if (!userStore.isLoggedIn) return
  
  try {
    const [userStats, defectStats, trendStats] = await Promise.all([
      getUserStats(),
      getDefectStats(30),
      getTrendStats(7)
    ])
    
    stats.value = userStats
    defectTypeData.value = defectStats.defect_types
    trendData.value = trendStats
    
    initDefectTypeChart()
    initTrendChart()
  } catch (error) {
    console.error('加载数据失败:', error)
  }
}

onMounted(() => {
  loadData()
  window.addEventListener('resize', () => {
    defectChartInstance?.resize()
    trendChartInstance?.resize()
  })
})

onUnmounted(() => {
  defectChartInstance?.dispose()
  trendChartInstance?.dispose()
})
</script>

<style scoped>
.home-page {
  width: 100%;
  min-width: 100%;
  padding-bottom: 40px;
}

.hero-section {
  background: linear-gradient(135deg, #1a237e 0%, #3949ab 50%, #5c6bc0 100%);
  padding: 60px 40px;
  color: white;
  text-align: center;
  width: 100%;
  min-width: 100%;
  box-sizing: border-box;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto 50px;
}

.hero-title {
  font-size: 42px;
  font-weight: 700;
  margin-bottom: 20px;
  text-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

.hero-desc {
  font-size: 18px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 30px;
}

.hero-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
}

.hero-actions .el-button {
  padding: 12px 30px;
  font-size: 16px;
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 30px;
  flex-wrap: wrap;
  max-width: 1000px;
  margin: 0 auto;
}

.stat-card {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 25px 40px;
  display: flex;
  align-items: center;
  gap: 15px;
  min-width: 180px;
}

.stat-card .el-icon {
  font-size: 36px;
  color: #64b5f6;
}

.stat-info {
  display: flex;
  flex-direction: column;
  text-align: left;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  margin-top: 5px;
}

.features-section {
  padding: 60px 40px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  min-width: 100%;
  box-sizing: border-box;
}

.section-title {
  text-align: center;
  font-size: 32px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 40px;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 25px;
}

.feature-card {
  background: white;
  border-radius: 16px;
  padding: 35px 25px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
}

.feature-icon {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  background: linear-gradient(135deg, #409EFF, #36cfc9);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.feature-icon .el-icon {
  font-size: 32px;
  color: white;
}

.feature-card h3 {
  font-size: 20px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 12px;
}

.feature-card p {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

.charts-section {
  padding: 0 40px 40px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  min-width: 100%;
  box-sizing: border-box;
}

.chart-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 25px;
}

.chart-card {
  background: white;
  border-radius: 16px;
  padding: 25px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.chart-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 20px;
}

.chart-container {
  height: 300px;
}

.activity-section {
  padding: 0 40px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  min-width: 100%;
  box-sizing: border-box;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header .section-title {
  margin-bottom: 0;
  text-align: left;
}

.activity-list {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.activity-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 45px;
  height: 45px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.activity-icon.image {
  background: #e3f2fd;
  color: #2196f3;
}

.activity-icon.video {
  background: #f3e5f5;
  color: #9c27b0;
}

.activity-icon .el-icon {
  font-size: 22px;
}

.activity-content {
  flex: 1;
}

.activity-content h4 {
  font-size: 15px;
  font-weight: 500;
  color: #1a1a2e;
  margin-bottom: 5px;
}

.activity-content p {
  font-size: 13px;
  color: #999;
}

.activity-time {
  font-size: 13px;
  color: #999;
}

@media (max-width: 1024px) {
  .features-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .chart-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 28px;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }
  
  .hero-stats {
    gap: 15px;
  }
  
  .stat-card {
    padding: 20px 25px;
    min-width: 140px;
  }
}
</style>
