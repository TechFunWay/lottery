<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { footballMatchApi, footballBetApi } from '../api'
import type { FootballMatch, FootballMatchStatus } from '../types'
import { Plus, Trash2, Edit2, X, RefreshCw, Download, CheckCircle, AlertCircle } from 'lucide-vue-next'

const matches = ref<FootballMatch[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingId = ref<number | null>(null)
const filterLeague = ref('')
const filterStatus = ref('')
const deleteConfirm = ref(false)
const deleteId = ref<number | null>(null)
const fetching = ref(false)

const toast = ref({
  show: false,
  type: 'info' as 'success' | 'error' | 'info',
  message: ''
})

const showToast = (type: 'success' | 'error' | 'info', message: string) => {
  toast.value = { show: true, type, message }
  setTimeout(() => { toast.value.show = false }, 3000)
}

const form = ref({
  match_id: '',
  issue_number: '',
  league: '',
  home_team: '',
  away_team: '',
  match_time: new Date().toISOString().split('T')[0],
  home_score: 0,
  away_score: 0,
  half_home_score: 0,
  half_away_score: 0,
  handicap: 0,
  status: '未开赛' as FootballMatchStatus,
})

const errors = ref<Record<string, string>>({})

const validateForm = () => {
  errors.value = {}
  if (!form.value.match_id.trim()) errors.value.match_id = '比赛ID为必填项'
  if (!form.value.home_team.trim()) errors.value.home_team = '主队为必填项'
  if (!form.value.away_team.trim()) errors.value.away_team = '客队为必填项'
  return Object.keys(errors.value).length === 0
}

const loadMatches = async () => {
  loading.value = true
  const res = await footballMatchApi.list({
    league: filterLeague.value,
    status: filterStatus.value,
  }).catch(() => null)
  if (res) matches.value = (res as any).data || []
  loading.value = false
}

onMounted(loadMatches)

const openModal = (item?: FootballMatch) => {
  errors.value = {}
  if (item) {
    editingId.value = item.id
    form.value = {
      match_id: item.match_id,
      issue_number: item.issue_number,
      league: item.league,
      home_team: item.home_team,
      away_team: item.away_team,
      match_time: item.match_time?.split('T')[0] || '',
      home_score: item.home_score,
      away_score: item.away_score,
      half_home_score: item.half_home_score,
      half_away_score: item.half_away_score,
      handicap: item.handicap,
      status: item.status,
    }
  } else {
    editingId.value = null
    form.value = {
      match_id: '',
      issue_number: '',
      league: '',
      home_team: '',
      away_team: '',
      match_time: new Date().toISOString().split('T')[0],
      home_score: 0,
      away_score: 0,
      half_home_score: 0,
      half_away_score: 0,
      handicap: 0,
      status: '未开赛',
    }
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
}

const saveMatch = async () => {
  if (!validateForm()) return
  const payload = { ...form.value }
  if (!payload.issue_number) payload.issue_number = payload.match_id

  try {
    if (editingId.value) {
      await footballMatchApi.update(editingId.value, payload)
    } else {
      await footballMatchApi.create(payload)
    }
    closeModal()
    loadMatches()
    showToast('success', editingId.value ? '修改成功' : '新增成功')
  } catch (e: any) {
    const msg = e.response?.data?.error || '保存失败'
    showToast('error', msg)
  }
}

const deleteMatch = (id: number) => {
  deleteId.value = id
  deleteConfirm.value = true
}

const confirmDelete = async () => {
  if (!deleteId.value) return
  try {
    await footballMatchApi.delete(deleteId.value)
    deleteConfirm.value = false
    deleteId.value = null
    loadMatches()
    showToast('success', '删除成功')
  } catch (e) {
    showToast('error', '删除失败')
  }
}

const fetchMatches = async () => {
  fetching.value = true
  try {
    const res = await footballMatchApi.fetch()
    const data = (res as any).data
    showToast('success', `获取 ${data?.total || 0} 场比赛，新增 ${data?.saved_count || 0} 场`)
    loadMatches()
  } catch (e: any) {
    showToast('error', '抓取失败：' + (e.response?.data?.error || e.message))
  } finally {
    fetching.value = false
  }
}

const fetchResults = async () => {
  fetching.value = true
  try {
    const res = await footballMatchApi.fetchResults()
    const data = (res as any).data
    showToast('success', `更新 ${data?.updated_count || 0} 场比赛结果`)
    loadMatches()
    await footballBetApi.recheck()
  } catch (e: any) {
    showToast('error', '抓取失败：' + (e.response?.data?.error || e.message))
  } finally {
    fetching.value = false
  }
}

const matchStatusColors: Record<string, string> = {
  '未开赛': 'bg-slate-100 text-slate-600',
  '进行中': 'bg-blue-100 text-blue-600',
  '已完赛': 'bg-emerald-100 text-emerald-600',
  '已取消': 'bg-red-100 text-red-500',
  '延期': 'bg-amber-100 text-amber-600',
}
</script>

<template>
  <div class="animate-fade-in">
    <div class="mb-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-800">比赛管理</h1>
          <p class="text-slate-400 text-sm mt-1">共 {{ matches.length }} 场比赛</p>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="fetchMatches"
            :disabled="fetching"
            class="flex items-center gap-1.5 px-3 py-2 bg-amber-500 hover:bg-amber-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-amber-500/30 cursor-pointer disabled:opacity-50"
          >
            <Download class="w-4 h-4" />
            <span class="hidden sm:inline">抓取赛程</span>
          </button>
          <button
            @click="fetchResults"
            :disabled="fetching"
            class="flex items-center gap-1.5 px-3 py-2 bg-emerald-500 hover:bg-emerald-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-emerald-500/30 cursor-pointer disabled:opacity-50"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': fetching }" />
            <span class="hidden sm:inline">获取结果</span>
          </button>
          <button
            @click="openModal()"
            class="flex items-center gap-1.5 px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer"
          >
            <Plus class="w-4 h-4" />
            <span class="hidden sm:inline">新增</span>
          </button>
        </div>
      </div>
    </div>

    <div class="flex flex-wrap gap-3 mb-6">
      <select v-model="filterStatus" @change="loadMatches"
        class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer">
        <option value="">全部状态</option>
        <option value="未开赛">未开赛</option>
        <option value="进行中">进行中</option>
        <option value="已完赛">已完赛</option>
        <option value="已取消">已取消</option>
        <option value="延期">延期</option>
      </select>
    </div>

    <div class="bg-white rounded-2xl card-shadow overflow-hidden">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="matches.length === 0" class="text-center py-16 text-slate-400">
        <p>暂无比赛记录</p>
      </div>

      <template v-else>
        <div class="md:hidden p-4 space-y-3">
          <div v-for="item in matches" :key="item.id" class="bg-slate-50 rounded-xl p-4">
            <div class="flex items-start justify-between mb-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-xs text-slate-400">{{ item.league }}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="matchStatusColors[item.status]">
                  {{ item.status }}
                </span>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                  <Edit2 class="w-4 h-4" />
                </button>
                <button @click="deleteMatch(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div class="text-sm font-medium text-slate-800 mb-1">
              {{ item.home_team }} vs {{ item.away_team }}
            </div>
            <div v-if="item.status === '已完赛'" class="text-lg font-bold text-slate-700 mb-1">
              {{ item.home_score }} : {{ item.away_score }}
            </div>
            <div class="flex items-center gap-4 text-xs text-slate-500">
              <span>{{ item.match_time?.split('T')[0] }}</span>
              <span v-if="item.handicap !== 0" class="text-amber-500">让球: {{ item.handicap }}</span>
            </div>
          </div>
        </div>

        <div class="hidden md:block overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">联赛</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">场次</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">主队</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">比分</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">客队</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">让球</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">状态</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">比赛时间</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-slate-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="item in matches" :key="item.id" class="hover:bg-slate-50">
                <td class="px-4 py-3 text-sm text-slate-600">{{ item.league || '-' }}</td>
                <td class="px-4 py-3 text-sm text-slate-500">{{ item.issue_number }}</td>
                <td class="px-4 py-3 text-sm font-medium text-slate-700">{{ item.home_team }}</td>
                <td class="px-4 py-3 text-sm font-bold text-slate-800">
                  <span v-if="item.status === '已完赛'">{{ item.home_score }} : {{ item.away_score }}</span>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="px-4 py-3 text-sm font-medium text-slate-700">{{ item.away_team }}</td>
                <td class="px-4 py-3 text-sm text-amber-500">{{ item.handicap !== 0 ? item.handicap : '-' }}</td>
                <td class="px-4 py-3">
                  <span class="px-2 py-1 rounded-full text-xs font-medium" :class="matchStatusColors[item.status]">
                    {{ item.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-slate-500">{{ item.match_time?.split('T')[0] }}</td>
                <td class="px-4 py-3 text-right">
                  <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button @click="deleteMatch(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeModal"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto animate-slide-up">
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800">{{ editingId ? '编辑' : '新增' }}比赛</h2>
          <button @click="closeModal" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">
                比赛ID<span class="text-red-500 ml-0.5">*</span>
              </label>
              <input v-model="form.match_id" placeholder="如：周一001"
                :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                  errors.match_id ? 'border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']" />
              <p v-if="errors.match_id" class="mt-1 text-xs text-red-500">⚠ {{ errors.match_id }}</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">场次号</label>
              <input v-model="form.issue_number" placeholder="默认同比赛ID"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">联赛</label>
            <input v-model="form.league" placeholder="如：英超"
              class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">
                主队<span class="text-red-500 ml-0.5">*</span>
              </label>
              <input v-model="form.home_team" placeholder="主队名称"
                :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                  errors.home_team ? 'border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">
                客队<span class="text-red-500 ml-0.5">*</span>
              </label>
              <input v-model="form.away_team" placeholder="客队名称"
                :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                  errors.away_team ? 'border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">比赛时间</label>
              <input v-model="form.match_time" type="date"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">状态</label>
              <select v-model="form.status"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
                <option value="未开赛">未开赛</option>
                <option value="进行中">进行中</option>
                <option value="已完赛">已完赛</option>
                <option value="已取消">已取消</option>
                <option value="延期">延期</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">主队得分</label>
              <input v-model.number="form.home_score" type="number" min="0"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">客队得分</label>
              <input v-model.number="form.away_score" type="number" min="0"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">半场主队</label>
              <input v-model.number="form.half_home_score" type="number" min="0"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">半场客队</label>
              <input v-model.number="form.half_away_score" type="number" min="0"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">让球数</label>
              <input v-model.number="form.handicap" type="number" step="0.5"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>
        </div>

        <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex justify-end gap-3">
          <button @click="closeModal" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button @click="saveMatch" class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer">保存</button>
        </div>
      </div>
    </div>

    <div v-if="deleteConfirm" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="deleteConfirm = false; deleteId = null"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-sm animate-slide-up">
        <div class="p-6 text-center">
          <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Trash2 class="w-6 h-6 text-red-500" />
          </div>
          <h3 class="text-lg font-semibold text-slate-800 mb-2">确认删除</h3>
          <p class="text-slate-600 text-sm">确定要删除这场比赛吗？</p>
        </div>
        <div class="border-t border-slate-100 px-6 py-4 flex gap-3">
          <button @click="deleteConfirm = false; deleteId = null"
            class="flex-1 px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button @click="confirmDelete"
            class="flex-1 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-xl font-medium transition-colors cursor-pointer">删除</button>
        </div>
      </div>
    </div>

    <Transition name="toast">
      <div v-if="toast.show"
        class="fixed top-20 left-1/2 -translate-x-1/2 z-[100] px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 animate-slide-up"
        :class="{
          'bg-emerald-500 text-white': toast.type === 'success',
          'bg-red-500 text-white': toast.type === 'error',
          'bg-blue-500 text-white': toast.type === 'info'
        }">
        <CheckCircle v-if="toast.type === 'success'" class="w-5 h-5" />
        <AlertCircle v-else-if="toast.type === 'error'" class="w-5 h-5" />
        <AlertCircle v-else class="w-5 h-5" />
        <span class="font-medium">{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>
