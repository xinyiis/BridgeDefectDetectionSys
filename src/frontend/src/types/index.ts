export interface User {
  user_id: string
  username: string
  real_name: string
  phone?: string
  email?: string
  role: number
  status: number
  create_time?: string
  last_login_time?: string
}

export interface Bridge {
  bridge_id: string
  bridge_name: string
  bridge_code: string
  address: string
  longitude: number
  latitude: number
  bridge_type: string
  build_year: number
  length: number
  width: number
  status: string
  model_3d_path?: string
  remark?: string
  user_id?: string
  create_time?: string
  update_time?: string
}

export interface Drone {
  drone_id: string
  drone_code: string
  drone_name: string
  model: string
  camera_param: string
  status: string
  remark?: string
  user_id?: string
  create_time?: string
}

export interface DetectionTask {
  task_id: string
  bridge_id?: string
  task_type: string
  user_id?: string
  drone_id?: string
  model_name: string
  image_path?: string
  video_stream_url?: string
  start_time?: string
  end_time?: string
  status: string
  pixel_ratio: number
  total_defect?: number
  cost_time?: number
}

export interface Defect {
  defect_id: string
  task_id: string
  defect_type: string
  defect_position: string
  defect_image?: string
  defect_video?: string
  length?: number
  width?: number
  area?: number
  pixel_ratio?: number
  level: string
  detect_time: string
  cost_time?: number
  user_id?: string
  remark?: string
}

export interface AnalysisReport {
  analysis_id: string
  defect_id: string
  image_path?: string
  prompt: string
  analysis_content: string
  analysis_time: string
  cost_time?: number
  model_version?: string
}

export interface Report {
  report_id: string
  report_name: string
  report_type: string
  user_id?: string
  create_time: string
  file_path?: string
  bridge_id?: string
  start_time?: string
  end_time?: string
  status: string
}

export interface StatsData {
  bridge_count: number
  drone_count: number
  task_count: number
  defect_count: number
}

export interface DefectTypeStat {
  type: string
  count: number
}

export interface TrendData {
  dates: string[]
  task_counts: number[]
  defect_counts: number[]
}

export interface SystemStats {
  user_count: number
  admin_count: number
  normal_user_count: number
  bridge_count: number
  drone_count: number
  defect_count: number
  today_task_count: number
  total_task_count: number
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}
