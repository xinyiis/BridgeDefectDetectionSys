import request from '../utils/request'

export const login = (data) => request.post('/api/login', data)
export const register = (data) => request.post('/api/register', data)
export const getUserInfo = () => request.get('/api/user/info')
export const logout = () => request.post('/api/logout')