<template>
  <div class="page-wrapper">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>返回
        </el-button>
        <h1 class="page-title">{{ bridge?.bridge_name }}</h1>
      </div>
      <el-button type="primary" @click="handleEdit">
        <el-icon><Edit /></el-icon>编辑
      </el-button>
    </div>
    
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>基本信息</span>
              <el-tag :type="getStatusType(bridge?.status)">{{ bridge?.status }}</el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="桥梁编号">{{ bridge?.bridge_code }}</el-descriptions-item>
            <el-descriptions-item label="桥梁类型">{{ bridge?.bridge_type }}</el-descriptions-item>
            <el-descriptions-item label="建造年份">{{ bridge?.build_year }}</el-descriptions-item>
            <el-descriptions-item label="长度">{{ bridge?.length }}m</el-descriptions-item>
            <el-descriptions-item label="宽度">{{ bridge?.width }}m</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ bridge?.create_time }}</el-descriptions-item>
          </el-descriptions>
          <el-descriptions :column="1" border class="mt-4">
            <el-descriptions-item label="详细地址">{{ bridge?.address }}</el-descriptions-item>
            <el-descriptions-item label="备注">{{ bridge?.remark || '无' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
        
        <el-card class="mt-4">
          <template #header>
            <span>检测记录</span>
          </template>
          <el-empty description="暂无检测记录" />
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>位置信息</span>
          </template>
          <div class="location-info">
            <p><el-icon><Location /></el-icon> 经度: {{ bridge?.longitude }}</p>
            <p><el-icon><Location /></el-icon> 纬度: {{ bridge?.latitude }}</p>
          </div>
          <div class="map-placeholder">
            <el-icon><MapLocation /></el-icon>
            <span>地图展示区域</span>
          </div>
        </el-card>
        
        <el-card class="mt-4">
          <template #header>
            <span>统计信息</span>
          </template>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-label">检测次数</span>
              <span class="stat-value">0</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">病害数量</span>
              <span class="stat-value">0</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Edit, Location, MapLocation } from '@element-plus/icons-vue'
import { getBridgeDetail } from '@/api/bridge'
import type { Bridge } from '@/types'

const route = useRoute()
const router = useRouter()

const bridge = ref<Bridge>()

const getStatusType = (status?: string) => {
  const map: Record<string, string> = {
    '正常': 'success',
    '异常': 'danger',
    '维修中': 'warning'
  }
  return map[status || ''] || 'info'
}

const loadData = async () => {
  const id = route.params.id as string
  bridge.value = await getBridgeDetail(id)
}

const handleEdit = () => {
  // 编辑逻辑
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

.info-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mt-4 {
  margin-top: 16px;
}

.location-info {
  margin-bottom: 15px;
}

.location-info p {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 8px 0;
  color: #666;
}

.map-placeholder {
  height: 200px;
  background: #f5f7fa;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
  gap: 10px;
}

.map-placeholder .el-icon {
  font-size: 48px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-label {
  display: block;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.stat-value {
  display: block;
  font-size: 28px;
  font-weight: 600;
  color: #409EFF;
}
</style>
