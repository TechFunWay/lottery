<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { statsApi } from '../api'
import type { OverviewStats, PrizeDistribution, TrendData } from '../types'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { TrendingUp, PieChart as PieChartIcon, BarChart3, Target } from 'lucide-vue-next'

use([CanvasRenderer, LineChart, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent])

const overview = ref<OverviewStats | null>(null)
const prizeDistribution = ref<PrizeDistribution[]>([])
const trends = ref<TrendData[]>([])
const loading = ref(true)
const filterType = ref('')

const loadData = async () => {
  loading.value = true
  try {
    const [overviewRes, prizesRes, trendsRes] = await Promise.all([
      statsApi.overview(filterType.value).catch(() => null),
      statsApi.prizes(filterType.value).catch(() => null),
      statsApi.trends(filterType.value, 12).catch(() => null)
    ])
    if (overviewRes?.data) overview.value = overviewRes.data
    // prizes 可能返回 null（无中奖记录），保证是数组
    prizeDistribution.value = Array.isArray(prizesRes?.data) ? prizesRes!.data : []
    if (trendsRes?.data) trends.value = trendsRes.data
  } catch (e) {
    console.error('统计数据加载失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

// 盈亏趋势图配置
const profitTrendOption = computed(() => {
  if (!trends.value.length) return {}
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['投入', '中奖', '净盈亏'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: trends.value.map(t => t.month) },
    yAxis: { type: 'value' },
    series: [
      {
        name: '投入',
        type: 'bar',
        data: trends.value.map(t => t.investment),
        itemStyle: { color: '#94a3b8' }
      },
      {
        name: '中奖',
        type: 'bar',
        data: trends.value.map(t => t.win_amount),
        itemStyle: { color: '#10b981' }
      },
      {
        name: '净盈亏',
        type: 'line',
        data: trends.value.map(t => t.win_amount - t.investment),
        itemStyle: { color: '#3b82f6' },
        lineStyle: { width: 3 }
      }
    ]
  }
})

// 奖级分布饼图配置
const prizePieOption = computed(() => {
  if (!prizeDistribution.value.length) return {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c}次 ({d}%)' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false, position: 'center' },
      emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
      data: prizeDistribution.value.map(p => ({ name: p.prize_name, value: p.count }))
    }]
  }
})

// 中奖率趋势图配置
const winRateOption = computed(() => {
  if (!trends.value.length) return {}
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: trends.value.map(t => t.month) },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [{
      name: '中奖率',
      type: 'line',
      smooth: true,
      data: trends.value.map(t => t.total_bets > 0 ? (t.win_count / t.total_bets * 100).toFixed(1) : 0),
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(59, 130, 246, 0.3)' }, { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }] } },
      itemStyle: { color: '#3b82f6' },
      lineStyle: { width: 3 }
    }]
  }
})

const formatMoney = (v: number) => {
  if (!v && v !== 0) return '¥0.00'
  return `¥${Math.abs(v).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
</script>

<template>
  <div class="animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-slate-800">统计分析</h1>
      <select
        v-model="filterType"
        @change="loadData"
        class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
      >
        <option value="">全部类型</option>
        <option value="双色球">双色球</option>
        <option value="大乐透">大乐透</option>
        <option value="福彩3D">福彩3D</option>
        <option value="排列3">排列3</option>
        <option value="排列5">排列5</option>
        <option value="七乐彩">七乐彩</option>
      </select>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else>
      <!-- Overview Cards -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div class="bg-white rounded-2xl p-5 card-shadow">
          <div class="flex items-center gap-2 mb-2">
            <TrendingUp class="w-5 h-5 text-blue-500" />
            <span class="text-sm text-slate-400">总投入</span>
          </div>
          <div class="text-2xl font-bold text-slate-800">{{ formatMoney(overview?.total_investment || 0) }}</div>
        </div>

        <div class="bg-white rounded-2xl p-5 card-shadow">
          <div class="flex items-center gap-2 mb-2">
            <Target class="w-5 h-5 text-emerald-500" />
            <span class="text-sm text-slate-400">总中奖</span>
          </div>
          <div class="text-2xl font-bold text-emerald-600">{{ formatMoney(overview?.total_winning || 0) }}</div>
        </div>

        <div class="bg-white rounded-2xl p-5 card-shadow">
          <div class="flex items-center gap-2 mb-2">
            <BarChart3 class="w-5 h-5 text-amber-500" />
            <span class="text-sm text-slate-400">净盈亏</span>
          </div>
          <div class="text-2xl font-bold" :class="(overview?.net_profit || 0) >= 0 ? 'text-emerald-600' : 'text-red-500'">
            {{ (overview?.net_profit || 0) >= 0 ? '+' : '-' }}{{ formatMoney(overview?.net_profit || 0) }}
          </div>
        </div>

        <div class="bg-white rounded-2xl p-5 card-shadow">
          <div class="flex items-center gap-2 mb-2">
            <PieChartIcon class="w-5 h-5 text-purple-500" />
            <span class="text-sm text-slate-400">中奖率</span>
          </div>
          <div class="text-2xl font-bold text-purple-600">{{ (overview?.win_rate || 0).toFixed(1) }}%</div>
        </div>
      </div>

        <!-- Charts -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- 盈亏趋势 -->
        <div class="bg-white rounded-2xl p-6 card-shadow">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">盈亏趋势（近12个月）</h3>
          <v-chart v-if="trends.length > 0" class="w-full h-72" :option="profitTrendOption" autoresize />
          <div v-else class="h-72 flex items-center justify-center text-slate-400 text-sm">暂无数据</div>
        </div>

        <!-- 奖级分布 -->
        <div class="bg-white rounded-2xl p-6 card-shadow">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">奖级分布</h3>
          <v-chart v-if="prizeDistribution.length > 0" class="w-full h-72" :option="prizePieOption" autoresize />
          <div v-else class="h-72 flex items-center justify-center text-slate-400 text-sm">暂无中奖数据</div>
        </div>

        <!-- 中奖率趋势 -->
        <div class="bg-white rounded-2xl p-6 card-shadow lg:col-span-2">
          <h3 class="text-lg font-semibold text-slate-800 mb-4">中奖率趋势</h3>
          <v-chart v-if="trends.length > 0" class="w-full h-64" :option="winRateOption" autoresize />
          <div v-else class="h-64 flex items-center justify-center text-slate-400 text-sm">暂无数据</div>
        </div>
      </div>
    </div>
  </div>
</template>
