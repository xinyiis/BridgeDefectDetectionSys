<template>
  <div class="login-container">
    <div class="auth-card">
      <h2>{{ isLoginMode ? '系统登录' : '用户注册' }}</h2>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="formData.username" type="text" required placeholder="请输入用户名" />
        </div>

        <div class="form-group">
          <label>密码</label>
          <input v-model="formData.password" type="password" required placeholder="请输入密码" />
        </div>

        <div class="form-group" v-if="!isLoginMode">
          <label>邮箱</label>
          <input v-model="formData.email" type="email" required placeholder="请输入邮箱" />
        </div>

        <button type="submit" class="primary-btn">
          {{ isLoginMode ? '测试登录' : '测试注册' }}
        </button>
      </form>

      <div class="toggle-mode">
        <a href="#" @click.prevent="isLoginMode = !isLoginMode">
          {{ isLoginMode ? '没有账号？切换到注册' : '已有账号？切换到登录' }}
        </a>
      </div>
    </div>

    <div class="debug-panel">
      <h3>接口调试结果：</h3>
      <div class="action-buttons">
        <button @click="testUserInfo">测试获取当前用户信息</button>
        <button @click="testLogout" class="danger">测试退出登录</button>
      </div>
      <pre class="result-box">{{ testResult || '等待操作...' }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { login, register, getUserInfo, logout } from '../api/user'

const isLoginMode = ref(true)
const formData = reactive({
  username: '',
  password: '',
  email: ''
})

const testResult = ref('')

// 处理登录/注册提交
const handleSubmit = async () => {
  try {
    testResult.value = '请求中...'
    let res
    if (isLoginMode.value) {
      res = await login({ username: formData.username, password: formData.password })
    } else {
      res = await register(formData)
    }
    testResult.value = `[${isLoginMode.value ? '登录' : '注册'}响应]: \n` + JSON.stringify(res, null, 2)
  } catch (error) {
    testResult.value = `[请求失败]: \n` + error.message
  }
}

// 测试获取信息
const testUserInfo = async () => {
  try {
    testResult.value = '请求中...'
    const res = await getUserInfo()
    testResult.value = `[获取用户信息响应]: \n` + JSON.stringify(res, null, 2)
  } catch (error) {
    testResult.value = `[获取失败]: \n` + error.message
  }
}

// 测试登出
const testLogout = async () => {
  try {
    testResult.value = '请求中...'
    const res = await logout()
    testResult.value = `[登出响应]: \n` + JSON.stringify(res, null, 2)
  } catch (error) {
    testResult.value = `[登出失败]: \n` + error.message
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 50px;
  font-family: sans-serif;
}
.auth-card {
  width: 350px;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  background: white;
  border: 1px solid #eee;
}
.form-group {
  margin-bottom: 15px;
}
.form-group label {
  display: block;
  margin-bottom: 5px;
  font-size: 14px;
  color: #333;
}
.form-group input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
}
.primary-btn {
  width: 100%;
  padding: 12px;
  background-color: #409eff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  margin-top: 10px;
}
.primary-btn:hover { background-color: #66b1ff; }
.toggle-mode {
  text-align: center;
  margin-top: 15px;
  font-size: 14px;
}
.toggle-mode a {
  color: #409eff;
  text-decoration: none;
}
.debug-panel {
  margin-top: 30px;
  width: 600px;
  background: #f4f4f5;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}
.action-buttons {
  margin-bottom: 15px;
  display: flex;
  gap: 10px;
}
.action-buttons button {
  padding: 8px 15px;
  cursor: pointer;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: white;
}
.action-buttons button.danger {
  color: white;
  background-color: #f56c6c;
  border-color: #f56c6c;
}
.result-box {
  white-space: pre-wrap;
  word-wrap: break-word;
  background: #282c34;
  color: #abb2bf;
  padding: 15px;
  border-radius: 4px;
  min-height: 150px;
  margin: 0;
}
</style>