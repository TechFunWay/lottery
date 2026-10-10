<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { LotteryType } from '../types'

const props = defineProps<{
  modelValue: string
  lotteryType: LotteryType
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'betCountChange': [count: number]
  'betTypeChange': [betType: string]
}>()

const digits = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

const digitCount = computed(() => (props.lotteryType === '排列5' ? 5 : 3))
const availablePlays = computed(() =>
  props.lotteryType === '排列5' ? ['直选'] : ['直选', '组选3', '组选6', '定位胆']
)
const positionLabels = computed(() =>
  digitCount.value === 5 ? ['万位', '千位', '百位', '十位', '个位'] : ['百位', '十位', '个位']
)

const play = ref('直选')
const positions = ref<number[][]>([]) // 直选 / 定位胆：每位候选数字
const group = ref<number[]>([])       // 组选：数字集合

const isGroupPlay = computed(() => play.value === '组选3' || play.value === '组选6')

const initState = () => {
  positions.value = Array.from({ length: digitCount.value }, () => [])
  group.value = []
  if (!availablePlays.value.includes(play.value)) play.value = '直选'
}

let internalUpdate = false

watch(() => props.lotteryType, () => {
  initState()
}, { immediate: true })

// 编辑/复制回填
watch(() => props.modelValue, (val) => {
  if (internalUpdate) return
  if (!val || val === '' || val === '[]' || val === '{}') return
  try {
    const parsed = JSON.parse(val)
    if (Array.isArray(parsed.positions)) {
      play.value = parsed.play || '直选'
      positions.value = Array.from({ length: digitCount.value }, (_, i) =>
        Array.isArray(parsed.positions[i]) ? [...parsed.positions[i]] : []
      )
    } else if (Array.isArray(parsed.group)) {
      play.value = parsed.play || '组选6'
      group.value = [...parsed.group]
    } else if (Array.isArray(parsed.numbers)) {
      // 旧格式 { numbers, bet_type }
      const bt = parsed.bet_type || '直选'
      play.value = bt
      if (bt === '组选3' || bt === '组选6') {
        group.value = [...parsed.numbers]
      } else {
        positions.value = Array.from({ length: digitCount.value }, (_, i) =>
          parsed.numbers[i] != null ? [parsed.numbers[i]] : []
        )
      }
    } else if (Array.isArray(parsed)) {
      play.value = '直选'
      positions.value = Array.from({ length: digitCount.value }, (_, i) =>
        parsed[i] != null ? [parsed[i]] : []
      )
    }
  } catch {
    // ignore
  }
}, { immediate: true })

const comb = (n: number, k: number): number => {
  if (n < k || k < 0) return 0
  let r = 1
  for (let i = 0; i < k; i++) r = (r * (n - i)) / (i + 1)
  return Math.round(r)
}

// 注数
const betCount = computed(() => {
  if (isGroupPlay.value) {
    const n = group.value.length
    if (play.value === '组选6') return n >= 3 ? comb(n, 3) : 0
    return n >= 2 ? n * (n - 1) : 0 // 组选3
  }
  if (play.value === '定位胆') {
    return positions.value.reduce((s, p) => s + p.length, 0)
  }
  // 直选：各位候选数乘积
  let prod = 1
  for (const p of positions.value) {
    if (p.length === 0) return 0
    prod *= p.length
  }
  return prod
})

// 用于显示的投注方式标签
const betTypeLabel = computed(() => {
  if (play.value === '直选') {
    return positions.value.some(p => p.length > 1) ? '直选复式' : '直选'
  }
  return play.value
})

const updateValue = () => {
  let result: any
  if (isGroupPlay.value) {
    result = { play: play.value, group: [...group.value].sort((a, b) => a - b) }
  } else {
    result = { play: play.value, positions: positions.value.map(p => [...p].sort((a, b) => a - b)) }
  }
  internalUpdate = true
  emit('update:modelValue', JSON.stringify(result))
  emit('betTypeChange', betTypeLabel.value)
  setTimeout(() => { internalUpdate = false }, 0)
}

watch(betCount, (n) => emit('betCountChange', n))
watch(play, () => updateValue())

const togglePosition = (pos: number, digit: number) => {
  const arr = positions.value[pos]
  const idx = arr.indexOf(digit)
  if (idx === -1) arr.push(digit)
  else arr.splice(idx, 1)
  updateValue()
}

const toggleGroup = (digit: number) => {
  const idx = group.value.indexOf(digit)
  if (idx === -1) group.value.push(digit)
  else group.value.splice(idx, 1)
  updateValue()
}

const isPosOn = (pos: number, d: number) => positions.value[pos]?.includes(d) ?? false
const isGroupOn = (d: number) => group.value.includes(d)

const hasAnyValue = computed(() =>
  group.value.length > 0 || positions.value.some(p => p.length > 0)
)

const clearAll = () => {
  positions.value = positions.value.map(() => [])
  group.value = []
  updateValue()
}
</script>

<template>
  <div class="space-y-4">
    <!-- 玩法切换 + 清空 -->
    <div class="flex items-center justify-between gap-2 flex-wrap">
      <div v-if="availablePlays.length > 1" class="flex gap-1.5 flex-wrap">
        <button
          v-for="p in availablePlays"
          :key="p"
          type="button"
          @click="play = p"
          :class="play === p ? 'bg-blue-500 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          class="px-3 py-1.5 rounded-lg text-sm font-medium transition-all cursor-pointer"
        >
          {{ p }}
        </button>
      </div>
      <button
        v-if="hasAnyValue"
        type="button"
        @click="clearAll"
        class="text-xs text-slate-400 hover:text-red-500 transition-colors cursor-pointer ml-auto"
      >
        清空号码
      </button>
    </div>

    <!-- 定位胆说明 -->
    <p v-if="play === '定位胆'" class="text-xs text-slate-400">
      选择要投注的位置和号码，未选的位置表示不投；每位每号 2 元一注。
    </p>

    <!-- 直选 / 定位胆：按位选号 -->
    <div v-if="!isGroupPlay" class="space-y-3">
      <div v-for="(label, pos) in positionLabels" :key="pos">
        <div class="flex items-center gap-2 mb-1.5">
          <span class="text-sm font-medium text-slate-600 w-12 shrink-0">{{ label }}</span>
          <span v-if="positions[pos]?.length" class="text-xs text-slate-400">
            已选 {{ [...positions[pos]].sort((a, b) => a - b).join(' ') }}
          </span>
        </div>
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="d in digits"
            :key="d"
            type="button"
            @click="togglePosition(pos, d)"
            :class="isPosOn(pos, d)
              ? 'bg-red-500 text-white border-red-500'
              : 'bg-white text-slate-600 border-slate-200 hover:border-red-300'"
            class="w-9 h-9 rounded-lg border-2 text-sm font-bold transition-all cursor-pointer"
          >
            {{ d }}
          </button>
        </div>
      </div>
    </div>

    <!-- 组选：选号集合 -->
    <div v-else>
      <div class="flex items-center gap-2 mb-1.5">
        <span class="text-sm font-medium text-slate-600">选号</span>
        <span class="text-xs text-slate-400">
          {{ play === '组选6' ? '至少选 3 个不同号' : '至少选 2 个不同号' }}
        </span>
      </div>
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="d in digits"
          :key="d"
          type="button"
          @click="toggleGroup(d)"
          :class="isGroupOn(d)
            ? 'bg-red-500 text-white border-red-500'
            : 'bg-white text-slate-600 border-slate-200 hover:border-red-300'"
          class="w-9 h-9 rounded-lg border-2 text-sm font-bold transition-all cursor-pointer"
        >
          {{ d }}
        </button>
      </div>
    </div>

    <!-- 注数信息 -->
    <div class="flex items-center gap-3 px-4 py-2 bg-amber-50 border border-amber-200 rounded-xl">
      <span class="text-amber-600 text-sm font-medium">{{ betTypeLabel }}</span>
      <span class="text-amber-500 text-xs">|</span>
      <span class="text-amber-700 text-sm">共 <strong>{{ betCount }}</strong> 注</span>
      <span class="text-amber-500 text-xs">金额 ¥{{ betCount * 2 }}</span>
      <span v-if="betCount === 0 && hasAnyValue" class="text-red-400 text-xs ml-auto">请完成选号</span>
    </div>
  </div>
</template>
