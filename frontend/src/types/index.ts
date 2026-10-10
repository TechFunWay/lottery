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

export type FootballKeySource = 'user' | 'admin' | 'env' | 'builtin' | 'none'

export interface FootballConfigStatus {
  configured: boolean
  source: FootballKeySource
  masked_key: string
  registration_url?: string
}

export interface FootballTestResult {
  success: boolean
  message: string
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
  multiple: number
  append: boolean
  periods: number
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
  manual_amount: number | null
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

// 系统版本信息
export interface VersionInfo {
  name: string
  version: string
  buildTime: string
  gitCommit: string
  status: string
}

// 系统升级记录
export interface SystemUpgrade {
  id: number
  version: string
  name: string
  status: number
  startTime: string
  endTime: string
  remark: string
  createdAt: string
}

// ===== 竞彩足球类型 =====

export type FootballPlayType = '胜平负' | '让球胜平负' | '比分' | '总进球' | '半全场'
export type FootballMatchStatus = '未开赛' | '进行中' | '已完赛' | '已取消' | '延期'
export type FootballBetType = '单关' | '2串1' | '3串1' | '4串1' | '5串1' | '6串1' | '7串1' | '8串1'
export type FootballBetStatus = '待开奖' | '已中奖' | '未中奖' | '部分中奖'

export interface FootballMatch {
  id: number
  match_id: string
  issue_number: string
  league: string
  home_team: string
  away_team: string
  match_time: string
  status: FootballMatchStatus
  home_score: number
  away_score: number
  half_home_score: number
  half_away_score: number
  handicap: number
  source: 'manual' | 'auto'
  created_at: string
  updated_at: string
}

export interface FootballSelection {
  match_id: string
  play_type: FootballPlayType
  selection: string
  odds: number
  handicap?: number
}

export interface FootballBet {
  id: number
  user_id: number
  bet_type: FootballBetType
  amount: number
  multiple: number
  status: FootballBetStatus
  selections: string
  remark: string
  win_amount: number
  created_at: string
  updated_at: string
}

export interface FootballOverview {
  total_bets: number
  total_amount: number
  total_win: number
  net_profit: number
  win_count: number
  win_rate: number
}

export const FOOTBALL_PLAY_TYPES: { type: FootballPlayType; name: string; options: string[] }[] = [
  {
    type: '胜平负',
    name: '胜平负',
    options: ['3', '1', '0']
  },
  {
    type: '让球胜平负',
    name: '让球胜平负',
    options: ['3', '1', '0']
  },
  {
    type: '比分',
    name: '比分',
    options: [
      '1:0', '2:0', '2:1', '3:0', '3:1', '3:2',
      '4:0', '4:1', '4:2', '5:0', '5:1', '5:2',
      '0:1', '0:2', '1:2', '0:3', '1:3', '2:3',
      '0:4', '1:4', '2:4', '0:5', '1:5', '2:5',
      '0:0', '1:1', '2:2', '3:3', '胜其他', '平其他', '负其他'
    ]
  },
  {
    type: '总进球',
    name: '总进球',
    options: ['0', '1', '2', '3', '4', '5', '6', '7+']
  },
  {
    type: '半全场',
    name: '半全场',
    options: ['胜胜', '胜平', '胜负', '平胜', '平平', '平负', '负胜', '负平', '负负']
  }
]

export const FOOTBALL_BET_TYPES: FootballBetType[] = ['单关', '2串1', '3串1', '4串1', '5串1', '6串1', '7串1', '8串1']

export const WDL_LABELS: Record<string, string> = {
  '3': '主胜',
  '1': '平',
  '0': '客胜'
}
