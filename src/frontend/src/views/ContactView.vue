<template>
  <div class="page-wrapper">
    <div class="page-header">
      <h1 class="page-title">联系方式</h1>
    </div>
    
    <el-row :gutter="30">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>联系信息</span>
          </template>
          <div class="contact-list">
            <div class="contact-item">
              <el-icon><Location /></el-icon>
              <div class="contact-info">
                <h4>公司地址</h4>
                <p>浙江省杭州市余杭塘路2318号勤园23号楼</p>
              </div>
            </div>
            <div class="contact-item">
              <el-icon><Phone /></el-icon>
              <div class="contact-info">
                <h4>联系电话</h4>
                <p>0571-28860710</p>
              </div>
            </div>
            <div class="contact-item">
              <el-icon><Message /></el-icon>
              <div class="contact-info">
                <h4>邮编</h4>
                <p>311121</p>
              </div>
            </div>
            <div class="contact-item">
              <el-icon><Clock /></el-icon>
              <div class="contact-info">
                <h4>工作时间</h4>
                <p>周一至周五 9:00-18:00</p>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>在线咨询</span>
          </template>
          <el-form :model="form" label-width="80px">
            <el-form-item label="姓名">
              <el-input v-model="form.name" placeholder="请输入您的姓名" />
            </el-form-item>
            <el-form-item label="电话">
              <el-input v-model="form.phone" placeholder="请输入您的电话" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="form.email" placeholder="请输入您的邮箱" />
            </el-form-item>
            <el-form-item label="留言">
              <el-input v-model="form.message" type="textarea" rows="5" placeholder="请输入您的留言" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSubmit">提交留言</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Location, Phone, Message, Clock } from '@element-plus/icons-vue'

const form = reactive({
  name: '',
  phone: '',
  email: '',
  message: ''
})

const handleSubmit = () => {
  ElMessage.success('留言提交成功，我们会尽快与您联系！')
  form.name = ''
  form.phone = ''
  form.email = ''
  form.message = ''
}
</script>

<style scoped>
.page-wrapper {
  padding: 20px;
  max-width: 1000px;
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

.contact-list {
  padding: 10px 0;
}

.contact-item {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  padding: 20px 0;
  border-bottom: 1px solid #eee;
}

.contact-item:last-child {
  border-bottom: none;
}

.contact-item .el-icon {
  font-size: 24px;
  color: #409EFF;
  margin-top: 3px;
}

.contact-info h4 {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 5px;
}

.contact-info p {
  font-size: 14px;
  color: #666;
}
</style>
