<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Heart } from 'lucide-vue-next'
import NavBar from './components/NavBar.vue'
import SupportModal from './components/SupportModal.vue'

const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')
const showSupportModal = ref(false)
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <NavBar v-if="!isLoginPage" />
    <main :class="isLoginPage ? 'pb-8' : 'pt-20 pb-8 px-4 sm:px-6 lg:px-8'">
      <div :class="isLoginPage ? '' : 'max-w-7xl mx-auto'">
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
    <SupportModal v-model="showSupportModal" />
  </div>
</template>
