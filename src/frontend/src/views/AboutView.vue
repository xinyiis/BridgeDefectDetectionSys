<template>
  <div class="page-wrapper">
    <div class="about-hero">
      <h1 class="hero-title">关于我们</h1>
      <p class="hero-desc">专注于桥梁表观病害智能检测技术研发</p>
    </div>
    
    <div class="about-content">
      <el-card class="intro-card">
        <h2>公司简介</h2>
        <p>
          我们是一家专注于桥梁健康监测与智能检测技术的高科技企业。公司依托先进的低空无人机技术、
          计算机视觉算法和深度学习技术，为桥梁养护管理提供全方位的智能化解决方案。
        </p>
        <p>
          我们的团队由来自清华大学、北京航空航天大学等知名高校的博士、硕士组成，
          在桥梁工程、人工智能、计算机视觉等领域拥有深厚的技术积累。
        </p>
      </el-card>
      
      <el-row :gutter="20" class="feature-row">
        <el-col :span="8">
          <div class="feature-box">
            <el-icon class="feature-icon"><VideoCamera /></el-icon>
            <h3>无人机巡检</h3>
            <p>利用低空无人机进行高效、安全的桥梁巡检，覆盖传统人工难以到达的区域</p>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-box">
            <el-icon class="feature-icon"><Cpu /></el-icon>
            <h3>AI智能识别</h3>
            <p>基于深度学习的病害自动识别算法，准确识别裂缝、剥落、露筋等表观病害</p>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-box">
            <el-icon class="feature-icon"><DataAnalysis /></el-icon>
            <h3>数据分析</h3>
            <p>提供病害发展趋势分析、桥梁健康评估等数据驱动的决策支持</p>
          </div>
        </el-col>
      </el-row>
      
      <el-card class="team-card">
        <template #header>
          <span>核心团队</span>
        </template>
        <el-row :gutter="30">
          <el-col :span="8" v-for="member in team" :key="member.name">
            <div class="team-member">
              <el-avatar :size="80" :src="member.avatar" />
              <h4>{{ member.name }}</h4>
              <p class="position">{{ member.position }}</p>
              <p class="desc">{{ member.description }}</p>
            </div>
          </el-col>
        </el-row>
      </el-card>
      
      <el-card class="contact-card">
        <template #header>
          <span>联系我们</span>
        </template>
        <div class="contact-grid">
          <div class="contact-item">
            <el-icon><Location /></el-icon>
            <span>北京市海淀区中关村大街1号科技大厦18层</span>
          </div>
          <div class="contact-item">
            <el-icon><Phone /></el-icon>
            <span>400-888-8888</span>
          </div>
          <div class="contact-item">
            <el-icon><Message /></el-icon>
            <span>contact@bridge-detection.com</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { VideoCamera, Cpu, DataAnalysis, Location, Phone, Message } from '@element-plus/icons-vue'

const team = [
  {
    name: '张博士',
    position: '首席科学家',
    description: '清华大学土木工程博士，20年桥梁工程经验',
    avatar: ''
  },
  {
    name: '李教授',
    position: '技术总监',
    description: '北京航空航天大学计算机博士，AI算法专家',
    avatar: ''
  },
  {
    name: '王工程师',
    position: '产品总监',
    description: '15年软件产品开发经验，专注工程领域信息化',
    avatar: ''
  }
]
</script>

<style scoped>
.page-wrapper {
  padding-bottom: 40px;
}

.about-hero {
  background: linear-gradient(135deg, #1a237e 0%, #3949ab 100%);
  padding: 80px 20px;
  text-align: center;
  color: white;
}

.hero-title {
  font-size: 42px;
  font-weight: 700;
  margin-bottom: 15px;
}

.hero-desc {
  font-size: 18px;
  color: rgba(255, 255, 255, 0.8);
}

.about-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 20px;
}

.intro-card {
  margin-bottom: 40px;
}

.intro-card h2 {
  font-size: 24px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 20px;
}

.intro-card p {
  font-size: 15px;
  line-height: 1.8;
  color: #666;
  margin-bottom: 15px;
}

.feature-row {
  margin-bottom: 40px;
}

.feature-box {
  text-align: center;
  padding: 40px 30px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.feature-icon {
  font-size: 48px;
  color: #409EFF;
  margin-bottom: 20px;
}

.feature-box h3 {
  font-size: 20px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 15px;
}

.feature-box p {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

.team-card {
  margin-bottom: 40px;
}

.team-member {
  text-align: center;
  padding: 20px;
}

.team-member h4 {
  font-size: 18px;
  font-weight: 600;
  margin: 15px 0 5px;
}

.team-member .position {
  font-size: 14px;
  color: #409EFF;
  margin-bottom: 10px;
}

.team-member .desc {
  font-size: 13px;
  color: #666;
  line-height: 1.6;
}

.contact-card {
  margin-bottom: 40px;
}

.contact-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 30px;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 12px;
}

.contact-item .el-icon {
  font-size: 24px;
  color: #409EFF;
}

.contact-item span {
  font-size: 14px;
  color: #666;
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 28px;
  }
  
  .contact-grid {
    grid-template-columns: 1fr;
  }
}
</style>
