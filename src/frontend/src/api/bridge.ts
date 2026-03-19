import request from '@/utils/request'
import type { Bridge } from '@/types'

// 1. 列表查询参数 (对应文档 6.1)
export interface BridgeListParams {
  page?: number
  page_size?: number
  status?: string
}

// 2. 列表返回结构 (对应文档 6.1)
export interface BridgeListData {
  total: number
  list: Bridge[]
}

// 3. 添加桥梁参数 (对应文档 6.3)
// 注意：文档中包含 model_3d_file (文件类型)
export interface CreateBridgeParams {
  bridge_name: string
  bridge_code: string
  address: string
  longitude: number
  latitude: number
  bridge_type: string
  build_year: number
  length: number
  width: number
  status?: string
  model_3d_file?: File // 对应文档中的 model_3d_file (file 类型)
  remark?: string
}

/**
 * 6.1 获取桥梁列表
 */
export const getBridgeList = (params?: BridgeListParams): Promise<BridgeListData> => {
  return request.get('/bridges', { params })
}

/**
 * 6.2 获取桥梁详情
 */
export const getBridgeDetail = (bridgeId: string | number): Promise<Bridge> => {
  return request.get(`/bridges/${bridgeId}`)
}

/**
 * 6.3 添加桥梁 (关键：使用 FormData 处理文件上传)
 */
export const createBridge = (params: CreateBridgeParams): Promise<{ id: number; bridge_name: string }> => {
  const formData = new FormData()

  // 将参数对象转为 FormData
  Object.keys(params).forEach(key => {
    const value = (params as any)[key]
    if (value !== undefined && value !== null) {
      formData.append(key, value)
    }
  })

  return request.post('/bridges', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

/**
 * 6.4 更新桥梁信息 (JSON 格式)
 * 注意：文档提示 bridge_code 不可修改
 */
export const updateBridge = (bridgeId: string | number, params: Partial<CreateBridgeParams>): Promise<{ id: number; bridge_name: string }> => {
  return request.put(`/bridges/${bridgeId}`, params)
}

/**
 * 6.5 删除桥梁
 */
export const deleteBridge = (bridgeId: string | number): Promise<void> => {
  return request.delete(`/bridges/${bridgeId}`)
}