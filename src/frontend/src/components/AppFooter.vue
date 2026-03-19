<template>
  <footer class="footer">
    <div class="footer-container">
      <div class="footer-grid">
        <div class="footer-section">
          <h3 class="section-title">
            <el-icon><DocumentChecked /></el-icon>
            典型案例
          </h3>
          <ul class="link-list">
            <li><router-link to="/cases">桥梁裂缝检测案例</router-link></li>
            <li><router-link to="/cases">混凝土剥落识别</router-link></li>
            <li><router-link to="/cases">钢筋锈蚀分析</router-link></li>
            <li><router-link to="/cases">结构变形监测</router-link></li>
            <li><router-link to="/cases">无人机巡检方案</router-link></li>
          </ul>
        </div>
        
        <div class="footer-section">
          <h3 class="section-title">
            <el-icon><Phone /></el-icon>
            联系方式
          </h3>
          <ul class="contact-list">
            <li>
              <el-icon><Location /></el-icon>
              <span>浙江省杭州市余杭塘路2318号勤园23号楼</span>
            </li>
            <li>
              <el-icon><Phone /></el-icon>
              <span>0571-28860710</span>
            </li>
            <li>
              <el-icon><Message /></el-icon>
              <span>contact@bridge-detection.com</span>
            </li>
            <li>
              <el-icon><Clock /></el-icon>
              <span>周一至周五 9:00-18:00</span>
            </li>
          </ul>
        </div>
        
        <div class="footer-section">
          <h3 class="section-title">
            <el-icon><Lock /></el-icon>
            安全中心
          </h3>
          <ul class="link-list">
            <li><router-link to="/security">隐私政策</router-link></li>
            <li><router-link to="/security">用户协议</router-link></li>
            <li><router-link to="/security">数据安全</router-link></li>
            <li><router-link to="/security">系统安全</router-link></li>
            <li><router-link to="/security">安全公告</router-link></li>
          </ul>
        </div>
        
        <div class="footer-section">
          <h3 class="section-title">
            <el-icon><InfoFilled /></el-icon>
            关于我们
          </h3>
          <p class="about-text">
            专注于桥梁表观病害智能检测技术研发，利用低空无人机视觉和深度学习算法，为桥梁养护提供精准、高效的智能化解决方案。
          </p>
          <div class="social-links">
            <a href="#" class="social-icon"><el-icon><Platform /></el-icon></a>
            <a href="#" class="social-icon"><el-icon><ChatDotRound /></el-icon></a>
            <a href="#" class="social-icon"><el-icon><Share /></el-icon></a>
          </div>
        </div>
      </div>
      
      <div class="footer-bottom">
        <div class="copyright">
          <p>&copy; 2026 桥梁表观病害智能检测系统 版权所有</p>
          <p>京ICP备XXXXXXXX号-1 | 京公网安备XXXXXXXXXXX号</p>
        </div>
        <div class="footer-links">
          <router-link to="/about">关于我们</router-link>
          <span class="divider">|</span>
          <router-link to="/contact">联系我们</router-link>
          <span class="divider">|</span>
          <router-link to="/security">帮助中心</router-link>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import {
  DocumentChecked,
  Phone,
  Location,
  Message,
  Clock,
  InfoFilled,
  Platform,
  ChatDotRound,
  Share,
  Lock
} from '@element-plus/icons-vue'
</script>

<style scoped>
.footer {
  background: linear-gradient(135deg, #1a237e 0%, #0d1642 100%);
  color: rgba(255, 255, 255, 0.8);
  padding: 50px 0 0;
}

.footer-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
}

.footer-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 40px;
  padding-bottom: 40px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.footer-section {
  min-width: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: white;
  margin-bottom: 20px;
}

.section-title .el-icon {
  font-size: 20px;
  color: #64b5f6;
}

.link-list {
  list-style: none;
}

.link-list li {
  margin-bottom: 12px;
}

.link-list a {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-size: 14px;
  transition: all 0.3s;
  display: inline-block;
}

.link-list a:hover {
  color: #64b5f6;
  transform: translateX(4px);
}

.contact-list {
  list-style: none;
}

.contact-list li {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
  font-size: 14px;
}

.contact-list .el-icon {
  font-size: 16px;
  color: #64b5f6;
  margin-top: 2px;
  flex-shrink: 0;
}

.about-text {
  font-size: 14px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 20px;
}

.social-links {
  display: flex;
  gap: 12px;
}

.social-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  text-decoration: none;
  transition: all 0.3s;
}

.social-icon:hover {
  background: #64b5f6;
  transform: translateY(-2px);
}

.footer-bottom {
  padding: 20px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
}

.copyright {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
}

.copyright p {
  margin: 4px 0;
}

.footer-links {
  display: flex;
  align-items: center;
  gap: 15px;
}

.footer-links a {
  color: rgba(255, 255, 255, 0.6);
  text-decoration: none;
  font-size: 13px;
  transition: color 0.3s;
}

.footer-links a:hover {
  color: #64b5f6;
}

.divider {
  color: rgba(255, 255, 255, 0.3);
}

@media (max-width: 1024px) {
  .footer-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .footer-grid {
    grid-template-columns: 1fr;
  }
  
  .footer-bottom {
    flex-direction: column;
    text-align: center;
  }
}
</style>
