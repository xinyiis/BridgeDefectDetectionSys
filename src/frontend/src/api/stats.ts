import request from '@/utils/request'
import type { StatsData, DefectTypeStat, TrendData, SystemStats } from '@/types'

export const getUserStats = (): Promise<StatsData> => {
  return request.get('/stats/user')
}

export const getDefectStats = (days?: number): Promise<{ defect_types: DefectTypeStat[] }> => {
  return request.get('/stats/defect', { params: { days } })
}

export const getTrendStats = (days?: number): Promise<TrendData> => {
  return request.get('/stats/trend', { params: { days } })
}

export const getSystemStats = (): Promise<SystemStats> => {
  return request.get('/stats/system')
}
