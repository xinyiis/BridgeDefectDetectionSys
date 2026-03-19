import request from '@/utils/request'
import type { DetectionTask, Defect } from '@/types'

export interface UploadImageParams {
  image: File
  bridge_id?: string
  model_name: string
  pixel_ratio: number
}

export interface UploadResult {
  task_id: string
  image_path: string
  status: string
}

export interface DetectionResult {
  task_id: string
  status: string
  total_defect: number
  cost_time: number
  defects: Defect[]
}

export interface CreateVideoTaskParams {
  drone_id: string
  bridge_id: string
  model_name: string
  pixel_ratio: number
  video_stream_url: string
}

export interface VideoTaskResult {
  task_id: string
  status: string
  websocket_url?: string
}

export const uploadImage = (params: UploadImageParams): Promise<UploadResult> => {
  const formData = new FormData()
  formData.append('image', params.image)
  if (params.bridge_id) formData.append('bridge_id', params.bridge_id)
  formData.append('model_name', params.model_name)
  formData.append('pixel_ratio', params.pixel_ratio.toString())
  
  return request.post('/detection/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export const startDetection = (taskId: string): Promise<{ task_id: string; status: string }> => {
  return request.post(`/detection/${taskId}/start`)
}

export const getDetectionResult = (taskId: string): Promise<DetectionResult> => {
  return request.get(`/detection/${taskId}/result`)
}

export const createVideoTask = (params: CreateVideoTaskParams): Promise<{ task_id: string; status: string }> => {
  return request.post('/detection/video', params)
}

export const startVideoDetection = (taskId: string): Promise<VideoTaskResult> => {
  return request.post(`/detection/video/${taskId}/start`)
}

export const stopVideoDetection = (taskId: string): Promise<{ task_id: string; status: string; total_defect: number; cost_time: number }> => {
  return request.post(`/detection/video/${taskId}/stop`)
}
