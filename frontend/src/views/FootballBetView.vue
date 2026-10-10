<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { footballBetApi, footballMatchApi } from '../api'
import FootballBetForm from '../components/FootballBetForm.vue'
import type { FootballBet, FootballMatch, FootballSelection } from '../types'
import { WDL_LABELS } from '../types'
import { Plus, Trash2, Edit2, X, RefreshCw, CheckCircle, AlertCircle } from 'lucide-vue-next'

const bets = ref<FootballBet[]>([])
const matches = ref<FootballMatch[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingId = ref<number | null>(null)
const filterStatus = ref('')
const deleteConfirm = ref(false)
const deleteId = ref<number | null>(null)
const rechecking = ref(false)

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
  bet_type: '单关',
  amount: 2,
  multiple: 1,
  selections: '[]',
  remark: '',
})

const errors = ref<Record<string, string>>({})

const validateForm = () => {
  errors.value = {}
  if (form.value.amount <= 0) errors.value.amount = '金额必须大于0'
  const sels = parseSelections(form.value.selections)
  if (sels.length === 0) errors.value.selections = '请至少添加一场比赛'
  for (const sel of sels) {
    if (!sel.match_id) { errors.value.selections = '请选择比赛'; break }
    if (!sel.selection) { errors.value.selections = '请选择投注结果'; break }
  }
  return Object.keys(errors.value).length === 0
}

const loadBets = async () => {
  loading.value = true
  const res = await footballBetApi.list({ status: filterStatus.value }).catch(() => null)
  if (res) bets.value = (res as any).data || []
  const matchRes = await footballMatchApi.list({ size: 500 }).catch(() => null)
  if (matchRes) matches.value = (matchRes as any).data || []
  loading.value = false
}

onMounted(loadBets)

const openModal = (item?: FootballBet) => {
  errors.value = {}
  if (item) {
    editingId.value = item.id
    form.value = {
      bet_type: item.bet_type,
      amount: item.amount,
      multiple: item.multiple,
      selections: item.selections,
      remark: item.remark,
    }
  } else {
    editingId.value = null
    form.value = {
      bet_type: '单关',
      amount: 2,
      multiple: 1,
      selections: '[]',
      remark: '',
    }
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
}

const saveBet = async () => {
  if (!validateForm()) return
  try {
    if (editingId.value) {
      await footballBetApi.update(editingId.value, form.value)
    } else {
      await footballBetApi.create(form.value)
    }
    closeModal()
    loadBets()
    showToast('success', editingId.value ? '修改成功' : '新增成功')
  } catch (e: any) {
    showToast('error', e.response?.data?.error || '保存失败')
  }
}

const deleteBet = (id: number) => {
  deleteId.value = id
  deleteConfirm.value = true
}

const confirmDelete = async () => {
  if (!deleteId.value) return
  try {
    await footballBetApi.delete(deleteId.value)
    deleteConfirm.value = false
    deleteId.value = null
    loadBets()
    showToast('success', '删除成功')
  } catch (e) {
    showToast('error', '删除失败')
  }
}

const recheckBets = async () => {
  rechecking.value = true
  try {
    await footballBetApi.recheck()
    showToast('success', '正在重新检查中奖记录...')
    setTimeout(() => loadBets(), 2000)
  } catch (e) {
    showToast('error', '重新检查失败')
  } finally {
    rechecking.value = false
  }
}

const parseSelections = (json: string): FootballSelection[] => {
  try { return JSON.parse(json) } catch { return [] }
}

const getMatchInfo = (matchId: string) => {
  return matches.value.find(m => m.match_id === matchId)
}

const statusColors: Record<string, string> = {
  '待开奖': 'bg-amber-100 text-amber-600',
  '已中奖': 'bg-emerald-100 text-emerald-600',
  '未中奖': 'bg-slate-100 text-slate-500',
  '部分中奖': 'bg-blue-100 text-blue-600',
}

const getOptionLabel = (playType: string, option: string) => {
  if (playType === '胜平负' || playType === '让球胜平负') {
    return WDL_LABELS[option] || option
  }
  return option
}

const isSelectionHit = (sel: FootballSelection): boolean | null => {
  const match = getMatchInfo(sel.match_id)
  if (!match || match.status !== '已完赛') return null

  if (sel.play_type === '胜平负') {
    const result = match.home_score > match.away_score ? '3' : match.home_score === match.away_score ? '1' : '0'
    return sel.selection === result
  }
  if (sel.play_type === '让球胜平负') {
    const handicap = sel.handicap || match.handicap
    const adjusted = match.home_score + handicap
    const result = adjusted > match.away_score ? '3' : adjusted === match.away_score ? '1' : '0'
    return sel.selection === result
  }
  if (sel.play_type === '比分') {
    const result = `${match.home_score}:${match.away_score}`
    return sel.selection === result
  }
  if (sel.play_type === '总进球') {
    const total = match.home_score + match.away_score
    const result = total >= 7 ? '7+' : String(total)
    return sel.selection === result
  }
  if (sel.play_type === '半全场') {
    const halfR = match.half_home_score > match.half_away_score ? '胜' : match.half_home_score === match.half_away_score ? '平' : '负'
    const fullR = match.home_score > match.away_score ? '胜' : match.home_score === match.away_score ? '平' : '负'
    return sel.selection === halfR + fullR
  }
  return null
}

const totalAmount = () => bets.value.reduce((sum, b) => sum + b.amount, 0)
const totalWin = () => bets.value.reduce((sum, b) => sum + b.win_amount, 0)
</script>

<template>
  <div class="animate-fade-in">
    <div class="mb-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-800">投注记录</h1>
          <p class="text-slate-400 text-sm mt-1">共 {{ bets.length }} 条记录，累计投注 ¥{{ totalAmount().toFixed(2) }}，中奖 ¥{{ totalWin().toFixed(2) }}</p>
        </div>
        <div class="flex items-center gap-2">
          <button @click="recheckBets" :disabled="rechecking"
            class="flex items-center gap-1.5 px-3 py-2 bg-amber-50 hover:bg-amber-100 text-amber-600 text-sm rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-50">
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': rechecking }" />
            <span class="hidden sm:inline">重新检查</span>
          </button>
          <button @click="openModal()"
            class="flex items-center gap-1.5 px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded-xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer">
            <Plus class="w-4 h-4" />
            <span class="hidden sm:inline">新增</span>
          </button>
        </div>
      </div>
    </div>

    <div class="flex flex-wrap gap-3 mb-6">
      <select v-model="filterStatus" @change="loadBets"
        class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer">
        <option value="">全部状态</option>
        <option value="待开奖">待开奖</option>
        <option value="已中奖">已中奖</option>
        <option value="未中奖">未中奖</option>
      </select>
    </div>

    <div class="bg-white rounded-2xl card-shadow overflow-hidden">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="bets.length === 0" class="text-center py-16 text-slate-400">
        <p>暂无投注记录</p>
      </div>

      <template v-else>
        <div class="md:hidden p-4 space-y-3">
          <div v-for="bet in bets" :key="bet.id" class="bg-slate-50 rounded-xl p-4">
            <div class="flex items-start justify-between mb-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium text-slate-800 text-sm">{{ bet.bet_type }}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusColors[bet.status]">
                  {{ bet.status }}
                </span>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button @click="openModal(bet)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                  <Edit2 class="w-4 h-4" />
                </button>
                <button @click="deleteBet(bet.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>

            <div class="space-y-2 mb-2">
              <div v-for="(sel, idx) in parseSelections(bet.selections)" :key="idx" class="text-sm">
                <div class="flex items-center gap-2">
                  <span class="text-slate-500">{{ sel.match_id }}</span>
                  <span class="text-xs px-1.5 py-0.5 bg-slate-200 rounded">{{ sel.play_type }}</span>
                  <span class="font-medium"
                    :class="isSelectionHit(sel) === true ? 'text-emerald-600' : isSelectionHit(sel) === false ? 'text-red-400' : 'text-slate-700'">
                    {{ getOptionLabel(sel.play_type, sel.selection) }}
                  </span>
                  <span v-if="sel.odds > 0" class="text-xs text-amber-500">@{{ sel.odds }}</span>
                  <span v-if="isSelectionHit(sel) === true" class="text-xs text-emerald-500">✓</span>
                  <span v-else-if="isSelectionHit(sel) === false" class="text-xs text-red-400">✗</span>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-4 text-xs text-slate-500">
              <span>¥{{ bet.amount }} × {{ bet.multiple }}</span>
              <span v-if="bet.win_amount > 0" class="font-bold text-emerald-600">中奖 ¥{{ bet.win_amount.toLocaleString() }}</span>
              <span>{{ bet.created_at?.split('T')[0] }}</span>
            </div>
          </div>
        </div>

        <div class="hidden md:block overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">投注类型</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">投注详情</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">金额</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">状态</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">奖金</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">日期</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-slate-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="bet in bets" :key="bet.id" class="hover:bg-slate-50">
                <td class="px-4 py-3 text-sm font-medium text-slate-700">{{ bet.bet_type }}</td>
                <td class="px-4 py-3">
                  <div class="space-y-1">
                    <div v-for="(sel, idx) in parseSelections(bet.selections)" :key="idx" class="text-xs flex items-center gap-1.5">
                      <span class="text-slate-500">{{ sel.match_id }}</span>
                      <span class="px-1 py-0.5 bg-slate-100 rounded text-slate-500">{{ sel.play_type }}</span>
                      <span class="font-medium"
                        :class="isSelectionHit(sel) === true ? 'text-emerald-600' : isSelectionHit(sel) === false ? 'text-red-400' : 'text-slate-700'">
                        {{ getOptionLabel(sel.play_type, sel.selection) }}
                      </span>
                      <span v-if="sel.odds > 0" class="text-amber-500">@{{ sel.odds }}</span>
                      <span v-if="isSelectionHit(sel) === true" class="text-emerald-500">✓</span>
                      <span v-else-if="isSelectionHit(sel) === false" class="text-red-400">✗</span>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3 text-sm text-slate-600">¥{{ bet.amount }} × {{ bet.multiple }}</td>
                <td class="px-4 py-3">
                  <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusColors[bet.status]">
                    {{ bet.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm">
                  <span v-if="bet.win_amount > 0" class="font-bold text-emerald-600">¥{{ bet.win_amount.toLocaleString() }}</span>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="px-4 py-3 text-sm text-slate-500">{{ bet.created_at?.split('T')[0] }}</td>
                <td class="px-4 py-3 text-right">
                  <button @click="openModal(bet)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button @click="deleteBet(bet.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
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
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between z-10">
          <h2 class="text-lg font-semibold text-slate-800">{{ editingId ? '编辑' : '新增' }}投注</h2>
          <button @click="closeModal" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">投注类型</label>
              <select v-model="form.bet_type"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
                <option v-for="bt in ['单关', '2串1', '3串1', '4串1', '5串1', '6串1', '7串1', '8串1']" :key="bt" :value="bt">{{ bt }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">倍数</label>
              <input v-model.number="form.multiple" type="number" min="1"
                class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">投注金额 (元)</label>
            <input v-model.number="form.amount" type="number" step="0.5" min="0"
              :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                errors.amount ? 'border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']" />
            <p v-if="errors.amount" class="mt-1 text-xs text-red-500">⚠ {{ errors.amount }}</p>
          </div>

          <div>
            <FootballBetForm v-model="form.selections" />
            <p v-if="errors.selections" class="mt-1 text-xs text-red-500">⚠ {{ errors.selections }}</p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">备注</label>
            <input v-model="form.remark" placeholder="可选"
              class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
          </div>
        </div>

        <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex justify-end gap-3">
          <button @click="closeModal" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button @click="saveBet" class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer">保存</button>
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
          <p class="text-slate-600 text-sm">确定要删除这条投注记录吗？</p>
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
