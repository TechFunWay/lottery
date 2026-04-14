<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { statsApi, purchaseApi, drawApi } from '../api'
import type { OverviewStats, WinningRecord, LotteryType } from '../types'
import { LOTTERY_CONFIGS } from '../types'
import { TrendingUp, TrendingDown, DollarSign, Target, Trophy, ShoppingCart, Sparkles, X, CheckCircle, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const overview = ref<OverviewStats | null>(null)
const recentWinnings = ref<WinningRecord[]>([])
const loading = ref(true)

// Toast 提示
const toast = ref({
  show: false,
  type: 'info' as 'success' | 'error' | 'info',
  message: ''
})

const showToast = (type: 'success' | 'error' | 'info', message: string) => {
  toast.value = { show: true, type, message }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

// 生成幸运号弹窗
const showLuckyModal = ref(false)
const selectedLotteryType = ref<LotteryType>('双色球')
const generatedNumbers = ref('')
const generating = ref(false)
const issueNumber = ref('')
const drawRecords = ref<any[]>([])

// 加载开奖记录
const loadDrawRecords = async () => {
  const res = await drawApi.list({ size: 500 }).catch(() => null)
  if (res) drawRecords.value = res.data || []
}

// 验证期号：必须是未来的期号（大于等于当前最大期号）
const validateIssueNumber = (): boolean => {
  const inputIssue = issueNumber.value.trim()
  if (!inputIssue) {
    showToast('error', '请输入期号')
    return false
  }

  // 获取当前彩票类型的所有期号
  const currentTypeRecords = drawRecords.value
    .filter(d => d.lottery_type === selectedLotteryType.value)
    .map(d => d.issue_number)
    .sort()

  if (currentTypeRecords.length === 0) {
    // 没有历史期号，需要用户手动确认（可能是新一期的第一注）
    return confirm('未找到当前期号记录，确定要添加到新期号吗？')
  }

  // 获取最大期号
  const maxIssue = currentTypeRecords[currentTypeRecords.length - 1]

  // 验证期号必须大于等于当前最大期号
  if (inputIssue < maxIssue) {
    showToast('error', `期号不能小于当前期号 ${maxIssue}，请输入未来期号`)
    return false
  }

  return true
}

// 随机生成号码
const generateLuckyNumbers = (type: LotteryType): string => {
  const config = LOTTERY_CONFIGS.find(c => c.type === type)
  if (!config) return '{}'

  const generateUnique = (min: number, max: number, count: number): number[] => {
    const nums = new Set<number>()
    while (nums.size < count) {
      nums.add(Math.floor(Math.random() * (max - min + 1)) + min)
    }
    return Array.from(nums).sort((a, b) => a - b)
  }

  // 生成红球/前区/主号
  const main = generateUnique(config.redRange.min, config.redRange.max, config.redRange.count)

  // 3D、排列3、排列5 只有主号码
  if (!config.blueRange) {
    return JSON.stringify({ numbers: main })
  }

  // 有蓝球/后区
  const blue = generateUnique(config.blueRange.min, config.blueRange.max, config.blueRange.count)

  // 双色球/七星彩格式：{ red: [...], blue: [...] }
  if (type === '双色球' || type === '七星彩') {
    return JSON.stringify({ red: main, blue })
  }
  // 大乐透格式：{ front: [...], back: [...] }
  if (type === '大乐透') {
    return JSON.stringify({ front: main, back: blue })
  }
  // 七乐彩
  return JSON.stringify({ numbers: main })
}

// 重新生成
const regenerateNumbers = () => {
  generating.value = true
  setTimeout(() => {
    generatedNumbers.value = generateLuckyNumbers(selectedLotteryType.value)
    generating.value = false
  }, 300)
}

// 打开弹窗
const openLuckyModal = async () => {
  selectedLotteryType.value = '双色球'
  generatedNumbers.value = generateLuckyNumbers('双色球')
  issueNumber.value = ''
  await loadDrawRecords()
  showLuckyModal.value = true
}

// 确认添加购买记录
const confirmAddPurchase = async () => {
  if (!generatedNumbers.value || !issueNumber.value.trim()) return
  // 验证期号必须是未来期号
  if (!validateIssueNumber()) {
    return
  }
  try {
    await purchaseApi.create({
      lottery_type: selectedLotteryType.value,
      issue_number: issueNumber.value.trim(),
      purchase_date: new Date().toISOString().split('T')[0],
      numbers: generatedNumbers.value,
      bet_type: '单式',
      amount: 2,
      remark: '幸运号'
    })
    showLuckyModal.value = false
    router.push('/purchase')
  } catch (e) {
    console.error('添加失败', e)
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const [overviewRes, winningsRes] = await Promise.all([
      statsApi.overview().catch(() => null),
      statsApi.recentWinnings(5).catch(() => null)
    ])
    if (overviewRes?.data) overview.value = overviewRes.data
    recentWinnings.value = Array.isArray(winningsRes?.data) ? winningsRes!.data : []
  } catch (e) {
    console.error('首页数据加载失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const formatMoney = (v: number) => {
  if (!v && v !== 0) return '¥0.00'
  return `¥${Math.abs(v).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

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
    } else if (n.numbers) {
      // 排列3、排列5、福彩3D、七乐彩、七星彩：{ numbers: [...] }
      n.numbers.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    } else if (n.main) {
      n.main.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
    } else if (Array.isArray(n)) {
      n.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    }
    return balls
  } catch { return [] }
}

const prizeColor = (level: number) => {
  if (level === 1) return 'text-amber-600 bg-amber-50 border-amber-200'
  if (level === 2) return 'text-slate-600 bg-slate-50 border-slate-200'
  if (level <= 4) return 'text-orange-600 bg-orange-50 border-orange-200'
  return 'text-blue-600 bg-blue-50 border-blue-200'
}
</script>

<template>
  <div class="animate-fade-in min-h-screen flex flex-col pb-20 md:pb-24">
    <!-- Hero Section -->
    <div class="gradient-primary rounded-2xl p-8 mb-8 text-white shadow-lg shadow-blue-500/20">
      <h1 class="text-3xl font-bold mb-1">彩彩助手</h1>
      <p class="text-blue-100 text-sm mb-6">记录每一次幸运，分析每一次投注</p>
      <div class="flex items-end gap-4">
        <div class="text-blue-100 text-sm pb-1 shrink-0">总盈亏</div>
        <div class="text-4xl md:text-5xl font-bold truncate">
          {{ overview ? (overview.net_profit >= 0 ? '+' : '') + formatMoney(overview.net_profit) : '...' }}
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 card-shadow hover:card-shadow-hover transition-all duration-300">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center">
            <ShoppingCart class="w-5 h-5 text-blue-500" />
          </div>
          <span class="text-xs text-slate-400">总投入</span>
        </div>
        <div class="text-2xl font-bold text-slate-800">
          {{ overview ? formatMoney(overview.total_investment) : '--' }}
        </div>
        <div class="text-xs text-slate-400 mt-1">共 {{ overview?.total_bets ?? 0 }} 次投注</div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow hover:card-shadow-hover transition-all duration-300">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-emerald-50 flex items-center justify-center">
            <DollarSign class="w-5 h-5 text-emerald-500" />
          </div>
          <span class="text-xs text-slate-400">总中奖</span>
        </div>
        <div class="text-2xl font-bold text-emerald-600">
          {{ overview ? formatMoney(overview.total_winning) : '--' }}
        </div>
        <div class="text-xs text-slate-400 mt-1">中奖 {{ overview?.win_count ?? 0 }} 次</div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow hover:card-shadow-hover transition-all duration-300">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center"
            :class="(overview?.net_profit ?? 0) >= 0 ? 'bg-emerald-50' : 'bg-red-50'">
            <component
              :is="(overview?.net_profit ?? 0) >= 0 ? TrendingUp : TrendingDown"
              class="w-5 h-5"
              :class="(overview?.net_profit ?? 0) >= 0 ? 'text-emerald-500' : 'text-red-500'"
            />
          </div>
          <span class="text-xs text-slate-400">净盈亏</span>
        </div>
        <div class="text-2xl font-bold"
          :class="(overview?.net_profit ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-500'">
          {{ overview ? ((overview.net_profit >= 0 ? '+' : '') + formatMoney(overview.net_profit)) : '--' }}
        </div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow hover:card-shadow-hover transition-all duration-300">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-amber-50 flex items-center justify-center">
            <Target class="w-5 h-5 text-amber-500" />
          </div>
          <span class="text-xs text-slate-400">中奖率</span>
        </div>
        <div class="text-2xl font-bold text-amber-600">
          {{ overview ? overview.win_rate.toFixed(1) + '%' : '--' }}
        </div>
      </div>
    </div>

    <!-- Recent Winnings -->
    <div class="bg-white rounded-2xl card-shadow p-6">
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-2">
          <Trophy class="w-5 h-5 text-amber-500" />
          <h2 class="text-lg font-semibold text-slate-800">最近中奖</h2>
        </div>
        <button
          @click="router.push('/winnings')"
          class="text-sm text-blue-500 hover:text-blue-600 cursor-pointer"
        >
          查看全部 →
        </button>
      </div>

      <div v-if="loading" class="flex justify-center py-8">
        <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="recentWinnings.length === 0" class="text-center py-12 text-slate-400">
        <Trophy class="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p>暂无中奖记录，加油！</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="win in recentWinnings"
          :key="win.id"
          class="flex items-center justify-between p-4 rounded-xl bg-slate-50 hover:bg-slate-100 transition-colors"
        >
          <div class="flex items-center gap-3">
            <span class="px-2 py-1 rounded-lg text-xs font-medium border"
              :class="prizeColor(win.prize_level)">
              {{ win.prize_name }}
            </span>
            <div>
              <div class="text-sm font-medium text-slate-700">
                {{ win.lottery_type }} · {{ win.issue_number }}期
              </div>
              <div class="flex flex-wrap gap-0.5 mt-0.5">
                <span v-for="(ball, idx) in formatNumbers(win.purchase?.numbers || '{}')" :key="idx"
                  class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                  :class="ball.num === '|'
                    ? 'text-slate-300 text-xs bg-transparent'
                    : ball.type === 'blue'
                      ? 'bg-blue-500 text-white'
                      : 'bg-slate-200 text-slate-600'">
                  {{ ball.num }}
                </span>
              </div>
            </div>
          </div>
          <div class="text-right">
            <div class="text-base font-bold text-emerald-600">¥{{ win.prize_amount.toLocaleString() }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions - Fixed Bottom -->
    <div class="fixed bottom-0 left-0 right-0 z-50 bg-white/95 backdrop-blur-sm border-t border-slate-200 px-4 py-3 md:hidden">
      <div class="max-w-xl mx-auto grid grid-cols-3 gap-3">
        <button
          @click="router.push('/purchase')"
          class="flex items-center justify-center gap-1.5 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer text-sm"
        >
          <ShoppingCart class="w-4 h-4" />
          <span>录入购买</span>
        </button>
        <button
          @click="router.push('/draw')"
          class="flex items-center justify-center gap-1.5 py-2.5 bg-emerald-500 hover:bg-emerald-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-emerald-500/30 cursor-pointer text-sm"
        >
          <Trophy class="w-4 h-4" />
          <span>录入开奖</span>
        </button>
        <button
          @click="openLuckyModal"
          class="flex items-center justify-center gap-1.5 py-2.5 bg-amber-500 hover:bg-amber-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-amber-500/30 cursor-pointer text-sm"
        >
          <Sparkles class="w-4 h-4" />
          <span>幸运号</span>
        </button>
      </div>
    </div>

    <!-- Desktop Quick Actions -->
    <div class="hidden md:block fixed bottom-6 left-1/2 -translate-x-1/2 w-full max-w-xl px-4">
      <div class="grid grid-cols-3 gap-2">
        <button
          @click="router.push('/purchase')"
          class="flex items-center justify-center gap-1.5 py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-2xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer text-sm"
        >
          <ShoppingCart class="w-4 h-4" />
          <span>录入购买</span>
        </button>
        <button
          @click="router.push('/draw')"
          class="flex items-center justify-center gap-1.5 py-3 bg-emerald-500 hover:bg-emerald-600 text-white rounded-2xl font-medium transition-all duration-200 shadow-lg shadow-emerald-500/30 cursor-pointer text-sm"
        >
          <Trophy class="w-4 h-4" />
          <span>录入开奖</span>
        </button>
        <button
          @click="openLuckyModal"
          class="flex items-center justify-center gap-1.5 py-3 bg-amber-500 hover:bg-amber-600 text-white rounded-2xl font-medium transition-all duration-200 shadow-lg shadow-amber-500/30 cursor-pointer text-sm"
        >
          <Sparkles class="w-4 h-4" />
          <span>幸运号</span>
        </button>
      </div>
    </div>

    <!-- 生成幸运号弹窗 -->
    <div v-if="showLuckyModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showLuckyModal = false"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-sm animate-slide-up">
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between rounded-t-2xl">
          <h2 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
            <Sparkles class="w-5 h-5 text-amber-500" />
            生成幸运号
          </h2>
          <button @click="showLuckyModal = false" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <!-- 选择彩票类型 -->
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-2">选择彩票类型</label>
            <div class="relative">
              <select
                v-model="selectedLotteryType"
                @change="regenerateNumbers"
                class="w-full px-4 py-2.5 pr-10 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-amber-400 cursor-pointer"
              >
                <option v-for="config in LOTTERY_CONFIGS" :key="config.type" :value="config.type">
                  {{ config.name }}
                </option>
              </select>
            </div>
          </div>

          <!-- 期号 -->
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-2">期号</label>
            <input
              v-model="issueNumber"
              type="text"
              placeholder="请输入期号"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-amber-400"
            />
          </div>

          <!-- 生成的号码 -->
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-2">生成的幸运号</label>
            <div class="bg-gradient-to-br from-amber-50 to-orange-50 rounded-xl p-4 text-center">
              <div v-if="generating" class="py-4">
                <div class="w-8 h-8 border-2 border-amber-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
              </div>
              <div v-else class="flex flex-wrap items-center justify-center gap-1.5">
                <template v-for="(ball, idx) in formatNumbers(generatedNumbers)" :key="idx">
                  <span v-if="ball.num === '|'" class="text-slate-300 w-4 text-center">|</span>
                  <span v-else
                    class="inline-flex items-center justify-center w-6 h-6 rounded-full text-sm font-bold"
                    :class="ball.type === 'blue' ? 'bg-blue-500 text-white' : ball.type === 'main' ? 'bg-amber-500 text-white' : 'bg-red-500 text-white'">
                    {{ ball.num }}
                  </span>
                </template>
              </div>
            </div>
          </div>

          <!-- 重新生成按钮 -->
          <button
            @click="regenerateNumbers"
            :disabled="generating"
            class="w-full py-2.5 text-amber-600 hover:bg-amber-50 rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-50"
          >
            重新生成
          </button>
        </div>

        <div class="border-t border-slate-100 px-6 py-4 flex gap-3">
          <button
            @click="showLuckyModal = false"
            class="flex-1 px-4 py-2.5 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer"
          >
            取消
          </button>
          <button
            @click="confirmAddPurchase"
            class="flex-1 px-4 py-2.5 bg-amber-500 hover:bg-amber-600 text-white rounded-xl font-medium transition-colors cursor-pointer"
          >
            添加到购买记录
          </button>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <Transition>
      <div
        v-if="toast.show"
        class="fixed top-20 left-1/2 -translate-x-1/2 z-[100] px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-slide-up"
        :class="{
          'bg-emerald-500 text-white': toast.type === 'success',
          'bg-red-500 text-white': toast.type === 'error',
          'bg-blue-500 text-white': toast.type === 'info'
        }"
      >
        <CheckCircle v-if="toast.type === 'success'" class="w-5 h-5" />
        <AlertCircle v-else-if="toast.type === 'error'" class="w-5 h-5" />
        <AlertCircle v-else class="w-5 h-5" />
        <span class="font-medium">{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>
