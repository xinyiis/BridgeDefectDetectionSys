import request from '@/utils/request'
import type { Report } from '@/types'

// 1. 创建报表参数 (对应文档 11.1)
export interface CreateReportParams {
  report_name: string
  report_type: 'bridge_inspection' | 'defect_analysis' | 'health_comparison'
  bridge_id?: number         // 桥梁检测和缺陷分析必填
  bridge_ids?: number[]      // 健康对比必填，至少2个
  start_time: string         // 格式：2006-01-02
  end_time: string           // 格式：2006-01-02
}

// 2. 报表列表查询参数 (对应文档 11.2)
export interface ReportListParams {
  page?: number
  page_size?: number
  report_type?: string
  bridge_id?: number
}

// 3. 列表返回结构 (对应文档 11.2)
export interface ReportListData {
  total: number
  page: number
  page_size: number
  list: Report[]
}

/**
 * 11.1 生成检测报告 (异步任务)
 * POST /api/v1/reports
 */
export const createReport = (params: CreateReportParams): Promise<Report> => {
  // 文档显示返回的是 Report 对象，初始 status 为 generating
  return request.post('/reports', params)
}

/**
 * 11.2 获取报表列表
 * GET /api/v1/reports
 */
export const getReportList = (params?: ReportListParams): Promise<ReportListData> => {
  return request.get('/reports', { params })
}

/**
 * 11.3 获取报表详情 (用于轮询 status)
 * GET /api/v1/reports/:id
 */
export const getReportDetail = (reportId: string | number): Promise<Report> => {
  return request.get(`/reports/${reportId}`)
}

/**
 * 11.4 下载报表文件
 * GET /api/v1/reports/:id/download
 */
export const downloadReport = (reportId: string | number) => {
  // 注意：下载文件通常需要设置 responseType 为 blob
  return request.get(`/reports/${reportId}/download`, {
    responseType: 'blob'
  })
}

/**
 * 11.5 删除报表
 * DELETE /api/v1/reports/:id
 */
export const deleteReport = (reportId: string | number): Promise<void> => {
  return request.delete(`/reports/${reportId}`)
}