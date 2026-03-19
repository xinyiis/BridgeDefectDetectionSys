import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  // 【关键改动 1】开启凭证携带，允许浏览器自动发送和接收 Cookie
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
})

request.interceptors.request.use(
    (config) => {
      // 【关键改动 2】删掉手动塞 Token 的逻辑
      // 因为使用 Cookie 模式，浏览器会自动在 Request Headers 中带上 Cookie 字段
      // 不再需要 Authorization: Bearer xxxx
      return config
    },
    (error) => {
      return Promise.reject(error)
    }
)

request.interceptors.response.use(
    (response: AxiosResponse) => {
      const { code, message, data } = response.data
      // 对齐文档 2.0：业务成功码为 200
      if (code === 200) {
        return data
      } else {
        ElMessage.error(message || '请求失败')
        return Promise.reject(new Error(message))
      }
    },
    (error) => {
      const { response } = error
      if (response) {
        switch (response.status) {
          case 401:
            // 401 处理逻辑保持你刚才修正的版本，非常稳健
            if (!error.config.url.includes('/auth/login')) {
              ElMessage.error('登录已过期，请重新登录')
              const userStore = useUserStore()
              userStore.logout()
              if (window.location.pathname !== '/login') {
                window.location.href = '/login'
              }
            } else {
              // 登录接口报 401 提示具体的业务错误（如：用户不存在）
              ElMessage.error(response.data?.message || '用户名或密码错误')
            }
            break
          case 403:
            ElMessage.error('权限不足')
            break
          case 404:
            ElMessage.error('请求的资源不存在')
            break
          case 500:
            ElMessage.error('服务器内部错误')
            break
          default:
            ElMessage.error(response.data?.message || '网络错误')
        }
      } else {
        ElMessage.error('网络连接失败')
      }
      return Promise.reject(error)
    }
)

export default request