<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { drawApi } from '../api'
import { LOTTERY_CONFIGS } from '../types'
import type { DrawAnalysis } from '../types'
import FrequencyChart from '../components/analysis/FrequencyChart.vue'
import OmissionChart from '../components/analysis/OmissionChart.vue'
import TrendChart from '../components/analysis/TrendChart.vue'
import MetricsChart from '../components/analysis/MetricsChart.vue'

// 仅 7 种数字彩（LOTTERY_CONFIGS 前 7 项即彩票，足球玩法在其后）
const lotteryTypes = LOTTERY_CONFIGS.map(c => c.type)

const selectedType = ref<string>('双色球')
const count = ref<number>(50)
const customCount = ref<string>('')
const presets = [
  { label: '近30期', value: 30 },
  { label: '近50期', value: 50 },
  { label: '近100期', value: 100 },
  { label: '全部', value: 0 },
]

const analysis = ref<DrawAnalysis | null>(null)
const loading = ref(false)
const error = ref('')

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await drawApi.analysis(selectedType.value, count.value)
    analysis.value = res.data
  } catch (e: any) {
    error.value = e?.response?.data?.error || '加载失败'
    analysis.value = null
  } finally {
    loading.value = false
  }
}

const selectType = (t: string) => {
  selectedType.value = t
  load()
}
const selectPreset = (v: number) => {
  count.value = v
  customCount.value = ''
  load()
}
const applyCustom = () => {
  const n = parseInt(customCount.value, 10)
  if (!isNaN(n) && n > 0) {
    count.value = n
    load()
  }
}

onMounted(load)
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 控制栏 -->
    <div class="bg-white rounded-lg shadow p-4 space-y-3">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="t in lotteryTypes"
          :key="t"
          @click="selectType(t)"
          class="px-3 py-1 rounded text-sm border"
          :class="selectedType === t ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-white text-slate-600 border-slate-300'"
        >{{ t }}</button>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="p in presets"
          :key="p.value"
          @click="selectPreset(p.value)"
          class="px-3 py-1 rounded text-sm border"
          :class="count === p.value && customCount === '' ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-white text-slate-600 border-slate-300'"
        >{{ p.label }}</button>
        <input
          v-model="customCount"
          @keyup.enter="applyCustom"
          type="number"
          min="1"
          placeholder="自定义期数"
          class="w-28 px-2 py-1 text-sm border border-slate-300 rounded"
        />
      </div>
    </div>

    <!-- 状态 -->
    <div v-if="loading" class="text-center text-slate-400 py-12">加载中…</div>
    <div v-else-if="error" class="text-center text-red-500 py-12">{{ error }}</div>
    <div v-else-if="!analysis || analysis.issue_count === 0" class="text-center text-slate-400 py-12">
      暂无开奖数据，请先到
      <router-link to="/draw" class="text-emerald-600 underline">开奖管理</router-link>
      抓取或录入。
    </div>

    <!-- 图表区（Task 8-10 填充） -->
    <template v-else>
      <div class="text-sm text-slate-500">共 {{ analysis.issue_count }} 期</div>
      <div v-for="zone in analysis.zones" :key="zone.name" class="space-y-3">
        <h3 class="font-semibold text-slate-700">{{ zone.name }}</h3>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-3 overflow-x-auto">
          <FrequencyChart :zone="zone" />
          <OmissionChart :zone="zone" />
        </div>
        <TrendChart :zone="zone" :issues="analysis.issues" />
      </div>
      <div v-if="analysis.metrics.length" class="space-y-3">
        <h3 class="font-semibold text-slate-700">和值 / 跨度</h3>
        <div class="overflow-x-auto">
          <MetricsChart :metrics="analysis.metrics" />
        </div>
      </div>
    </template>
  </div>
</template>
