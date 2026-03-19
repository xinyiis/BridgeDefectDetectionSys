import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { login as loginApi, register as registerApi, logout as logoutApi } from '@/api/auth'
import type { LoginParams, RegisterParams } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  // 1. 用户信息响应式变量
  const userInfo = ref<User | null>(null)

  // 2. 计算属性：判断登录状态和权限
  // 增加判定：只有当 userInfo 存在且包含基本信息时才算已登录
  const isLoggedIn = computed(() => {
    return userInfo.value !== null && (!!userInfo.value.id || !!userInfo.value.user_id || !!userInfo.value.username)
  })

  const isAdmin = computed(() => userInfo.value?.role === 'admin' || userInfo.value?.role === 1)
  const username = computed(() => userInfo.value?.username || '')
  const realName = computed(() => userInfo.value?.real_name || '')

  // 3. 存储用户信息的方法
  const setUserInfo = (info: User) => {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  // 4. 登录逻辑 (重点修改：全兼容数据解构)
  const login = async (params: LoginParams) => {
    const data = await loginApi(params)

    // 兼容性提取：
    // 如果后端回的是 { user: {...} } 则取 data.user
    // 如果后端直接回的是 {...} 则取 data 本身
    const userData = data?.user || data

    // 严谨校验：确保 userData 不是 undefined 或空对象
    if (userData && (userData.id || userData.user_id || userData.username)) {
      setUserInfo(userData)
      return data
    } else {
      console.error('登录解析失败！后端返回的数据结构是：', data)
      throw new Error('未获取到有效的用户信息，请检查后端返回结构')
    }
  }

  // 5. 注册逻辑
  const register = async (params: RegisterParams) => {
    const data = await registerApi(params)
    return data
  }

  // 6. 登出逻辑
  const logout = async () => {
    try {
      await logoutApi()
    } finally {
      userInfo.value = null
      localStorage.removeItem('userInfo')
    }
  }

  // 7. 初始化逻辑（从本地恢复状态）
  const initUserInfo = () => {
    const storedUserInfo = localStorage.getItem('userInfo')
    if (storedUserInfo) {
      try {
        const parsed = JSON.parse(storedUserInfo)
        // 只有解析出来确实有内容才赋值
        if (parsed && typeof parsed === 'object') {
          userInfo.value = parsed
        }
      } catch (e) {
        console.error('解析本地用户信息失败:', e)
        userInfo.value = null
        localStorage.removeItem('userInfo')
      }
    }
  }

  return {
    userInfo,
    isLoggedIn,
    isAdmin,
    username,
    realName,
    login,
    register,
    logout,
    initUserInfo,
    setUserInfo
  }
})