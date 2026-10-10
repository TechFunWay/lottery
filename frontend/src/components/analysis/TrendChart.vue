<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { ScatterChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import type { AnalysisZone } from '../../types'

use([CanvasRenderer, ScatterChart, GridComponent, TooltipComponent, TitleComponent])

const props = defineProps<{ zone: AnalysisZone; issues: string[] }>()

const option = computed(() => {
  // 散点：[期号索引, 号码]
  const points: [number, number][] = []
  props.zone.trend.forEach((nums, issueIdx) => {
    nums.forEach(n => points.push([issueIdx, n]))
  })
  const yData: number[] = []
  for (let n = props.zone.min; n <= props.zone.max; n++) yData.push(n)
  return {
    title: { text: `${props.zone.name} · 走势`, left: 'center', textStyle: { fontSize: 13 } },
    tooltip: {
      trigger: 'item',
      formatter: (p: any) => `${props.issues[p.value[0]] ?? ''}<br/>号码 ${String(p.value[1] + props.zone.min).padStart(2, '0')}`,
    },
    grid: { left: 36, right: 10, top: 40, bottom: 40 },
    xAxis: {
      type: 'category',
      data: props.issues,
      axisLabel: { fontSize: 8, rotate: 45, interval: Math.ceil(props.issues.length / 20) },
    },
    yAxis: {
      type: 'category',
      data: yData.map(n => String(n).padStart(2, '0')),
      axisLabel: { fontSize: 8 },
    },
    series: [{
      type: 'scatter',
      symbolSize: 8,
      data: points.map(([x, y]) => [x, y - props.zone.min]),
      itemStyle: { color: '#10b981' },
    }],
  }
})
</script>

<template>
  <div class="bg-white rounded-lg shadow p-2">
    <VChart :option="option" autoresize style="height: 360px; min-width: 600px" />
  </div>
</template>
