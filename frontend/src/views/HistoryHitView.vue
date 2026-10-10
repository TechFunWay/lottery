<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { purchaseApi, drawApi } from '../api'
import type { PurchaseRecord, DrawResult } from '../types'
import { Search, Target, Calendar, ChevronDown, ChevronUp } from 'lucide-vue-next'

const purchases = ref<PurchaseRecord[]>([])
const drawResults = ref<DrawResult[]>([])
const loading = ref(false)
const filterType = ref('')
const expandedIds = ref<Set<number>>(new Set())

// 解析号码JSON
const parseNumbers = (json: string): any => {
  try {
    return JSON.parse(json)
  } catch {
    return {}
  }
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
  } catch {
    return []
  }
}

// 判断号码是否命中
const isHit = (num: number, purchaseJson: string, drawJson: string, isBlue: boolean = false): boolean => {
  const draw = parseNumbers(drawJson)
  const drawNums = isBlue
    ? (draw.blue || draw.back || [])
    : (draw.red || draw.front || draw.main || draw.numbers || [])
  return drawNums.includes(num)
}

// 格式化号码为带圆圈的样式（带命中标记）
interface NumberBallWithHit extends NumberBall {
  isHit: boolean
}

const formatNumbersWithHit = (purchaseJson: string, drawJson: string): NumberBallWithHit[] => {
  try {
    const purchase = parseNumbers(purchaseJson)
    const draw = parseNumbers(drawJson)

    const drawRedSet = new Set<number>()
    const drawBlueSet = new Set<number>()

    if (draw.red) draw.red.forEach((n: number) => drawRedSet.add(n))
    if (draw.blue) draw.blue.forEach((n: number) => drawBlueSet.add(n))
    if (draw.front) draw.front.forEach((n: number) => drawRedSet.add(n))
    if (draw.back) draw.back.forEach((n: number) => drawBlueSet.add(n))
    if (draw.main) draw.main.forEach((n: number) => drawRedSet.add(n))
    if (draw.numbers) draw.numbers.forEach((n: number) => drawRedSet.add(n))

    const balls: NumberBallWithHit[] = []

    if (purchase.red) {
      purchase.red.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red', isHit: drawRedSet.has(x) }))
      balls.push({ num: '|', type: 'main', isHit: false })
      purchase.blue.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue', isHit: drawBlueSet.has(x) }))
    } else if (purchase.front) {
      purchase.front.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red', isHit: drawRedSet.has(x) }))
      balls.push({ num: '|', type: 'main', isHit: false })
      purchase.back.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue', isHit: drawBlueSet.has(x) }))
    } else if (purchase.main) {
      purchase.main.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red', isHit: drawRedSet.has(x) }))
    } else if (purchase.numbers) {
      purchase.numbers.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main', isHit: drawRedSet.has(x) }))
    }

    return balls
  } catch {
    return []
  }
}

// 计算中奖金额（按照后端规则）
const calculatePrize = (purchaseJson: string, drawJson: string): number => {
  const purchase = parseNumbers(purchaseJson)
  const draw = parseNumbers(drawJson)

  const drawRedSet = new Set<number>()
  const drawBlueSet = new Set<number>()

  if (draw.red) draw.red.forEach((n: number) => drawRedSet.add(n))
  if (draw.blue) draw.blue.forEach((n: number) => drawBlueSet.add(n))
  if (draw.front) draw.front.forEach((n: number) => drawRedSet.add(n))
  if (draw.back) draw.back.forEach((n: number) => drawBlueSet.add(n))
  if (draw.main) draw.main.forEach((n: number) => drawRedSet.add(n))
  if (draw.special) draw.special.forEach((n: number) => drawBlueSet.add(n))
  if (draw.numbers) draw.numbers.forEach((n: number) => drawRedSet.add(n))

  let redHits = 0
  let blueHits = 0

  if (purchase.red) {
    // 双色球
    redHits = purchase.red.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.blue && purchase.blue.length > 0 && draw.blue && draw.blue.length > 0 && purchase.blue[0] === draw.blue[0] ? 1 : 0

    // 双色球中奖规则
    if (redHits === 6 && blueHits === 1) return 5000000  // 一等奖
    if (redHits === 6 && blueHits === 0) return 200000   // 二等奖
    if (redHits === 5 && blueHits === 1) return 3000     // 三等奖
    if (redHits === 5 && blueHits === 0) return 200      // 四等奖
    if (redHits === 4 && blueHits === 1) return 200      // 四等奖
    if (redHits === 4 && blueHits === 0) return 10       // 五等奖
    if (redHits === 3 && blueHits === 1) return 10       // 五等奖
    if (redHits === 2 && blueHits === 1) return 5        // 六等奖
    if (redHits === 1 && blueHits === 1) return 5        // 六等奖
    if (redHits === 0 && blueHits === 1) return 5        // 六等奖
  } else if (purchase.front) {
    // 大乐透
    redHits = purchase.front.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.back ? purchase.back.filter((n: number) => drawBlueSet.has(n)).length : 0

    // 大乐透中奖规则
    if (redHits === 5 && blueHits === 2) return 10000000 // 一等奖
    if (redHits === 5 && blueHits === 1) return 500000   // 二等奖
    if (redHits === 5 && blueHits === 0) return 10000    // 三等奖
    if (redHits === 4 && blueHits === 2) return 3000     // 四等奖
    if (redHits === 4 && blueHits === 1) return 300      // 五等奖
    if (redHits === 3 && blueHits === 2) return 300      // 五等奖
    if (redHits === 4 && blueHits === 0) return 100      // 六等奖
    if (redHits === 3 && blueHits === 1) return 100      // 六等奖
    if (redHits === 2 && blueHits === 2) return 100      // 六等奖
    if (redHits === 3 && blueHits === 0) return 15       // 七等奖
    if (redHits === 2 && blueHits === 1) return 15       // 七等奖
    if (redHits === 1 && blueHits === 2) return 15       // 七等奖
    if (redHits === 0 && blueHits === 2) return 15       // 七等奖
  } else if (purchase.main) {
    // 七乐彩
    redHits = purchase.main.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.special && purchase.special.length > 0 && draw.special && draw.special.length > 0 && purchase.special[0] === draw.special[0] ? 1 : 0

    // 七乐彩中奖规则
    if (redHits === 7 && blueHits === 0) return 5000000  // 一等奖
    if (redHits === 6 && blueHits === 1) return 10000    // 二等奖
    if (redHits === 6 && blueHits === 0) return 1000     // 三等奖
    if (redHits === 5 && blueHits === 1) return 100      // 四等奖
    if (redHits === 5 && blueHits === 0) return 100      // 四等奖
    if (redHits === 4 && blueHits === 1) return 30       // 五等奖
    if (redHits === 4 && blueHits === 0) return 30       // 五等奖
    if (redHits === 3 && blueHits === 1) return 10       // 六等奖
    if (redHits === 7 && blueHits === 0) return 5        // 七等奖（仅特别号）
  } else if (purchase.numbers) {
    // 福彩3D、排列3、排列5等 - 需要考虑直选/组选，这里简化处理
    // 实际上需要根据 bet_type 来计算，这里只能做简化
    // 3D/排列3：直选1040/1000，组选173/167
    // 排列5：直选100000
    // 这个简化版本可能不准确，但历史命中页面主要用于参考
    if (redHits >= 3) return 1000  // 简化处理
    if (redHits >= 2) return 100
    if (redHits >= 1) return 10
  }

  return 0
}

// 计算奖项等级
const getPrizeLevel = (purchaseJson: string, drawJson: string): { level: number; name: string } => {
  const purchase = parseNumbers(purchaseJson)
  const draw = parseNumbers(drawJson)

  const drawRedSet = new Set<number>()
  const drawBlueSet = new Set<number>()

  if (draw.red) draw.red.forEach((n: number) => drawRedSet.add(n))
  if (draw.blue) draw.blue.forEach((n: number) => drawBlueSet.add(n))
  if (draw.front) draw.front.forEach((n: number) => drawRedSet.add(n))
  if (draw.back) draw.back.forEach((n: number) => drawBlueSet.add(n))
  if (draw.main) draw.main.forEach((n: number) => drawRedSet.add(n))
  if (draw.special) draw.special.forEach((n: number) => drawBlueSet.add(n))
  if (draw.numbers) draw.numbers.forEach((n: number) => drawRedSet.add(n))

  let redHits = 0
  let blueHits = 0

  if (purchase.red) {
    // 双色球
    redHits = purchase.red.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.blue && purchase.blue.length > 0 && draw.blue && draw.blue.length > 0 && purchase.blue[0] === draw.blue[0] ? 1 : 0

    if (redHits === 6 && blueHits === 1) return { level: 1, name: '一等奖' }
    if (redHits === 6 && blueHits === 0) return { level: 2, name: '二等奖' }
    if (redHits === 5 && blueHits === 1) return { level: 3, name: '三等奖' }
    if (redHits === 5 && blueHits === 0) return { level: 4, name: '四等奖' }
    if (redHits === 4 && blueHits === 1) return { level: 4, name: '四等奖' }
    if (redHits === 4 && blueHits === 0) return { level: 5, name: '五等奖' }
    if (redHits === 3 && blueHits === 1) return { level: 5, name: '五等奖' }
    if (redHits === 2 && blueHits === 1) return { level: 6, name: '六等奖' }
    if (redHits === 1 && blueHits === 1) return { level: 6, name: '六等奖' }
    if (redHits === 0 && blueHits === 1) return { level: 6, name: '六等奖' }
  } else if (purchase.front) {
    // 大乐透
    redHits = purchase.front.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.back ? purchase.back.filter((n: number) => drawBlueSet.has(n)).length : 0

    if (redHits === 5 && blueHits === 2) return { level: 1, name: '一等奖' }
    if (redHits === 5 && blueHits === 1) return { level: 2, name: '二等奖' }
    if (redHits === 5 && blueHits === 0) return { level: 3, name: '三等奖' }
    if (redHits === 4 && blueHits === 2) return { level: 4, name: '四等奖' }
    if (redHits === 4 && blueHits === 1) return { level: 5, name: '五等奖' }
    if (redHits === 3 && blueHits === 2) return { level: 5, name: '五等奖' }
    if (redHits === 4 && blueHits === 0) return { level: 6, name: '六等奖' }
    if (redHits === 3 && blueHits === 1) return { level: 6, name: '六等奖' }
    if (redHits === 2 && blueHits === 2) return { level: 6, name: '六等奖' }
    if (redHits === 3 && blueHits === 0) return { level: 7, name: '七等奖' }
    if (redHits === 2 && blueHits === 1) return { level: 7, name: '七等奖' }
    if (redHits === 1 && blueHits === 2) return { level: 7, name: '七等奖' }
    if (redHits === 0 && blueHits === 2) return { level: 7, name: '七等奖' }
  } else if (purchase.main) {
    // 七乐彩
    redHits = purchase.main.filter((n: number) => drawRedSet.has(n)).length
    blueHits = purchase.special && purchase.special.length > 0 && draw.special && draw.special.length > 0 && purchase.special[0] === draw.special[0] ? 1 : 0

    if (redHits === 7 && blueHits === 0) return { level: 1, name: '一等奖' }
    if (redHits === 6 && blueHits === 1) return { level: 2, name: '二等奖' }
    if (redHits === 6 && blueHits === 0) return { level: 3, name: '三等奖' }
    if (redHits === 5 && blueHits === 1) return { level: 4, name: '四等奖' }
    if (redHits === 5 && blueHits === 0) return { level: 4, name: '四等奖' }
    if (redHits === 4 && blueHits === 1) return { level: 5, name: '五等奖' }
    if (redHits === 4 && blueHits === 0) return { level: 5, name: '五等奖' }
    if (redHits === 3 && blueHits === 1) return { level: 6, name: '六等奖' }
    if (redHits === 7 && blueHits === 0) return { level: 7, name: '七等奖' }
  }

  return { level: 0, name: '未中奖' }
}

// 获取奖项样式类
const getPrizeLevelClass = (lotteryType: string, prize: number): string => {
  if (prize <= 0) return 'bg-slate-100 text-slate-500'
  if (prize >= 5000000) return 'bg-amber-100 text-amber-700 border border-amber-300'  // 一等奖
  if (prize >= 1000000) return 'bg-slate-100 text-slate-700 border border-slate-300'  // 二等奖
  if (prize >= 10000) return 'bg-orange-100 text-orange-700 border border-orange-300'  // 三等奖
  if (prize >= 1000) return 'bg-blue-100 text-blue-700 border border-blue-300'  // 四等奖
  if (prize >= 100) return 'bg-emerald-100 text-emerald-700 border border-emerald-300'  // 五等奖/六等奖
  return 'bg-indigo-100 text-indigo-700 border border-indigo-300'  // 七等奖
}

// 命中结果详情
interface HitDetail {
  draw: DrawResult
  redHits: number
  blueHits: number
  prize: number
  hitRedBalls: number[]
  hitBlueBalls: number[]
}

// 购买记录的命中历史
interface PurchaseHitHistory {
  purchase: PurchaseRecord
  hitDetails: HitDetail[]
  totalPrize: number
}

const loadData = async () => {
  loading.value = true
  // 加载购买记录（获取全部）
  const purchaseRes = await purchaseApi.list({ size: 100 }).catch(() => null)
  if (purchaseRes) purchases.value = purchaseRes.data || []

  // 加载开奖记录
  const drawRes = await drawApi.list({ size: 500 }).catch(() => null)
  if (drawRes) drawResults.value = drawRes.data || []

  loading.value = false
}

onMounted(loadData)

// 计算每笔购买记录的命中历史
const purchaseHitHistory = computed<PurchaseHitHistory[]>(() => {
  const result: PurchaseHitHistory[] = []

  const filteredPurchases = filterType.value
    ? purchases.value.filter(p => p.lottery_type === filterType.value)
    : purchases.value

  for (const purchase of filteredPurchases) {
    const purchaseDate = purchase.purchase_date.split('T')[0]

    // 找出所有开奖记录（同一彩票类型），包含购买日期之前和之后的
    // 这样可以看到这个号码在历史上所有期号的命中情况
    const relevantDraws = drawResults.value
      .filter(d => d.lottery_type === purchase.lottery_type)
      .sort((a, b) => b.draw_date.localeCompare(a.draw_date))

    const hitDetails: HitDetail[] = []

    for (const draw of relevantDraws) {
      const purchaseNums = parseNumbers(purchase.numbers)
      const drawNums = parseNumbers(draw.numbers)

      const drawRedSet = new Set<number>()
      const drawBlueSet = new Set<number>()

      if (drawNums.red) drawNums.red.forEach((n: number) => drawRedSet.add(n))
      if (drawNums.blue) drawNums.blue.forEach((n: number) => drawBlueSet.add(n))
      if (drawNums.front) drawNums.front.forEach((n: number) => drawRedSet.add(n))
      if (drawNums.back) drawNums.back.forEach((n: number) => drawBlueSet.add(n))
      if (drawNums.main) drawNums.main.forEach((n: number) => drawRedSet.add(n))
      if (drawNums.numbers) drawNums.numbers.forEach((n: number) => drawRedSet.add(n))

      let hitRedBalls: number[] = []
      let hitBlueBalls: number[] = []

      if (purchaseNums.red) {
        hitRedBalls = purchaseNums.red.filter((n: number) => drawRedSet.has(n))
        hitBlueBalls = purchaseNums.blue ? purchaseNums.blue.filter((n: number) => drawBlueSet.has(n)) : []
      } else if (purchaseNums.front) {
        hitRedBalls = purchaseNums.front.filter((n: number) => drawRedSet.has(n))
        hitBlueBalls = purchaseNums.back ? purchaseNums.back.filter((n: number) => drawBlueSet.has(n)) : []
      } else if (purchaseNums.main) {
        hitRedBalls = purchaseNums.main.filter((n: number) => drawRedSet.has(n))
      } else if (purchaseNums.numbers) {
        hitRedBalls = purchaseNums.numbers.filter((n: number) => drawRedSet.has(n))
      }

      const redHits = hitRedBalls.length
      const blueHits = hitBlueBalls.length
      const prize = calculatePrize(purchase.numbers, draw.numbers)

      // 只有中奖（有奖金）的才记录
      if (prize > 0) {
        hitDetails.push({
          draw,
          redHits,
          blueHits,
          prize,
          hitRedBalls,
          hitBlueBalls
        })
      }
    }

    const totalPrize = hitDetails.reduce((sum, h) => sum + h.prize, 0)

    // 有中奖记录就展示
    if (hitDetails.length > 0) {
      result.push({
        purchase,
        hitDetails,
        totalPrize
      })
    }
  }

  return result
})

// 统计信息 - 基于已过滤的数据
const stats = computed(() => {
  const history = purchaseHitHistory.value
  const totalHits = history.reduce((sum, h) => sum + h.hitDetails.length, 0)
  const totalPrize = history.reduce((sum, h) => sum + h.totalPrize, 0)
  const winCount = history.filter(h => h.totalPrize > 0).length
  return {
    totalPurchases: purchases.value.length,
    hitCount: history.length,
    totalHits,
    totalPrize,
    winCount
  }
})

const lotteryTypes = ['双色球', '大乐透', '福彩3D', '排列3', '排列5', '七乐彩']

// 展开/收起切换
const toggleExpand = (id: number) => {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}
</script>

<template>
  <div class="animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 flex items-center gap-2">
          <Target class="w-6 h-6 text-blue-500" />
          历史命中
        </h1>
        <p class="text-slate-400 text-sm mt-1">追踪每笔购买号码在所有历史开奖中的命中情况（假如一直买这个号码，后面的中奖情况）</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-6">
      <div class="relative">
        <select
          v-model="filterType"
          class="appearance-none px-4 py-2 pr-10 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
        >
          <option value="">全部类型</option>
          <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
        </select>
        <ChevronDown class="w-4 h-4 text-slate-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none" />
      </div>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-xl p-4 card-shadow">
        <div class="text-xs text-slate-400 mb-1">购买记录</div>
        <div class="text-2xl font-bold text-slate-800">{{ stats.totalPurchases }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 card-shadow">
        <div class="text-xs text-slate-400 mb-1">有中奖记录</div>
        <div class="text-2xl font-bold text-blue-600">{{ stats.hitCount }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 card-shadow">
        <div class="text-xs text-slate-400 mb-1">实际中奖</div>
        <div class="text-2xl font-bold text-emerald-600">{{ stats.winCount }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 card-shadow">
        <div class="text-xs text-slate-400 mb-1">累计奖金</div>
        <div class="text-2xl font-bold text-amber-600">¥{{ stats.totalPrize.toLocaleString() }}</div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="purchaseHitHistory.length === 0" class="text-center py-16 text-slate-400">
      <Search class="w-12 h-12 mx-auto mb-3 opacity-30" />
      <p>暂无中奖记录</p>
    </div>

    <!-- List -->
    <div v-else class="space-y-3">
      <div
        v-for="item in purchaseHitHistory"
        :key="item.purchase.id"
        class="bg-white rounded-2xl p-4 card-shadow"
      >
        <!-- 默认显示：期号、投注号码、中奖次数、奖金 -->
        <div
          class="flex items-center justify-between cursor-pointer"
          @click="toggleExpand(item.purchase.id)"
        >
          <div class="flex items-center gap-3 flex-1 min-w-0">
            <!-- 期号 -->
            <div class="shrink-0">
              <div class="text-xs text-slate-400 mb-0.5">期号</div>
              <div class="text-sm font-medium text-slate-700">{{ item.purchase.issue_number }}</div>
            </div>

            <!-- 投注号码 -->
            <div class="flex items-center gap-1 min-w-0">
              <span v-if="hasRed(item.purchase.numbers)" v-for="(ball, idx) in getRedBalls(item.purchase.numbers)" :key="'r'+idx"
                class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold bg-slate-200 text-slate-600 shrink-0">
                {{ String(ball).padStart(2, '0') }}
              </span>
              <span v-if="hasBlue(item.purchase.numbers)" class="text-slate-300 w-5 text-center shrink-0">|</span>
              <span v-if="hasBlue(item.purchase.numbers)" v-for="(ball, idx) in getBlueBalls(item.purchase.numbers)" :key="'b'+idx"
                class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold bg-blue-500 text-white shrink-0">
                {{ String(ball).padStart(2, '0') }}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-3 shrink-0">
            <div class="text-right">
              <div class="text-xs text-slate-400">中奖 {{ item.hitDetails.length }} 期</div>
              <div class="text-lg font-bold text-amber-600">¥{{ item.totalPrize.toLocaleString() }}</div>
            </div>
            <ChevronDown
              class="w-5 h-5 text-slate-400 transition-transform"
              :class="{ 'rotate-180': expandedIds.has(item.purchase.id) }"
            />
          </div>
        </div>

        <!-- 展开详情 -->
        <div v-if="expandedIds.has(item.purchase.id)" class="border-t border-slate-100 mt-3 pt-3 space-y-3">
          <div v-for="hit in item.hitDetails" :key="hit.draw.id" class="bg-slate-50 rounded-xl p-3">
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-slate-700">第 {{ hit.draw.issue_number }} 期</span>
                <span class="text-xs text-slate-400">{{ hit.draw.draw_date.split('T')[0] }}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                  :class="getPrizeLevelClass(item.purchase.lottery_type, hit.prize)">
                  {{ getPrizeLevel(item.purchase.numbers, hit.draw.numbers).name }}
                </span>
              </div>
              <span v-if="hit.prize > 0" class="text-sm font-bold text-emerald-600">¥{{ hit.prize.toLocaleString() }}</span>
            </div>

            <!-- 开奖号码 -->
            <div class="flex items-center gap-2 mb-2">
              <span class="text-xs text-slate-400 shrink-0 w-10">开奖</span>
              <div class="flex flex-wrap items-center gap-1">
                <template v-for="(ball, idx) in formatNumbers(hit.draw.numbers)" :key="idx">
                  <span v-if="ball.num === '|'" class="text-slate-300 w-5 text-center">|</span>
                  <span v-else
                    class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                    :class="ball.type === 'blue' ? 'bg-blue-500 text-white' : 'bg-red-500 text-white'">
                    {{ ball.num }}
                  </span>
                </template>
              </div>
            </div>

            <!-- 投注号码（带命中标记） -->
            <div class="flex items-center gap-2">
              <span class="text-xs text-slate-400 shrink-0 w-10">投注</span>
              <div class="flex flex-wrap items-center gap-1">
                <template v-for="(ball, idx) in formatNumbersWithHit(item.purchase.numbers, hit.draw.numbers)" :key="idx">
                  <span v-if="ball.num === '|'" class="text-slate-300 w-5 text-center">|</span>
                  <span v-else
                    class="inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold"
                    :class="{
                      'bg-red-500 text-white': ball.isHit && ball.type === 'red',
                      'bg-blue-500 text-white': ball.isHit && ball.type === 'blue',
                      'bg-slate-200 text-slate-500': !ball.isHit
                    }">
                    {{ ball.num }}
                  </span>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
