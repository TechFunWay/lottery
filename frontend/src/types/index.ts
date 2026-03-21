// 用户类型
export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'disabled'

// 用户信息
export interface User {
  id: number
  username: string
  email: string
  role: UserRole
  status: UserStatus
  last_login: string | null
  created_at: string
  updated_at: string
}

// 认证响应
export interface AuthResponse {
  data: {
    user: User
    token: string
  }
  message: string
}

// 系统配置
export interface SystemConfig {
  id?: number
  key: string
  value: string
  remark?: string
}

// 公开配置（无需登录）
export interface PublicConfigs {
  allow_register: boolean
}

// 彩票类型
export type LotteryType = '双色球' | '大乐透' | '福彩3D' | '排列3' | '排列5' | '七乐彩' | '七星彩'

// 投注方式
export type BetType = '单式' | '复式' | '胆拖'

// 购买记录
export interface PurchaseRecord {
  id: number
  lottery_type: LotteryType
  issue_number: string
  purchase_date: string
  numbers: string
  bet_type: BetType
  amount: number
  remark: string
  status: '待开奖' | '已开奖' | '未中奖' | '已中奖'
  created_at: string
  updated_at: string
}

// 开奖结果
export interface DrawResult {
  id: number
  lottery_type: LotteryType
  issue_number: string
  draw_date: string
  numbers: string
  source: 'manual' | 'auto'
  created_at: string
}

// 中奖记录
export interface WinningRecord {
  id: number
  purchase_id: number
  draw_id: number
  lottery_type: LotteryType
  issue_number: string
  prize_level: number
  prize_name: string
  prize_amount: number
  purchase: PurchaseRecord
  draw: DrawResult
  created_at: string
}

// 盈亏总览
export interface OverviewStats {
  total_investment: number
  total_winning: number
  net_profit: number
  total_bets: number
  win_count: number
  win_rate: number
}

// 奖级分布
export interface PrizeDistribution {
  prize_name: string
  count: number
  amount: number
}

// 号码频率
export interface NumberFrequency {
  number: number
  lottery_type: string
  position: string
  count: number
}

// 趋势数据
export interface TrendData {
  month: string
  total_bets: number
  win_count: number
  investment: number
  win_amount: number
}

// 彩票规则配置
export interface LotteryConfig {
  type: LotteryType
  name: string
  redRange: { min: number; max: number; count: number }
  blueRange?: { min: number; max: number; count: number }
  specialRange?: { min: number; max: number; count: number }
  description: string
}

export const LOTTERY_CONFIGS: LotteryConfig[] = [
  {
    type: '双色球',
    name: '双色球',
    redRange: { min: 1, max: 33, count: 6 },
    blueRange: { min: 1, max: 16, count: 1 },
    description: '6个红球(01-33) + 1个蓝球(01-16)'
  },
  {
    type: '大乐透',
    name: '大乐透',
    redRange: { min: 1, max: 35, count: 5 },
    blueRange: { min: 1, max: 12, count: 2 },
    description: '5个前区(01-35) + 2个后区(01-12)'
  },
  {
    type: '福彩3D',
    name: '福彩3D',
    redRange: { min: 0, max: 9, count: 3 },
    description: '3个数字(0-9)，支持直选/组选'
  },
  {
    type: '排列3',
    name: '排列3',
    redRange: { min: 0, max: 9, count: 3 },
    description: '3个数字(0-9)，位置对应'
  },
  {
    type: '排列5',
    name: '排列5',
    redRange: { min: 0, max: 9, count: 5 },
    description: '5个数字(0-9)，位置对应'
  },
  {
    type: '七乐彩',
    name: '七乐彩',
    redRange: { min: 1, max: 30, count: 7 },
    description: '7个主号(01-30)'
  },
  {
    type: '七星彩',
    name: '七星彩',
    redRange: { min: 0, max: 9, count: 6 },
    blueRange: { min: 0, max: 9, count: 1 },
    description: '6个红球(0-9) + 1个蓝球(0-9)'
  }
]
