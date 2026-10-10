<script setup lang="ts">
import { ref } from 'vue'
import { X, ExternalLink } from 'lucide-vue-next'
import api from '../api'

defineProps<{
  modelValue: boolean
  // 是否显示底部操作按钮（暂不支持 / 已支持）
  showActions?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  /** 用户点【暂不支持】或关闭：本次运行内不再提示 */
  dismiss: []
  /** 【已支持】上报成功：永不再提示 */
  supported: []
}>()

const sending = ref(false)
const errorText = ref('')

const close = () => {
  emit('dismiss')
  emit('update:modelValue', false)
}

const hideOnError = (e: Event) => {
  const img = e.currentTarget as HTMLImageElement
  img.style.display = 'none'
}

// 【已支持】发送匿名支持计数，成功才通知父组件永久记住
const confirmSupported = async () => {
  sending.value = true
  errorText.value = ''
  let ok = false
  try {
    // api 实例的响应拦截器直接返回响应体（{ok:true}），类型上仍标注为 AxiosResponse
    const data = await api.post('/donate/support') as unknown as { ok?: boolean }
    ok = data?.ok === true
  } catch {
    ok = false
  }
  sending.value = false
  if (ok) {
    emit('supported')
    emit('update:modelValue', false)
  } else {
    errorText.value = '发送失败，请稍后重试（不影响你的支持 ❤）'
  }
}
</script>

<template>
  <teleport to="body">
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        @click.self="close"
      >
        <transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <!-- 精简为一屏展示（mobile-ui 规范 2026-10-01） -->
          <div
            v-if="modelValue"
            class="bg-white rounded-2xl shadow-xl w-full max-w-[320px] relative overflow-visible text-center"
          >
            <!-- 关闭按钮 -->
            <button
              @click="close"
              class="absolute top-2 right-2.5 bg-none border-none text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-all px-2 py-1 rounded-md text-base leading-none z-10"
            >
              ✕
            </button>

            <div class="px-5 pt-5 pb-4">
              <!-- 标题 -->
              <div class="text-base font-bold text-red-500 mb-1.5">
                ☕ 请站长喝杯咖啡？
              </div>

              <!-- 精简文案 -->
              <p class="text-xs leading-relaxed text-slate-500 mb-3">
                应用免费无广告、数据保存在本机。若对你有帮助，欢迎赞赏——<strong class="text-slate-700">金额随意，1 元也是心意</strong>。
              </p>

              <!-- 二维码（仅微信） -->
              <div class="flex justify-center my-2">
                <div class="text-center">
                  <img
                    src="/img/wechat-qr.png"
                    alt="微信收款码"
                    class="w-[150px] h-auto rounded-lg border border-slate-200 object-contain bg-white"
                    @error="hideOnError"
                  />
                  <span class="block mt-1 text-[11px] text-slate-500">微信扫码</span>
                </div>
              </div>

              <!-- 底部链接 -->
              <div class="flex justify-center gap-3 mt-3 pt-2.5 border-t border-slate-100">
                <a
                  href="http://techfunway.wycto.cn"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-[11px] text-blue-500 hover:text-blue-700 hover:underline transition-colors"
                >
                  🏠 博主首页
                </a>
                <a
                  href="http://techfunway.wycto.cn/fnapp/lottery/"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-[11px] text-blue-500 hover:text-blue-700 hover:underline transition-colors"
                >
                  📖 应用文档
                </a>
              </div>

              <!-- 匿名计数说明与操作按钮 -->
              <template v-if="showActions">
                <p class="text-[11px] text-slate-400 mt-2.5 leading-4">
                  已支付可点「已支持」发送一次匿名计数（与支付记录无关）
                </p>
                <p v-if="errorText" class="text-xs text-rose-500 mt-1.5">{{ errorText }}</p>
                <div class="flex justify-center gap-2.5 mt-2.5">
                  <button
                    class="flex-1 px-3 py-2 text-sm text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors"
                    @click="close"
                  >
                    暂不支持
                  </button>
                  <button
                    class="flex-1 px-3 py-2 text-sm text-white bg-gradient-to-r from-indigo-500 to-violet-500 hover:opacity-90 rounded-lg transition-opacity disabled:opacity-50"
                    :disabled="sending"
                    @click="confirmSupported"
                  >
                    {{ sending ? '发送中…' : '❤ 已支持' }}
                  </button>
                </div>
              </template>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>
