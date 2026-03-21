<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { authApi, configApi } from '../api'
import { hashPassword } from '../utils/crypto'
import { User, Lock, ArrowRight, Gift, Mail } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()

// 表单状态
const username = ref('')
const password = ref('')
const email = ref('')
const loading = ref(false)
const isRegisterMode = ref(false)
const rememberMe = ref(false)  // 记住密码

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
  } finally {
    loading.value = false
  }
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
    registerError.value = err.response?.data?.error || '注册失败，请重试'
  } finally {
    loading.value = false
  }
}

const switchToRegister = () => {
  isRegisterMode.value = true
  loginError.value = ''
  username.value = ''
  password.value = ''
  email.value = ''
}

const switchToLogin = () => {
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
          彩票助手
        </h1>
        <p class="text-slate-500 mt-2">{{ isRegisterMode ? (adminExists ? '创建新账号' : '创建管理员账号') : '欢迎回来' }}</p>
      </div>

      <!-- ========== 登录卡片 ========== -->
      <div v-if="!isRegisterMode" class="bg-white rounded-2xl shadow-xl shadow-slate-200/50 p-8">
        <h2 class="text-2xl font-bold text-slate-800 mb-6">登录</h2>

        <div v-if="loginError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {{ loginError }}
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
        <h2 class="text-2xl font-bold text-slate-800 mb-6">注册</h2>

        <!-- 成功状态 -->
        <div v-if="registerSuccess" class="text-center py-4">
          <div class="inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
            <User class="w-8 h-8 text-green-600" />
          </div>
          <h3 class="text-xl font-semibold text-slate-800 mb-2">注册成功！</h3>
          <p class="text-slate-500">正在跳转到首页...</p>
        </div>

        <template v-else>
          <div v-if="registerError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
            {{ registerError }}
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

          <!-- 邮箱（可选） -->
          <div class="mb-4">
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

          <button
            @click="handleRegister"
            :disabled="loading"
            class="w-full py-3 px-4 bg-gradient-to-r from-blue-600 to-emerald-500 text-white font-medium rounded-lg hover:shadow-lg hover:shadow-blue-500/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
            <template v-else>
              <span>注册</span>
              <ArrowRight class="w-4 h-4" />
            </template>
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
  </div>
</template>
