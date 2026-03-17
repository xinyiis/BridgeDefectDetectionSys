import axios from 'axios'

const request = axios.create({
    baseURL: '', // Vite 代理会拦截 /api，所以这里留空即可
    timeout: 5000,
    withCredentials: true // 【关键】允许跨域携带 Cookie，保证 Session 正常工作
})

// 响应拦截器：统一返回 data 部分，简化页面逻辑
request.interceptors.response.use(
    response => response.data,
    error => {
        console.error('接口请求报错:', error)
        return Promise.reject(error)
    }
)

export default request