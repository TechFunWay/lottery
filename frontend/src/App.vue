<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from './components/NavBar.vue'

const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <NavBar v-if="!isLoginPage" />
    <main :class="isLoginPage ? 'pb-8' : 'pt-20 pb-8 px-4 sm:px-6 lg:px-8'">
      <div :class="isLoginPage ? '' : 'max-w-7xl mx-auto'">
        <router-view />
      </div>
    </main>
  </div>
</template>
