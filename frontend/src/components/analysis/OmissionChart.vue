<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { AnalysisZone } from '../../types'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent])

const props = defineProps<{ zone: AnalysisZone }>()

const option = computed(() => {
  const nums = props.zone.omission.map(o => String(o.num).padStart(2, '0'))
  const current = props.zone.omission.map(o => o.current)
  const max = props.zone.omission.map(o => o.max)
  return {
    title: { text: `${props.zone.name} · 遗漏值`, left: 'center', textStyle: { fontSize: 13 } },
    tooltip: { trigger: 'axis' },
    legend: { bottom: 0, data: ['当前遗漏', '历史最大'], textStyle: { fontSize: 10 } },
    grid: { left: 30, right: 10, top: 40, bottom: 40 },
    xAxis: { type: 'category', data: nums, axisLabel: { fontSize: 9 } },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      { name: '当前遗漏', type: 'bar', data: current, itemStyle: { color: '#10b981' } },
      { name: '历史最大', type: 'bar', data: max, itemStyle: { color: '#cbd5e1' } },
    ],
  }
})
</script>

<template>
  <div class="bg-white rounded-lg shadow p-2">
    <VChart :option="option" autoresize style="height: 240px; min-width: 480px" />
  </div>
</template>
