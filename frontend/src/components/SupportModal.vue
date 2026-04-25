<script setup lang="ts">
import { X, ExternalLink } from 'lucide-vue-next'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const close = () => {
  emit('update:modelValue', false)
}

const hideOnError = (e: Event) => {
  const img = e.currentTarget as HTMLImageElement
  img.style.display = 'none'
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
          <div
            v-if="modelValue"
            class="bg-white rounded-2xl shadow-xl w-full max-w-[380px] relative overflow-visible text-center"
          >
            <!-- 关闭按钮 -->
            <button
              @click="close"
              class="absolute top-2.5 right-3 bg-none border-none text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-all px-2 py-1 rounded-md text-lg leading-none z-10"
            >
              ✕
            </button>

            <div class="px-6 pt-7 pb-5">
              <!-- 标题 -->
              <div class="text-xl font-bold text-red-500 mb-3">
                ☕ 请站长喝杯咖啡？
              </div>

              <!-- 搞怪文案 -->
              <div class="text-sm leading-7 text-slate-500 mb-4">
                <p>嘿～看到这里的都是真爱！🎉</p>
                <p>愿意支持的随意一下，感觉麻烦的直接跳过～</p>
                <p class="text-xs text-slate-400 italic mt-1">反正我也就随便说说，你也就随便看看 😜</p>
              </div>

              <!-- 二维码 -->
              <div class="flex justify-center gap-6 my-4">
                <div class="text-center">
                  <img
                    src="/img/wechat-qr.png"
                    alt="微信收款码"
                    class="w-[150px] h-[150px] rounded-lg border border-slate-200 object-contain bg-white"
                    @error="hideOnError"
                  />
                  <span class="block mt-1.5 text-[13px] text-slate-500">微信</span>
                </div>
                <div class="text-center">
                  <img
                    src="/img/alipay-qr.jpg"
                    alt="支付宝收款码"
                    class="w-[150px] h-[150px] rounded-lg border border-slate-200 object-contain bg-white"
                    @error="hideOnError"
                  />
                  <span class="block mt-1.5 text-[13px] text-slate-500">支付宝</span>
                </div>
              </div>

              <!-- 底部链接 -->
              <div class="flex justify-center gap-4 mt-4 pt-4 border-t border-slate-100">
                <a
                  href="http://techfunway.wycto.cn"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-[13px] text-blue-500 hover:text-blue-700 hover:underline transition-colors"
                >
                  🏠 博主首页
                </a>
                <a
                  href="http://techfunway.wycto.cn/fnapp/lottery/"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-[13px] text-blue-500 hover:text-blue-700 hover:underline transition-colors"
                >
                  📖 应用文档
                </a>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>
