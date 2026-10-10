// 号码文本批量解析：把一行一注的文本解析为与 NumberInput/DigitNumberInput
// 产出格式一致的 numbers JSON（小红球 {red,blue} / 大乐透 {front,back} /
// 七乐彩 {main,special} / 数字型 {numbers} / 七星彩 裸数组）。
import type { LotteryType } from '../types'
import { LOTTERY_CONFIGS } from '../types'

export interface ParsedLine {
  /** 行号（从 1 开始，对应文本域中的行） */
  line: number
  /** 原始文本 */
  raw: string
  /** 解析成功时的 numbers JSON 字符串 */
  numbers?: string
  /** 该行注数（复式按组合数展开） */
  betCount: number
  /** 解析失败原因 */
  error?: string
}

export interface ParseResult {
  lines: ParsedLine[]
  /** 成功行数 */
  okCount: number
  /** 失败行数 */
  failCount: number
  /** 总注数（各成功行 betCount 之和） */
  totalBets: number
}

/** 数字型（福彩3D/排列3/排列5）：每个位置可以给多个候选数，解析为直选复式 */
export interface DigitLine {
  play: '直选' | '组选3' | '组选6'
  positions: number[][]
}

const isDigitLottery = (type: LotteryType) => type === '福彩3D' || type === '排列3' || type === '排列5'

/** 分隔符：任意空白/中英文逗号/顿号/中英文分号/制表符（竖线与 + 保留作红蓝分区用） */
const SPLIT_RE = /[\s,，、;;]+/

/** 一行里把红蓝分区切开：支持 "01 05 + 06"、"01 05 | 06"、"01 05 06(蓝)" */
const splitMainBlue = (text: string): { main: string[]; blue: string[] } => {
  let mainPart = text
  let bluePart = ''
  if (text.includes('+')) {
    const idx = text.indexOf('+')
    mainPart = text.slice(0, idx)
    bluePart = text.slice(idx + 1)
  } else if (text.includes('|')) {
    const idx = text.indexOf('|')
    mainPart = text.slice(0, idx)
    bluePart = text.slice(idx + 1)
  } else if (text.includes('//')) {
    const idx = text.indexOf('//')
    mainPart = text.slice(0, idx)
    bluePart = text.slice(idx + 2)
  }
  const toks = (s: string) => s.split(SPLIT_RE).map(t => t.trim()).filter(t => t !== '')
  return { main: toks(mainPart), blue: toks(bluePart) }
}

/** 令牌转数字：接受 "01" "1" "12" 形式 */
const toNum = (tok: string): number | null => {
  if (!/^\d{1,2}$/.test(tok)) return null
  const n = parseInt(tok, 10)
  return Number.isNaN(n) ? null : n
}

const pad2 = (n: number) => String(n).padStart(2, '0')

/** 乐透型（双色球/大乐透/七乐彩）单行解析 */
const parseLottoLine = (type: LotteryType, text: string): { numbers?: string; betCount: number; error?: string } => {
  const config = LOTTERY_CONFIGS.find(c => c.type === type)
  if (!config) return { betCount: 0, error: '未知彩种' }

  const { main: mainToks, blue: blueToks } = splitMainBlue(text)
  if (mainToks.length === 0) return { betCount: 0, error: '没有号码' }

  const main: number[] = []
  for (const tok of mainToks) {
    const n = toNum(tok)
    if (n === null) return { betCount: 0, error: `无法识别的号码 "${tok}"` }
    if (n < config.redRange.min || n > config.redRange.max) {
      return { betCount: 0, error: `${n} 超出主号范围 ${config.redRange.min}-${config.redRange.max}` }
    }
    if (main.includes(n)) return { betCount: 0, error: `号码 ${n} 重复` }
    main.push(n)
  }
  if (main.length < config.redRange.count) {
    return { betCount: 0, error: `主号至少 ${config.redRange.count} 个，当前 ${main.length} 个` }
  }
  // 上限：双色球 20、大乐透 18（与 NumberInput 加号上限一致）；七乐彩不支持复式
  const mainMax = type === '双色球' ? 20 : type === '大乐透' ? 18 : config.redRange.count
  if (main.length > mainMax) {
    return { betCount: 0, error: `主号最多 ${mainMax} 个` }
  }
  main.sort((a, b) => a - b)

  // 蓝区：双色球/大乐透用配置（必填）；七乐彩没有 blueRange，
  // 特别号与主号同池（01-30），可省略，存 {main, special} 与算奖规则对齐
  const isQiLeCai = type === '七乐彩'
  const blueCfg = isQiLeCai
    ? { min: config.redRange.min, max: config.redRange.max, count: 1 }
    : config.blueRange
  const blueOptional = isQiLeCai
  const blue: number[] = []
  if (blueCfg && blueToks.length > 0) {
    for (const tok of blueToks) {
      const n = toNum(tok)
      if (n === null) return { betCount: 0, error: `无法识别的蓝区号码 "${tok}"` }
      if (n < blueCfg.min || n > blueCfg.max) {
        return { betCount: 0, error: `蓝区 ${n} 超出范围 ${blueCfg.min}-${blueCfg.max}` }
      }
      if (blue.includes(n)) return { betCount: 0, error: `蓝区号码 ${n} 重复` }
      blue.push(n)
    }
    if (blue.length < blueCfg.count) {
      return { betCount: 0, error: `蓝区至少 ${blueCfg.count} 个，当前 ${blue.length} 个` }
    }
    const blueMax = type === '双色球' ? 16 : type === '大乐透' ? 12 : blueCfg.count
    if (blue.length > blueMax) {
      return { betCount: 0, error: `蓝区最多 ${blueMax} 个` }
    }
    blue.sort((a, b) => a - b)
  } else if (blueCfg && !blueOptional) {
    return { betCount: 0, error: `缺少蓝区号码，请用 + 或 | 分隔` }
  }

  let numbers: string
  if (type === '双色球') {
    numbers = JSON.stringify({ red: main, blue })
  } else if (type === '大乐透') {
    numbers = JSON.stringify({ front: main, back: blue })
  } else {
    // 七乐彩 {main, special}
    numbers = JSON.stringify({ main, special: blue })
  }

  // 注数：主号组合 × 蓝区组合（七乐彩不支持复式，主号必须恰好 count 个）
  const comb = (n: number, k: number): number => {
    if (n < k || k < 0) return 0
    let r = 1
    for (let i = 0; i < k; i++) r = (r * (n - i)) / (i + 1)
    return Math.round(r)
  }
  const mainComb = isQiLeCai ? (main.length === config.redRange.count ? 1 : 0) : comb(main.length, config.redRange.count)
  const blueComb = blueCfg ? comb(blue.length, blueCfg.count) : 1
  return { numbers, betCount: mainComb * blueComb }
}

/** 七星彩：6+1，号码 0-9 允许重复，存裸数组 [a,b,c,d,e,f,蓝] */
const parseQiXingCaiLine = (text: string): { numbers?: string; betCount: number; error?: string } => {
  const { main, blue } = splitMainBlue(text)
  if (main.length === 0) return { betCount: 0, error: '没有号码' }

  // 无 + 分隔的 7 个数字：最后一个视为蓝号
  if (blue.length === 0 && main.length === 7) {
    return { numbers: JSON.stringify(main.map(toNum)), betCount: 1 }
  }
  if (blue.length === 0) {
    return { betCount: 0, error: '七星彩需要 6 个主号 + 1 个蓝号，可用 + 分隔' }
  }

  const nums: number[] = []
  for (const tok of main) {
    const n = toNum(tok)
    if (n === null || n > 9) return { betCount: 0, error: `无法识别的号码 "${tok}"（0-9）` }
    nums.push(n)
  }
  if (nums.length !== 6) return { betCount: 0, error: `前 6 位需要 6 个数字，当前 ${nums.length} 个` }
  const b = toNum(blue[0])
  if (b === null || b > 9) return { betCount: 0, error: `蓝区号码 "${blue[0]}" 无效（0-9）` }
  return { numbers: JSON.stringify([...nums, b]), betCount: 1 }
}

/** 数字型（福彩3D/排列3/排列5）单行解析 */
const parseDigitLine = (type: LotteryType, text: string): { numbers?: string; betCount: number; error?: string } => {
  const count = type === '排列5' ? 5 : 3
  // 玩法前缀支持："组选6 1 2 3"、"组三 1 2 3"
  let play: DigitLine['play'] = '直选'
  let body = text
  const m = text.match(/^(组选6|组六|组选3|组三)\s*/)
  if (m) {
    play = m[1].startsWith('组选6') || m[1] === '组六' ? '组选6' : '组选3'
    body = text.slice(m[0].length)
  }

  const toks = body.split(SPLIT_RE).map(t => t.trim()).filter(t => t !== '')
  if (toks.length === 0) return { betCount: 0, error: '没有号码' }

  const nums: number[] = []
  for (const tok of toks) {
    const n = toNum(tok)
    if (n === null || n > 9) return { betCount: 0, error: `无法识别的号码 "${tok}"（0-9）` }
    nums.push(n)
  }

  if (play === '组选6' || play === '组选3') {
    const uniq = Array.from(new Set(nums)).sort((a, b) => a - b)
    if (play === '组选6' && uniq.length < 3) return { betCount: 0, error: '组选6 至少 3 个不同号码' }
    if (play === '组选3' && uniq.length < 2) return { betCount: 0, error: '组选3 至少 2 个不同号码' }
    const comb = (n: number, k: number): number => {
      if (n < k || k < 0) return 0
      let r = 1
      for (let i = 0; i < k; i++) r = (r * (n - i)) / (i + 1)
      return Math.round(r)
    }
    const betCount = play === '组选6' ? (uniq.length >= 3 ? comb(uniq.length, 3) : 0) : uniq.length >= 2 ? uniq.length * (uniq.length - 1) : 0
    return { numbers: JSON.stringify({ play, group: uniq }), betCount }
  }

  // 直选：恰好 count 个数字，按位对应（不做位拆分，保持简单直觉）
  if (nums.length !== count) {
    return { betCount: 0, error: `${type} 直选需要 ${count} 个数字，当前 ${nums.length} 个` }
  }
  return { numbers: JSON.stringify(nums), betCount: 1 }
}

/**
 * 批量解析多行号码文本。
 * 每行一注；空行与以 # 开头的注释行跳过。
 */
export const parseNumberLines = (type: LotteryType, text: string): ParseResult => {
  const rawLines = text.split(/\r?\n/)
  const lines: ParsedLine[] = []
  let okCount = 0
  let totalBets = 0

  rawLines.forEach((raw, i) => {
    const trimmed = raw.trim()
    if (trimmed === '' || trimmed.startsWith('#')) return
    let r: { numbers?: string; betCount: number; error?: string }
    if (type === '七星彩') {
      r = parseQiXingCaiLine(trimmed)
    } else if (isDigitLottery(type)) {
      r = parseDigitLine(type, trimmed)
    } else {
      r = parseLottoLine(type, trimmed)
    }
    if (r.error || !r.numbers) {
      lines.push({ line: i + 1, raw, betCount: 0, error: r.error || '解析失败' })
    } else {
      lines.push({ line: i + 1, raw, numbers: r.numbers, betCount: r.betCount })
      okCount++
      totalBets += r.betCount
    }
  })

  return { lines, okCount, failCount: lines.length - okCount, totalBets }
}

/** 复式判定：任一成功行注数 > 1 */
export const hasFushi = (lines: ParsedLine[]): boolean => lines.some(l => l.numbers && l.betCount > 1)

/** 生成给用户的格式说明 */
export const formatHint = (type: LotteryType): string => {
  const config = LOTTERY_CONFIGS.find(c => c.type === type)
  if (!config) return ''
  if (type === '七星彩') return '每行一注：6 个数字 + 蓝号，如 1 2 3 4 5 6 + 7'
  if (isDigitLottery(type)) return `每行一注：${config.redRange.count} 个数字，如 ${Array.from({ length: config.redRange.count }, (_, i) => i).join(' ')}；组选在行首写 组选6 / 组选3`
  const sep = type === '七乐彩' ? '特别号' : '蓝区'
  return `每行一注：主号 + ${sep}，主号用空格分隔，蓝区用 + 分隔，如 ${Array.from({ length: config.redRange.count }, (_, i) => pad2(config.redRange.min + i * 3)).join(' ')} + ${pad2(config.blueRange?.min ?? 1)}`
}
