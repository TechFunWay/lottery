<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { footballBetApi, footballMatchApi } from '../api'
import type { FootballOverview, FootballBet, FootballMatch } from '../types'
import { TrendingUp, TrendingDown, DollarSign, Target, Trophy, ShoppingCart, Zap, RefreshCw, CheckCircle, AlertCircle, Download, Plus, Edit2, ArrowRight } from 'lucide-vue-next'

const router = useRouter()
const overview = ref<FootballOverview | null>(null)
const recentBets = ref<FootballBet[]>([])
const recentMatches = ref<FootballMatch[]>([])
const loading = ref(true)

const toast = ref({
  show: false,
  type: 'info' as 'success' | 'error' | 'info',
  message: ''
})

const showToast = (type: 'success' | 'error' | 'info', message: string) => {
  toast.value = { show: true, type, message }
  setTimeout(() => { toast.value.show = false }, 3000)
}

const loadData = async () => {
  loading.value = true
  try {
    const [overviewRes, betsRes, matchesRes] = await Promise.all([
      footballBetApi.overview().catch(() => null),
      footballBetApi.list({ size: 5 }).catch(() => null),
      footballMatchApi.list({ size: 5 }).catch(() => null),
    ])
    if (overviewRes?.data) overview.value = overviewRes.data
    if (betsRes?.data) recentBets.value = (betsRes as any).data || []
    if (matchesRes?.data) recentMatches.value = (matchesRes as any).data || []
  } catch (e) {
    console.error('竞彩足球数据加载失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

const formatMoney = (v: number) => {
  if (!v && v !== 0) return '¥0.00'
  return `¥${Math.abs(v).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

const statusColors: Record<string, string> = {
  '待开奖': 'bg-amber-100 text-amber-600',
  '已中奖': 'bg-emerald-100 text-emerald-600',
  '未中奖': 'bg-slate-100 text-slate-500',
  '部分中奖': 'bg-blue-100 text-blue-600',
}

const matchStatusColors: Record<string, string> = {
  '未开赛': 'bg-slate-100 text-slate-500',
  '进行中': 'bg-blue-100 text-blue-600',
  '已完赛': 'bg-emerald-100 text-emerald-600',
  '已取消': 'bg-red-100 text-red-500',
  '延期': 'bg-amber-100 text-amber-600',
}

const parseSelections = (json: string) => {
  try { return JSON.parse(json) } catch { return [] }
}

const rechecking = ref(false)
const recheckBets = async () => {
  rechecking.value = true
  try {
    await footballBetApi.recheck()
    showToast('success', '正在重新检查中奖记录...')
    setTimeout(() => loadData(), 2000)
  } catch (e) {
    showToast('error', '重新检查失败')
  } finally {
    rechecking.value = false
  }
}

const fetching = ref(false)
const fetchMatches = async () => {
  fetching.value = true
  try {
    const res = await footballMatchApi.fetch()
    const data = (res as any).data
    const topMessage = (res as any).message as string | undefined
    if (data?.empty) {
      showToast('info', topMessage || '当前数据源暂无可用赛程数据')
    } else {
      showToast('success', `获取 ${data?.total || 0} 场比赛，新增 ${data?.saved_count || 0} 场`)
    }
    loadData()
  } catch (e: any) {
    showToast('error', '抓取赛程失败：' + (e.response?.data?.error || e.message))
  } finally {
    fetching.value = false
  }
}

const fetchResults = async () => {
  fetching.value = true
  try {
    const res = await footballMatchApi.fetchResults()
    const data = (res as any).data
    const topMessage = (res as any).message as string | undefined
    if (data?.empty) {
      showToast('info', topMessage || '当前数据源暂无可用比赛结果')
    } else {
      showToast('success', `更新 ${data?.updated_count || 0} 场比赛结果`)
    }
    loadData()
    await footballBetApi.recheck()
  } catch (e: any) {
    showToast('error', '获取结果失败：' + (e.response?.data?.error || e.message))
  } finally {
    fetching.value = false
  }
}
</script>

<template>
  <div class="animate-fade-in">
    <div class="gradient-primary rounded-2xl p-8 mb-8 text-white shadow-lg shadow-blue-500/20">
      <h1 class="text-3xl font-bold mb-1">竞彩足球</h1>
      <p class="text-blue-100 text-sm mb-6">记录每一次竞猜，追踪每一场比赛</p>
      <div class="flex items-end gap-4">
        <div class="text-blue-100 text-sm pb-1 shrink-0">净盈亏</div>
        <div class="text-4xl md:text-5xl font-bold truncate">
          {{ overview ? ((overview.net_profit >= 0 ? '+' : '') + formatMoney(overview.net_profit)) : '...' }}
        </div>
      </div>
    </div>

    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 card-shadow">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center">
            <ShoppingCart class="w-5 h-5 text-blue-500" />
          </div>
          <span class="text-xs text-slate-400">总投注</span>
        </div>
        <div class="text-2xl font-bold text-slate-800">{{ overview ? formatMoney(overview.total_amount) : '--' }}</div>
        <div class="text-xs text-slate-400 mt-1">共 {{ overview?.total_bets ?? 0 }} 次投注</div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-emerald-50 flex items-center justify-center">
            <DollarSign class="w-5 h-5 text-emerald-500" />
          </div>
          <span class="text-xs text-slate-400">总中奖</span>
        </div>
        <div class="text-2xl font-bold text-emerald-600">{{ overview ? formatMoney(overview.total_win) : '--' }}</div>
        <div class="text-xs text-slate-400 mt-1">中奖 {{ overview?.win_count ?? 0 }} 次</div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl flex items-center justify-center"
            :class="(overview?.net_profit ?? 0) >= 0 ? 'bg-emerald-50' : 'bg-red-50'">
            <component
              :is="(overview?.net_profit ?? 0) >= 0 ? TrendingUp : TrendingDown"
              class="w-5 h-5"
              :class="(overview?.net_profit ?? 0) >= 0 ? 'text-emerald-500' : 'text-red-500'"
            />
          </div>
          <span class="text-xs text-slate-400">净盈亏</span>
        </div>
        <div class="text-2xl font-bold"
          :class="(overview?.net_profit ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-500'">
          {{ overview ? ((overview.net_profit >= 0 ? '+' : '') + formatMoney(overview.net_profit)) : '--' }}
        </div>
      </div>

      <div class="bg-white rounded-2xl p-5 card-shadow">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 rounded-xl bg-amber-50 flex items-center justify-center">
            <Target class="w-5 h-5 text-amber-500" />
          </div>
          <span class="text-xs text-slate-400">中奖率</span>
        </div>
        <div class="text-2xl font-bold text-amber-600">{{ overview ? overview.win_rate.toFixed(1) + '%' : '--' }}</div>
      </div>
    </div>

    <div class="mb-8">
      <h2 class="text-lg font-semibold text-slate-800 mb-4">快捷操作</h2>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <button @click="router.push('/football/matches')"
          class="bg-white rounded-2xl p-5 card-shadow hover:shadow-md transition-all duration-200 cursor-pointer group text-left">
          <div class="w-12 h-12 rounded-xl bg-emerald-50 flex items-center justify-center mb-3 group-hover:bg-emerald-100 transition-colors">
            <Plus class="w-6 h-6 text-emerald-500" />
          </div>
          <div class="font-semibold text-slate-800 mb-1">录入比赛</div>
          <div class="text-xs text-slate-400">手动添加或抓取赛程</div>
        </button>

        <button @click="router.push('/football/bets')"
          class="bg-white rounded-2xl p-5 card-shadow hover:shadow-md transition-all duration-200 cursor-pointer group text-left">
          <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center mb-3 group-hover:bg-blue-100 transition-colors">
            <Edit2 class="w-6 h-6 text-blue-500" />
          </div>
          <div class="font-semibold text-slate-800 mb-1">录入投注</div>
          <div class="text-xs text-slate-400">记录竞彩投注选择</div>
        </button>

        <button @click="fetchMatches" :disabled="fetching"
          class="bg-white rounded-2xl p-5 card-shadow hover:shadow-md transition-all duration-200 cursor-pointer group text-left disabled:opacity-50">
          <div class="w-12 h-12 rounded-xl bg-amber-50 flex items-center justify-center mb-3 group-hover:bg-amber-100 transition-colors">
            <Download class="w-6 h-6 text-amber-500" />
          </div>
          <div class="font-semibold text-slate-800 mb-1">抓取赛程</div>
          <div class="text-xs text-slate-400">自动获取近期比赛</div>
        </button>

        <button @click="fetchResults" :disabled="fetching"
          class="bg-white rounded-2xl p-5 card-shadow hover:shadow-md transition-all duration-200 cursor-pointer group text-left disabled:opacity-50">
          <div class="w-12 h-12 rounded-xl bg-purple-50 flex items-center justify-center mb-3 group-hover:bg-purple-100 transition-colors">
            <RefreshCw class="w-6 h-6 text-purple-500" :class="{ 'animate-spin': fetching }" />
          </div>
          <div class="font-semibold text-slate-800 mb-1">获取结果</div>
          <div class="text-xs text-slate-400">更新已完赛比分</div>
        </button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <div class="bg-white rounded-2xl card-shadow p-6">
        <div class="flex items-center justify-between mb-5">
          <div class="flex items-center gap-2">
            <Zap class="w-5 h-5 text-amber-500" />
            <h2 class="text-lg font-semibold text-slate-800">近期比赛</h2>
          </div>
          <button @click="router.push('/football/matches')"
            class="flex items-center gap-1 text-sm text-blue-500 hover:text-blue-600 font-medium cursor-pointer">
            比赛管理 <ArrowRight class="w-4 h-4" />
          </button>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="recentMatches.length === 0" class="text-center py-8">
          <div class="text-slate-400 text-sm mb-3">暂无比赛记录</div>
          <button @click="router.push('/football/matches')"
            class="px-4 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-600 rounded-xl text-sm font-medium transition-colors cursor-pointer">
            去录入比赛 →
          </button>
        </div>

        <div v-else class="space-y-3">
          <div v-for="match in recentMatches" :key="match.id"
            class="flex items-center justify-between p-3 rounded-xl bg-slate-50">
            <div class="flex-1">
              <div class="text-sm font-medium text-slate-700">
                {{ match.home_team }} vs {{ match.away_team }}
              </div>
              <div class="text-xs text-slate-400 mt-0.5">
                {{ match.league }} · {{ match.match_time?.split('T')[0] }}
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span v-if="match.status === '已完赛'" class="text-sm font-bold text-slate-700">
                {{ match.home_score }}:{{ match.away_score }}
              </span>
              <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="matchStatusColors[match.status]">
                {{ match.status }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-2xl card-shadow p-6">
        <div class="flex items-center justify-between mb-5">
          <div class="flex items-center gap-2">
            <Trophy class="w-5 h-5 text-amber-500" />
            <h2 class="text-lg font-semibold text-slate-800">近期投注</h2>
          </div>
          <button @click="router.push('/football/bets')"
            class="flex items-center gap-1 text-sm text-blue-500 hover:text-blue-600 font-medium cursor-pointer">
            投注记录 <ArrowRight class="w-4 h-4" />
          </button>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="recentBets.length === 0" class="text-center py-8">
          <div class="text-slate-400 text-sm mb-3">暂无投注记录</div>
          <button @click="router.push('/football/bets')"
            class="px-4 py-2 bg-blue-50 hover:bg-blue-100 text-blue-600 rounded-xl text-sm font-medium transition-colors cursor-pointer">
            去录入投注 →
          </button>
        </div>

        <div v-else class="space-y-3">
          <div v-for="bet in recentBets" :key="bet.id"
            class="flex items-center justify-between p-3 rounded-xl bg-slate-50">
            <div class="flex-1">
              <div class="text-sm font-medium text-slate-700">
                {{ bet.bet_type }} · {{ parseSelections(bet.selections).length }}场
              </div>
              <div class="text-xs text-slate-400 mt-0.5">
                ¥{{ bet.amount }} · {{ bet.created_at?.split('T')[0] }}
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span v-if="bet.win_amount > 0" class="text-sm font-bold text-emerald-600">
                ¥{{ bet.win_amount.toLocaleString() }}
              </span>
              <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusColors[bet.status]">
                {{ bet.status }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <button
        @click="recheckBets"
        :disabled="rechecking"
        class="flex items-center gap-2 px-4 py-2 bg-amber-50 hover:bg-amber-100 text-amber-600 rounded-xl text-sm font-medium transition-colors cursor-pointer disabled:opacity-50"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': rechecking }" />
        重新检查中奖
      </button>
    </div>

    <Transition name="toast">
      <div
        v-if="toast.show"
        class="fixed top-20 left-1/2 -translate-x-1/2 z-[100] px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-slide-up"
        :class="{
          'bg-emerald-500 text-white': toast.type === 'success',
          'bg-red-500 text-white': toast.type === 'error',
          'bg-blue-500 text-white': toast.type === 'info'
        }"
      >
        <CheckCircle v-if="toast.type === 'success'" class="w-5 h-5" />
        <AlertCircle v-else-if="toast.type === 'error'" class="w-5 h-5" />
        <AlertCircle v-else class="w-5 h-5" />
        <span class="font-medium">{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>
