<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { purchaseApi } from '../api'
import type { LotteryType } from '../types'
import { LOTTERY_CONFIGS } from '../types'
import { parseNumberLines, formatHint } from '../utils/numberParse'
import { Plus, X, AlertCircle, CheckCircle } from 'lucide-vue-next'

const props = defineProps<{
  /** 打开时用于校验期号的开奖记录期号列表（按类型过滤） */
  drawIssues: Record<string, string[]>
}>()

const emit = defineEmits<{
  close: []
  /** 导入成功（已刷新列表由父组件处理） */
  imported: [count: number]
}>()

const lotteryTypes = LOTTERY_CONFIGS.map(c => c.type)
const lotteryType = ref<LotteryType>('双色球')
const issueNumber = ref('')
const purchaseDate = ref(new Date().toISOString().split('T')[0])
const multiple = ref(1)
const periods = ref(1)
const append = ref(false)
const remark = ref('')
const text = ref('')
const importing = ref(false)

const textareaRef = ref<HTMLTextAreaElement | null>(null)

watch(lotteryType, () => {
  // 切换彩种后重新解析
  nextTick(() => textareaRef.value?.focus())
})

const result = computed(() => parseNumberLines(lotteryType.value, text.value))
const okLines = computed(() => result.value.lines.filter(l => l.numbers))
const failLines = computed(() => result.value.lines.filter(l => l.error))

// 金额 = 总注数 × 2元 × 倍数（追加 +1 元/注）
const totalAmount = computed(() => {
  const unit = append.value && lotteryType.value === '大乐透' ? 3 : 2
  return result.value.totalBets * unit * multiple.value
})

// 大乐透追加只对复试生效的提示与金额有关，解析不感知 append
const canSubmit = computed(() =>
  !!issueNumber.value.trim() &&
  result.value.okCount > 0 &&
  result.value.failCount === 0 &&
  !importing.value &&
  totalAmount.value > 0
)

const validateIssue = (): string => {
  const input = issueNumber.value.trim()
  if (!input) return '期号为必填项'
  const issues = (props.drawIssues[lotteryType.value] || []).slice().sort()
  if (issues.length > 0 && input < issues[issues.length - 1]) {
    return `期号不能小于当前最大期号 ${issues[issues.length - 1]}`
  }
  return ''
}

const issueError = ref('')

const doImport = async () => {
  issueError.value = validateIssue()
  if (issueError.value) return
  if (result.value.failCount > 0) return
  if (importing.value) return

  importing.value = true
  try {
    const items = okLines.value.map(l => ({
      lottery_type: lotteryType.value,
      issue_number: issueNumber.value.trim(),
      purchase_date: purchaseDate.value,
      numbers: l.numbers!,
      bet_type: l.betCount > 1 ? '复式' : '单式',
      amount: (() => {
        const unit = append.value && lotteryType.value === '大乐透' ? 3 : 2
        return l.betCount * unit * multiple.value * periods.value
      })(),
      multiple: multiple.value,
      append: append.value && lotteryType.value === '大乐透',
      periods: periods.value,
      remark: remark.value
    }))
    const res: any = await purchaseApi.createBatch(items)
    emit('imported', res?.total ?? items.length)
  } catch (e: any) {
    issueError.value = e?.response?.data?.error || '导入失败，请重试'
  } finally {
    importing.value = false
  }
}

const close = () => emit('close')
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="close"></div>
    <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[92vh] flex flex-col animate-slide-up">
      <div class="sticky top-0 bg-white border-b border-slate-100 px-6 py-4 flex items-center justify-between rounded-t-2xl shrink-0">
        <h2 class="text-lg font-semibold text-slate-800 flex items-center gap-2">
          <span class="inline-flex items-center justify-center w-8 h-8 rounded-xl bg-blue-50">
            <Plus class="w-4.5 h-4.5 text-blue-500" />
          </span>
          批量录入
        </h2>
        <button @click="close" class="p-1.5 text-slate-400 hover:text-slate-600 cursor-pointer">
          <X class="w-5 h-5" />
        </button>
      </div>

      <div class="p-6 space-y-4 overflow-y-auto">
        <!-- 彩种与期号日期 -->
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">彩票类型</label>
            <select v-model="lotteryType" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="t in lotteryTypes" :key="t" :value="t">{{ t }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">
              期号<span class="text-red-500 ml-0.5">*</span>
            </label>
            <input
              v-model="issueNumber"
              placeholder="如：2024015"
              @input="issueError = ''"
              :class="['w-full px-4 py-2.5 border rounded-xl focus:outline-none transition-colors',
                issueError ? 'border-red-400 focus:border-red-400 bg-red-50' : 'border-slate-200 focus:border-blue-400']"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">购买日期</label>
            <input v-model="purchaseDate" type="date" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
          </div>
        </div>

        <!-- 倍数 / 期数 / 追加 -->
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">倍数</label>
            <select v-model.number="multiple" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="n in 99" :key="n" :value="n">{{ n }} 倍</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-600 mb-1.5">期数（连买）</label>
            <select v-model.number="periods" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 cursor-pointer">
              <option v-for="n in 10" :key="n" :value="n">{{ n }} 期</option>
            </select>
          </div>
          <div v-if="lotteryType === '大乐透'" class="flex items-end pb-1">
            <label class="flex items-center gap-2 text-sm text-slate-600 cursor-pointer">
              <input v-model="append" type="checkbox" class="w-4 h-4 text-blue-500 border-slate-300 rounded focus:ring-blue-400 cursor-pointer" />
              追加 (+1元/注)
            </label>
          </div>
        </div>

        <!-- 号码文本域 -->
        <div>
          <div class="flex items-center justify-between mb-1.5">
            <label class="block text-sm font-medium text-slate-600">
              号码<span class="text-red-500 ml-0.5">*</span>
              <span class="text-slate-400 font-normal ml-2">每行一注，支持复式</span>
            </label>
            <button
              type="button"
              v-if="text"
              @click="text = ''"
              class="text-xs text-slate-400 hover:text-red-500 transition-colors cursor-pointer"
            >清空</button>
          </div>
          <textarea
            ref="textareaRef"
            v-model="text"
            rows="8"
            :placeholder="formatHint(lotteryType)"
            class="w-full px-4 py-3 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400 font-mono text-sm leading-relaxed resize-y"
          ></textarea>
          <p class="mt-1 text-xs text-slate-400">{{ formatHint(lotteryType) }}；空行与 # 开头的行会被忽略</p>
        </div>

        <!-- 解析预览 -->
        <div v-if="text.trim()" class="space-y-2">
          <div class="flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm"
            :class="failLines.length > 0 ? 'bg-red-50 border border-red-200' : 'bg-emerald-50 border border-emerald-200'">
            <CheckCircle v-if="failLines.length === 0" class="w-4 h-4 text-emerald-500 shrink-0" />
            <AlertCircle v-else class="w-4 h-4 text-red-500 shrink-0" />
            <span :class="failLines.length > 0 ? 'text-red-600' : 'text-emerald-700'">
              识别 {{ result.okCount }} 注 / 共 {{ result.totalBets }} 注（含复式）
              <template v-if="failLines.length > 0">，{{ failLines.length }} 行有问题</template>
              · 金额 ¥{{ totalAmount }} × {{ periods }} 期 = ¥{{ totalAmount * periods }}
            </span>
          </div>

          <!-- 错误行列表 -->
          <div v-if="failLines.length > 0" class="max-h-36 overflow-y-auto space-y-1 px-1">
            <div v-for="l in failLines" :key="l.line" class="flex items-start gap-2 text-xs text-red-600 bg-red-50 rounded-lg px-3 py-1.5">
              <span class="shrink-0 font-mono text-red-400">第{{ l.line }}行</span>
              <span class="font-mono text-slate-500 break-all shrink-0 max-w-[40%] truncate" :title="l.raw">{{ l.raw }}</span>
              <span>{{ l.error }}</span>
            </div>
          </div>

          <!-- 成功行预览 -->
          <div v-if="okLines.length > 0 && failLines.length === 0" class="max-h-36 overflow-y-auto space-y-1 px-1">
            <div v-for="l in okLines" :key="l.line" class="flex items-center gap-2 text-xs text-slate-600 bg-slate-50 rounded-lg px-3 py-1.5">
              <span class="shrink-0 font-mono text-slate-400">第{{ l.line }}行</span>
              <span class="font-mono break-all">{{ l.numbers }}</span>
              <span v-if="l.betCount > 1" class="shrink-0 px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-medium">复式 {{ l.betCount }} 注</span>
            </div>
          </div>
        </div>

        <!-- 备注 -->
        <div>
          <label class="block text-sm font-medium text-slate-600 mb-1.5">备注</label>
          <input v-model="remark" placeholder="可选，会写入每条记录" class="w-full px-4 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-400" />
        </div>

        <p v-if="issueError" class="text-xs text-red-500 flex items-center gap-1">
          <span>⚠</span> {{ issueError }}
        </p>
      </div>

      <div class="sticky bottom-0 bg-white border-t border-slate-100 px-6 py-4 flex items-center justify-between gap-3 rounded-b-2xl shrink-0">
        <span class="text-sm text-slate-500">
          共 <strong class="text-slate-700">{{ result.totalBets * periods }}</strong> 注
          · <strong class="text-slate-700">¥{{ totalAmount * periods }}</strong>
        </span>
        <div class="flex gap-3">
          <button @click="close" class="px-5 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-medium transition-colors cursor-pointer">取消</button>
          <button
            @click="doImport"
            :disabled="!canSubmit"
            class="px-5 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-medium transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {{ importing ? '导入中...' : `导入 ${result.okCount} 条` }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
