<script setup lang="ts">
import { ref, computed } from 'vue'
import { footballMatchApi } from '../api'
import type { FootballMatch, FootballPlayType, FootballSelection } from '../types'
import { FOOTBALL_PLAY_TYPES, FOOTBALL_BET_TYPES, WDL_LABELS } from '../types'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const matches = ref<FootballMatch[]>([])
const loading = ref(false)

const loadMatches = async () => {
  loading.value = true
  try {
    const res = await footballMatchApi.list({ size: 100 })
    matches.value = (res as any).data || []
  } catch (e) {
    console.error('加载比赛失败', e)
  } finally {
    loading.value = false
  }
}

loadMatches()

const currentSelections = computed<FootballSelection[]>({
  get: () => {
    try { return JSON.parse(props.modelValue || '[]') } catch { return [] }
  },
  set: (val) => {
    emit('update:modelValue', JSON.stringify(val))
  }
})

const addSelection = () => {
  currentSelections.value = [...currentSelections.value, {
    match_id: '',
    play_type: '胜平负',
    selection: '',
    odds: 0,
  }]
}

const removeSelection = (index: number) => {
  const newSelections = [...currentSelections.value]
  newSelections.splice(index, 1)
  currentSelections.value = newSelections
}

const updateSelection = (index: number, field: keyof FootballSelection, value: any) => {
  const newSelections = [...currentSelections.value]
  newSelections[index] = { ...newSelections[index], [field]: value }
  if (field === 'play_type') {
    newSelections[index].selection = ''
  }
  if (field === 'match_id') {
    const match = matches.value.find(m => m.match_id === value)
    if (match) {
      newSelections[index].handicap = match.handicap
    }
  }
  currentSelections.value = newSelections
}

const getPlayTypeOptions = (playType: FootballPlayType) => {
  const config = FOOTBALL_PLAY_TYPES.find(p => p.type === playType)
  return config?.options || []
}

const getMatchLabel = (matchId: string) => {
  const match = matches.value.find(m => m.match_id === matchId)
  if (!match) return matchId
  return `${match.home_team} vs ${match.away_team}`
}

const getOptionLabel = (playType: FootballPlayType, option: string) => {
  if (playType === '胜平负' || playType === '让球胜平负') {
    return WDL_LABELS[option] || option
  }
  return option
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <label class="text-sm font-medium text-slate-600">投注选项</label>
      <button @click="addSelection"
        class="px-3 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-600 rounded-lg text-sm font-medium transition-colors cursor-pointer">
        + 添加场次
      </button>
    </div>

    <div v-if="loading" class="text-center py-4 text-slate-400 text-sm">加载比赛中...</div>

    <div v-else-if="currentSelections.length === 0" class="text-center py-6 text-slate-400 text-sm">
      请点击"添加场次"选择比赛
    </div>

    <div v-else class="space-y-3">
      <div v-for="(sel, index) in currentSelections" :key="index"
        class="bg-slate-50 rounded-xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium text-slate-500">第 {{ index + 1 }} 场</span>
          <button @click="removeSelection(index)"
            class="text-red-400 hover:text-red-500 text-xs cursor-pointer">移除</button>
        </div>

        <div>
          <label class="block text-xs text-slate-500 mb-1">选择比赛</label>
          <select :value="sel.match_id" @change="updateSelection(index, 'match_id', ($event.target as HTMLSelectElement).value)"
            class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white">
            <option value="">请选择比赛</option>
            <option v-for="match in matches" :key="match.match_id" :value="match.match_id">
              {{ match.match_id }} - {{ match.home_team }} vs {{ match.away_team }}
              <template v-if="match.status === '已完赛'"> ({{ match.home_score }}:{{ match.away_score }})</template>
            </option>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-slate-500 mb-1">玩法</label>
            <select :value="sel.play_type" @change="updateSelection(index, 'play_type', ($event.target as HTMLSelectElement).value)"
              class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400 cursor-pointer bg-white">
              <option v-for="pt in FOOTBALL_PLAY_TYPES" :key="pt.type" :value="pt.type">{{ pt.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-slate-500 mb-1">赔率</label>
            <input type="number" step="0.01" min="0" :value="sel.odds"
              @input="updateSelection(index, 'odds', parseFloat(($event.target as HTMLInputElement).value) || 0)"
              class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400"
              placeholder="如：1.85" />
          </div>
        </div>

        <div v-if="sel.play_type === '让球胜平负'">
          <label class="block text-xs text-slate-500 mb-1">让球数</label>
          <input type="number" step="0.5" :value="sel.handicap || 0"
            @input="updateSelection(index, 'handicap', parseFloat(($event.target as HTMLInputElement).value) || 0)"
            class="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-blue-400"
            placeholder="如：-1" />
        </div>

        <div>
          <label class="block text-xs text-slate-500 mb-1">选择结果</label>
          <div class="flex flex-wrap gap-2">
            <button v-for="opt in getPlayTypeOptions(sel.play_type)" :key="opt"
              @click="updateSelection(index, 'selection', opt)"
              :class="[
                'px-3 py-1.5 rounded-lg text-sm font-medium transition-all cursor-pointer border',
                sel.selection === opt
                  ? 'bg-blue-500 text-white border-blue-500'
                  : 'bg-white text-slate-600 border-slate-200 hover:border-blue-300'
              ]">
              {{ getOptionLabel(sel.play_type, opt) }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="currentSelections.length > 0" class="mt-3 p-3 bg-slate-50 rounded-lg">
      <span class="text-xs text-slate-400">投注预览：</span>
      <span class="text-xs text-slate-600 ml-2 font-mono">{{ modelValue }}</span>
    </div>
  </div>
</template>
