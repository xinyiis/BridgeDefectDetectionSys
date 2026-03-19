import request from '@/utils/request'
import type { AnalysisReport } from '@/types'

// 根据文档 10.1，定义创建分析的参数
export interface CreateAnalysisParams {
  defect_id: number    // 文档标注为 int
  prompt: string      // 用户输入的 Prompt
}

// 根据文档 10.1 响应格式修正
export interface AnalysisResult {
  id: number           // 文档返回的是 id
  analysis_content: string
  analysis_time: string
  cost_time: number
}

/**
 * 10.1 智能分析
 * POST /api/v1/analysis
 */
export const createAnalysis = (params: CreateAnalysisParams): Promise<AnalysisResult> => {
  // 路径对齐文档：baseURL(/api/v1) + /analysis
  return request.post('/analysis', params)
}

/**
 * 10.2 获取分析报告详情
 * GET /api/v1/analysis/{id}
 */
export const getAnalysisDetail = (analysisId: number | string): Promise<AnalysisReport> => {
  return request.get(`/analysis/${analysisId}`)
}