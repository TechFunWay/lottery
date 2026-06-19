<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { drawApi, winningApi } from '../api'
import NumberInput from '../components/NumberInput.vue'
import Pagination from '../components/Pagination.vue'
import type { DrawResult, LotteryType } from '../types'
import { LOTTERY_CONFIGS } from '../types'
import { Plus, Trash2, Edit2, X, RefreshCw, Award, AlertCircle, CheckCircle, Download } from 'lucide-vue-next'

const draws = ref<DrawResult[]>([])
const loading = ref(false)
const showModal = ref(false)
const showBatchModal = ref(false)
const editingId = ref<number | null>(null)
const filterType = ref('')
const deleteConfirm = ref(false)
const deleteId = ref<number | null>(null)

// 批量抓取表单
const batchForm = ref({
  lottery_type: '双色球' as LotteryType,
  start_date: '',
  end_date: '',
  count: 10
})
const batchLoading = ref(false)

// 通用弹窗
const toast = ref({
  show: false,
  type: 'info' as 'success' | 'error' | 'info',
  message: ''
})

const showToast = (type: 'success' | 'error' | 'info', message: string) => {
  toast.value = { show: true, type, message }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

const form = ref({
  lottery_type: '双色球' as LotteryType,
  issue_number: '',
  draw_date: new Date().toISOString().split('T')[0],
  numbers: '',
  fu_yun_award: false
})

// 表单校验错误
const errors = reactive({
  issue_number: '',
  numbers: ''
})

const validateForm = () => {
  errors.issue_number = form.value.issue_number.trim() ? '' : '期号为必填项'
  errors.numbers = form.value.numbers ? '' : '请输入号码'
  return !errors.issue_number && !errors.numbers
}

const lotteryTypes = LOTTERY_CONFIGS.map(c => c.type)
const fetchableTypes = ['双色球', '大乐透', '福彩3D', '排列3', '排列5', '七星彩']

const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

watch([currentPage, pageSize], () => {
  loadDraws()
})

const loadDraws = async () => {
  loading.value = true
  const res = await drawApi.list({
    lottery_type: filterType.value,
    page: currentPage.value,
    size: pageSize.value
  }).catch(() => null)
  if (res) {
    draws.value = res.data || []
    total.value = res.total || 0
  }
  loading.value = false
}

onMounted(loadDraws)

const openModal = (item?: DrawResult) => {
  // 清空校验错误
  errors.issue_number = ''
  errors.numbers = ''
  if (item) {
    editingId.value = item.id
    form.value = {
      lottery_type: item.lottery_type,
      issue_number: item.issue_number,
      draw_date: item.draw_date.split('T')[0],
      numbers: item.numbers,
      fu_yun_award: item.fu_yun_award ?? false
    }
  } else {
    editingId.value = null
    form.value = {
      lottery_type: '双色球',
      issue_number: '',
      draw_date: new Date().toISOString().split('T')[0],
      numbers: '',
      fu_yun_award: false
    }
  }
  showModal.value = true
  // 延迟触发一次 watch 确保 NumberInput 正确回填
  setTimeout(() => {
    if (item && form.value.numbers) {
      const temp = form.value.numbers
      form.value.numbers = ''
      form.value.numbers = temp
    }
  }, 0)
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
}

const saveDraw = async () => {
  if (!validateForm()) return
  const payload = {
    lottery_type: form.value.lottery_type,
    issue_number: form.value.issue_number,
    draw_date: form.value.draw_date,
    numbers: form.value.numbers,
    // 福运奖仅双色球适用
    fu_yun_award: form.value.lottery_type === '双色球' ? form.value.fu_yun_award : false
  }
  try {
    if (editingId.value) {
      await drawApi.update(editingId.value, payload)
    } else {
      await drawApi.create(payload)
    }
    closeModal()
    loadDraws()
    // 自动触发中奖检查
    await winningApi.recheck()
    showToast('success', editingId.value ? '修改成功' : '录入成功')
  } catch (e: any) {
    console.error('保存失败', e)
    const errorMsg = e.response?.data?.error || '保存失败，请重试'
    showToast('error', errorMsg)
  }
}

const deleteDraw = (id: number) => {
  deleteId.value = id
  deleteConfirm.value = true
}

const confirmDelete = async () => {
  if (!deleteId.value) return
  try {
    await drawApi.delete(deleteId.value)
    deleteConfirm.value = false
    deleteId.value = null
    loadDraws()
    showToast('success', '删除成功')
  } catch (e) {
    console.error('删除失败', e)
    showToast('error', '删除失败，请重试')
  }
}

const fetchDraw = async () => {
  if (!form.value.lottery_type) {
    showToast('info', '请选择彩票类型')
    return
  }
  loading.value = true
  const res = await drawApi.fetchLatest(form.value.lottery_type, form.value.issue_number).catch((e: any) => {
    showToast('error', '抓取失败：' + (e.response?.data?.error || e.message))
    return null
  })
  loading.value = false
  if (res) {
    const draw = res.data
    form.value.issue_number = draw.issue_number
    form.value.draw_date = draw.draw_date.split('T')[0]
    form.value.numbers = draw.numbers
    // 自动检查中奖
    await winningApi.recheck()
    showToast('success', '抓取成功！')
  }
}

// 格式化号码为带圆圈的样式
interface NumberBall {
  num: string
  type: 'red' | 'blue' | 'main'
}

const formatNumbers = (json: string): NumberBall[] => {
  try {
    const n = JSON.parse(json)
    const balls: NumberBall[] = []
    if (n.red) {
      n.red.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
      balls.push({ num: '|', type: 'main' })
      n.blue.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue' }))
    } else if (n.front) {
      n.front.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
      balls.push({ num: '|', type: 'main' })
      n.back.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'blue' }))
    } else if (n.main) {
      n.main.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'red' }))
    } else if (n.numbers) {
      n.numbers.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    } else if (Array.isArray(n)) {
      n.forEach((x: number) => balls.push({ num: String(x).padStart(2, '0'), type: 'main' }))
    }
    return balls
  } catch { return [] }
}

const sourceColor = (source: string) => {
  return source === 'auto' ? 'bg-blue-100 text-blue-600' : 'bg-amber-100 text-amber-600'
}

// 批量抓取开奖结果
const openBatchModal = () => {
  batchForm.value = {
    lottery_type: '双色球',
    start_date: '',
    end_date: new Date().toISOString().split('T')[0],
    count: 10
  }
  showBatchModal.value = true
}

const closeBatchModal = () => {
  showBatchModal.value = false
}

const fetchBatchDraws = async () => {
  if (!batchForm.value.lottery_type) {
    showToast('error', '请选择彩票类型')
    return
  }
  if (!batchForm.value.start_date && !batchForm.value.end_date && batchForm.value.count <= 0) {
    showToast('error', '请设置时间范围或获取数量')
    return
  }
  
  batchLoading.value = true
  try {
    const res = await drawApi.fetchBatch({
      lottery_type: batchForm.value.lottery_type,
      start_date: batchForm.value.start_date,
      end_date: batchForm.value.end_date,
      count: batchForm.value.count
    })
    const data = res.data
    let msg = `成功获取 ${data?.total || 0} 条，新增 ${data?.count || 0} 条`
    if (data?.exist_count > 0) {
      msg += `，已存在 ${data.exist_count} 条`
    }
    closeBatchModal()
    loadDraws()
    // 自动触发中奖检查
    await winningApi.recheck()
    showToast('success', msg)
  } catch (e: any) {
    console.error('批量抓取失败', e)
    showToast('error', '批量抓取失败：' + (e.response?.data?.error || e.message))
  } finally {
    batchLoading.value = false
  }
}
</script>

<template>
  <div class="animate-fade-in">
    <!-- Header -->
    <div class="mb-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-800">开奖管理</h1>
          <p class="text-slate-400 text-sm mt-1">共 {{ draws.length }} 条开奖记录</p>
        </div>
        <div class="flex items-center gap-3">
          <button
            @click="openBatchModal()"
            class="flex items-center gap-2 px-3 py-2.5 bg-emerald-500 hover:bg-emerald-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-emerald-500/30 cursor-pointer"
          >
            <Download class="w-4 h-4" />
            <span class="hidden sm:inline">批量获取</span>
          </button>
          <button
            @click="openModal()"
            class="flex items-center gap-2 px-4 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-all duration-200 shadow-lg shadow-blue-500/30 cursor-pointer"
          >
            <Plus class="w-4 h-4" />
            <span class="hidden sm:inline">录入开奖</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-6">
      <select
        v-model="filterType"
        @change="loadDraws"
        class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
      >
        <option value="">全部类型</option>
        <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-2xl card-shadow overflow-hidden">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <div v-else-if="draws.length === 0" class="text-center py-16 text-slate-400">
        <Award class="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p>暂无开奖记录</p>
      </div>

      <template v-else>
        <!-- 移动端：卡片列表 -->
        <div class="md:hidden p-4 space-y-3">
          <div v-for="item in draws" :key="item.id" class="bg-slate-50 rounded-xl p-4">
            <div class="flex items-start justify-between mb-2">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium text-slate-800 text-sm">{{ item.lottery_type }}</span>
                <span class="text-slate-400 text-xs">期号 {{ item.issue_number }}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="sourceColor(item.source)">
                  {{ item.source === 'auto' ? '自动' : '手动' }}
                </span>
                <span v-if="item.fu_yun_award" class="px-2 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-600">
                  福运奖
                </span>
              </div>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                  <Edit2 class="w-4 h-4" />
                </button>
                <button @click="deleteDraw(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div class="flex flex-wrap gap-1.5 mb-2">
              <span v-for="(ball, idx) in formatNumbers(item.numbers)" :key="idx"
                class="inline-flex items-center justify-center w-7 h-7 rounded-full text-sm font-bold"
                :class="ball.num === '|'
                  ? 'text-slate-300 text-xs bg-transparent'
                  : ball.type === 'blue'
                    ? 'bg-blue-500 text-white'
                    : 'bg-red-500 text-white'">
                {{ ball.num }}
              </span>
            </div>
            <div class="flex items-center gap-4 text-xs text-slate-500">
              <span>{{ item.draw_date.split('T')[0] }}</span>
            </div>
          </div>
        </div>

        <!-- PC 端：表格 -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">彩票类型</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">期号</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">开奖号码</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">来源</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500">开奖日期</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-slate-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="item in draws" :key="item.id" class="hover:bg-slate-50">
                <td class="px-4 py-3 text-sm font-medium text-slate-700">{{ item.lottery_type }}</td>
                <td class="px-4 py-3 text-sm text-slate-600">{{ item.issue_number }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-wrap gap-1">
                    <span v-for="(ball, idx) in formatNumbers(item.numbers)" :key="idx"
                      class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold"
                      :class="ball.num === '|'
                        ? 'text-slate-300 text-xs bg-transparent'
                        : ball.type === 'blue'
                          ? 'bg-blue-500 text-white'
                          : 'bg-red-500 text-white'">
                      {{ ball.num }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span class="px-2 py-1 rounded-full text-xs font-medium" :class="sourceColor(item.source)">
                      {{ item.source === 'auto' ? '自动抓取' : '手动录入' }}
                    </span>
                    <span v-if="item.fu_yun_award" class="px-2 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-600">
                      福运奖
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 text-sm text-slate-500">{{ item.draw_date.split('T')[0] }}</td>
                <td class="px-4 py-3 text-right">
                  <button @click="openModal(item)" class="p-1.5 text-slate-400 hover:text-blue-500 cursor-pointer">
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button @click="deleteDraw(item.id)" class="p-1.5 text-slate-400 hover:text-red-500 cursor-pointer">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <Pagination
          v-if="total > 0"
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
        />
      </template>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeModal"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto animate-slide-up">
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800">{{ editingId ? '编辑' : '录入' }}开奖结果</h2>
          <button @click="closeModal" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">彩票类型</label>
            <select v-model="form.lottery_type" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">
                期号<span class="text-red-500 ml-0.5">*</span>
              </label>
              <input
                v-model="form.issue_number"
                placeholder="如：2024015"
                @input="errors.issue_number = ''"
                :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                  errors.issue_number ? 'border-red-400 focus:border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']"
              />
              <p v-if="errors.issue_number" class="mt-1 text-xs text-red-500 flex items-center gap-1">
                <span>⚠</span> {{ errors.issue_number }}
              </p>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">开奖日期</label>
              <input v-model="form.draw_date" type="date" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">
              开奖号码<span class="text-red-500 ml-0.5">*</span>
            </label>
            <NumberInput v-model="form.numbers" :lottery-type="form.lottery_type" @update:modelValue="errors.numbers = ''" />
            <p v-if="errors.numbers" class="mt-1 text-xs text-red-500 flex items-center gap-1">
              <span>⚠</span> {{ errors.numbers }}
            </p>
          </div>

          <!-- 双色球福运奖标记 -->
          <div v-if="form.lottery_type === '双色球'" class="rounded-xl bg-amber-50 border border-amber-100 p-3">
            <label class="flex items-start gap-2.5 cursor-pointer">
              <input
                v-model="form.fu_yun_award"
                type="checkbox"
                class="mt-0.5 w-4 h-4 rounded border-slate-300 text-amber-500 focus:ring-amber-400 cursor-pointer"
              />
              <span class="text-sm text-slate-700">
                本期触发福运奖（奖池≥15亿派奖）
                <span class="block text-xs text-slate-400 mt-0.5">勾选后，中3个红球（蓝球未中）将判中福运奖 5 元</span>
              </span>
            </label>
          </div>

          <div v-if="fetchableTypes.includes(form.lottery_type)" class="pt-2">
            <button
              @click="fetchDraw"
              :disabled="loading"
              class="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-600 rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-50"
            >
              <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
              自动抓取开奖结果
            </button>
            <p class="text-xs text-slate-400 mt-2 text-center">支持从官方接口自动获取 {{ form.lottery_type }} 开奖结果</p>
          </div>
        </div>

        <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex justify-end gap-3">
          <button @click="closeModal" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button @click="saveDraw" class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer">保存</button>
        </div>
      </div>
    </div>

    <!-- 批量抓取弹窗 -->
    <div v-if="showBatchModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeBatchModal"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-md animate-slide-up">
        <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-800">批量获取开奖号码</h2>
          <button @click="closeBatchModal" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">彩票类型</label>
            <select v-model="batchForm.lottery_type" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="t in fetchableTypes" :key="t" :value="t">{{ t }}</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">开始日期</label>
              <input v-model="batchForm.start_date" type="date" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-600 mb-1.5">结束日期</label>
              <input v-model="batchForm.end_date" type="date" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">获取数量（最近N期）</label>
            <input v-model.number="batchForm.count" type="number" min="1" max="100" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
            <p class="text-xs text-slate-400 mt-1">设置时间范围或获取数量，两者选其一</p>
          </div>
        </div>

        <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex justify-end gap-3">
          <button @click="closeBatchModal" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button 
            @click="fetchBatchDraws" 
            :disabled="batchLoading"
            class="px-5 py-2 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white rounded-xl font-medium transition-colors cursor-pointer flex items-center gap-2"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': batchLoading }" />
            {{ batchLoading ? '获取中...' : '开始获取' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="deleteConfirm" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="deleteConfirm = false; deleteId = null"></div>
      <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-sm animate-slide-up">
        <div class="p-6 text-center">
          <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Trash2 class="w-6 h-6 text-red-500" />
          </div>
          <h3 class="text-lg font-semibold text-slate-800 mb-2">确认删除</h3>
          <p class="text-slate-600 text-sm">确定要删除这条开奖记录吗？删除后将无法恢复。</p>
        </div>
        <div class="border-t border-slate-100 px-6 py-4 flex gap-3">
          <button
            @click="deleteConfirm = false; deleteId = null"
            class="flex-1 px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer"
          >
            取消
          </button>
          <button
            @click="confirmDelete"
            class="flex-1 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-xl font-medium transition-colors cursor-pointer"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Toast 提示 -->
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
