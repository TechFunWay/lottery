<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { AnalysisZone } from '../../types'

use([CanvasRenderer, BarChart, GridComponent, TooltipComponent, TitleComponent])

const props = defineProps<{ zone: AnalysisZone }>()

const option = computed(() => {
  const nums = props.zone.frequency.map(f => String(f.num).padStart(2, '0'))
  const counts = props.zone.frequency.map(f => f.count)
  const maxCount = Math.max(1, ...counts)
  return {
    title: { text: `${props.zone.name} · 出现频率`, left: 'center', textStyle: { fontSize: 13 } },
    tooltip: { trigger: 'axis' },
    grid: { left: 30, right: 10, top: 40, bottom: 30 },
    xAxis: { type: 'category', data: nums, axisLabel: { fontSize: 9 } },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      type: 'bar',
      data: counts.map(c => ({
        value: c,
        // 冷热渐变：出现越多越暖（红），越少越冷（蓝）
        itemStyle: { color: `hsl(${210 - Math.round((c / maxCount) * 210)},70%,55%)` },
      })),
    }],
  }
})
</script>

<template>
  <div class="bg-white rounded-lg shadow p-2">
    <VChart :option="option" autoresize style="height: 240px; min-width: 480px" />
  </div>
</template>
