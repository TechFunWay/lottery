import axios from 'axios'
import type { PurchaseRecord, DrawResult, OverviewStats, PrizeDistribution, TrendData, WinningRecord, User, AuthResponse, SystemConfig, PublicConfigs } from '../types'

// API 基础路径配置
// 开发环境: http://localhost:8902/api
// 生产环境: /api (相对路径，通过 Nginx 代理或 Go 后端提供)
const getBaseURL = () => {
  // 优先使用环境变量
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }

  // 生产环境使用相对路径
  if (import.meta.env.PROD) {
    return '/api'
  }

  // 开发环境使用完整地址
  return 'http://localhost:8902/api'
}

const api = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
})

// 请求拦截器 - 添加JWT令牌
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // 清除token并跳转到登录页
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)


// ===== 购买记录 API =====
export interface CreatePurchasePayload {
  lottery_type: string
  issue_number: string
  purchase_date: string
  numbers: string
  bet_type: string
  amount: number
  remark?: string
}

export const purchaseApi = {
  create: (data: CreatePurchasePayload) => api.post('/purchases', data),
  list: (params?: { lottery_type?: string; status?: string; page?: number; size?: number }) =>
    api.get('/purchases', { params }),
  update: (id: number, data: Partial<CreatePurchasePayload>) => api.put(`/purchases/${id}`, data),
  delete: (id: number) => api.delete(`/purchases/${id}`),
}

// ===== 开奖结果 API =====
export interface CreateDrawPayload {
  lottery_type: string
  issue_number: string
  draw_date: string
  numbers: string
}

export const drawApi = {
  create: (data: CreateDrawPayload) => api.post('/draws', data),
  list: (params?: { lottery_type?: string; page?: number; size?: number }) =>
    api.get('/draws', { params }),
  update: (id: number, data: Partial<CreateDrawPayload>) => api.put(`/draws/${id}`, data),
  delete: (id: number) => api.delete(`/draws/${id}`),
  fetchLatest: (lottery_type: string, issue?: string) =>
    api.get('/draws/fetch', { params: { lottery_type, issue } }),
  fetchBatch: (params: { lottery_type: string; start_date?: string; end_date?: string; count?: number }) =>
    api.post('/draws/fetch-batch', params),
}

// ===== 中奖记录 API =====
export const winningApi = {
  list: (params?: { lottery_type?: string; page?: number; size?: number }) =>
    api.get('/winnings', { params }),
  recheck: () => api.post('/winnings/recheck'),
}

// ===== 统计分析 API =====
export const statsApi = {
  overview: (lottery_type?: string): Promise<{ data: OverviewStats }> =>
    api.get('/statistics/overview', { params: { lottery_type } }),
  prizes: (lottery_type?: string): Promise<{ data: PrizeDistribution[] }> =>
    api.get('/statistics/prizes', { params: { lottery_type } }),
  numbers: (lottery_type?: string) =>
    api.get('/statistics/numbers', { params: { lottery_type } }),
  trends: (lottery_type?: string, months?: number): Promise<{ data: TrendData[] }> =>
    api.get('/statistics/trends', { params: { lottery_type, months } }),
  recentWinnings: (limit?: number): Promise<{ data: WinningRecord[] }> =>
    api.get('/statistics/recent-winnings', { params: { limit } }),
}

// ===== 认证 API =====
export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email?: string
}

export const authApi = {
  login: (data: LoginRequest): Promise<AuthResponse> =>
    api.post('/auth/login', data),
  register: (data: RegisterRequest): Promise<AuthResponse> =>
    api.post('/auth/register', data),
  getCurrentUser: (): Promise<{ data: User }> =>
    api.get('/auth/me'),
  checkAdmin: (): Promise<{ exists: boolean }> =>
    api.get('/auth/check-admin'),
  changePassword: (data: { old_password: string; new_password: string }): Promise<{ message: string }> =>
    api.put('/auth/password', data),
}

// ===== 用户管理 API (仅管理员) =====
export const userApi = {
  getAll: (): Promise<{ data: User[] }> =>
    api.get('/users'),
  update: (id: number, data: { status?: string; role?: string; password?: string }) =>
    api.put(`/users/${id}`, data),
  delete: (id: number) =>
    api.delete(`/users/${id}`),
}

// ===== 系统配置 API =====
export const configApi = {
  // 公开配置，无需登录
  getPublic: (): Promise<{ data: PublicConfigs }> =>
    api.get('/configs/public'),
  // 完整配置列表（管理员）
  getAll: (): Promise<{ data: SystemConfig[] }> =>
    api.get('/configs'),
  // 批量更新配置（管理员）
  updateBatch: (configs: SystemConfig[]): Promise<{ message: string }> =>
    api.put('/configs', configs),
}

export default api
