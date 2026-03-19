import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: () => import('@/views/HomeView.vue'),
      meta: { title: '首页' }
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/LoginView.vue'),
      meta: { title: '登录', guest: true }
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { title: '注册', guest: true }
    },
    {
      path: '/bridges',
      name: 'Bridges',
      component: () => import('@/views/bridge/BridgeListView.vue'),
      meta: { title: '桥梁管理', requiresAuth: true }
    },
    {
      path: '/bridges/:id',
      name: 'BridgeDetail',
      component: () => import('@/views/bridge/BridgeDetailView.vue'),
      meta: { title: '桥梁详情', requiresAuth: true }
    },
    {
      path: '/drones',
      name: 'Drones',
      component: () => import('@/views/drone/DroneListView.vue'),
      meta: { title: '无人机管理', requiresAuth: true }
    },
    {
      path: '/detection/image',
      name: 'ImageDetection',
      component: () => import('@/views/detection/ImageDetectionView.vue'),
      meta: { title: '图片识别', requiresAuth: true }
    },
    {
      path: '/detection/video',
      name: 'VideoDetection',
      component: () => import('@/views/detection/VideoDetectionView.vue'),
      meta: { title: '视频流识别', requiresAuth: true }
    },
    {
      path: '/detection/result/:id',
      name: 'DetectionResult',
      component: () => import('@/views/detection/DetectionResultView.vue'),
      meta: { title: '检测结果', requiresAuth: true }
    },
    {
      path: '/analysis',
      name: 'Analysis',
      component: () => import('@/views/analysis/AnalysisView.vue'),
      meta: { title: '智能分析', requiresAuth: true }
    },
    {
      path: '/reports',
      name: 'Reports',
      component: () => import('@/views/report/ReportListView.vue'),
      meta: { title: '报表管理', requiresAuth: true }
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('@/views/user/UserListView.vue'),
      meta: { title: '用户管理', requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/profile',
      name: 'Profile',
      component: () => import('@/views/user/ProfileView.vue'),
      meta: { title: '个人中心', requiresAuth: true }
    },
    {
      path: '/cases',
      name: 'Cases',
      component: () => import('@/views/CaseView.vue'),
      meta: { title: '典型案例' }
    },
    {
      path: '/contact',
      name: 'Contact',
      component: () => import('@/views/ContactView.vue'),
      meta: { title: '联系方式' }
    },
    {
      path: '/security',
      name: 'Security',
      component: () => import('@/views/SecurityView.vue'),
      meta: { title: '安全中心' }
    },
    {
      path: '/about',
      name: 'About',
      component: () => import('@/views/AboutView.vue'),
      meta: { title: '关于我们' }
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  // 1. 【核心修正】如果 Store 里的 userInfo 为空，先尝试从 LocalStorage 初始化一次
  // 这样可以保证页面刷新时，登录状态能瞬间找回
  if (!userStore.userInfo) {
    userStore.initUserInfo()
  }

  const isLoggedIn = userStore.isLoggedIn
  const isAdmin = userStore.isAdmin

  // 2. 处理“必须登录”的路由
  if (to.meta.requiresAuth && !isLoggedIn) {
    ElMessage.warning('请先登录')
    next('/login')
    return
  }

  // 3. 处理“访客”路由（已登录用户访问登录/注册页直接跳走）
  if (to.meta.guest && isLoggedIn) {
    next('/')
    return
  }

  // 4. 处理“管理员权限”路由
  if (to.meta.requiresAdmin && !isAdmin) {
    ElMessage.error('权限不足')
    next('/')
    return
  }

  // 5. 设置标题并放行
  document.title = `${to.meta.title || ''} - 桥梁表观病害智能检测系统`
  next()
})

export default router
