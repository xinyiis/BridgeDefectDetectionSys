import request from '@/utils/request'
import type { Drone } from '@/types'

// 1. 无人机列表查询参数 (对应文档 7.1)
export interface DroneListParams {
  page?: number
  page_size?: number
}

// 2. 列表返回结构 (对应文档 7.1)
export interface DroneListData {
  total: number
  page: number
  page_size: number
  list: Drone[]
}

// 3. 添加/更新参数 (对应文档 7.3 & 7.4)
export interface CreateDroneParams {
  name: string          // 文档 7.3 使用的是 name 而不是 drone_name
  model?: string        // 文档标注为“否”，即可选
  stream_url?: string   // 文档中包含视频流地址
}

/**
 * 7.1 获取无人机列表
 */
export const getDroneList = (params?: DroneListParams): Promise<DroneListData> => {
  return request.get('/drones', { params })
}

/**
 * 7.2 获取无人机详情
 */
export const getDroneDetail = (id: string | number): Promise<Drone> => {
  return request.get(`/drones/${id}`)
}

/**
 * 7.3 添加无人机
 */
export const createDrone = (params: CreateDroneParams): Promise<{ id: number; name: string; model: string }> => {
  return request.post('/drones', params)
}

/**
 * 7.4 更新无人机信息
 */
export const updateDrone = (id: string | number, params: Partial<CreateDroneParams>): Promise<{ id: number; name: string }> => {
  return request.put(`/drones/${id}`, params)
}

/**
 * 7.5 删除无人机
 */
export const deleteDrone = (id: string | number): Promise<void> => {
  return request.delete(`/drones/${id}`)
}