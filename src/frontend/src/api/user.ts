import request from '@/utils/request'
import type { User } from '@/types'

export interface UserListParams {
  page?: number
  page_size?: number
}

export interface UserListData {
  total: number
  list: User[]
}

export interface UpdateUserParams {
  real_name?: string
  phone?: string
  email?: string
  password?: string
}

// 1. 获取用户列表 (后端路径是 /api/v1/admin/users)
export const getUserList = (params?: UserListParams): Promise<UserListData> => {
  return request.get('/admin/users', { params })
}

// 2. 获取用户详情 (后端路径是 /api/v1/admin/users/:id)
export const getUserDetail = (userId: string): Promise<User> => {
  return request.get(`/admin/users/${userId}`)
}

// 3. 更新个人信息 (后端有 /api/v1/user/profile)
// 如果是管理员修改别人，通常是 /api/v1/admin/users/:id
export const updateUser = (userId: string, params: UpdateUserParams): Promise<User> => {
  return request.put(`/admin/users/${userId}`, params)
}

// 4. 提升管理员 (后端路径是 /api/v1/admin/users/promote)
export const promoteUser = (userId: string): Promise<{ user_id: string; role: number }> => {
  // 注意：我看你后端日志 promote 是 POST /api/v1/admin/users/promote
  // 可能需要传 body 而不是拼在 URL 里，请根据后端逻辑确认
  return request.post('/admin/users/promote', { user_id: userId })
}
