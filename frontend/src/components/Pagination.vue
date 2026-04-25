<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

interface Props {
  currentPage: number
  pageSize: number
  total: number
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:currentPage', page: number): void
  (e: 'update:pageSize', size: number): void
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

const startItem = computed(() => (props.currentPage - 1) * props.pageSize + 1)
const endItem = computed(() => Math.min(props.currentPage * props.pageSize, props.total))

const pageNumbers = computed(() => {
  const pages: (number | string)[] = []
  const total = totalPages.value
  const current = props.currentPage

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    if (current <= 3) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 2) {
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) pages.push(i)
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    }
  }
  return pages
})

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value || page === props.currentPage) return
  emit('update:currentPage', page)
}

const changePageSize = (size: number) => {
  emit('update:pageSize', size)
  emit('update:currentPage', 1)
}
</script>

<template>
  <div class="flex flex-col sm:flex-row items-center justify-between gap-4 px-4 py-4 border-t border-slate-100">
    <div class="text-sm text-slate-500">
      共 {{ total }} 条，第 {{ startItem }}-{{ endItem }} 条
    </div>

    <div class="flex items-center gap-2">
      <select
        :value="pageSize"
        @change="changePageSize(Number(($event.target as HTMLSelectElement).value))"
        class="px-2 py-1.5 bg-white border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer"
      >
        <option :value="10">10条/页</option>
        <option :value="20">20条/页</option>
        <option :value="50">50条/页</option>
      </select>

      <button
        @click="goToPage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="p-1.5 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
      >
        <ChevronLeft class="w-4 h-4" />
      </button>

      <div class="flex items-center gap-1">
        <button
          v-for="page in pageNumbers"
          :key="page"
          @click="typeof page === 'number' ? goToPage(page) : undefined"
          :disabled="page === '...'"
          :class="[
            'min-w-[2rem] h-8 px-2 rounded-lg text-sm font-medium transition-colors',
            page === currentPage
              ? 'bg-blue-500 text-white'
              : page === '...'
                ? 'text-slate-400 cursor-default'
                : 'text-slate-600 hover:bg-slate-50 cursor-pointer'
          ]"
        >
          {{ page }}
        </button>
      </div>

      <button
        @click="goToPage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="p-1.5 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
      >
        <ChevronRight class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>
