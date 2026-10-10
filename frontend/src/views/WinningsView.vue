<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { winningApi } from '../api'
import Pagination from '../components/Pagination.vue'
import type { WinningRecord } from '../types'
import { Trophy, Gift, TrendingUp, Edit2, RotateCcw, X } from 'lucide-vue-next'

const winnings = ref<WinningRecord[]>([])
const loading = ref(false)
const filterType = ref('')

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

watch([currentPage, pageSize], () => {
  loadWinnings()
})

const loadWinnings = async () => {
  loading.value = true
  const res = await winningApi.list({
    lottery_type: filterType.value,
    page: currentPage.value,
    size: pageSize.value
  }).catch(() => null)
  if (res) {
    winnings.value = res.data || []
    total.value = res.total || 0
  }
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

// ===== 数字型彩票（福彩3D/排列3/排列5）展示 =====
const digitTypes = ['福彩3D', '排列3', '排列5']
const isDigitType = (type: string) => digitTypes.includes(type)

const digitPosLabels = (count: number) =>
  count === 5 ? ['万', '千', '百', '十', '个'] : ['百', '十', '个']

const formatDigitBet = (json: string): string => {
  const n = parseNumbers(json) as any
  const play = n.play || n.bet_type || '直选'
  if (Array.isArray(n.group) || play === '组选3' || play === '组选6') {
    const g = (n.group || n.numbers || []) as number[]
    return `${play}：${[...g].sort((a, b) => a - b).join(' ')}`
  }
  let pos: number[][] = []
  if (Array.isArray(n.positions)) pos = n.positions
  else if (Array.isArray(n.numbers)) pos = n.numbers.map((x: number) => [x])
  else if (Array.isArray(n)) pos = (n as number[]).map((x: number) => [x])
  const labels = digitPosLabels(pos.length)
  if (play === '定位胆') {
    const parts = pos
      .map((p, i) => (p && p.length ? `${labels[i]}:${[...p].sort((a, b) => a - b).join(',')}` : null))
      .filter(Boolean)
    return `定位胆 ${parts.join('  ')}`
  }
  return `${play}：` + pos.map(p => (p && p.length ? [...p].sort((a, b) => a - b).join(',') : '-')).join(' | ')
}

const getDrawDigits = (json: string): number[] => {
  const n = parseNumbers(json) as any
  if (Array.isArray(n)) return n
  if (Array.isArray(n.numbers)) return n.numbers
  return []
}

const prizeColor = (level: number) => {
  if (level === 1) return 'text-amber-600 bg-amber-50 border-amber-200'
  if (level === 2) return 'text-slate-600 bg-slate-50 border-slate-200'
  if (level <= 4) return 'text-orange-600 bg-orange-50 border-orange-200'
  return 'text-blue-600 bg-blue-50 border-blue-200'
}

// 有效奖金：手动调整过则用手动值，否则用系统计算值
const effAmount = (w: WinningRecord): number =>
  w.manual_amount != null ? w.manual_amount : w.prize_amount

const isAdjusted = (w: WinningRecord): boolean => w.manual_amount != null

const totalWinning = () => winnings.value.reduce((sum, w) => sum + effAmount(w), 0)

// ===== 金额编辑 =====
const editModal = ref(false)
const editingWin = ref<WinningRecord | null>(null)
const editAmount = ref<number>(0)
const saving = ref(false)

const toast = ref({ show: false, type: 'info' as 'success' | 'error', message: '' })
const showToast = (type: 'success' | 'error', message: string) => {
  toast.value = { show: true, type, message }
  setTimeout(() => { toast.value.show = false }, 3000)
}

const openEdit = (win: WinningRecord) => {
  editingWin.value = win
  editAmount.value = effAmount(win)
  editModal.value = true
}

const closeEdit = () => {
  editModal.value = false
  editingWin.value = null
}

const saveEdit = async () => {
  if (!editingWin.value) return
  if (editAmount.value < 0 || isNaN(editAmount.value)) {
    showToast('error', '请输入有效金额')
    return
  }
  saving.value = true
  try {
    await winningApi.update(editingWin.value.id, editAmount.value)
    closeEdit()
    await loadWinnings()
    showToast('success', '金额已调整')
  } catch (e) {
    console.error('调整失败', e)
    showToast('error', '调整失败，请重试')
  } finally {
    saving.value = false
  }
}

// 还原为系统计算金额
const resetAmount = async (win: WinningRecord) => {
  saving.value = true
  try {
    await winningApi.update(win.id, null)
    if (editModal.value) closeEdit()
    await loadWinnings()
    showToast('success', '已还原为系统金额')
  } catch (e) {
    console.error('还原失败', e)
    showToast('error', '还原失败，请重试')
  } finally {
    saving.value = false
  }
}
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
            <div class="flex items-center justify-end gap-1.5">
              <div class="text-xl font-bold text-emerald-600">¥{{ effAmount(win).toLocaleString() }}</div>
              <button @click="openEdit(win)" class="p-1 text-slate-300 hover:text-blue-500 cursor-pointer" title="调整金额">
                <Edit2 class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-if="isAdjusted(win)" class="flex items-center justify-end gap-1 mt-0.5">
              <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-medium">已调整</span>
              <span class="text-[10px] text-slate-400 line-through">¥{{ win.prize_amount.toLocaleString() }}</span>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-400">期号</span>
            <span class="text-slate-600 font-medium">{{ win.issue_number }}</span>
          </div>
          <!-- 数字型彩票 -->
          <div v-if="isDigitType(win.lottery_type)" class="flex items-start justify-between text-sm gap-2">
            <span class="text-slate-400 shrink-0">投注号码</span>
            <span class="text-slate-700 font-mono text-right break-all">{{ formatDigitBet(win.purchase?.numbers || '{}') }}</span>
          </div>
          <div class="flex items-center justify-between text-sm" v-if="!isDigitType(win.lottery_type)">
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

    <Pagination
      v-if="total > 0"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
    />

    <!-- 金额调整弹窗 -->
    <div v-if="editModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeEdit"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-sm animate-slide-up">
        <div class="border-b border-slate-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800">调整中奖金额</h2>
          <button @click="closeEdit" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <p class="text-xs text-slate-400">
            适用于活动翻倍等情况。系统计算金额为
            <span class="font-medium text-slate-600">¥{{ editingWin?.prize_amount.toLocaleString() }}</span>，
            调整后「重新检查中奖」不会覆盖。
          </p>
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">中奖金额 (元)</label>
            <input
              v-model.number="editAmount"
              type="number"
              step="0.01"
              min="0"
              class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400"
            />
          </div>
        </div>
        <div class="border-t border-slate-100 px-6 py-4 flex items-center gap-3">
          <button
            v-if="editingWin && isAdjusted(editingWin)"
            @click="resetAmount(editingWin)"
            :disabled="saving"
            class="flex items-center gap-1.5 px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-50"
          >
            <RotateCcw class="w-4 h-4" />
            还原
          </button>
          <div class="flex-1"></div>
          <button @click="closeEdit" class="px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button
            @click="saveEdit"
            :disabled="saving"
            class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-50"
          >
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <Transition name="toast">
      <div
        v-if="toast.show"
        class="fixed top-20 left-1/2 -translate-x-1/2 z-[100] px-6 py-3 rounded-xl shadow-lg flex items-center gap-3"
        :class="{
          'bg-emerald-500 text-white': toast.type === 'success',
          'bg-red-500 text-white': toast.type === 'error'
        }"
      >
        <span class="font-medium">{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>
