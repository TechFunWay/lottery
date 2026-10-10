<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { AnalysisMetric } from '../../types'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent])

const props = defineProps<{ metrics: AnalysisMetric[] }>()

const latest = computed(() => props.metrics[props.metrics.length - 1])

const option = computed(() => ({
  title: { text: '和值 / 跨度走势', left: 'center', textStyle: { fontSize: 13 } },
  tooltip: { trigger: 'axis' },
  legend: { bottom: 0, data: ['和值', '跨度'], textStyle: { fontSize: 10 } },
  grid: { left: 36, right: 10, top: 40, bottom: 40 },
  xAxis: {
    type: 'category',
    data: props.metrics.map(m => m.issue),
    axisLabel: { fontSize: 8, rotate: 45, interval: Math.ceil(props.metrics.length / 20) },
  },
  yAxis: { type: 'value' },
  series: [
    { name: '和值', type: 'line', smooth: true, data: props.metrics.map(m => m.sum), itemStyle: { color: '#10b981' } },
    { name: '跨度', type: 'line', smooth: true, data: props.metrics.map(m => m.span), itemStyle: { color: '#f59e0b' } },
  ],
}))
</script>

<template>
  <div class="bg-white rounded-lg shadow p-2">
    <VChart :option="option" autoresize style="height: 280px; min-width: 600px" />
    <div v-if="latest" class="flex gap-4 justify-center text-xs text-slate-500 pb-2">
      <span>最近一期（{{ latest.issue }}）奇偶比 <b class="text-slate-700">{{ latest.oddEven }}</b></span>
      <span v-if="latest.bigSmall">大小比 <b class="text-slate-700">{{ latest.bigSmall }}</b></span>
    </div>
  </div>
</template>
