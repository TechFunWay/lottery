<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { LotteryType } from '../types'
import { LOTTERY_CONFIGS } from '../types'

const props = defineProps<{
  modelValue: string
  lotteryType: LotteryType
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'betCountChange': [count: number]
}>()

const config = computed(() => LOTTERY_CONFIGS.find(c => c.type === props.lotteryType))

// 主号码数组（支持动态数量）
const mainNumbers = ref<(number | null)[]>([])
const blueNumbers = ref<(number | null)[]>([])

// 错误提示
const errors = ref<string[]>([])

// 初始化号码数组
const initNumbers = () => {
  if (!config.value) return
  const count = config.value.redRange.count
  mainNumbers.value = Array(count).fill(null)
  if (config.value.blueRange) {
    blueNumbers.value = Array(config.value.blueRange.count).fill(null)
  } else {
    blueNumbers.value = []
  }
}

watch(() => props.lotteryType, () => {
  initNumbers()
}, { immediate: true })

// 标记是否是组件内部触发的更新，避免 watch 和 handleInput 互相覆盖
let internalUpdate = false

// 监听外部值变化（编辑回填时触发）
watch(() => props.modelValue, (val) => {
  if (internalUpdate) return
  if (!val || !config.value) return
  if (val === '' || val === '[]') return
  try {
    const parsed = JSON.parse(val)

    if (parsed.red !== undefined) {
      mainNumbers.value = parsed.red.map((n: number) => n)
      blueNumbers.value = parsed.blue ? parsed.blue.map((n: number) => n) : []
    } else if (parsed.front !== undefined) {
      mainNumbers.value = parsed.front.map((n: number) => n)
      blueNumbers.value = parsed.back ? parsed.back.map((n: number) => n) : []
    } else if (parsed.main !== undefined) {
      mainNumbers.value = parsed.main.map((n: number) => n)
      blueNumbers.value = parsed.special ? parsed.special.map((n: number) => n) : []
    } else if (parsed.numbers !== undefined) {
      mainNumbers.value = parsed.numbers.map((n: number) => n)
      blueNumbers.value = []
    } else if (Array.isArray(parsed)) {
      mainNumbers.value = parsed.map((n: number) => n)
      blueNumbers.value = []
    }
  } catch (e) {
    // ignore
  }
}, { immediate: true })

// 计算组合数 C(n, k)
const comb = (n: number, k: number): number => {
  if (n < k || k < 0) return 0
  if (k === 0 || k === n) return 1
  let res = 1
  for (let i = 0; i < k; i++) {
    res = res * (n - i) / (i + 1)
  }
  return Math.round(res)
}

// 计算当前注数
const betCount = computed(() => {
  if (!config.value) return 0

  const mainFilled = mainNumbers.value.filter(n => n !== null).length
  const blueFilled = blueNumbers.value.filter(n => n !== null).length

  const mainNeed = config.value.redRange.count
  const blueNeed = config.value.blueRange?.count ?? 0

  // 只有双色球和大乐透支持复式
  const supportsMultiple = props.lotteryType === '双色球' || props.lotteryType === '大乐透'
  if (!supportsMultiple) {
    return (mainFilled === mainNeed && (blueNeed === 0 || blueFilled === blueNeed)) ? 1 : 0
  }

  if (mainFilled < mainNeed) return 0
  if (blueNeed > 0 && blueFilled < blueNeed) return 0

  const mainCombs = comb(mainFilled, mainNeed)
  const blueCombs = blueNeed > 0 ? comb(blueFilled, blueNeed) : 1

  return mainCombs * blueCombs
})

// 是否复式投注
const isMultiple = computed(() => {
  if (!config.value) return false
  const supportsMultiple = props.lotteryType === '双色球' || props.lotteryType === '大乐透'
  if (!supportsMultiple) return false

  const mainFilled = mainNumbers.value.filter(n => n !== null).length
  const blueFilled = blueNumbers.value.filter(n => n !== null).length

  return mainFilled > config.value.redRange.count || blueFilled > (config.value.blueRange?.count ?? 0)
})

watch(betCount, (count) => {
  emit('betCountChange', count)
})

const updateValue = () => {
  if (!config.value) return
  const mains = mainNumbers.value.filter(n => n !== null) as number[]
  const blues = blueNumbers.value.filter(n => n !== null) as number[]

  let result: any
  const lt = props.lotteryType
  if (lt === '双色球') {
    result = { red: mains, blue: blues }
  } else if (lt === '大乐透') {
    result = { front: mains, back: blues }
  } else if (lt === '福彩3D') {
    result = { numbers: mains, bet_type: '直选' }
  } else if (lt === '七乐彩') {
    result = { main: mains, special: blues }
  } else {
    result = mains
  }

  internalUpdate = true
  emit('update:modelValue', JSON.stringify(result))
  setTimeout(() => { internalUpdate = false }, 0)
}

const handleInput = (index: number, type: 'main' | 'blue', event: Event) => {
  const input = event.target as HTMLInputElement
  const rawVal = input.value.trim()

  errors.value = []

  if (rawVal === '') {
    if (type === 'main') {
      mainNumbers.value[index] = null
    } else {
      blueNumbers.value[index] = null
    }
    updateValue()
    return
  }

  const val = parseInt(rawVal)
  if (isNaN(val)) {
    errors.value = ['请输入有效数字']
    return
  }

  // 范围校验
  const min = type === 'main' ? (config.value?.redRange.min ?? 0) : (config.value?.blueRange?.min ?? 0)
  const max = type === 'main' ? (config.value?.redRange.max ?? 99) : (config.value?.blueRange?.max ?? 99)

  if (val < min || val > max) {
    errors.value = [`${val} 超出范围 ${min}-${max}`]
    return
  }

  // 检查重复号码（福彩3D、排列3、排列5、七星彩允许重复）
  const allowDuplicate = props.lotteryType === '福彩3D' || props.lotteryType === '排列3' || props.lotteryType === '排列5' || props.lotteryType === '七星彩'
  if (!allowDuplicate) {
    const currentArray = type === 'main' ? mainNumbers.value : blueNumbers.value
    const existingIndex = currentArray.findIndex((n, i) => i !== index && n === val)
    if (existingIndex !== -1) {
      errors.value = [`号码 ${val} 已存在`]
      return
    }
  }

  if (type === 'main') {
    mainNumbers.value[index] = val
  } else {
    blueNumbers.value[index] = val
  }
  updateValue()
}

const handleBlur = (index: number, type: 'main' | 'blue', event: Event) => {
  const input = event.target as HTMLInputElement
  const rawVal = input.value.trim()

  if (rawVal === '') {
    if (type === 'main') {
      mainNumbers.value[index] = null
    } else {
      blueNumbers.value[index] = null
    }
    updateValue()
    return
  }

  const val = parseInt(rawVal)
  if (isNaN(val)) {
    errors.value = ['请输入有效数字']
    input.value = ''
    if (type === 'main') {
      mainNumbers.value[index] = null
    } else {
      blueNumbers.value[index] = null
    }
    updateValue()
    return
  }

  const min = type === 'main' ? (config.value?.redRange.min ?? 0) : (config.value?.blueRange?.min ?? 0)
  const max = type === 'main' ? (config.value?.redRange.max ?? 99) : (config.value?.blueRange?.max ?? 99)

  if (val < min || val > max) {
    errors.value = [`${val} 超出范围 ${min}-${max}`]
    input.value = ''
    if (type === 'main') {
      mainNumbers.value[index] = null
    } else {
      blueNumbers.value[index] = null
    }
    updateValue()
    return
  }

  // 检查重复号码（福彩3D、排列3、排列5、七星彩允许重复）
  const allowDuplicate = props.lotteryType === '福彩3D' || props.lotteryType === '排列3' || props.lotteryType === '排列5' || props.lotteryType === '七星彩'
  if (!allowDuplicate) {
    const currentArray = type === 'main' ? mainNumbers.value : blueNumbers.value
    const existingIndex = currentArray.findIndex((n, i) => i !== index && n === val)
    if (existingIndex !== -1) {
      errors.value = [`号码 ${val} 已存在`]
      input.value = ''
      if (type === 'main') {
        mainNumbers.value[index] = null
      } else {
        blueNumbers.value[index] = null
      }
      updateValue()
      return
    }
  }

  // 有效值，自动排序（福彩3D、排列3、排列5、七星彩不排序）
  const noSort = props.lotteryType === '福彩3D' || props.lotteryType === '排列3' || props.lotteryType === '排列5' || props.lotteryType === '七星彩'
  if (!noSort) {
    if (type === 'main') {
      const filled = mainNumbers.value.filter(n => n !== null).sort((a, b) => a! - b!)
      // 保持数组长度不变，用 null 填充
      const newArr = Array(mainNumbers.value.length).fill(null)
      filled.forEach((n, i) => { if (n !== null) newArr[i] = n })
      mainNumbers.value = newArr
    } else if (blueNumbers.value.length > 0) {
      const filled = blueNumbers.value.filter(n => n !== null).sort((a, b) => a! - b!)
      const newArr = Array(blueNumbers.value.length).fill(null)
      filled.forEach((n, i) => { if (n !== null) newArr[i] = n })
      blueNumbers.value = newArr
    }
  }

  errors.value = []
  updateValue()
}

// 添加号码输入框
const addMainSlot = () => {
  if (!config.value) return
  const supportsMultiple = props.lotteryType === '双色球' || props.lotteryType === '大乐透'
  if (!supportsMultiple) return
  const maxCount = props.lotteryType === '双色球' ? 20 : 18
  if (mainNumbers.value.length >= maxCount) {
    errors.value = [`最多选择 ${maxCount} 个号码`]
    return
  }
  mainNumbers.value.push(null)
}

const addBlueSlot = () => {
  if (!config.value) return
  const supportsMultiple = props.lotteryType === '双色球' || props.lotteryType === '大乐透'
  if (!supportsMultiple) return
  const maxCount = props.lotteryType === '双色球' ? 16 : 12
  if (blueNumbers.value.length >= maxCount) {
    errors.value = [`最多选择 ${maxCount} 个号码`]
    return
  }
  blueNumbers.value.push(null)
}

// 删除最后一个空输入框
const removeMainSlot = () => {
  if (mainNumbers.value.length <= (config.value?.redRange.count ?? 1)) return
  // 从后往前找第一个为 null 的
  const lastNullIndex = mainNumbers.value.reduceRight((acc, n, i) => {
    if (acc === -1 && n === null) return i
    return acc
  }, -1)
  if (lastNullIndex !== -1) {
    mainNumbers.value.splice(lastNullIndex as number, 1)
  } else {
    // 如果没有空的，删除最后一个
    mainNumbers.value.pop()
  }
  updateValue()
}

const removeBlueSlot = () => {
  if (blueNumbers.value.length <= (config.value?.blueRange?.count ?? 1)) return
  const lastNullIndex = blueNumbers.value.reduceRight((acc, n, i) => {
    if (acc === -1 && n === null) return i
    return acc
  }, -1)
  if (lastNullIndex !== -1) {
    blueNumbers.value.splice(lastNullIndex as number, 1)
  } else {
    blueNumbers.value.pop()
  }
  updateValue()
}

const mainLabel = computed(() => {
  const lt = props.lotteryType
  if (lt === '大乐透') return '前区'
  if (lt === '七乐彩') return '主号'
  return '红球'
})

const blueLabel = computed(() => {
  const lt = props.lotteryType
  if (lt === '大乐透') return '后区'
  if (lt === '七乐彩') return '特别号'
  return '蓝球'
})

const mainMax = computed(() => config.value?.redRange.max ?? 99)
const mainMin = computed(() => config.value?.redRange.min ?? 0)
const blueMax = computed(() => config.value?.blueRange?.max ?? 99)
const blueMin = computed(() => config.value?.blueRange?.min ?? 0)

// 是否支持复式
const supportsMultiple = computed(() => props.lotteryType === '双色球' || props.lotteryType === '大乐透')
</script>

<template>
  <div class="space-y-4">
    <!-- 错误提示 -->
    <div v-if="errors.length > 0" class="flex items-center gap-2 px-4 py-3 bg-red-50 border border-red-200 rounded-xl">
      <span class="text-red-500 text-sm">⚠</span>
      <span class="text-red-600 text-sm">{{ errors[0] }}</span>
    </div>

    <!-- 主号码 -->
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="text-sm font-medium text-slate-600">
          {{ mainLabel }}
          <span class="text-slate-400 ml-1">
            ({{ mainMin }}-{{ mainMax }}，至少{{ config?.redRange.count }}个)
          </span>
        </label>
        <div v-if="supportsMultiple" class="flex items-center gap-1">
          <button
            @click="removeMainSlot"
            :disabled="mainNumbers.length <= (config?.redRange.count ?? 1)"
            class="w-6 h-6 flex items-center justify-center rounded-md bg-slate-100 text-slate-500 hover:bg-slate-200 text-xs disabled:opacity-30 cursor-pointer"
          >
            −
          </button>
          <span class="text-xs text-slate-400 w-8 text-center">{{ mainNumbers.length }}</span>
          <button
            @click="addMainSlot"
            class="w-6 h-6 flex items-center justify-center rounded-md bg-slate-100 text-slate-500 hover:bg-slate-200 text-xs cursor-pointer"
          >
            +
          </button>
        </div>
      </div>
      <div class="flex flex-wrap gap-2">
        <div
          v-for="(num, i) in mainNumbers"
          :key="i"
          class="relative"
        >
          <input
            type="number"
            :value="num ?? ''"
            :min="mainMin"
            :max="mainMax"
            placeholder="--"
            @input="handleInput(i, 'main', $event)"
            @blur="handleBlur(i, 'main', $event)"
            class="w-14 h-12 text-center text-sm font-bold rounded-xl border-2 border-slate-200 focus:border-blue-400 focus:outline-none transition-all
              bg-gradient-to-b from-red-50 to-white text-red-600
              hover:border-red-300 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
          />
        </div>
      </div>
    </div>

    <!-- 蓝球/后区 -->
    <div v-if="config?.blueRange">
      <div class="flex items-center justify-between mb-2">
        <label class="text-sm font-medium text-slate-600">
          {{ blueLabel }}
          <span class="text-slate-400 ml-1">
            ({{ blueMin }}-{{ blueMax }}，至少{{ config?.blueRange.count }}个)
          </span>
        </label>
        <div v-if="supportsMultiple" class="flex items-center gap-1">
          <button
            @click="removeBlueSlot"
            :disabled="blueNumbers.length <= (config?.blueRange?.count ?? 1)"
            class="w-6 h-6 flex items-center justify-center rounded-md bg-slate-100 text-slate-500 hover:bg-slate-200 text-xs disabled:opacity-30 cursor-pointer"
          >
            −
          </button>
          <span class="text-xs text-slate-400 w-8 text-center">{{ blueNumbers.length }}</span>
          <button
            @click="addBlueSlot"
            class="w-6 h-6 flex items-center justify-center rounded-md bg-slate-100 text-slate-500 hover:bg-slate-200 text-xs cursor-pointer"
          >
            +
          </button>
        </div>
      </div>
      <div class="flex flex-wrap gap-2">
        <div
          v-for="(num, i) in blueNumbers"
          :key="i"
        >
          <input
            type="number"
            :value="num ?? ''"
            :min="blueMin"
            :max="blueMax"
            placeholder="--"
            @input="handleInput(i, 'blue', $event)"
            @blur="handleBlur(i, 'blue', $event)"
            class="w-14 h-12 text-center text-sm font-bold rounded-xl border-2 border-slate-200 focus:border-blue-400 focus:outline-none transition-all
              bg-gradient-to-b from-blue-50 to-white text-blue-600
              hover:border-blue-300 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
          />
        </div>
      </div>
    </div>

    <!-- 注数信息 -->
    <div v-if="supportsMultiple" class="flex items-center gap-3 px-4 py-2 bg-amber-50 border border-amber-200 rounded-xl">
      <span class="text-amber-600 text-sm font-medium">
        {{ isMultiple ? '复式投注' : '单式投注' }}
      </span>
      <span class="text-amber-500 text-xs">|</span>
      <span class="text-amber-700 text-sm">
        共 <strong>{{ betCount }}</strong> 注
      </span>
      <span v-if="isMultiple" class="text-amber-500 text-xs">
        (金额 ¥{{ betCount * 2 }})
      </span>
    </div>

    <!-- 号码预览 -->
    <div v-if="modelValue && modelValue !== '[]'" class="mt-3 p-3 bg-slate-50 rounded-lg">
      <span class="text-xs text-slate-400">号码预览：</span>
      <span class="text-xs text-slate-600 ml-2 font-mono">{{ modelValue }}</span>
    </div>
  </div>
</template>
