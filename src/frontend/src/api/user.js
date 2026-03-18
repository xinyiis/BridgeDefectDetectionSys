import request from '../utils/request'

// 注意：这里必须和后端的路由组前缀 /api/v1 完全一致
export const register = (data) => request.post('/api/v1/auth/register', data)
export const login = (data) => request.post('/api/v1/auth/login', data)
export const logout = () => request.post('/api/v1/auth/logout')
export const getUserInfo = () => request.get('/api/v1/user/profile') // 注意后端这里叫 profile