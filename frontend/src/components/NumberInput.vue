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
}>()

const config = computed(() => LOTTERY_CONFIGS.find(c => c.type === props.lotteryType))

// 主号码数组
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
  if (internalUpdate) return  // 内部更新触发的，跳过，防止循环覆盖
  if (!val || !config.value) return
  // 如果 val 是空字符串，不做处理（避免初始化时清空）
  if (val === '' || val === '[]') return
  try {
    const parsed = JSON.parse(val)
    const blueCount = config.value.blueRange?.count ?? 0

    const padTo = (arr: number[], count: number) =>
      [...arr, ...Array(Math.max(0, count - arr.length)).fill(null)]

    if (parsed.red !== undefined) {
      mainNumbers.value = padTo(parsed.red ?? [], config.value.redRange.count)
      blueNumbers.value = padTo(parsed.blue ?? [], blueCount)
    } else if (parsed.front !== undefined) {
      mainNumbers.value = padTo(parsed.front ?? [], config.value.redRange.count)
      blueNumbers.value = padTo(parsed.back ?? [], blueCount)
    } else if (parsed.main !== undefined) {
      mainNumbers.value = padTo(parsed.main ?? [], config.value.redRange.count)
      blueNumbers.value = padTo(parsed.special ?? [], blueCount)
    } else if (parsed.numbers !== undefined) {
      mainNumbers.value = padTo(parsed.numbers ?? [], config.value.redRange.count)
    } else if (Array.isArray(parsed)) {
      mainNumbers.value = padTo(parsed, config.value.redRange.count)
    }
  } catch (e) {
    // ignore
  }
}, { immediate: true })

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

  // 标记为内部更新，避免 watch modelValue 触发后覆盖蓝球槽位
  internalUpdate = true
  emit('update:modelValue', JSON.stringify(result))
  // 下一个 tick 后重置标记，保证外部真实赋值（如编辑回填）仍能触发 watch
  setTimeout(() => { internalUpdate = false }, 0)
}

const handleInput = (index: number, type: 'main' | 'blue', event: Event) => {
  const input = event.target as HTMLInputElement
  const rawVal = input.value.trim()

  // 清空当前错误
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

  // 只更新临时值，不做校验（等失去焦点时校验）
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

  // 获取当前号码范围
  const min = type === 'main' ? (config.value?.redRange.min ?? 0) : (config.value?.blueRange?.min ?? 0)
  const max = type === 'main' ? (config.value?.redRange.max ?? 99) : (config.value?.blueRange?.max ?? 99)

  // 范围检查
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

  // 重复检查
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

  // 有效值，自动排序
  if (type === 'main') {
    const filled = mainNumbers.value.filter(n => n !== null).sort((a, b) => a! - b!)
    mainNumbers.value = [...filled, ...Array(config.value!.redRange.count - filled.length).fill(null)]
  } else if (blueNumbers.value.length > 0) {
    const filled = blueNumbers.value.filter(n => n !== null).sort((a, b) => a! - b!)
    blueNumbers.value = [...filled, ...Array(config.value!.blueRange!.count - filled.length).fill(null)]
  }

  errors.value = []
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
      <label class="block text-sm font-medium text-slate-600 mb-2">
        {{ mainLabel }}
        <span class="text-slate-400 ml-1">
          ({{ mainMin }}-{{ mainMax }}，共{{ config?.redRange.count }}个)
        </span>
      </label>
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
      <label class="block text-sm font-medium text-slate-600 mb-2">
        {{ blueLabel }}
        <span class="text-slate-400 ml-1">
          ({{ blueMin }}-{{ blueMax }}，共{{ config?.blueRange.count }}个)
        </span>
      </label>
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

    <!-- 号码预览 -->
    <div v-if="modelValue && modelValue !== '[]'" class="mt-3 p-3 bg-slate-50 rounded-lg">
      <span class="text-xs text-slate-400">号码预览：</span>
      <span class="text-xs text-slate-600 ml-2 font-mono">{{ modelValue }}</span>
    </div>
  </div>
</template>
