<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Home, ShoppingCart, Award, BarChart3, Gift, Menu, X, Target, Users, LogOut, User as UserIcon, Settings, KeyRound } from 'lucide-vue-next'
import type { User, VersionInfo } from '../types'
import { authApi, systemApi } from '../api'
import { hashPassword } from '../utils/crypto'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const currentUser = ref<User | null>(null)
const showUserMenu = ref(false)
const showPasswordModal = ref(false)
const versionInfo = ref<VersionInfo | null>(null)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordLoading = ref(false)
const passwordError = ref('')
const passwordSuccess = ref(false)

const isLoggedIn = computed(() => currentUser.value !== null)
const isAdmin = computed(() => currentUser.value?.role === 'admin')

const navItems = [
  { name: '仪表盘', path: '/', icon: Home },
  { name: '购买记录', path: '/purchase', icon: ShoppingCart },
  { name: '开奖管理', path: '/draw', icon: Award },
  { name: '中奖记录', path: '/winnings', icon: Gift },
  { name: '历史命中', path: '/history-hit', icon: Target },
  { name: '统计分析', path: '/statistics', icon: BarChart3 },
]

const adminItems = []

const isActive = (path: string) => route.path === path

const toggleMenu = () => {
  menuOpen.value = !menuOpen.value
}

const closeMenu = () => {
  menuOpen.value = false
  showUserMenu.value = false
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  currentUser.value = null
  showUserMenu.value = false
  router.push('/login')
}

// 打开修改密码弹窗
const openChangePassword = () => {
  showUserMenu.value = false
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  passwordError.value = ''
  passwordSuccess.value = false
  showPasswordModal.value = true
}

// 修改密码
const handleChangePassword = async () => {
  if (!oldPassword.value || !newPassword.value || !confirmPassword.value) {
    passwordError.value = '请填写所有字段'
    return
  }
  if (newPassword.value.length < 6) {
    passwordError.value = '新密码长度至少为6位'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = '两次输入的新密码不一致'
    return
  }
  passwordLoading.value = true
  passwordError.value = ''
  try {
    const hashedOld = hashPassword(oldPassword.value)
    const hashedNew = hashPassword(newPassword.value)
    await authApi.changePassword({ old_password: hashedOld, new_password: hashedNew })
    passwordSuccess.value = true
    // 清除记住的密码
    localStorage.removeItem('rememberedCredentials')
    setTimeout(() => {
      showPasswordModal.value = false
      // 修改成功后退出登录，用新密码重新登录
      handleLogout()
    }, 1500)
  } catch (err: any) {
    passwordError.value = err.response?.data?.error || '密码修改失败'
  } finally {
    passwordLoading.value = false
  }
}

const loadUser = () => {
  const userStr = localStorage.getItem('user')
  currentUser.value = userStr ? JSON.parse(userStr) : null
}

// 加载版本信息
const loadVersionInfo = async () => {
  try {
    const data = await systemApi.getVersion()
    versionInfo.value = data
  } catch (err) {
    console.error('Failed to load version info:', err)
    // 如果API调用失败，设置默认版本信息
    versionInfo.value = {
      name: '彩彩助手',
      version: 'v1.0.0',
      buildTime: 'unknown',
      gitCommit: 'unknown',
      status: 'running'
    }
  }
}

// 每次路由变化都重新同步登录状态
watch(() => route.path, () => {
  loadUser()
  closeMenu()
})

// 点击页面其他区域关闭下拉菜单
const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!target.closest('.user-menu-wrapper')) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  loadUser()
  loadVersionInfo()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <nav class="fixed top-0 left-0 right-0 z-50 glass border-b border-slate-200/50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl gradient-primary flex items-center justify-center shadow-lg shadow-blue-500/30">
            <Gift class="w-5 h-5 text-white" />
          </div>
          <div class="flex flex-col">
            <span class="text-xl font-bold bg-gradient-to-r from-blue-600 to-emerald-500 bg-clip-text text-transparent">
              彩彩助手
            </span>
            <span class="text-xs text-slate-500" v-if="versionInfo">
              {{ versionInfo.version }}
            </span>
          </div>
        </div>

        <!-- Desktop Navigation -->
        <div class="hidden md:flex items-center gap-1">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200"
            :class="isActive(item.path)
              ? 'bg-blue-50 text-blue-600 shadow-sm'
              : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'"
          >
            <component :is="item.icon" class="w-4 h-4" />
            {{ item.name }}
          </router-link>

          <!-- User menu / Login button -->
          <div v-if="isLoggedIn" class="relative ml-2 user-menu-wrapper">
            <button
              @click="showUserMenu = !showUserMenu"
              class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 bg-slate-100 hover:bg-slate-200 text-slate-700"
            >
              <UserIcon class="w-4 h-4" />
              <span>{{ currentUser?.username }}</span>
            </button>

            <!-- Dropdown -->
            <div
              v-if="showUserMenu"
              class="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-slate-200 py-2"
            >
              <div class="px-4 py-2 border-b border-slate-100">
                <p class="text-sm font-medium text-slate-800">{{ currentUser?.username }}</p>
                <p class="text-xs text-slate-500">{{ isAdmin ? '管理员' : '普通用户' }}</p>
              </div>

              <!-- 后台管理（仅管理员） -->
              <router-link
                v-if="isAdmin"
                to="/admin"
                @click="showUserMenu = false"
                class="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 flex items-center gap-2"
              >
                <Settings class="w-4 h-4" />
                后台管理
              </router-link>

              <!-- 修改密码 -->
              <button
                @click="openChangePassword"
                class="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 flex items-center gap-2"
              >
                <KeyRound class="w-4 h-4" />
                修改密码
              </button>

              <button
                @click="handleLogout"
                class="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
              >
                <LogOut class="w-4 h-4" />
                退出登录
              </button>
            </div>
          </div>

          <router-link
            v-else
            to="/login"
            class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 bg-blue-600 hover:bg-blue-700 text-white"
          >
            <UserIcon class="w-4 h-4" />
            登录
          </router-link>
        </div>

        <!-- Mobile menu button -->
        <button
          class="md:hidden p-2 rounded-lg text-slate-600 hover:bg-slate-100 transition-colors cursor-pointer"
          @click="toggleMenu"
          aria-label="打开菜单"
        >
          <Menu v-if="!menuOpen" class="w-6 h-6" />
          <X v-else class="w-6 h-6" />
        </button>
      </div>
    </div>

    <!-- Mobile Dropdown Menu -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div v-if="menuOpen" class="md:hidden bg-white border-t border-slate-100 shadow-lg">
        <div class="px-4 py-3 space-y-1">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            @click="closeMenu"
            class="flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200"
            :class="isActive(item.path)
              ? 'bg-blue-50 text-blue-600'
              : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'"
          >
            <component :is="item.icon" class="w-5 h-5" />
            {{ item.name }}
          </router-link>

          <!-- Login/Logout -->
          <div v-if="isLoggedIn" class="border-t border-slate-100 pt-3">
            <div class="px-4 py-2 text-sm text-slate-600">
              <span class="font-medium">{{ currentUser?.username }}</span>
              <span class="text-slate-400">({{ isAdmin ? '管理员' : '普通用户' }})</span>
            </div>

            <!-- 后台管理（仅管理员） -->
            <router-link
              v-if="isAdmin"
              to="/admin"
              @click="closeMenu"
              class="flex items-center gap-3 px-4 py-3 w-full rounded-xl text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
            >
              <Settings class="w-5 h-5" />
              后台管理
            </router-link>

            <!-- 修改密码 -->
            <button
              @click="openChangePassword"
              class="flex items-center gap-3 px-4 py-3 w-full rounded-xl text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
            >
              <KeyRound class="w-5 h-5" />
              修改密码
            </button>

            <button
              @click="handleLogout"
              class="flex items-center gap-3 px-4 py-3 w-full rounded-xl text-sm font-medium text-red-600 hover:bg-red-50 transition-colors"
            >
              <LogOut class="w-5 h-5" />
              退出登录
            </button>
          </div>

          <router-link
            v-else
            to="/login"
            @click="closeMenu"
            class="flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium bg-blue-600 text-white"
          >
            <UserIcon class="w-5 h-5" />
            登录
          </router-link>
        </div>
      </div>
    </transition>
  </nav>

  <!-- Mobile menu backdrop -->
  <div
    v-if="menuOpen"
    class="fixed inset-0 z-40 bg-black/20 md:hidden"
    @click="closeMenu"
  />

  <!-- 修改密码弹窗 -->
  <teleport to="body">
    <div
      v-if="showPasswordModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="showPasswordModal = false"
    >
      <div class="bg-white rounded-2xl shadow-xl w-full max-w-md p-6">
        <h3 class="text-xl font-bold text-slate-800 mb-4 flex items-center gap-2">
          <KeyRound class="w-5 h-5 text-blue-600" />
          修改密码
        </h3>

        <!-- 成功状态 -->
        <div v-if="passwordSuccess" class="text-center py-6">
          <div class="inline-flex items-center justify-center w-14 h-14 bg-green-100 rounded-full mb-3">
            <svg class="w-7 h-7 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <p class="text-green-600 font-medium">密码修改成功</p>
          <p class="text-slate-500 text-sm mt-1">即将退出登录，请使用新密码重新登录</p>
        </div>

        <template v-else>
          <div v-if="passwordError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
            {{ passwordError }}
          </div>

          <!-- 旧密码 -->
          <div class="mb-4">
            <label class="block text-sm font-medium text-slate-700 mb-1">旧密码</label>
            <input
              v-model="oldPassword"
              type="password"
              placeholder="请输入旧密码"
              class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
            />
          </div>

          <!-- 新密码 -->
          <div class="mb-4">
            <label class="block text-sm font-medium text-slate-700 mb-1">新密码</label>
            <input
              v-model="newPassword"
              type="password"
              placeholder="请输入新密码（至少6位）"
              class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
            />
          </div>

          <!-- 确认新密码 -->
          <div class="mb-6">
            <label class="block text-sm font-medium text-slate-700 mb-1">确认新密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
              class="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
              @keyup.enter="handleChangePassword"
            />
          </div>

          <div class="flex gap-3">
            <button
              @click="showPasswordModal = false"
              class="flex-1 px-4 py-2.5 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 transition-colors"
            >
              取消
            </button>
            <button
              @click="handleChangePassword"
              :disabled="passwordLoading"
              class="flex-1 px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <span v-if="passwordLoading" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              <span>确认修改</span>
            </button>
          </div>
        </template>
      </div>
    </div>
  </teleport>


</template>
