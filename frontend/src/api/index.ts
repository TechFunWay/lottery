import axios from 'axios'
import type { PurchaseRecord, DrawResult, OverviewStats, PrizeDistribution, TrendData, WinningRecord, User, AuthResponse, SystemConfig, PublicConfigs, FootballMatch, FootballBet, FootballOverview } from '../types'

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
    // 如果是登录接口失败，不跳转（让调用者处理错误显示）
    const isLoginRequest = error.config?.url?.includes('/auth/login')
    if (error.response?.status === 401 && !isLoginRequest) {
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
  multiple?: number
  append?: boolean
  periods?: number
  remark?: string
}

export const purchaseApi = {
  create: (data: CreatePurchasePayload) => api.post('/purchases', data),
  list: (params?: { lottery_type?: string; status?: string; page?: number; size?: number }): Promise<ListResponse<PurchaseRecord>> =>
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

export interface ListResponse<T> {
  data: T[]
  total: number
}

export const drawApi = {
  create: (data: CreateDrawPayload) => api.post('/draws', data),
  list: (params?: { lottery_type?: string; page?: number; size?: number }): Promise<ListResponse<DrawResult>> =>
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
  list: (params?: { lottery_type?: string; page?: number; size?: number }): Promise<ListResponse<WinningRecord>> =>
    api.get('/winnings', { params }),
  recheck: () => api.post('/winnings/recheck'),
  // 手动调整中奖金额：传 number 设置，传 null 还原为系统计算值
  update: (id: number, manual_amount: number | null) =>
    api.put(`/winnings/${id}`, { manual_amount }),
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

// ===== 竞彩足球数据源 API =====
export const footballConfigApi = {
  // 公开状态(system-level)
  getSystemStatus: (): Promise<{ data: FootballConfigStatus }> =>
    api.get('/football/config/status'),
  // 当前用户的有效 key 状态
  getMyStatus: (): Promise<{ data: FootballConfigStatus }> =>
    api.get('/football/config/me'),
  // 设置/清除当前用户的 key(空 = 清除)
  setMyKey: (key: string): Promise<{ message: string }> =>
    api.put('/football/config/me', { key }),
  // 管理员设置全局 key
  setGlobalKey: (key: string): Promise<{ message: string }> =>
    api.put('/football/config/global', { key }),
  // 测试 key(body.key 留空则用当前用户有效 key)
  testKey: (key: string): Promise<{ data: FootballTestResult }> =>
    api.post('/football/config/test', { key }),
}

// ===== 系统信息 API =====
export interface VersionInfo {
  name: string
  version: string
  buildTime: string
  gitCommit: string
  status: string
}

export interface UpgradeHistory {
  id: number
  version: string
  name: string
  status: number
  startTime: string
  endTime: string
  remark: string
  createdAt: string
}

export const systemApi = {
  // 获取版本信息
  getVersion: (): Promise<VersionInfo> =>
    api.get('/version'),
  // 获取当前数据库版本
  getCurrentVersion: (): Promise<{ version: string }> =>
    api.get('/version/current'),
  // 获取升级历史
  getUpgradeHistory: (): Promise<{ data: UpgradeHistory[] }> =>
    api.get('/version/history'),
}

// ===== 竞彩足球 API =====
export interface CreateFootballMatchPayload {
  match_id: string
  issue_number?: string
  league?: string
  home_team: string
  away_team: string
  match_time: string
  home_score?: number
  away_score?: number
  half_home_score?: number
  half_away_score?: number
  handicap?: number
  status?: string
}

export interface CreateFootballBetPayload {
  bet_type: string
  amount: number
  multiple?: number
  selections: string
  remark?: string
}

export const footballMatchApi = {
  create: (data: CreateFootballMatchPayload) => api.post('/football/matches', data),
  list: (params?: { league?: string; status?: string; page?: number; size?: number }) =>
    api.get('/football/matches', { params }),
  update: (id: number, data: CreateFootballMatchPayload) => api.put(`/football/matches/${id}`, data),
  delete: (id: number) => api.delete(`/football/matches/${id}`),
  fetch: () => api.get('/football/matches/fetch'),
  fetchResults: () => api.post('/football/matches/fetch-results'),
}

export const footballBetApi = {
  create: (data: CreateFootballBetPayload) => api.post('/football/bets', data),
  list: (params?: { status?: string; page?: number; size?: number }) =>
    api.get('/football/bets', { params }),
  update: (id: number, data: CreateFootballBetPayload) => api.put(`/football/bets/${id}`, data),
  delete: (id: number) => api.delete(`/football/bets/${id}`),
  recheck: () => api.post('/football/bets/recheck'),
  overview: (): Promise<{ data: FootballOverview }> =>
    api.get('/football/overview'),
}

export default api
