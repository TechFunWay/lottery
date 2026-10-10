<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { authApi, configApi } from '../api'
import { hashPassword } from '../utils/crypto'
import { User, Lock, ArrowRight, Gift, Mail, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()

// 表单状态
const username = ref('')
const password = ref('')
const email = ref('')
const loading = ref(false)
const isRegisterMode = ref(false)
const rememberMe = ref(false)  // 记住密码
const fnosEnabled = import.meta.env.VITE_FNOS_APP === 'true'
const fnosBinding = ref(false)
const fnosUsername = ref('')
const fnosMode = ref<'register' | 'bind'>('register')

// ===== 飞牛授权确认 =====
// 应用运行在直连端口上，NAS 身份头只在网关域存在。登录页点击「飞牛授权登录」
// 整页跳到网关跳板页（fnos-entry.html）签发一次性票据后跳回本页，先弹窗让用户
// 确认当前 NAS 账号，确认后才凭票据（X-FnOS-Ticket 头，见 api 拦截器）完成登录/绑定。
const showFnOSConfirm = ref(false)
const fnosConfirmUsername = ref('')
// 票据过期等失败后置为 true，在当前卡片上提供「重新授权」入口
const fnosNeedsReauth = ref(false)

function clearFnOSTicket() {
  sessionStorage.removeItem('fnos_ticket')
  sessionStorage.removeItem('fnos_ticket_username')
}

// 票据只能由网关连接签发（直连端口会 401），因此同源签发成功即可确认当前页面
// 运行在网关域——不能用 URL 路径判断：直连端口的页面路径同样带 /app/ 前缀。
async function tryIssueTicketLocally(): Promise<boolean> {
  try {
    const res = await authApi.fnosTicket()
    if (res.data?.ticket) {
      sessionStorage.setItem('fnos_ticket', res.data.ticket)
      if (res.data.fnos_username) sessionStorage.setItem('fnos_ticket_username', res.data.fnos_username)
      return true
    }
  } catch { /* 直连端口：走网关跳板页 */ }
  return false
}

// 错误/成功提示
const loginError = ref('')
const registerError = ref('')
const registerSuccess = ref(false)

// 系统配置
const allowRegister = ref(true)   // 是否允许注册（来自后端配置）
const adminExists = ref(true)     // 管理员是否存在

// 初始化：获取公开配置 + 管理员状态 + 加载记住的密码
onMounted(async () => {
  // 加载记住的密码
  loadRememberedCredentials()

  // 网关跳板页带回票据：弹出授权确认框（未登录时才有意义）
  if (fnosEnabled && sessionStorage.getItem('fnos_ticket') && !localStorage.getItem('token')) {
    openFnOSConfirm()
  }

  try {
    const [cfgRes, adminRes] = await Promise.all([
      configApi.getPublic(),
      authApi.checkAdmin(),
    ])
    allowRegister.value = cfgRes.data.allow_register
    adminExists.value = adminRes.exists
    // 若尚未有管理员，直接跳到注册页
    if (!adminRes.exists) {
      isRegisterMode.value = true
    }
  } catch (e) {
    console.error('初始化配置失败:', e)
  }
})

// 点击「飞牛授权登录」：先尝试在当前页面直接签发票据（网关域打开时直接成功）；
// 直连端口则整页跳到网关跳板页签发票据后跳回本页。
async function handleFnOSLogin() {
  loading.value = true
  if (await tryIssueTicketLocally()) {
    loading.value = false
    openFnOSConfirm()
    return
  }
  const gatewayURL = localStorage.getItem('fnos_gateway_url')
  if (!gatewayURL) {
    loading.value = false
    const msg = '请先从飞牛桌面打开本应用，再使用飞牛授权登录'
    loginError.value = msg
    registerError.value = msg
    return
  }
  window.location.href = gatewayURL
}

function openFnOSConfirm() {
  fnosConfirmUsername.value = sessionStorage.getItem('fnos_ticket_username') || ''
  showFnOSConfirm.value = true
}

function cancelFnOSConfirm() {
  showFnOSConfirm.value = false
  clearFnOSTicket()
}

// 切换飞牛账号：跳到飞牛桌面的登录页，登录成功后经 redirect_uri 回到本页。
// 桌面登录页始终在网关域：优先取记忆的网关地址；没有记忆但本页同源能签发票据，
// 说明当前就在网关域，直接用当前 origin。
async function switchFnOSAccount() {
  showFnOSConfirm.value = false
  clearFnOSTicket()
  const stored = localStorage.getItem('fnos_gateway_url')
  let desktopOrigin: string | null = null
  if (stored) {
    desktopOrigin = new URL(stored).origin
  } else if (await tryIssueTicketLocally()) {
    desktopOrigin = window.location.origin
    localStorage.setItem('fnos_gateway_url', `${desktopOrigin}${import.meta.env.BASE_URL}fnos-entry.html`)
  }
  if (!desktopOrigin) {
    loginError.value = '无法定位飞牛桌面入口，请从飞牛桌面重新打开本应用'
    return
  }
  const appLoginPath = `${window.location.origin}${import.meta.env.BASE_URL}login`
  window.location.href = `${desktopOrigin}/login?redirect_uri=${encodeURIComponent(appLoginPath)}`
}

async function loginWithTicket() {
  loading.value = true
  try {
    const response = await authApi.fnosLogin({ remember: rememberMe.value })
    if (response.data.binding_required) {
      showFnOSConfirm.value = false
      fnosBinding.value = true
      fnosUsername.value = response.data.fnos_username || ''
      fnosMode.value = response.data.suggested_mode === 'register' ? 'register' : 'bind'
      username.value = response.data.suggested_username || ''
      password.value = ''
      email.value = ''
      isRegisterMode.value = true
      return
    }
    clearFnOSTicket()
    showFnOSConfirm.value = false
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('user', JSON.stringify(response.data.user))
    router.push((route.query.redirect as string) || '/')
  } catch (err: any) {
    clearFnOSTicket()
    showFnOSConfirm.value = false
    const msg = err.response?.data?.error || '飞牛 NAS 登录失败，请重新点击飞牛授权登录'
    // 未绑定跳转创建/绑定页后再失败（如票据过期），错误要落在注册卡片上
    if (fnosBinding.value) {
      fnosNeedsReauth.value = true
      registerError.value = msg
    } else {
      loginError.value = msg
    }
  } finally {
    loading.value = false
  }
}

// 加载记住的凭据
const loadRememberedCredentials = () => {
  const remembered = localStorage.getItem('rememberedCredentials')
  if (remembered) {
    try {
      const { username: savedUsername, password: savedPassword } = JSON.parse(remembered)
      username.value = savedUsername
      password.value = savedPassword
      rememberMe.value = true
    } catch (e) {
      console.error('加载记住的密码失败:', e)
    }
  }
}

// 保存记住的凭据
const saveRememberedCredentials = () => {
  if (rememberMe.value) {
    // 保存明文密码
    localStorage.setItem('rememberedCredentials', JSON.stringify({
      username: username.value,
      password: password.value
    }))
  } else {
    // 清除记住的凭据
    localStorage.removeItem('rememberedCredentials')
  }
}

// 显示注册按钮：（没有管理员）或（有管理员 且 允许注册）
const showRegisterEntry = () => !adminExists.value || allowRegister.value

// 登录
const handleLogin = async () => {
  if (!username.value || !password.value) {
    loginError.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  loginError.value = ''
  try {
    // 始终加密密码（因为存储的是明文）
    const hashed = hashPassword(password.value)

    const response = await authApi.login({ username: username.value, password: hashed })
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('user', JSON.stringify(response.data.user))

    // 保存记住的凭据（明文）
    saveRememberedCredentials()

    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: any) {
    loginError.value = err.response?.data?.error || '登录失败，请检查用户名和密码'
    return
  } finally {
    loading.value = false
  }
}

const completeLogin = (data: any) => {
  localStorage.setItem('token', data.token)
  localStorage.setItem('user', JSON.stringify(data.user))
  clearFnOSTicket()
  router.push((route.query.redirect as string) || '/')
}

// 注册
const handleRegister = async () => {
  if (!username.value || !password.value) {
    registerError.value = '请输入用户名和密码'
    return
  }
  if (password.value.length < 6) {
    registerError.value = '密码长度至少为6位'
    return
  }
  loading.value = true
  registerError.value = ''
  try {
    const hashed = hashPassword(password.value)
    if (fnosBinding.value) {
      const response = await authApi.bindFnOS({
        mode: fnosMode.value,
        username: username.value,
        password: hashed,
        email: email.value,
        remember: rememberMe.value,
      })
      completeLogin(response.data)
      return
    }
    const response = await authApi.register({
      username: username.value,
      password: hashed,
      email: email.value,
    })
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('user', JSON.stringify(response.data.user))
    registerSuccess.value = true
    setTimeout(() => router.push('/'), 1500)
  } catch (err: any) {
    // 绑定/注册用的票据已失效（401）时，引导用户重新走飞牛授权
    if (fnosBinding.value && err.response?.status === 401) {
      fnosNeedsReauth.value = true
      clearFnOSTicket()
    }
    registerError.value = err.response?.data?.error || '注册失败，请重试'
    return
  } finally {
    loading.value = false
  }
}

const switchToRegister = () => {
  fnosBinding.value = false
  fnosNeedsReauth.value = false
  isRegisterMode.value = true
  loginError.value = ''
  username.value = ''
  password.value = ''
  email.value = ''
}

const switchToLogin = () => {
  fnosBinding.value = false
  fnosNeedsReauth.value = false
  isRegisterMode.value = false
  registerError.value = ''
  // 切换回登录页时，重新加载记住的凭据
  loadRememberedCredentials()
  email.value = ''
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 via-white to-emerald-50 flex items-center justify-center p-4">
    <div class="w-full max-w-md">

      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl gradient-primary shadow-lg shadow-blue-500/30 mb-4">
          <Gift class="w-10 h-10 text-white" />
        </div>
        <h1 class="text-3xl font-bold bg-gradient-to-r from-blue-600 to-emerald-500 bg-clip-text text-transparent">
          彩彩助手
        </h1>
        <p class="text-slate-500 mt-2">{{ isRegisterMode ? (adminExists ? '创建新账号' : '创建管理员账号') : '欢迎回来' }}</p>
      </div>

      <!-- ========== 登录卡片 ========== -->
      <div v-if="!isRegisterMode" class="bg-white rounded-2xl shadow-xl shadow-slate-200/50 p-8">
        <h2 class="text-2xl font-bold text-slate-800 mb-6">登录</h2>

        <!-- 错误提示 - 醒目样式 -->
        <div v-if="loginError" class="mb-4 p-4 bg-red-100 border-2 border-red-400 rounded-xl text-red-700 text-base font-medium flex items-center gap-3 animate-shake">
          <AlertCircle class="w-5 h-5 flex-shrink-0" />
          <span>{{ loginError }}</span>
        </div>

        <!-- 用户名 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-slate-700 mb-2">用户名</label>
          <div class="relative">
            <User class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              v-model="username"
              type="text"
              placeholder="请输入用户名"
              class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              @keyup.enter="handleLogin"
            />
          </div>
        </div>

        <!-- 密码 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-slate-700 mb-2">密码</label>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input
              v-model="password"
              type="password"
              placeholder="请输入密码"
              class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              @keyup.enter="handleLogin"
            />
          </div>
        </div>

        <!-- 记住密码 -->
        <div class="mb-6 flex items-center">
          <input
            id="remember-me"
            v-model="rememberMe"
            type="checkbox"
            class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500 cursor-pointer"
          />
          <label for="remember-me" class="ml-2 text-sm text-slate-600 cursor-pointer select-none">
            记住密码
          </label>
        </div>

        <button
          @click="handleLogin"
          :disabled="loading"
          class="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-emerald-500 text-white font-medium rounded-lg hover:shadow-lg hover:shadow-blue-500/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
          <template v-else>
            <span>登录</span>
            <ArrowRight class="w-4 h-4" />
          </template>
        </button>

        <button
          v-if="fnosEnabled"
          type="button"
          @click="handleFnOSLogin"
          :disabled="loading"
          class="mt-3 w-full min-h-12 py-3 px-4 border border-slate-300 bg-transparent text-slate-700 font-medium rounded-lg hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ loading ? '正在获取飞牛账号…' : '使用飞牛 NAS 账号授权登录' }}
        </button>

        <!-- 注册入口 -->
        <div v-if="showRegisterEntry()" class="mt-6 text-center">
          <p class="text-slate-600 text-sm">
            {{ adminExists ? '没有账号？' : '首次使用，请先' }}
            <button @click="switchToRegister" class="text-blue-600 hover:text-blue-700 font-medium">
              {{ adminExists ? '立即注册' : '注册管理员账号' }}
            </button>
          </p>
        </div>
      </div>

      <!-- ========== 注册卡片 ========== -->
      <div v-else class="bg-white rounded-2xl shadow-xl shadow-slate-200/50 p-8">
        <h2 class="text-2xl font-bold text-slate-800 mb-6">
          {{ fnosBinding ? '绑定飞牛 NAS 账号' : (adminExists ? '注册' : '创建管理员账号') }}
        </h2>

        <!-- 成功状态 -->
        <div v-if="registerSuccess" class="text-center py-4">
          <div class="inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
            <User class="w-8 h-8 text-green-600" />
          </div>
          <h3 class="text-xl font-semibold text-slate-800 mb-2">注册成功！</h3>
          <p class="text-slate-500">正在跳转到首页...</p>
        </div>

        <template v-else>
          <div v-if="fnosBinding" class="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-xl text-blue-800 text-sm">
            当前飞牛 NAS 用户 <strong>{{ fnosUsername || '已登录用户' }}</strong> 尚未绑定应用账号。
          </div>
          <!-- 错误提示 - 醒目样式 -->
          <div v-if="registerError" class="mb-4 p-4 bg-red-100 border-2 border-red-400 rounded-xl text-red-700 text-base font-medium flex items-center gap-3 animate-shake">
            <AlertCircle class="w-5 h-5 flex-shrink-0" />
            <span>{{ registerError }}</span>
          </div>

          <!-- 用户名 -->
          <div class="mb-4">
            <label class="block text-sm font-medium text-slate-700 mb-2">用户名</label>
            <div class="relative">
              <User class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                v-model="username"
                type="text"
                placeholder="请输入用户名"
                class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              />
            </div>
          </div>

          <!-- 密码 -->
          <div class="mb-6">
            <label class="block text-sm font-medium text-slate-700 mb-2">密码</label>
            <div class="relative">
              <Lock class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                v-model="password"
                type="password"
                placeholder="请输入密码（至少6位）"
                class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
                @keyup.enter="handleRegister"
              />
            </div>
          </div>

          <!-- 邮箱（可选） -->
          <div v-if="!fnosBinding || fnosMode === 'register'" class="mb-4">
            <label class="block text-sm font-medium text-slate-700 mb-2">
              邮箱
              <span class="text-slate-400 font-normal ml-1">（可选）</span>
            </label>
            <div class="relative">
              <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                v-model="email"
                type="email"
                placeholder="请输入邮箱"
                class="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              />
            </div>
          </div>

          <button
            @click="handleRegister"
            :disabled="loading"
            class="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-emerald-500 text-white font-medium rounded-lg hover:shadow-lg hover:shadow-blue-500/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            <template v-else>
              <span>{{ fnosBinding ? (fnosMode === 'register' ? '创建并绑定' : '验证并绑定') : '注册' }}</span>
              <ArrowRight class="w-4 h-4" />
            </template>
          </button>

          <!-- 飞牛授权入口：登录/注册两张卡片都提供 -->
          <button
            v-if="fnosEnabled && !fnosBinding"
            type="button"
            @click="handleFnOSLogin"
            :disabled="loading"
            class="mt-3 w-full min-h-12 py-3 px-4 border border-slate-300 bg-transparent text-slate-700 font-medium rounded-lg hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loading ? '正在获取飞牛账号…' : '使用飞牛 NAS 账号授权登录' }}
          </button>

          <button
            v-if="fnosBinding"
            type="button"
            @click="fnosMode = fnosMode === 'register' ? 'bind' : 'register'"
            class="mt-3 w-full min-h-11 text-sm font-medium text-blue-600 hover:text-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 rounded-lg"
          >
            {{ fnosMode === 'register' ? '已有应用账号？验证并绑定' : '没有应用账号？创建并绑定' }}
          </button>

          <!-- 票据失效后的重新授权入口 -->
          <button
            v-if="fnosEnabled && fnosNeedsReauth"
            type="button"
            @click="handleFnOSLogin"
            :disabled="loading"
            class="mt-3 w-full min-h-12 py-3 px-4 border border-blue-300 bg-blue-50 text-blue-700 font-medium rounded-lg hover:bg-blue-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loading ? '正在获取飞牛账号…' : '重新进行飞牛授权' }}
          </button>

          <!-- 返回登录 -->
          <div v-if="adminExists" class="mt-6 text-center">
            <button @click="switchToLogin" class="text-slate-500 hover:text-slate-700 text-sm">
              已有账号？返回登录
            </button>
          </div>
        </template>
      </div>

    </div>

    <!-- 飞牛授权确认框：拿到票据后先让用户确认当前 NAS 账号 -->
    <div
      v-if="showFnOSConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4"
      @click.self="cancelFnOSConfirm"
    >
      <div class="w-full max-w-sm bg-white rounded-2xl shadow-2xl p-6 text-center">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl gradient-primary shadow-lg shadow-blue-500/30 mb-4">
          <User class="w-7 h-7 text-white" />
        </div>
        <h3 class="text-lg font-bold text-slate-800 mb-1">使用飞牛账号授权登录</h3>
        <p class="text-sm text-slate-500 mb-4">当前 NAS 账号</p>
        <div class="mb-5 px-4 py-3 bg-blue-50 border border-blue-200 rounded-xl text-blue-800 font-semibold">
          {{ fnosConfirmUsername || '飞牛用户' }}
        </div>
        <button
          @click="loginWithTicket"
          :disabled="loading"
          class="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-emerald-500 text-white font-medium rounded-lg hover:shadow-lg hover:shadow-blue-500/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
          <template v-else>
            <span>确认并登录</span>
            <ArrowRight class="w-4 h-4" />
          </template>
        </button>
        <div class="mt-3 flex items-center justify-center gap-4 text-sm">
          <button @click="switchFnOSAccount" class="text-blue-600 hover:text-blue-700 font-medium">切换飞牛账号</button>
          <button @click="cancelFnOSConfirm" class="text-slate-500 hover:text-slate-700">取消</button>
        </div>
        <p class="mt-3 text-xs text-slate-400">取消将不保留任何登录状态</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 错误提示抖动动画 */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  10%, 30%, 50%, 70%, 90% { transform: translateX(-4px); }
  20%, 40%, 60%, 80% { transform: translateX(4px); }
}

.animate-shake {
  animation: shake 0.5s ease-in-out;
}
</style>
