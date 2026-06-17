<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { userApi, configApi, footballConfigApi } from '../api'
import { hashPassword } from '../utils/crypto'
import type { User, SystemConfig, FootballTestResult } from '../types'
import {
  Users, Lock, Unlock, Shield, ShieldCheck, Trash2, RefreshCw,
  Settings, ToggleLeft, ToggleRight, Save, KeyRound, X, CheckCircle2, AlertCircle, TestTube2, Info, ExternalLink,
} from 'lucide-vue-next'

// ===== 用户管理 =====
const users = ref<User[]>([])
const loadingUsers = ref(false)
const userError = ref('')
const userSuccess = ref('')
const updating = ref<number | null>(null)
const deleting = ref<number | null>(null)

// 重置密码弹窗
const resettingUser = ref<User | null>(null)
const newPassword = ref('')
const resettingPassword = ref(false)

// 弹窗提示
const showAlert = ref(false)
const alertMessage = ref('')
const alertType = ref<'success' | 'error'>('success')

const fetchUsers = async () => {
  loadingUsers.value = true
  userError.value = ''
  try {
    const res = await userApi.getAll()
    users.value = res.data
  } catch (err: any) {
    userError.value = err.response?.data?.error || '获取用户列表失败'
  } finally {
    loadingUsers.value = false
  }
}

const toggleUserStatus = async (user: User) => {
  if (deleting.value) return
  updating.value = user.id
  userError.value = ''
  userSuccess.value = ''
  try {
    const newStatus = user.status === 'active' ? 'disabled' : 'active'
    await userApi.update(user.id, { status: newStatus })
    user.status = newStatus
    userSuccess.value = `${user.username} 已${newStatus === 'active' ? '启用' : '禁用'}`
    setTimeout(() => (userSuccess.value = ''), 3000)
  } catch (err: any) {
    userError.value = err.response?.data?.error || '更新失败'
  } finally {
    updating.value = null
  }
}

const deleteUser = async (user: User) => {
  if (!confirm(`确定要删除用户 "${user.username}" 吗？此操作不可恢复。`)) return
  deleting.value = user.id
  userError.value = ''
  userSuccess.value = ''
  try {
    await userApi.delete(user.id)
    users.value = users.value.filter(u => u.id !== user.id)
    showAlertMessage(`用户 "${user.username}" 已删除`, 'success')
  } catch (err: any) {
    showAlertMessage(err.response?.data?.error || '删除失败', 'error')
  } finally {
    deleting.value = null
  }
}

// 打开重置密码弹窗
const openResetPassword = (user: User) => {
  resettingUser.value = user
  newPassword.value = ''
}

// 关闭重置密码弹窗
const closeResetPassword = () => {
  resettingUser.value = null
  newPassword.value = ''
}

// 确认重置密码
const confirmResetPassword = async () => {
  if (!resettingUser.value) return
  if (!newPassword.value || newPassword.value.length < 6) {
    showAlertMessage('密码长度至少为6位', 'error')
    return
  }

  resettingPassword.value = true
  try {
    // 前端先 MD5 加密一次，后端会再做一次 MD5+盐
    const hashedPassword = hashPassword(newPassword.value)
    await userApi.update(resettingUser.value.id, { password: hashedPassword })
    showAlertMessage(`用户 "${resettingUser.value.username}" 的密码已重置`, 'success')
    closeResetPassword()
  } catch (err: any) {
    showAlertMessage(err.response?.data?.error || '重置密码失败', 'error')
  } finally {
    resettingPassword.value = false
  }
}

// 显示弹窗提示
const showAlertMessage = (message: string, type: 'success' | 'error') => {
  alertMessage.value = message
  alertType.value = type
  showAlert.value = true
  setTimeout(() => {
    showAlert.value = false
  }, 3000)
}

// ===== 系统配置 =====
const configs = ref<SystemConfig[]>([])
const loadingConfig = ref(false)
const savingConfig = ref(false)
const configError = ref('')
const configSuccess = ref('')

// 本地配置状态（用于双向绑定）
const allowRegister = ref(true)

const REMARK_MAP: Record<string, string> = {
  allow_register: '允许用户自主注册',
}

const fetchConfigs = async () => {
  loadingConfig.value = true
  configError.value = ''
  try {
    const res = await configApi.getAll()
    configs.value = res.data
    // 同步到本地状态
    res.data.forEach(cfg => {
      if (cfg.key === 'allow_register') allowRegister.value = cfg.value === 'true'
    })
  } catch (err: any) {
    configError.value = err.response?.data?.error || '获取配置失败'
  } finally {
    loadingConfig.value = false
  }
}

const saveConfigs = async () => {
  savingConfig.value = true
  try {
    const updates: SystemConfig[] = [
      { key: 'allow_register', value: allowRegister.value ? 'true' : 'false' },
    ]
    await configApi.updateBatch(updates)
    showAlertMessage('配置已保存', 'success')
  } catch (err: any) {
    showAlertMessage(err.response?.data?.error || '保存配置失败', 'error')
  } finally {
    savingConfig.value = false
  }
}

// ===== 公共 =====
const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

const getRoleBadge = (role: string) =>
  role === 'admin'
    ? { class: 'bg-purple-100 text-purple-700 border-purple-200', icon: ShieldCheck, label: '管理员' }
    : { class: 'bg-slate-100 text-slate-700 border-slate-200', icon: Shield, label: '普通用户' }

const getStatusBadge = (status: string) =>
  status === 'active'
    ? { class: 'bg-green-100 text-green-700 border-green-200', icon: Unlock, label: '正常' }
    : { class: 'bg-red-100 text-red-700 border-red-200', icon: Lock, label: '禁用' }

const isAdmin = computed(() => {
  const userStr = localStorage.getItem('user')
  if (!userStr) return false
  return JSON.parse(userStr).role === 'admin'
})

// ===== 竞彩足球数据源(全局 key) =====
const globalKeyInput = ref('')
const globalKeyStatus = ref<{ configured: boolean; source: string; masked_key: string } | null>(null)
const globalKeyLoading = ref(false)
const globalKeySaving = ref(false)
const globalKeyClearing = ref(false)
const globalKeyTesting = ref(false)
const globalKeyTestResult = ref<FootballTestResult | null>(null)

const loadGlobalKeyStatus = async () => {
  globalKeyLoading.value = true
  try {
    const res = await footballConfigApi.getSystemStatus()
    globalKeyStatus.value = res.data
  } catch (err: any) {
    console.error('加载全局 Key 状态失败:', err)
  } finally {
    globalKeyLoading.value = false
  }
}

const saveGlobalKey = async () => {
  if (!globalKeyInput.value.trim()) {
    showAlertMessage('请输入 Key', 'error')
    return
  }
  globalKeySaving.value = true
  try {
    const res = await footballConfigApi.setGlobalKey(globalKeyInput.value.trim())
    showAlertMessage(res.message, 'success')
    globalKeyInput.value = ''
    globalKeyTestResult.value = null
    await loadGlobalKeyStatus()
  } catch (err: any) {
    showAlertMessage(err.response?.data?.error || '保存失败', 'error')
  } finally {
    globalKeySaving.value = false
  }
}

const testGlobalKey = async () => {
  globalKeyTesting.value = true
  globalKeyTestResult.value = null
  try {
    const res = await footballConfigApi.testKey(globalKeyInput.value.trim())
    globalKeyTestResult.value = res.data
  } catch (err: any) {
    globalKeyTestResult.value = { success: false, message: err.response?.data?.error || '测试失败' }
  } finally {
    globalKeyTesting.value = false
  }
}

const clearGlobalKey = async () => {
  if (!confirm('确定要清除全局 Key 吗?清除后仅自配用户的 Key 仍可工作。')) return
  globalKeyClearing.value = true
  try {
    const res = await footballConfigApi.setGlobalKey('')
    showAlertMessage(res.message, 'success')
    globalKeyTestResult.value = null
    await loadGlobalKeyStatus()
  } catch (err: any) {
    showAlertMessage(err.response?.data?.error || '清除失败', 'error')
  } finally {
    globalKeyClearing.value = false
  }
}

// 当前激活的 tab
const activeTab = ref<'users' | 'configs'>('users')

onMounted(() => {
  if (isAdmin.value) {
    fetchUsers()
    fetchConfigs()
    loadGlobalKeyStatus()
  }
})
</script>

<template>
  <div v-if="isAdmin">
    <!-- 标题 + Tab -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-3">
      <div class="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-4">
        <h1 class="text-xl sm:text-2xl font-bold text-slate-800 flex items-center gap-2">
          <Users class="w-5 h-5 sm:w-6 sm:h-6 text-blue-600" />
          管理中心
        </h1>
        <!-- Tab 切换 -->
        <div class="flex gap-0.5 bg-slate-100 rounded-lg p-0.5 w-fit">
          <button
            @click="activeTab = 'users'"
            :class="[
              'flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all',
              activeTab === 'users'
                ? 'bg-white text-blue-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-800',
            ]"
          >
            <Users class="w-3.5 h-3.5" />
            用户管理
          </button>
          <button
            @click="activeTab = 'configs'"
            :class="[
              'flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all',
              activeTab === 'configs'
                ? 'bg-white text-blue-600 shadow-sm'
                : 'text-slate-600 hover:text-slate-800',
            ]"
          >
            <Settings class="w-3.5 h-3.5" />
            系统配置
          </button>
        </div>
      </div>

      <button
        v-if="activeTab === 'users'"
        @click="fetchUsers"
        :disabled="loadingUsers"
        class="flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg transition-colors text-sm disabled:opacity-50 w-fit"
      >
        <RefreshCw :class="{ 'animate-spin': loadingUsers }" class="w-3.5 h-3.5" />
        刷新
      </button>
    </div>

    <!-- ===== 用户管理面板 ===== -->
    <div v-if="activeTab === 'users'">
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <div v-if="loadingUsers" class="flex items-center justify-center py-12">
          <div class="w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
        </div>

        <div v-else-if="users.length === 0" class="flex flex-col items-center justify-center py-12 text-slate-400">
          <Users class="w-12 h-12 mb-3 opacity-50" />
          <p>暂无用户</p>
        </div>

        <!-- 桌面端：表格布局 -->
        <div v-if="users.length > 0" class="hidden sm:block overflow-x-auto sm:overflow-visible">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-slate-50 text-slate-500 text-left">
                <th class="px-4 py-3 font-medium">用户</th>
                <th class="px-4 py-3 font-medium">邮箱</th>
                <th class="px-4 py-3 font-medium">角色</th>
                <th class="px-4 py-3 font-medium">状态</th>
                <th class="px-4 py-3 font-medium">最后登录</th>
                <th class="px-4 py-3 font-medium">注册时间</th>
                <th class="px-4 py-3 font-medium text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2.5">
                    <div class="w-8 h-8 rounded-lg gradient-primary flex-shrink-0 flex items-center justify-center text-white text-xs font-bold">
                      {{ user.username.charAt(0).toUpperCase() }}
                    </div>
                    <span class="font-medium text-slate-800">{{ user.username }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 text-slate-500">{{ user.email || '-' }}</td>
                <td class="px-4 py-3">
                  <span
                    :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium border', getRoleBadge(user.role).class]"
                  >
                    <component :is="getRoleBadge(user.role).icon" class="w-3 h-3" />
                    {{ getRoleBadge(user.role).label }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium border', getStatusBadge(user.status).class]"
                  >
                    <component :is="getStatusBadge(user.status).icon" class="w-3 h-3" />
                    {{ getStatusBadge(user.status).label }}
                  </span>
                </td>
                <td class="px-4 py-3 text-slate-500 whitespace-nowrap">{{ formatDate(user.last_login) }}</td>
                <td class="px-4 py-3 text-slate-500 whitespace-nowrap">{{ formatDate(user.created_at) }}</td>
                <td class="px-4 py-3">
                  <div v-if="user.role !== 'admin'" class="flex items-center gap-1.5 justify-end">
                    <button
                      @click="openResetPassword(user)"
                      :disabled="updating === user.id || deleting === user.id"
                      class="flex items-center gap-1 px-2 py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-md text-xs font-medium transition-all disabled:opacity-50"
                    >
                      <KeyRound class="w-3 h-3" />
                      重置密码
                    </button>
                    <button
                      @click="toggleUserStatus(user)"
                      :disabled="updating === user.id || deleting === user.id"
                      :class="[
                        'flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium transition-all disabled:opacity-50',
                        user.status === 'active'
                          ? 'bg-amber-50 hover:bg-amber-100 text-amber-700'
                          : 'bg-green-50 hover:bg-green-100 text-green-700',
                      ]"
                    >
                      <RefreshCw v-if="updating === user.id" class="w-3 h-3 animate-spin" />
                      <Lock v-else-if="user.status === 'active'" class="w-3 h-3" />
                      <Unlock v-else class="w-3 h-3" />
                      {{ user.status === 'active' ? '禁用' : '启用' }}
                    </button>
                    <button
                      @click="deleteUser(user)"
                      :disabled="updating === user.id || deleting === user.id"
                      class="flex items-center gap-1 px-2 py-1 bg-red-50 hover:bg-red-100 text-red-700 rounded-md text-xs font-medium transition-all disabled:opacity-50"
                    >
                      <RefreshCw v-if="deleting === user.id" class="w-3 h-3 animate-spin" />
                      <Trash2 v-else class="w-3 h-3" />
                      删除
                    </button>
                  </div>
                  <span v-else class="text-xs text-slate-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 移动端：卡片列表 -->
        <div v-if="users.length > 0" class="sm:hidden divide-y divide-slate-100">
          <div v-for="user in users" :key="user.id" class="p-3 hover:bg-slate-50 transition-colors">
            <!-- 用户名 + 角色/状态 -->
            <div class="flex items-center gap-2 mb-1.5">
              <div class="w-8 h-8 rounded-lg gradient-primary flex-shrink-0 flex items-center justify-center text-white text-xs font-bold">
                {{ user.username.charAt(0).toUpperCase() }}
              </div>
              <span class="font-medium text-slate-800 text-sm">{{ user.username }}</span>
              <span
                :class="['inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-medium border', getRoleBadge(user.role).class]"
              >
                {{ getRoleBadge(user.role).label }}
              </span>
              <span
                :class="['inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-medium border', getStatusBadge(user.status).class]"
              >
                {{ getStatusBadge(user.status).label }}
              </span>
            </div>
            <!-- 时间信息 -->
            <div class="text-xs text-slate-400 mb-2 pl-10">
              <span>最后登录: {{ formatDate(user.last_login) }}</span>
              <span class="ml-3">注册: {{ formatDate(user.created_at) }}</span>
            </div>
            <!-- 操作按钮 -->
            <div v-if="user.role !== 'admin'" class="flex items-center gap-2 pl-10">
              <button
                @click="openResetPassword(user)"
                :disabled="updating === user.id || deleting === user.id"
                class="flex items-center gap-1 px-2.5 py-1 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-md text-xs font-medium transition-all disabled:opacity-50"
              >
                <KeyRound class="w-3 h-3" />
                重置密码
              </button>
              <button
                @click="toggleUserStatus(user)"
                :disabled="updating === user.id || deleting === user.id"
                :class="[
                  'flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium transition-all disabled:opacity-50',
                  user.status === 'active'
                    ? 'bg-amber-50 hover:bg-amber-100 text-amber-700'
                    : 'bg-green-50 hover:bg-green-100 text-green-700',
                ]"
              >
                <RefreshCw v-if="updating === user.id" class="w-3 h-3 animate-spin" />
                <Lock v-else-if="user.status === 'active'" class="w-3 h-3" />
                <Unlock v-else class="w-3 h-3" />
                {{ user.status === 'active' ? '禁用' : '启用' }}
              </button>
              <button
                @click="deleteUser(user)"
                :disabled="updating === user.id || deleting === user.id"
                class="flex items-center gap-1 px-2.5 py-1 bg-red-50 hover:bg-red-100 text-red-700 rounded-md text-xs font-medium transition-all disabled:opacity-50"
              >
                <RefreshCw v-if="deleting === user.id" class="w-3 h-3 animate-spin" />
                <Trash2 v-else class="w-3 h-3" />
                删除
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="mt-2 text-xs text-slate-400">共 {{ users.length }} 个用户</div>
    </div>

    <!-- ===== 系统配置面板 ===== -->
    <div v-if="activeTab === 'configs'">
      <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h2 class="font-semibold text-slate-800 flex items-center gap-2">
            <Settings class="w-5 h-5 text-slate-500" />
            全局配置
          </h2>
        </div>

        <div v-if="loadingConfig" class="flex items-center justify-center py-12">
          <div class="w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
        </div>

        <div v-else class="divide-y divide-slate-100">
          <!-- 允许注册 -->
          <div class="flex items-center justify-between px-6 py-5">
            <div>
              <div class="font-medium text-slate-800 mb-0.5">允许用户注册</div>
              <div class="text-sm text-slate-500">开启后，任何人均可在登录页自主注册账号</div>
            </div>
            <button
              @click="allowRegister = !allowRegister"
              :class="['flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
                allowRegister
                  ? 'bg-blue-50 text-blue-700 hover:bg-blue-100'
                  : 'bg-slate-100 text-slate-500 hover:bg-slate-200'
              ]"
            >
              <ToggleRight v-if="allowRegister" class="w-5 h-5" />
              <ToggleLeft v-else class="w-5 h-5" />
              {{ allowRegister ? '已开启' : '已关闭' }}
            </button>
          </div>
        </div>

        <!-- 保存按钮 -->
        <div class="px-6 py-4 border-t border-slate-100 flex justify-end">
          <button
            @click="saveConfigs"
            :disabled="savingConfig"
            class="flex items-center gap-2 px-6 py-2.5 bg-gradient-to-r from-blue-600 to-emerald-500 text-white font-medium rounded-lg hover:shadow-lg transition-all disabled:opacity-50"
          >
            <span v-if="savingConfig" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            <Save v-else class="w-4 h-4" />
            保存配置
          </button>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-2xl card-shadow overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
            <KeyRound class="w-5 h-5 text-amber-500" />
            全局 API-Football Key
          </h2>
          <p class="text-xs text-slate-500 mt-1">供未自配的用户降级使用 · 修改后立即生效,所有用户都受益</p>
        </div>
        <button @click="loadGlobalKeyStatus" :disabled="globalKeyLoading"
          class="p-1.5 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors disabled:opacity-50">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': globalKeyLoading }" />
        </button>
      </div>

      <div class="px-6 py-4 space-y-4">
        <div v-if="globalKeyStatus" class="grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
          <div class="flex items-center justify-between sm:block">
            <span class="text-slate-500">状态</span>
            <span v-if="globalKeyStatus.configured" class="ml-2 sm:ml-0 inline-flex items-center gap-1 text-emerald-600 font-medium">
              <CheckCircle2 class="w-3.5 h-3.5" />已配置
            </span>
            <span v-else class="ml-2 sm:ml-0 inline-flex items-center gap-1 text-amber-600 font-medium">
              <Info class="w-3.5 h-3.5" />未配置
            </span>
          </div>
          <div class="flex items-center justify-between sm:block">
            <span class="text-slate-500">来源</span>
            <span class="ml-2 sm:ml-0 text-slate-700 font-medium">
              {{ globalKeyStatus.source === 'admin' ? '数据库' : globalKeyStatus.source === 'env' ? '环境变量' : globalKeyStatus.source === 'builtin' ? '内置(随应用发布)' : '无' }}
            </span>
          </div>
          <div v-if="globalKeyStatus.masked_key" class="flex items-center justify-between sm:block">
            <span class="text-slate-500">当前 Key</span>
            <code class="ml-2 sm:ml-0 text-xs font-mono bg-slate-100 px-1.5 py-0.5 rounded">{{ globalKeyStatus.masked_key }}</code>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">新 Key</label>
          <input
            v-model="globalKeyInput"
            type="password"
            placeholder="留空保存即清除"
            class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:ring-2 focus:ring-amber-500 focus:border-transparent outline-none transition-all font-mono text-sm"
            @keyup.enter="saveGlobalKey"
          />
        </div>

        <div v-if="globalKeyTestResult" class="p-3 rounded-lg text-sm flex items-start gap-2"
          :class="globalKeyTestResult.success ? 'bg-emerald-50 border border-emerald-200 text-emerald-700' : 'bg-red-50 border border-red-200 text-red-700'">
          <CheckCircle2 v-if="globalKeyTestResult.success" class="w-4 h-4 flex-shrink-0 mt-0.5" />
          <Info v-else class="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span>{{ globalKeyTestResult.message }}</span>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button @click="saveGlobalKey" :disabled="globalKeySaving || !globalKeyInput.trim()"
            class="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
            <RefreshCw v-if="globalKeySaving" class="w-4 h-4 animate-spin" />
            <Save v-else class="w-4 h-4" />
            保存
          </button>
          <button @click="testGlobalKey" :disabled="globalKeyTesting || !globalKeyInput.trim()"
            class="inline-flex items-center gap-1.5 px-4 py-2 bg-amber-50 hover:bg-amber-100 text-amber-700 border border-amber-200 rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
            <RefreshCw v-if="globalKeyTesting" class="w-4 h-4 animate-spin" />
            <TestTube2 v-else class="w-4 h-4" />
            测试
          </button>
          <button v-if="globalKeyStatus?.configured" @click="clearGlobalKey" :disabled="globalKeyClearing"
            class="inline-flex items-center gap-1.5 px-4 py-2 bg-red-50 hover:bg-red-100 text-red-700 border border-red-200 rounded-lg text-sm font-medium transition-colors disabled:opacity-50">
            <RefreshCw v-if="globalKeyClearing" class="w-4 h-4 animate-spin" />
            <Trash2 v-else class="w-4 h-4" />
            清除全局 Key
          </button>
        </div>

        <div class="flex items-start gap-2 p-3 bg-amber-50 border border-amber-100 rounded-lg text-xs text-amber-800">
          <Info class="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span>
            还没注册?
            <a href="https://dashboard.api-football.com/prod/register" target="_blank" rel="noopener"
              class="inline-flex items-center gap-0.5 underline font-medium">
              api-football.com 注册免费 Key <ExternalLink class="w-3 h-3" />
            </a>
            (邮箱注册,无需信用卡,100 次/天)
          </span>
        </div>
      </div>
    </div>

  </div>

  <!-- 非管理员 -->
  <div v-else class="flex flex-col items-center justify-center py-16 text-slate-500">
    <Lock class="w-16 h-16 mb-4 opacity-50" />
    <h2 class="text-xl font-semibold mb-2">需要管理员权限</h2>
    <p>只有管理员可以访问此页面</p>
  </div>

  <!-- 弹窗提示 -->
  <transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0 translate-y-4"
    enter-to-class="opacity-100 translate-y-0"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100 translate-y-0"
    leave-to-class="opacity-0 translate-y-4"
  >
    <div
      v-if="showAlert"
      class="fixed top-24 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-6 py-4 rounded-xl shadow-2xl"
      :class="alertType === 'success' ? 'bg-white border-2 border-green-200' : 'bg-white border-2 border-red-200'"
    >
      <div class="flex-shrink-0" :class="alertType === 'success' ? 'text-green-600' : 'text-red-600'">
        <CheckCircle2 v-if="alertType === 'success'" class="w-6 h-6" />
        <AlertCircle v-else class="w-6 h-6" />
      </div>
      <p class="text-sm font-medium" :class="alertType === 'success' ? 'text-slate-800' : 'text-red-700'">
        {{ alertMessage }}
      </p>
    </div>
  </transition>

  <!-- 重置密码弹窗 -->
  <transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="resettingUser"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="closeResetPassword"
    >
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
        <!-- 标题 -->
        <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-slate-800">重置密码</h3>
          <button @click="closeResetPassword" class="text-slate-400 hover:text-slate-600 transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- 内容 -->
        <div class="p-6">
          <p class="text-slate-600 mb-4">
            为用户 <span class="font-medium text-slate-800">{{ resettingUser.username }}</span> 设置新密码
          </p>

          <div class="mb-4">
            <label class="block text-sm font-medium text-slate-700 mb-2">新密码</label>
            <div class="relative">
              <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                v-model="newPassword"
                type="password"
                placeholder="请输入新密码（至少6位）"
                class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
                @keyup.enter="confirmResetPassword"
              />
            </div>
          </div>
        </div>

        <!-- 按钮 -->
        <div class="px-6 py-4 border-t border-slate-200 flex justify-end gap-3">
          <button
            @click="closeResetPassword"
            class="px-4 py-2 text-slate-700 hover:bg-slate-100 rounded-lg transition-colors text-sm font-medium"
          >
            取消
          </button>
          <button
            @click="confirmResetPassword"
            :disabled="resettingPassword"
            class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium disabled:opacity-50"
          >
            <span v-if="resettingPassword" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            <span v-else>确认重置</span>
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>
