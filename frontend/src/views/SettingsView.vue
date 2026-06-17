<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { footballConfigApi } from '../api'
import type { FootballConfigStatus, FootballTestResult } from '../types'
import { KeyRound, Save, TestTube2, Trash2, ExternalLink, CheckCircle2, XCircle, Info, RefreshCw } from 'lucide-vue-next'

const status = ref<FootballConfigStatus | null>(null)
const loading = ref(false)
const newKey = ref('')
const saving = ref(false)
const clearing = ref(false)
const testing = ref(false)
const lastTest = ref<FootballTestResult | null>(null)
const message = ref<{ type: 'success' | 'error' | 'info'; text: string } | null>(null)

const sourceLabel = computed(() => {
  if (!status.value) return ''
  switch (status.value.source) {
    case 'user': return '个人配置'
    case 'admin': return '管理员全局配置'
    case 'env': return '环境变量'
    case 'builtin': return '内置(随应用发布)'
    default: return '未配置'
  }
})

const sourceBadge = computed(() => {
  if (!status.value) return 'bg-slate-100 text-slate-600'
  switch (status.value.source) {
    case 'user': return 'bg-emerald-100 text-emerald-700 border-emerald-200'
    case 'admin': return 'bg-blue-100 text-blue-700 border-blue-200'
    case 'env': return 'bg-purple-100 text-purple-700 border-purple-200'
    case 'builtin': return 'bg-slate-100 text-slate-700 border-slate-200'
    default: return 'bg-amber-100 text-amber-700 border-amber-200'
  }
})

const showMessage = (type: 'success' | 'error' | 'info', text: string) => {
  message.value = { type, text }
  setTimeout(() => { message.value = null }, 3500)
}

const loadStatus = async () => {
  loading.value = true
  try {
    const res = await footballConfigApi.getMyStatus()
    status.value = res.data
  } catch (err: any) {
    showMessage('error', err.response?.data?.error || '加载状态失败')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  const key = newKey.value.trim()
  if (!key) {
    showMessage('error', '请输入 Key')
    return
  }
  saving.value = true
  try {
    const res = await footballConfigApi.setMyKey(key)
    showMessage('success', res.message)
    newKey.value = ''
    await loadStatus()
  } catch (err: any) {
    showMessage('error', err.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleTest = async () => {
  testing.value = true
  lastTest.value = null
  try {
    const res = await footballConfigApi.testKey(newKey.value.trim())
    lastTest.value = res.data
  } catch (err: any) {
    lastTest.value = { success: false, message: err.response?.data?.error || '测试失败' }
  } finally {
    testing.value = false
  }
}

const handleClear = async () => {
  if (!confirm('确定要清除你自配的 Key 吗?清除后将降级使用管理员/环境变量配置。')) return
  clearing.value = true
  try {
    const res = await footballConfigApi.setMyKey('')
    showMessage('success', res.message)
    await loadStatus()
  } catch (err: any) {
    showMessage('error', err.response?.data?.error || '清除失败')
  } finally {
    clearing.value = false
  }
}

onMounted(loadStatus)
</script>

<template>
  <div class="max-w-3xl mx-auto p-4 sm:p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-slate-800 flex items-center gap-2">
        <KeyRound class="w-6 h-6 text-blue-600" />
        数据源设置
      </h1>
      <p class="text-sm text-slate-500 mt-1">配置 API-Football Key 以启用「一键抓取开奖结果」自动验证中奖</p>
    </div>

    <!-- Toast 提示 -->
    <div
      v-if="message"
      class="mb-4 p-3 rounded-lg text-sm"
      :class="{
        'bg-emerald-50 border border-emerald-200 text-emerald-700': message.type === 'success',
        'bg-red-50 border border-red-200 text-red-700': message.type === 'error',
        'bg-blue-50 border border-blue-200 text-blue-700': message.type === 'info',
      }"
    >
      {{ message.text }}
    </div>

    <!-- 当前状态卡片 -->
    <div class="bg-white rounded-2xl card-shadow p-6 mb-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-slate-800">当前状态</h2>
        <button
          @click="loadStatus"
          :disabled="loading"
          class="p-1.5 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors disabled:opacity-50"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
      </div>

      <div v-if="status" class="space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-sm text-slate-600">是否配置</span>
          <span v-if="status.configured" class="inline-flex items-center gap-1 text-sm font-medium text-emerald-600">
            <CheckCircle2 class="w-4 h-4" />
            已配置
          </span>
          <span v-else class="inline-flex items-center gap-1 text-sm font-medium text-amber-600">
            <XCircle class="w-4 h-4" />
            未配置
          </span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-sm text-slate-600">Key 来源</span>
          <span :class="['inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border', sourceBadge]">
            {{ sourceLabel }}
          </span>
        </div>
        <div v-if="status.masked_key" class="flex items-center justify-between">
          <span class="text-sm text-slate-600">当前 Key</span>
          <code class="text-xs text-slate-700 font-mono bg-slate-100 px-2 py-1 rounded">{{ status.masked_key }}</code>
        </div>
      </div>
      <div v-else class="text-sm text-slate-400">加载中...</div>
    </div>

    <!-- 配置 Key -->
    <div class="bg-white rounded-2xl card-shadow p-6 mb-6">
      <h2 class="text-lg font-semibold text-slate-800 mb-1">配置我的 Key</h2>
      <p class="text-xs text-slate-500 mb-4">留空则保存即清除;每人最多配置 1 个 Key,优先级最高</p>

      <label class="block text-sm font-medium text-slate-700 mb-1">API-Football Key</label>
      <input
        v-model="newKey"
        type="password"
        placeholder="例如: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all font-mono text-sm"
        @keyup.enter="handleSave"
      />

      <!-- 测试结果 -->
      <div v-if="lastTest" class="mt-3 p-3 rounded-lg text-sm flex items-start gap-2"
        :class="lastTest.success ? 'bg-emerald-50 border border-emerald-200 text-emerald-700' : 'bg-red-50 border border-red-200 text-red-700'">
        <CheckCircle2 v-if="lastTest.success" class="w-4 h-4 flex-shrink-0 mt-0.5" />
        <XCircle v-else class="w-4 h-4 flex-shrink-0 mt-0.5" />
        <span>{{ lastTest.message }}</span>
      </div>

      <div class="flex flex-wrap items-center gap-2 mt-4">
        <button
          @click="handleSave"
          :disabled="saving || !newKey.trim()"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
          <Save v-else class="w-4 h-4" />
          保存
        </button>
        <button
          @click="handleTest"
          :disabled="testing"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-amber-50 hover:bg-amber-100 text-amber-700 border border-amber-200 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          <RefreshCw v-if="testing" class="w-4 h-4 animate-spin" />
          <TestTube2 v-else class="w-4 h-4" />
          {{ newKey.trim() ? '测试输入的 Key' : '测试当前 Key' }}
        </button>
        <button
          v-if="status?.source === 'user'"
          @click="handleClear"
          :disabled="clearing"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-red-50 hover:bg-red-100 text-red-700 border border-red-200 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          <RefreshCw v-if="clearing" class="w-4 h-4 animate-spin" />
          <Trash2 v-else class="w-4 h-4" />
          清除我的 Key
        </button>
      </div>
    </div>

    <!-- 注册说明 -->
    <div class="bg-blue-50 border border-blue-200 rounded-2xl p-5 flex items-start gap-3">
      <Info class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
      <div class="text-sm text-blue-900">
        <p class="font-medium mb-1">如何获取 API-Football Key?</p>
        <ol class="list-decimal pl-5 space-y-1 text-blue-800">
          <li>访问 <a href="https://dashboard.api-football.com/prod/register" target="_blank" rel="noopener" class="inline-flex items-center gap-0.5 underline font-medium">api-football.com 注册 <ExternalLink class="w-3 h-3" /></a>(邮箱注册,无需信用卡)</li>
          <li>登录后在 Dashboard 复制 API Key</li>
          <li>粘贴到上方输入框,点击「测试」验证,「保存」即可</li>
        </ol>
        <p class="text-xs text-blue-700 mt-2">免费版: <b>100 次/天</b> · 10 次/分,够个人日常使用。多人共享时建议每人配自己的 Key 以避免限流。</p>
      </div>
    </div>
  </div>
</template>
