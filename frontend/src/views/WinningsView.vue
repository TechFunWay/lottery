<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { winningApi } from '../api'
import type { WinningRecord } from '../types'
import { Trophy, Gift, TrendingUp } from 'lucide-vue-next'

const winnings = ref<WinningRecord[]>([])
const loading = ref(false)
const filterType = ref('')

const loadWinnings = async () => {
  loading.value = true
  const res = await winningApi.list({ lottery_type: filterType.value }).catch(() => null)
  if (res) winnings.value = res.data || []
  loading.value = false
}

onMounted(loadWinnings)

// 格式化号码为带圆圈的样式
interface NumberBall {
  num: string
  type: 'red' | 'blue' | 'main'
}

const formatNumbers = (json: string): NumberBall[] => {
  try {
    const n = JSON.parse(json)
    const balls: NumberBall[] = []
    if (n.red) {
      n.red.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
      balls.push({ num: '|', type: 'main' })
      n.blue.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue' }))
    } else if (n.front) {
      n.front.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
      balls.push({ num: '|', type: 'main' })
      n.back.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue' }))
    } else if (n.main) {
      n.main.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
    } else if (n.numbers) {
      n.numbers.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    } else if (Array.isArray(n)) {
      n.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    }
    return balls
  } catch { return [] }
}

// 解析号码为数字数组
const parseNumbers = (json: string): { red?: number[], blue?: number[], front?: number[], back?: number[], main?: number[], numbers?: number[] } => {
  try {
    return JSON.parse(json)
  } catch { return {} }
}

// 判断是否有红球
const hasRed = (json: string): boolean => {
  const n = parseNumbers(json)
  return !!(n.red || n.front || n.main)
}

// 判断是否有蓝球
const hasBlue = (json: string): boolean => {
  const n = parseNumbers(json)
  return !!(n.blue || n.back)
}

// 获取红球数组
const getRedBalls = (json: string): number[] => {
  const n = parseNumbers(json)
  if (n.red) return n.red
  if (n.front) return n.front
  if (n.main) return n.main
  return []
}

// 获取蓝球数组
const getBlueBalls = (json: string): number[] => {
  const n = parseNumbers(json)
  if (n.blue) return n.blue
  if (n.back) return n.back
  return []
}

// 判断号码是否命中
const isHit = (num: number, purchaseJson: string, drawJson: string, isBlue: boolean = false): boolean => {
  const draw = parseNumbers(drawJson)
  const drawNums = isBlue
    ? (draw.blue || draw.back || [])
    : (draw.red || draw.front || draw.main || draw.numbers || [])
  return drawNums.includes(num)
}

const prizeColor = (level: number) => {
  if (level === 1) return 'text-amber-600 bg-amber-50 border-amber-200'
  if (level === 2) return 'text-slate-600 bg-slate-50 border-slate-200'
  if (level <= 4) return 'text-orange-600 bg-orange-50 border-orange-200'
  return 'text-blue-600 bg-blue-50 border-blue-200'
}

const totalWinning = () => winnings.value.reduce((sum, w) => sum + w.prize_amount, 0)
</script>

<template>
  <div class="animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-800">中奖记录</h1>
        <p class="text-slate-400 text-sm mt-1">共 {{ winnings.length }} 次中奖，累计奖金 ¥{{ totalWinning().toLocaleString() }}</p>
      </div>
      <div class="flex items-center gap-2 px-4 py-2 bg-emerald-50 rounded-xl">
        <TrendingUp class="w-5 h-5 text-emerald-500" />
        <span class="text-emerald-600 font-medium">¥{{ totalWinning().toLocaleString() }}</span>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-6">
      <select
        v-model="filterType"
        @change="loadWinnings"
        class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
      >
        <option value="">全部类型</option>
        <option value="双色球">双色球</option>
        <option value="大乐透">大乐透</option>
        <option value="福彩3D">福彩3D</option>
        <option value="排列3">排列3</option>
        <option value="排列5">排列5</option>
        <option value="七乐彩">七乐彩</option>
      </select>
    </div>

    <!-- Winnings Grid -->
    <div v-if="loading" class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else-if="winnings.length === 0" class="text-center py-16 text-slate-400">
      <Gift class="w-12 h-12 mx-auto mb-3 opacity-30" />
      <p>暂无中奖记录，继续加油！</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="win in winnings"
        :key="win.id"
        class="bg-white rounded-2xl p-5 card-shadow hover:card-shadow-hover transition-all duration-300"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-2">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-100 to-orange-100 flex items-center justify-center">
              <Trophy class="w-5 h-5 text-amber-500" />
            </div>
            <div>
              <span class="px-2 py-0.5 rounded-lg text-xs font-medium border" :class="prizeColor(win.prize_level)">
                {{ win.prize_name }}
              </span>
              <div class="text-xs text-slate-400 mt-0.5">{{ win.lottery_type }}</div>
            </div>
          </div>
          <div class="text-right">
            <div class="text-xl font-bold text-emerald-600">¥{{ win.prize_amount.toLocaleString() }}</div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-400">期号</span>
            <span class="text-slate-600 font-medium">{{ win.issue_number }}</span>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-400 shrink-0">投注号码</span>
            <div class="flex flex-wrap items-center gap-1">
              <span v-if="hasRed(win.purchase?.numbers || '{}')" v-for="(ball, idx) in getRedBalls(win.purchase?.numbers || '{}')" :key="'r'+idx"
                class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                :class="isHit(ball, win.purchase?.numbers || '{}', win.draw?.numbers || '{}') ? 'bg-red-500 text-white' : 'bg-slate-200 text-slate-600'">
                {{ String(ball).padStart(2, '0') }}
              </span>
              <span v-if="hasBlue(win.purchase?.numbers || '{}')" class="text-slate-300 w-5 text-center">|</span>
              <span v-if="hasBlue(win.purchase?.numbers || '{}')" v-for="(ball, idx) in getBlueBalls(win.purchase?.numbers || '{}')" :key="'b'+idx"
                class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                :class="isHit(ball, win.purchase?.numbers || '{}', win.draw?.numbers || '{}', true) ? 'bg-blue-500 text-white' : 'bg-slate-200 text-slate-600'">
                {{ String(ball).padStart(2, '0') }}
              </span>
            </div>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-400 shrink-0">开奖号码</span>
            <div class="flex flex-wrap gap-1">
              <template v-for="(ball, idx) in formatNumbers(win.draw?.numbers || '{}')" :key="idx">
                <span v-if="ball.num === '|'" class="text-slate-300 w-5 text-center">|</span>
                <span v-else
                  class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                  :class="ball.type === 'blue' ? 'bg-blue-500 text-white' : 'bg-red-500 text-white'">
                  {{ ball.num }}
                </span>
              </template>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-400">
          <span>购买金额 ¥{{ win.purchase?.amount || 0 }}</span>
          <span>{{ win.created_at.split('T')[0] }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
