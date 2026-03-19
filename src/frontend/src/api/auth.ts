import request from '@/utils/request'
import type { ApiResponse, User } from '@/types'

export interface LoginParams {
  username: string
  password: string
}

export interface RegisterParams {
  username: string
  password: string
  real_name: string
  phone: string
  email: string
}

export interface LoginData {
  token: string
  user: User
}

export const login = (params: LoginParams): Promise<LoginData> => {
  return request.post('/auth/login', params)
}

export const register = (params: RegisterParams): Promise<User> => {
  return request.post('/auth/register', params)
}

export const logout = (): Promise<void> => {
  return request.post('/auth/logout')
}
