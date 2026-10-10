<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Heart, X } from 'lucide-vue-next'
import api from './api'
import NavBar from './components/NavBar.vue'
import SupportModal from './components/SupportModal.vue'

const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')
const showSupportModal = ref(false)

// ===== 赞赏支持状态 =====
const currentUserRole = ref('')
const appVersion = ref('')
const serverStartedAt = ref('')
// 支持记录按版本号记忆（donate_supported_version）：支持过当前版本后
// 本版本内不再提示；升级到新版本后旧记录失效，横幅与弹窗重新出现
const supportedVersion = ref(localStorage.getItem('donate_supported_version') || '')
const donateSupported = computed(() => !!appVersion.value && supportedVersion.value === appVersion.value)
// 横幅本次会话是否已被手动关闭：关闭不持久化，刷新浏览器后再次出现
const donateBannerClosed = ref(false)

// 顶部横幅：管理员可见；未支持当前版本前常驻展示，支持过当前版本后隐藏
const showDonateBanner = computed(() => {
  return !isLoginPage.value
    && currentUserRole.value === 'admin'
    && !donateSupported.value
    && !donateBannerClosed.value
})

const dismissDonateBanner = () => {
  donateBannerClosed.value = true
}

// 启动弹窗：应用每次启动后首次进入时弹一次；
// 点【暂不支持】后同版本同次运行内不再弹，应用重启后再次提示；
// 【已支持】发送成功后本版本内不再提示，升级到新版本后会再次提示。
// 抑制记录必须「同版本 + 同次运行」才有效：升级换版本后旧记录失效。
const scheduleSupportModal = () => {
  if (isLoginPage.value) return
  if (currentUserRole.value !== 'admin') return
  if (donateSupported.value) return
  const dismissedStart = localStorage.getItem('donate_dismissed_start')
  const dismissedVersion = localStorage.getItem('donate_dismissed_version')
  const sameRun = !!serverStartedAt.value && dismissedStart === serverStartedAt.value
  const sameVersion = !!appVersion.value && dismissedVersion === appVersion.value
  if (sameRun && sameVersion) return
  if (sameRun && !appVersion.value) return
  showSupportModal.value = true
}

const initDonate = async () => {
  try {
    const userStr = localStorage.getItem('user')
    currentUserRole.value = userStr ? (JSON.parse(userStr)?.role || '') : ''
  } catch { currentUserRole.value = '' }
  if (isLoginPage.value || currentUserRole.value !== 'admin') return
  try {
    // 走 API 客户端（baseURL 已含飞牛网关前缀），原生 axios 的绝对路径在网关部署下会打到宿主系统
    const res = await api.get('/version') as { version?: string; startedAt?: string }
    appVersion.value = res?.version || ''
    serverStartedAt.value = res?.startedAt || ''
  } catch { /* 版本接口失败不影响使用 */ }
  setTimeout(() => scheduleSupportModal(), 1500)
}

// SPA 内登录后不会重新挂载 App：离开登录页时补一次初始化（仅一次）
const donateInited = ref(false)
initDonate()
watch(isLoginPage, (loginPage, wasLoginPage) => {
  if (wasLoginPage && !loginPage && !donateInited.value) {
    donateInited.value = true
    initDonate()
  }
})

// 关闭弹窗（暂不支持 / ✕ / 点遮罩）：同版本同次运行内不再弹出
const onSupportDismiss = () => {
  showSupportModal.value = false
  if (serverStartedAt.value) localStorage.setItem('donate_dismissed_start', serverStartedAt.value)
  if (appVersion.value) localStorage.setItem('donate_dismissed_version', appVersion.value)
}

// 已支持成功：记住当前版本，升级到新版本后会重新提示
const onSupportConfirmed = () => {
  supportedVersion.value = appVersion.value
  if (appVersion.value) localStorage.setItem('donate_supported_version', appVersion.value)
  showSupportModal.value = false
}
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <NavBar v-if="!isLoginPage" />

    <main :class="isLoginPage ? 'pb-8' : 'pt-20 pb-8 px-4 sm:px-6 lg:px-8'">
      <div :class="isLoginPage ? '' : 'max-w-7xl mx-auto'">
        <!-- 赞赏横幅：管理员可见，未支持当前版本前常驻展示（关闭不记忆，
             刷新后再次出现，升级到新版本后也会再次出现）；文档流内展示 -->
        <div
          v-if="!isLoginPage && showDonateBanner"
          class="flex items-center justify-between gap-3 px-4 py-2.5 mb-4 rounded-xl bg-gradient-to-r from-indigo-500 to-violet-500 text-white text-sm shadow-md"
        >
          <span>☕ 彩彩助手免费、无广告、数据都在你自己的设备上。如果它帮到了你，欢迎请站长喝杯咖啡——1 元也是心意，完全自愿，不影响任何功能。</span>
          <span class="flex items-center gap-2 shrink-0">
            <button
              class="underline underline-offset-2 hover:opacity-80"
              @click="showSupportModal = true"
            >去赞赏</button>
            <button class="hover:opacity-80" title="关闭" @click="dismissDonateBanner">
              <X class="w-4 h-4" />
            </button>
          </span>
        </div>
        <router-view />
      </div>
    </main>

    <!-- 支持按钮 -->
    <button
      v-if="!isLoginPage"
      @click="showSupportModal = true"
      class="fixed bottom-6 right-6 z-40 flex items-center gap-2 px-4 py-2.5 bg-white/80 backdrop-blur-sm text-slate-700 rounded-full shadow-lg border border-slate-200 hover:bg-white hover:shadow-xl hover:scale-105 transition-all duration-200"
    >
      <Heart class="w-4 h-4 text-rose-500" />
      <span class="text-sm font-medium">支持</span>
    </button>

    <!-- 支持弹窗 -->
    <SupportModal
      v-model="showSupportModal"
      show-actions
      @dismiss="onSupportDismiss"
      @supported="onSupportConfirmed"
    />
  </div>
</template>
