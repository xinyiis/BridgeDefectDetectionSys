<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">安全中心</h1>
    </div>
    
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>隐私政策</span>
          </template>
          <div class="policy-content">
            <h3>数据收集与使用</h3>
            <p>我们致力于保护用户的隐私和数据安全。本系统仅收集必要的用户信息，包括：</p>
            <ul>
              <li>基本账户信息（用户名、邮箱、手机号）</li>
              <li>桥梁检测相关数据</li>
              <li>系统使用日志</li>
            </ul>
            
            <h3>数据保护</h3>
            <p>我们采用行业标准的加密技术保护用户数据，包括：</p>
            <ul>
              <li>传输层加密（HTTPS）</li>
              <li>数据库存储加密</li>
              <li>访问权限控制</li>
            </ul>
            
            <h3>用户权利</h3>
            <p>用户享有以下权利：</p>
            <ul>
              <li>查看个人数据</li>
              <li>修改个人信息</li>
              <li>删除账户数据</li>
              <li>导出个人数据</li>
            </ul>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>用户协议</span>
          </template>
          <div class="policy-content">
            <h3>服务条款</h3>
            <p>使用本系统即表示您同意以下条款：</p>
            <ul>
              <li>遵守国家法律法规</li>
              <li>不得恶意使用系统</li>
              <li>保护账户安全</li>
              <li>合理使用系统资源</li>
            </ul>
            
            <h3>责任声明</h3>
            <p>系统提供的检测结果仅供参考，最终决策应结合专业工程师的判断。我们不承担因使用本系统而产生的任何直接或间接损失。</p>
            
            <h3>知识产权</h3>
            <p>本系统的所有内容，包括但不限于文字、图片、音频、视频、软件等，均受知识产权保护。未经授权，不得复制、修改、传播或商业使用。</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="mt-4">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>数据安全</span>
          </template>
          <div class="security-content">
            <h3>数据加密</h3>
            <p>所有敏感数据均采用AES-256加密算法进行加密存储，确保数据在传输和存储过程中的安全性。</p>
            
            <h3>访问控制</h3>
            <p>实施严格的访问控制机制：</p>
            <ul>
              <li>基于角色的权限管理</li>
              <li>多因素身份认证</li>
              <li>操作日志记录</li>
              <li>异常行为监测</li>
            </ul>
            
            <h3>备份与恢复</h3>
            <p>定期对系统数据进行备份，确保在意外情况下能够快速恢复数据，保障业务连续性。</p>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header">
            <span>系统安全</span>
          </template>
          <div class="security-content">
            <h3>安全架构</h3>
            <p>采用多层次安全防护架构：</p>
            <ul>
              <li>网络层防火墙保护</li>
              <li>应用层安全检测</li>
              <li>数据层加密存储</li>
              <li>物理层安全控制</li>
            </ul>
            
            <h3>漏洞管理</h3>
            <p>建立完善的漏洞管理流程：</p>
            <ul>
              <li>定期安全扫描</li>
              <li>漏洞评估与修复</li>
              <li>安全更新推送</li>
              <li>应急响应机制</li>
            </ul>
            
            <h3>合规认证</h3>
            <p>系统已通过多项安全认证，符合国家和行业安全标准要求。</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-card class="mt-4">
      <template #header>
        <span>安全公告</span>
      </template>
      <el-timeline>
        <el-timeline-item timestamp="2026-03-15" type="primary">
          <h4>系统安全更新</h4>
          <p>系统已完成最新安全补丁更新，提升了整体安全防护能力。</p>
        </el-timeline-item>
        <el-timeline-item timestamp="2026-03-10" type="success">
          <h4>隐私政策更新</h4>
          <p>根据最新法规要求，我们更新了隐私政策，更好地保护用户数据。</p>
        </el-timeline-item>
        <el-timeline-item timestamp="2026-03-05" type="warning">
          <h4>安全提醒</h4>
          <p>请定期更新密码，启用双因素认证以增强账户安全性。</p>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
// 安全中心页面逻辑
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 0;
}

.mt-4 {
  margin-top: 16px;
}

.policy-content,
.security-content {
  line-height: 1.8;
}

.policy-content h3,
.security-content h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a2e;
  margin: 20px 0 10px;
}

.policy-content p,
.security-content p {
  color: #666;
  margin-bottom: 15px;
}

.policy-content ul,
.security-content ul {
  margin-left: 20px;
  margin-bottom: 15px;
}

.policy-content li,
.security-content li {
  color: #666;
  margin-bottom: 5px;
}
</style>
