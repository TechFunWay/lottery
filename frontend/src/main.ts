import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './index.css'

// 飞牛授权走「网关跳板 + 一次性票据」：跳板页带回 fnos_ticket，这里先接住，
// 后续 /auth/fnos/login、/bind 请求由 api 拦截器附带 X-FnOS-Ticket。
function captureFnOSTicket() {
  const hash = window.location.hash.replace(/^#/, '')
  if (!hash) return
  const params = new URLSearchParams(hash)
  const ticket = params.get('fnos_ticket')
  if (!ticket) return
  sessionStorage.setItem('fnos_ticket', ticket)
  const fnosUsername = params.get('fnos_username')
  if (fnosUsername) sessionStorage.setItem('fnos_ticket_username', fnosUsername)
  params.delete('fnos_ticket')
  params.delete('fnos_username')
  const rest = params.toString()
  history.replaceState(null, '', window.location.pathname + window.location.search + (rest ? `#${rest}` : ''))
}

// 飞牛网关域 = 同 hostname 的默认端口（80/443，即 NAS 桌面门户），应用直连端口固定带
// 非默认端口（如 :8902）。直连端口与网关域是不同 origin，localStorage 不共享，
// 因此直连页面无法依赖记忆，必须按此规则构造网关跳板页地址。
// 供直连端口页面上的「飞牛授权登录」按钮跳转签票。若 referrer 已给出网关域则优先采用。
function rememberFnOSGateway() {
  const base = import.meta.env.BASE_URL || '/'
  if (document.referrer) {
    try {
      const desktop = new URL(document.referrer)
      if (desktop.protocol === 'http:' || desktop.protocol === 'https:') {
        if (desktop.hostname === window.location.hostname && desktop.origin !== window.location.origin) {
          localStorage.setItem('fnos_gateway_url', `${desktop.origin}${base}fnos-entry.html`)
          return
        }
      }
    } catch { /* 继续用默认端口规则 */ }
  }
  // 同 hostname 默认端口 = 网关域
  if (window.location.protocol === 'http:' || window.location.protocol === 'https:') {
    localStorage.setItem('fnos_gateway_url', `${window.location.protocol}//${window.location.hostname}${base}fnos-entry.html`)
  }
}

captureFnOSTicket()
rememberFnOSGateway()

createApp(App).use(router).mount('#app')
