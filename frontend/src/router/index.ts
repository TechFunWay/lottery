import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { title: '登录', requiresAuth: false }
    },
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
      meta: { title: '仪表盘', requiresAuth: true }
    },
    {
      path: '/purchase',
      name: 'purchase',
      component: () => import('../views/PurchaseView.vue'),
      meta: { title: '购买记录', requiresAuth: true }
    },
    {
      path: '/draw',
      name: 'draw',
      component: () => import('../views/DrawView.vue'),
      meta: { title: '开奖管理', requiresAuth: true }
    },
    {
      path: '/winnings',
      name: 'winnings',
      component: () => import('../views/WinningsView.vue'),
      meta: { title: '中奖记录', requiresAuth: true }
    },
    {
      path: '/history-hit',
      name: 'history-hit',
      component: () => import('../views/HistoryHitView.vue'),
      meta: { title: '历史命中', requiresAuth: true }
    },
    {
      path: '/statistics',
      name: 'statistics',
      component: () => import('../views/StatisticsView.vue'),
      meta: { title: '统计分析', requiresAuth: true }
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/UserManageView.vue'),
      meta: { title: '后台管理', requiresAuth: true, requiresAdmin: true }
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null
  const isAdmin = user?.role === 'admin'

  // 需要认证的页面
  if (to.meta.requiresAuth) {
    // 没有token，跳转到登录页
    if (!token) {
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
      return
    }

    // 需要管理员权限
    if (to.meta.requiresAdmin && !isAdmin) {
      next('/')
      return
    }
  }

  // 已登录用户访问登录页，跳转到首页
  if (to.path === '/login' && token) {
    next('/')
    return
  }

  next()
})

export default router
