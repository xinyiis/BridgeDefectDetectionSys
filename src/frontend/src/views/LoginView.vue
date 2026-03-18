<template>
  <div class="login-container">
    <div class="login-box">
      <h2>{{ isLoginMode ? '系统登录' : '用户注册' }}</h2>

      <div class="form-container">
        <div class="form-item">
          <label>用户名:</label>
          <input v-model="formData.username" type="text" placeholder="3-50位用户名" />
        </div>

        <div class="form-item">
          <label>密码:</label>
          <input v-model="formData.password" type="password" placeholder="至少6位密码" />
        </div>

        <template v-if="!isLoginMode">
          <div class="form-item">
            <label>真实姓名:</label>
            <input v-model="formData.real_name" type="text" placeholder="请输入真实姓名" />
          </div>

          <div class="form-item">
            <label>邮箱:</label>
            <input v-model="formData.email" type="email" placeholder="example@qq.com" />
          </div>

          <div class="form-item">
            <label>手机号 (可选):</label>
            <input v-model="formData.phone" type="text" placeholder="11位手机号" />
          </div>
        </template>

        <div class="actions">
          <button class="submit-btn" @click="handleSubmit">
            {{ isLoginMode ? '立即登录' : '提交注册' }}
          </button>
          <button class="switch-btn" @click="isLoginMode = !isLoginMode">
            {{ isLoginMode ? '没有账号？去注册' : '已有账号？去登录' }}
          </button>
        </div>
      </div>

      <div class="debug-panel">
        <h3>测试结果预览:</h3>
        <pre class="result-box">{{ testResult || '等待操作...' }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { register, login } from '../api/user' // 请确保你的api文件路径正确

const isLoginMode = ref(true)
const testResult = ref('')

// 严格适配后端 dto.RegisterRequest 结构体
const formData = reactive({
  username: '',
  password: '',
  real_name: '', // 后端必填：real_name
  email: '',     // 后端必填：需符合email格式
  phone: ''      // 后端可选：11位数字
})

const handleSubmit = async () => {
  if (!formData.username || !formData.password) {
    testResult.value = '错误: 请填写用户名和密码'
    return
  }

  try {
    testResult.value = '正在请求后端...'
    let res

    if (isLoginMode.value) {
      // 登录逻辑
      res = await login({
        username: formData.username,
        password: formData.password
      })
    } else {
      // 注册逻辑：必须包含 real_name 和 email
      // 后端有 min 长度限制，前端在这直接发给后端由后端校验
      res = await register({
        username: formData.username,
        password: formData.password,
        real_name: formData.real_name,
        email: formData.email,
        phone: formData.phone || undefined // 如果手机号为空，不传该字段
      })
    }

    testResult.value = `[请求成功]: \n${JSON.stringify(res, null, 2)}`
  } catch (error) {
    // 捕获后端的 400 详细错误信息 (err.Error())
    const serverMessage = error.response?.data?.message || error.message
    testResult.value = `[请求失败]: \n状态码: ${error.response?.status}\n原因: ${serverMessage}`
    console.error('Full Error:', error)
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f5f5;
  color: #333;
}
.login-box {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  width: 400px;
}
.form-item {
  margin-bottom: 15px;
  text-align: left;
}
.form-item label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}
.form-item input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}
.actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 20px;
}
.submit-btn {
  background-color: #409eff;
  color: white;
  border: none;
  padding: 10px;
  border-radius: 4px;
  cursor: pointer;
}
.switch-btn {
  background: none;
  border: none;
  color: #666;
  text-decoration: underline;
  cursor: pointer;
}
.debug-panel {
  margin-top: 20px;
  border-top: 1px solid #eee;
  padding-top: 10px;
}
.result-box {
  background: #282c34;
  color: #abb2bf;
  padding: 15px;
  border-radius: 4px;
  font-size: 12px;
  text-align: left;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 200px;
  overflow-y: auto;
}
</style>