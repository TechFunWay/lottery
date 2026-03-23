<script setup lang="ts">
import { ref, onMounted, computed, reactive, nextTick } from 'vue'
import { purchaseApi, winningApi, drawApi } from '../api'
import NumberInput from '../components/NumberInput.vue'
import type { PurchaseRecord, LotteryType, DrawResult } from '../types'
import { LOTTERY_CONFIGS } from '../types'
import { Plus, Trash2, Edit2, Search, X, ChevronDown, RefreshCw, AlertCircle, CheckCircle, Sparkles } from 'lucide-vue-next'

const purchases = ref<PurchaseRecord[]>([])
const drawResults = ref<DrawResult[]>([])
const winningRecords = ref<any[]>([])
const loading = ref(false)
const rechecking = ref(false)
const showModal = ref(false)
const editingId = ref<number | null>(null)
const filterType = ref('')
const filterStatus = ref('')
const deleteConfirm = ref(false)
const deleteId = ref<number | null>(null)

// 生成幸运号弹窗
const showLuckyModal = ref(false)
const selectedLotteryType = ref<LotteryType>('双色球')
const generatedNumbers = ref('')
const generating = ref(false)

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

  const main = generateUnique(config.redRange.min, config.redRange.max, config.redRange.count)

  if (!config.blueRange) {
    return JSON.stringify({ numbers: main })
  }

  const blue = generateUnique(config.blueRange.min, config.blueRange.max, config.blueRange.count)

  if (type === '双色球') {
    return JSON.stringify({ red: main, blue })
  }
  if (type === '大乐透') {
    return JSON.stringify({ front: main, back: blue })
  }
  return JSON.stringify({ red: main })
}

// 打开幸运号弹窗
const openLuckyModal = async () => {
  selectedLotteryType.value = '双色球'
  generatedNumbers.value = generateLuckyNumbers('双色球')
  issueNumber.value = ''
  // 加载开奖记录用于验证期号
  if (drawResults.value.length === 0) {
    const res = await drawApi.list({ size: 500 }).catch(() => null)
    if (res) drawResults.value = res.data || []
  }
  showLuckyModal.value = true
}

// 重新生成
const regenerateNumbers = () => {
  generating.value = true
  setTimeout(() => {
    generatedNumbers.value = generateLuckyNumbers(selectedLotteryType.value)
    generating.value = false
  }, 300)
}

// 期号输入
const issueNumber = ref('')

// 验证期号：必须是未来的期号（大于等于当前最大期号）
const validateIssueNumber = (): boolean => {
  const inputIssue = issueNumber.value.trim()
  if (!inputIssue) return false

  // 获取当前彩票类型的所有期号
  const currentTypeRecords = drawResults.value
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

// 确认添加到购买记录
const confirmAddFromLucky = async () => {
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
    loadPurchases()
    showToast('success', '添加成功')
  } catch (e) {
    console.error('添加失败', e)
    showToast('error', '添加失败')
  }
}

// 通用弹窗
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

const form = ref({
  lottery_type: '双色球' as LotteryType,
  issue_number: '',
  purchase_date: new Date().toISOString().split('T')[0],
  numbers: '',
  bet_type: '单式',
  amount: 2,
  remark: ''
})

const betTypes = ['单式', '复式', '胆拖']
const lotteryTypes = LOTTERY_CONFIGS.map(c => c.type)

// 表单校验错误
const errors = reactive({
  issue_number: '',
  numbers: ''
})

const validateForm = () => {
  errors.issue_number = form.value.issue_number.trim() ? '' : '期号为必填项'
  errors.numbers = form.value.numbers ? '' : '请输入号码'
  return !errors.issue_number && !errors.numbers
}

const statusColors: Record<string, string> = {
  '待开奖': 'bg-slate-100 text-slate-600',
  '已开奖': 'bg-blue-100 text-blue-600',
  '未中奖': 'bg-slate-100 text-slate-500',
  '已中奖': 'bg-emerald-100 text-emerald-600'
}

const loadPurchases = async () => {
  loading.value = true
  const res = await purchaseApi.list({
    lottery_type: filterType.value,
    status: filterStatus.value
  }).catch(() => null)
  if (res) purchases.value = res.data || []
  // 加载开奖记录用于号码比对
  const drawRes = await drawApi.list({ size: 100 }).catch(() => null)
  if (drawRes) drawResults.value = drawRes.data || []
  // 加载中奖记录用于显示奖金
  const winningRes = await winningApi.list({ size: 100 }).catch(() => null)
  if (winningRes) winningRecords.value = winningRes.data || []
  loading.value = false
}

onMounted(loadPurchases)

const openModal = (item?: PurchaseRecord) => {
  // 清空校验错误
  errors.issue_number = ''
  errors.numbers = ''
  if (item) {
    editingId.value = item.id
    form.value = {
      lottery_type: item.lottery_type,
      issue_number: item.issue_number,
      purchase_date: item.purchase_date.split('T')[0],
      numbers: item.numbers,
      bet_type: item.bet_type,
      amount: item.amount,
      remark: item.remark
    }
  } else {
    editingId.value = null
    form.value = {
      lottery_type: '双色球',
      issue_number: '',
      purchase_date: new Date().toISOString().split('T')[0],
      numbers: '',
      bet_type: '单式',
      amount: 2,
      remark: ''
    }
  }
  showModal.value = true
  // 延迟触发一次 watch 确保 NumberInput 正确回填（因为 NumberInput 是 v-if 渲染的）
  setTimeout(() => {
    if (item && form.value.numbers) {
      // 触发一次 numbers 的响应式更新
      const temp = form.value.numbers
      form.value.numbers = ''
      form.value.numbers = temp
    }
  }, 0)
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
}

const savePurchase = async () => {
  if (!validateForm()) return
  const payload = {
    lottery_type: form.value.lottery_type,
    issue_number: form.value.issue_number,
    purchase_date: form.value.purchase_date,
    numbers: form.value.numbers,
    bet_type: form.value.bet_type,
    amount: form.value.amount,
    remark: form.value.remark
  }
  const isEdit = !!editingId.value
  try {
    if (editingId.value) {
      await purchaseApi.update(editingId.value, payload)
    } else {
      await purchaseApi.create(payload)
    }
    closeModal()
    // 自动触发中奖检查
    await autoCheckWinning(payload.lottery_type, payload.issue_number)
    loadPurchases()
    showToast('success', isEdit ? '修改成功' : '新增成功')
  } catch (e) {
    console.error('保存失败', e)
    showToast('error', '保存失败，请重试')
  }
}

// 自动检查中奖（根据期号查找开奖记录并计算）
const autoCheckWinning = async (lotteryType: string, issueNumber: string) => {
  // 查找对应的开奖记录
  const draw = drawResults.value.find(
    d => d.lottery_type === lotteryType && d.issue_number === issueNumber
  )
  // 无论是否有对应开奖记录，都触发重新计算（可能有其他期号已开奖）
  try {
    await winningApi.recheck()
    console.log('自动重新计算中奖完成')
  } catch (e) {
    console.error('自动计算中奖失败', e)
  }
}

const deletePurchase = (id: number) => {
  deleteId.value = id
  deleteConfirm.value = true
}

const confirmDelete = async () => {
  if (!deleteId.value) return
  try {
    await purchaseApi.delete(deleteId.value)
    deleteConfirm.value = false
    deleteId.value = null
    loadPurchases()
    showToast('success', '删除成功')
  } catch (e) {
    console.error('删除失败', e)
    showToast('error', '删除失败，请重试')
  }
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

// 解析号码JSON
const parseNumbers = (json: string): any => {
  try { return JSON.parse(json) } catch { return {} }
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

// 获取购买记录对应的开奖号码
const getDrawNumbers = (purchase: PurchaseRecord): string | null => {
  const draw = drawResults.value.find(
    d => d.lottery_type === purchase.lottery_type && d.issue_number === purchase.issue_number
  )
  return draw?.numbers || null
}

// 获取购买记录对应的中奖奖金
const getWinningAmount = (purchase: PurchaseRecord): number => {
  const winning = winningRecords.value.find(
    w => w.purchase?.id === purchase.id
  )
  return winning?.prize_amount || 0
}

// 获取购买记录对应的中奖信息
const getWinningInfo = (purchase: PurchaseRecord) => {
  const winning = winningRecords.value.find(
    w => w.purchase?.id === purchase.id
  )
  return winning
}

// 生成高亮号码HTML
const highlightNumbers = (purchaseJson: string, drawJson: string | null, size: 'sm' | 'lg' = 'lg') => {
  const purchase = parseNumbers(purchaseJson)
  if (!drawJson) return formatNumbers(purchaseJson)

  const draw = parseNumbers(drawJson)
  // 收集开奖号码集合，分红蓝区
  const drawRedSet = new Set<number>()
  const drawBlueSet = new Set<number>()
  if (draw.red) draw.red.forEach((n: number) => drawRedSet.add(n))
  if (draw.blue) draw.blue.forEach((n: number) => drawBlueSet.add(n))
  if (draw.front) draw.front.forEach((n: number) => drawRedSet.add(n))
  if (draw.back) draw.back.forEach((n: number) => drawBlueSet.add(n))
  if (draw.main) draw.main.forEach((n: number) => drawRedSet.add(n))
  if (draw.numbers) draw.numbers.forEach((n: number) => drawRedSet.add(n))

  const sizeClass = size === 'sm' ? 'w-5 h-5 text-xs' : 'w-6 h-6 text-xs'
  const ball = (num: number, isBlue: boolean) => {
    const padded = String(num).padStart(2, '0')
    const hitSet = isBlue ? drawBlueSet : drawRedSet
    if (hitSet.has(num)) {
      return `<span class="inline-block bg-red-500 text-white rounded-full ${sizeClass} leading-[inherit] text-center font-bold mx-0.5">${padded}</span>`
    }
    if (isBlue) {
      return `<span class="inline-block bg-blue-500 text-white rounded-full ${sizeClass} leading-[inherit] text-center mx-0.5">${padded}</span>`
    }
    return `<span class="inline-block bg-slate-200 text-slate-600 rounded-full ${sizeClass} leading-[inherit] text-center mx-0.5">${padded}</span>`
  }

  if (purchase.red && purchase.blue) {
    const red = purchase.red.map((n: number) => ball(n, false)).join('')
    const blue = purchase.blue.map((n: number) => ball(n, true)).join('')
    return `${red} <span class="text-slate-300 mx-1">|</span> ${blue}`
  }
  if (purchase.front && purchase.back) {
    const front = purchase.front.map((n: number) => ball(n, false)).join('')
    const back = purchase.back.map((n: number) => ball(n, true)).join('')
    return `${front} <span class="text-slate-300 mx-1">|</span> ${back}`
  }
  if (purchase.main) return purchase.main.map((n: number) => ball(n, false)).join('')
  if (purchase.numbers) return purchase.numbers.map((n: number) => ball(n, false)).join('')
  return formatNumbers(purchaseJson)
}

const totalAmount = computed(() => {
  return purchases.value.reduce((sum, p) => sum + p.amount, 0)
})

const recheckWinnings = async () => {
  rechecking.value = true
  try {
    await winningApi.recheck()
    showToast('success', '正在重新计算中奖记录...')
    setTimeout(() => {
      loadPurchases()
    }, 2000)
  } catch (e) {
    console.error('重新检查失败', e)
    showToast('error', '重新检查失败')
  } finally {
    rechecking.value = false
  }
}
</script>

<template>
  <div class="animate-fade-in">
    <!-- Header -->
    <div class="mb-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-800">购买记录</h1>
          <p class="text-slate-400 text-sm mt-1">共 {{ purchases.length }} 条记录，累计投入 ¥{{ totalAmount.toFixed(2) }}</p>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="openLuckyModal"
            class="flex items-center gap-1.5 px-3 py-2 bg-amber-500 hover:bg-amber-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-amber-500/30 cursor-pointer"
          >
            <Sparkles class="w-4 h-4" />
            <span class="hidden sm:inline">幸运号</span>
          </button>
          <button
            @click="openModal()"
            class="flex items-center gap-1.5 px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer"
          >
            <Plus class="w-4 h-4" />
            <span class="hidden sm:inline">新增</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-6">
      <div class="relative">
        <select
          v-model="filterType"
          @change="loadPurchases"
          class="appearance-none px-4 py-2 pr-10 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
        >
          <option value="">全部类型</option>
          <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
        </select>
        <ChevronDown class="w-4 h-4 text-slate-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none" />
      </div>
      <div class="relative">
        <select
          v-model="filterStatus"
          @change="loadPurchases"
          class="appearance-none px-4 py-2 pr-10 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
        >
          <option value="">全部状态</option>
          <option value="待开奖">待开奖</option>
          <option value="已开奖">已开奖</option>
          <option value="已中奖">已中奖</option>
          <option value="未中奖">未中奖</option>
        </select>
        <ChevronDown class="w-4 h-4 text-slate-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none" />
      </div>
      <button
        @click="recheckWinnings"
        :disabled="rechecking"
        class="flex items-center gap-2 px-4 py-2 bg-amber-50 hover:bg-amber-100 text-amber-600 rounded-xl text-sm font-medium transition-colors cursor-pointer disabled:opacity-50"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': rechecking }" />
        重新检查中奖
      </button>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-2xl card-shadow overflow-hidden">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="purchases.length === 0" class="text-center py-16 text-slate-400">
        <Search class="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p>暂无购买记录</p>
      </div>

      <template v-else>
        <!-- 移动端：卡片列表 -->
        <div class="md:hidden p-4 space-y-3">
          <div v-for="item in purchases" :key="item.id" class="bg-slate-50 rounded-xl p-4">
            <div class="flex items-start justify-between mb-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium text-slate-800 text-sm">{{ item.lottery_type }}</span>
                <span class="text-slate-400 text-xs">期号 {{ item.issue_number }}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusColors[item.status]">
                  {{ item.status }}
                </span>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                  <Edit2 class="w-4 h-4" />
                </button>
                <button @click="deletePurchase(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div class="mb-2">
              <div v-if="getDrawNumbers(item)" class="space-y-1.5">
              <div class="flex items-center gap-1">
                <span class="text-xs text-slate-400 shrink-0 w-12">投注号码</span>
                <span v-if="hasRed(item.numbers)" v-for="(ball, idx) in getRedBalls(item.numbers)" :key="'r'+idx"
                  class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold"
                  :class="isHit(ball, item.numbers, getDrawNumbers(item)!) ? 'bg-red-500 text-white' : 'bg-slate-200 text-slate-600'">
                  {{ String(ball).padStart(2, '0') }}
                </span>
                <span v-if="hasBlue(item.numbers)" class="text-slate-300 w-6 text-center">|</span>
                <span v-if="hasBlue(item.numbers)" v-for="(ball, idx) in getBlueBalls(item.numbers)" :key="'b'+idx"
                  class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold"
                  :class="isHit(ball, item.numbers, getDrawNumbers(item)!, true) ? 'bg-blue-500 text-white' : 'bg-slate-200 text-slate-600'">
                  {{ String(ball).padStart(2, '0') }}
                </span>
              </div>
              <div class="flex items-center gap-1">
                <span class="text-xs text-slate-400 shrink-0 w-12">开奖号码</span>
                <template v-for="(ball, idx) in formatNumbers(getDrawNumbers(item)!)" :key="idx">
                  <span v-if="ball.num === '|'"
                    class="text-slate-300 w-6 text-center">|</span>
                  <span v-else
                    class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold"
                    :class="ball.type === 'blue'
                      ? 'bg-blue-500 text-white'
                      : 'bg-red-500 text-white'">
                    {{ ball.num }}
                  </span>
                </template>
              </div>
            </div>
            <div v-else class="flex flex-wrap gap-1">
                <span v-for="(ball, idx) in formatNumbers(item.numbers)" :key="idx"
                  class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold"
                  :class="ball.num === '|'
                    ? 'text-slate-300 text-xs bg-transparent'
                    : ball.type === 'blue'
                      ? 'bg-blue-500 text-white'
                      : 'bg-slate-200 text-slate-600'">
                  {{ ball.num }}
                </span>
              </div>
            </div>
            <div class="flex items-center gap-4 text-xs text-slate-500">
              <span>{{ item.bet_type }}</span>
              <span class="font-medium text-slate-700">¥{{ item.amount }}</span>
              <span v-if="getWinningAmount(item) > 0" class="font-bold text-emerald-600">中奖 ¥{{ getWinningAmount(item).toLocaleString() }}</span>
              <span>{{ item.purchase_date.split('T')[0] }}</span>
            </div>
          </div>
        </div>

        <!-- PC 端：表格 -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">彩票类型</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">期号</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">投注/开奖号码</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">投注方式</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">金额</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">状态</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">奖金</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">日期</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-slate-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="item in purchases" :key="item.id" class="hover:bg-slate-50">
                <td class="px-4 py-3 text-sm font-medium text-slate-700">{{ item.lottery_type }}</td>
                <td class="px-4 py-3 text-sm text-slate-600">{{ item.issue_number }}</td>
                <td class="px-4 py-3">
                  <div v-if="getDrawNumbers(item)" class="space-y-1">
                    <div class="flex items-center gap-1">
                      <span class="text-xs text-slate-400 shrink-0 w-10">投注</span>
                      <span v-if="hasRed(item.numbers)" v-for="(ball, idx) in getRedBalls(item.numbers)" :key="'r'+idx"
                        class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                        :class="isHit(ball, item.numbers, getDrawNumbers(item)!) ? 'bg-red-500 text-white' : 'bg-slate-200 text-slate-600'">
                        {{ String(ball).padStart(2, '0') }}
                      </span>
                      <span v-if="hasBlue(item.numbers)" class="text-slate-300 w-5 text-center">|</span>
                      <span v-if="hasBlue(item.numbers)" v-for="(ball, idx) in getBlueBalls(item.numbers)" :key="'b'+idx"
                        class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                        :class="isHit(ball, item.numbers, getDrawNumbers(item)!, true) ? 'bg-blue-500 text-white' : 'bg-slate-200 text-slate-600'">
                        {{ String(ball).padStart(2, '0') }}
                      </span>
                    </div>
                    <div class="flex items-center gap-1">
                      <span class="text-xs text-slate-400 shrink-0 w-10">开奖</span>
                      <template v-for="(ball, idx) in formatNumbers(getDrawNumbers(item)!)" :key="idx">
                        <span v-if="ball.num === '|'"
                          class="text-slate-300 w-5 text-center">|</span>
                        <span v-else
                          class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                          :class="ball.type === 'blue'
                            ? 'bg-blue-500 text-white'
                            : 'bg-red-500 text-white'">
                          {{ ball.num }}
                        </span>
                      </template>
                    </div>
                  </div>
                  <div v-else class="flex flex-wrap gap-0.5">
                    <span v-for="(ball, idx) in formatNumbers(item.numbers)" :key="idx"
                      class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                      :class="ball.num === '|'
                        ? 'text-slate-300 text-xs bg-transparent'
                        : ball.type === 'blue'
                          ? 'bg-blue-500 text-white'
                          : 'bg-slate-200 text-slate-600'">
                      {{ ball.num }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 text-sm text-slate-600">{{ item.bet_type }}</td>
                <td class="px-4 py-3 text-sm text-slate-600">¥{{ item.amount }}</td>
                <td class="px-4 py-3">
                  <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusColors[item.status]">
                    {{ item.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm">
                  <span v-if="getWinningAmount(item) > 0" class="font-bold text-emerald-600">¥{{ getWinningAmount(item).toLocaleString() }}</span>
                  <span v-else-if="getWinningInfo(item)" class="text-slate-400">-</span>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="px-4 py-3 text-sm text-slate-500">{{ item.purchase_date.split('T')[0] }}</td>
                <td class="px-4 py-3 text-right">
                  <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button @click="deletePurchase(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeModal"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto animate-slide-up">
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800">{{ editingId ? '编辑' : '新增' }}购买记录</h2>
          <button @click="closeModal" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">彩票类型</label>
            <select v-model="form.lottery_type" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">
                期号<span class="text-red-500 ml-0.5">*</span>
              </label>
              <input
                v-model="form.issue_number"
                placeholder="如：2024015"
                @input="errors.issue_number = ''"
                :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                  errors.issue_number ? 'border-red-400 focus:border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']"
              />
              <p v-if="errors.issue_number" class="mt-1 text-xs text-red-500 flex items-center gap-1">
                <span>⚠</span> {{ errors.issue_number }}
              </p>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">购买日期</label>
              <input v-model="form.purchase_date" type="date" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">
              号码<span class="text-red-500 ml-0.5">*</span>
            </label>
            <NumberInput v-model="form.numbers" :lottery-type="form.lottery_type" @update:modelValue="errors.numbers = ''" />
            <p v-if="errors.numbers" class="mt-1 text-xs text-red-500 flex items-center gap-1">
              <span>⚠</span> {{ errors.numbers }}
            </p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">投注方式</label>
              <select v-model="form.bet_type" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
                <option v-for="t in betTypes" :key="t" :value="t">{{ t }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">金额 (元)</label>
              <input v-model.number="form.amount" type="number" step="0.5" min="0" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">备注</label>
            <input v-model="form.remark" placeholder="可选" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
          </div>
        </div>

        <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex justify-end gap-3">
          <button @click="closeModal" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button @click="savePurchase" class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer">保存</button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="deleteConfirm" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="deleteConfirm = false; deleteId = null"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-sm animate-slide-up">
        <div class="p-6 text-center">
          <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Trash2 class="w-6 h-6 text-red-500" />
          </div>
          <h3 class="text-lg font-semibold text-slate-800 mb-2">确认删除</h3>
          <p class="text-slate-600 text-sm">确定要删除这条购买记录吗？删除后将无法恢复。</p>
        </div>
        <div class="border-t border-slate-100 px-6 py-4 flex gap-3">
          <button
            @click="deleteConfirm = false; deleteId = null"
            class="flex-1 px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer"
          >
            取消
          </button>
          <button
            @click="confirmDelete"
            class="flex-1 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-xl font-medium transition-colors cursor-pointer"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
    <Transition name="toast">
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
            @click="confirmAddFromLucky"
            class="flex-1 px-4 py-2.5 bg-amber-500 hover:bg-amber-600 text-white rounded-xl font-medium transition-colors cursor-pointer"
          >
            添加到购买记录
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
