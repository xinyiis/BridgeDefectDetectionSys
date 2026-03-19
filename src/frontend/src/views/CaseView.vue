<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">典型案例</h1>
    </div>
    
    <div class="cases-grid">
      <div v-for="item in cases" :key="item.id" class="case-card">
        <div class="case-image">
          <img :src="item.image" :alt="item.title" />
        </div>
        <div class="case-content">
          <h3 class="case-title">{{ item.title }}</h3>
          <p class="case-desc">{{ item.description }}</p>
          <div class="case-meta">
            <span><el-icon><Location /></el-icon>{{ item.location }}</span>
            <span><el-icon><Clock /></el-icon>{{ item.date }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Location, Clock } from '@element-plus/icons-vue'

const cases = [
  {
    id: 1,
    title: '跨江大桥裂缝检测',
    description: '使用无人机对跨江大桥进行全面巡检，成功识别出多处细微裂缝，为桥梁养护提供了重要依据。',
    location: '北京市朝阳区',
    date: '2026-01-15',
    image: '/images/case1.jpg'
  },
  {
    id: 2,
    title: '高速公路立交桥检测',
    description: '通过AI识别技术自动标注混凝土剥落位置，检测效率提升80%。',
    location: '上海市浦东新区',
    date: '2026-01-10',
    image: '/images/case2.jpg'
  },
  {
    id: 3,
    title: '山区拱桥结构监测',
    description: '利用多光谱成像技术检测钢筋锈蚀情况，为桥梁耐久性评估提供数据支持。',
    location: '四川省成都市',
    date: '2026-01-05',
    image: '/images/case3.jpg'
  },
  {
    id: 4,
    title: '城市人行天桥巡检',
    description: '结合3D建模技术，实时监测桥梁变形情况，保障市民出行安全。',
    location: '广州市天河区',
    date: '2025-12-28',
    image: '/images/case4.jpg'
  }
]
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 30px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0;
}

.cases-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 25px;
}

.case-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
}

.case-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
}

.case-image {
  height: 200px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.case-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.case-content {
  padding: 20px;
}

.case-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 10px;
}

.case-desc {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
  margin-bottom: 15px;
}

.case-meta {
  display: flex;
  gap: 20px;
  font-size: 13px;
  color: #999;
}

.case-meta span {
  display: flex;
  align-items: center;
  gap: 5px;
}

@media (max-width: 768px) {
  .cases-grid {
    grid-template-columns: 1fr;
  }
}
</style>
