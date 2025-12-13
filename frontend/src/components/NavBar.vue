<template>
  <!-- 顶部导航栏 -->
  <header class="bg-light-100/80 dark:bg-dark-300/80 backdrop-blur-md border-b border-light-200 dark:border-dark-100 py-3 fixed top-0 left-0 right-0 z-40 transition-all duration-300">
    <div class="container mx-auto px-4">
      <div class="flex justify-between items-center">
        <!-- 左侧Logo和菜单按钮 -->
        <div class="flex items-center gap-3">
          <button 
            ref="sidebarToggleRef"
            class="w-10 h-10 rounded-md bg-light-200 dark:bg-dark-100 text-secondary hover:bg-light-300 dark:hover:bg-dark-200 transition-all duration-200 flex items-center justify-center"
          >
            <i class="ri-align-justify"></i>
          </button>
          <div class="flex items-center gap-2 font-semibold text-xl select-none cursor-default group">
            <div class="w-10 h-10 rounded-md bg-gradient-to-br from-primary to-primary-dark flex items-center justify-center text-white font-bold shadow-lg shadow-primary/20 transition-transform duration-300 group-hover:rotate-6">雾</div>
            <span class="relative">
              <span class="bg-clip-text text-transparent bg-gradient-to-r from-gray-800 to-gray-600 dark:from-white dark:to-gray-300 group-hover:from-primary group-hover:to-purple-500 transition-all duration-300">初春图床</span>
              <span class="absolute -bottom-1 left-0 w-full h-[3px] bg-gradient-to-r from-primary/40 via-purple-400/40 to-transparent rounded-full opacity-60 group-hover:opacity-100 group-hover:scale-x-110 transition-all duration-300 origin-left"></span>
            </span>
          </div>
        </div>
        
        <!-- 右侧操作区 - 只保留主题切换 -->
        <div class="flex items-center gap-2 md:gap-4">
          <button 
            ref="themeToggleRef"
            class="flex-shrink-0 w-10 h-10 rounded-md bg-light-200 dark:bg-dark-100 text-secondary hover:bg-light-300 dark:hover:bg-dark-200 hover:text-primary transition-all duration-200 flex items-center justify-center"
          >
            <i class="ri-moon-clear-line dark:hidden"></i>
            <i class="ri-sun-line dark:inline-block hidden"></i>
          </button>
        </div>
      </div>
    </div>
  </header>

  <!-- 侧边栏 -->
  <div 
    ref="sidebarRef"
    class="fixed top-0 left-0 h-full w-64 bg-light-100 dark:bg-dark-300 border-r border-light-200 dark:border-dark-100 z-50 transition-transform duration-300 sidebar-closed flex flex-col"
  >
    <div class="p-4 border-b border-light-200 dark:border-dark-100">
        <h3 class="font-medium text-secondary">导航菜单</h3>
    </div>
    <nav class="p-2 flex-1">
        <ul class="space-y-1">
          <li v-for="item in navItems" :key="item.path">
            <router-link
              :to="item.path"
              :class="[
                'flex items-center px-3 py-3 rounded-md transition-all duration-200',
                isRouteActive(item.path) ? 'bg-primary/10 text-primary' : 'hover:bg-light-100 dark:hover:bg-dark-300 text-secondary hover:text-primary'
              ]"
              @click="handleNavClick"
            >
              <i :class="`ri-${item.icon} w-6 text-center`"></i>
              <span class="ml-3">{{ item.name }}</span>
            </router-link>
          </li>
        </ul>
    </nav>

    <!-- 侧边栏底部 - 用户信息和登录/登出 -->
    <div class="border-t border-light-200 dark:border-dark-100 p-3">
      <!-- 已登录状态 -->
      <div v-if="isLoggedIn" class="space-y-2">
        <!-- 用户信息区域 - 点击跳转账户设置 -->
        <router-link 
          v-if="!isTourist"
          to="/account"
          class="flex items-center gap-3 p-2 rounded-lg hover:bg-light-200 dark:hover:bg-dark-200 transition-colors cursor-pointer"
          @click="handleNavClick"
        >
          <div class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-400 flex items-center justify-center text-white text-sm font-bold overflow-hidden flex-shrink-0">
            <img v-if="userProfile.avatar" :src="userProfile.avatar" class="w-full h-full object-cover" alt="" />
            <span v-else>{{ (userProfile.nickname || userProfile.username || 'U').charAt(0).toUpperCase() }}</span>
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-medium text-sm truncate">{{ userProfile.nickname || userProfile.username || '用户' }}</p>
            <p class="text-xs text-secondary truncate">点击编辑资料</p>
          </div>
        </router-link>

        <!-- 游客状态显示 -->
        <div v-else class="flex items-center gap-3 p-2">
          <div class="w-10 h-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center text-white text-sm font-bold flex-shrink-0">
            <i class="ri-user-line"></i>
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-medium text-sm">游客</p>
            <p class="text-xs text-secondary">临时访客</p>
          </div>
        </div>

        <!-- 登出按钮 -->
        <button 
          @click="handleLogout"
          class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/30 transition-colors text-sm"
        >
          <i class="ri-logout-circle-r-line"></i>
          <span>退出登录</span>
        </button>
      </div>

      <!-- 未登录状态 -->
      <router-link 
        v-else
        to="/login"
        class="flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-primary hover:bg-primary-dark text-white transition-colors text-sm"
        @click="handleNavClick"
      >
        <i class="ri-login-circle-line"></i>
        <span>登录</span>
      </router-link>
    </div>
  </div>

  <!-- 侧边栏遮罩 -->
  <div
    ref="sidebarOverlayRef"
    class="fixed inset-0 bg-black/50 z-40 overlay-hidden transition-opacity duration-300 pt-16"
  ></div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// 1. 定义 ref 引用
const themeToggleRef = ref(null)
const sidebarToggleRef = ref(null)
const sidebarRef = ref(null)
const sidebarOverlayRef = ref(null)

// 用户状态
const isLoggedIn = ref(false)
const isTourist = ref(false)
const userProfile = reactive({
  username: '',
  nickname: '',
  avatar: ''
})

// 2. 导航菜单数据（移除账户设置，用底部用户区域替代）
const navItems = ref([
  { path: '/', icon: 'home-line', name: '首页' },
  { path: '/gallery', icon: 'nft-line', name: '画廊' },
  { path: '/stats', icon: 'numbers-fill', name: '统计' }
])

// 初始化时根据用户类型添加设置菜单
const initNavItems = () => {
  const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}')
  isLoggedIn.value = !!localStorage.getItem('authToken')
  isTourist.value = userInfo?.isTourist === true
  
  // 只有非游客才显示系统设置
  if (!isTourist.value && isLoggedIn.value) {
    if (!navItems.value.find(item => item.path === '/settings')) {
      navItems.value.push({ path: '/settings', icon: 'settings-line', name: '系统设置' })
    }
  }
  
  // 加载用户资料
  if (isLoggedIn.value && !isTourist.value) {
    fetchUserProfile()
  }
}

// 获取用户资料
const fetchUserProfile = async () => {
  try {
    const response = await fetch('/api/user/profile', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('authToken')}`
      }
    })
    
    if (response.ok) {
      const result = await response.json()
      if (result.code === 200 && result.data) {
        Object.assign(userProfile, result.data)
      }
    }
  } catch (error) {
    console.error('获取用户资料失败:', error)
  }
}

const isRouteActive = (targetPath) => {
  const exactMatchPaths = ['/', '/login', '/404']
  if (exactMatchPaths.includes(targetPath)) {
    return route.path === targetPath
  }
  return route.path.startsWith(targetPath)
}

// 3. 导航点击事件
const handleNavClick = () => {
  closeSidebar()
}

// 4. 主题切换功能
const storageKey = 'theme-preference'

const detectUserThemePreference = () => {
  if (typeof localStorage !== 'undefined' && localStorage.getItem(storageKey)) {
    return localStorage.getItem(storageKey)
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

const applyTheme = (theme) => {
  const htmlElement = document.documentElement
  if (theme === 'dark') {
    htmlElement.classList.add('dark')
  } else {
    htmlElement.classList.remove('dark')
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(storageKey, theme)
  }
}

// 5. 侧边栏控制功能
const openSidebar = () => {
  if (sidebarRef.value) {
    sidebarRef.value.classList.remove('sidebar-closed')
    sidebarRef.value.classList.add('sidebar-open')
  }
  if (sidebarOverlayRef.value) {
    sidebarOverlayRef.value.classList.remove('overlay-hidden')
    sidebarOverlayRef.value.classList.add('overlay-visible')
  }
  document.body.style.overflow = 'hidden'
}

const closeSidebar = () => {
  if (sidebarRef.value) {
    sidebarRef.value.classList.remove('sidebar-open')
    sidebarRef.value.classList.add('sidebar-closed')
  }
  if (sidebarOverlayRef.value) {
    sidebarOverlayRef.value.classList.remove('overlay-visible')
    sidebarOverlayRef.value.classList.add('overlay-hidden')
  }
  document.body.style.overflow = ''
}

// 6. 登出功能
const handleLogout = async () => {
  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem('authToken')
    localStorage.removeItem('userInfo')
  }
  closeSidebar()
  try {
    await fetch('/api/logout', {
      method: 'POST'
    })
    router.push('/login').catch(err => {
      console.log('跳转登录页失败：', err)
    })
  } catch (error) {
    console.error('登出失败:', error)
  }
}  

// 7. 组件挂载时初始化
onMounted(() => {
  // 初始化导航菜单和用户状态
  initNavItems()

  // 初始化主题
  const initialTheme = detectUserThemePreference()
  applyTheme(initialTheme)

  // 绑定主题切换事件
  if (themeToggleRef.value) {
    themeToggleRef.value.addEventListener('click', () => {
      const currentTheme = localStorage.getItem(storageKey) || 'light'
      const newTheme = currentTheme === 'dark' ? 'light' : 'dark'
      applyTheme(newTheme)
    })
  }

  // 绑定侧边栏打开事件
  if (sidebarToggleRef.value) {
    sidebarToggleRef.value.addEventListener('click', openSidebar)
  }

  // 绑定侧边栏遮罩关闭事件
  if (sidebarOverlayRef.value) {
    sidebarOverlayRef.value.addEventListener('click', closeSidebar)
  }

  // 窗口大小变化事件
  const handleResize = () => {
    if (window.innerWidth >= 1024) {
      closeSidebar()
    }
  }
  window.addEventListener('resize', handleResize)
})

// 8. 组件卸载时清理
onUnmounted(() => {
  // 移除主题切换事件
  if (themeToggleRef.value) {
    themeToggleRef.value.removeEventListener('click', () => {})
  }

  // 移除侧边栏打开事件
  if (sidebarToggleRef.value) {
    sidebarToggleRef.value.removeEventListener('click', openSidebar)
  }

  // 移除侧边栏遮罩关闭事件
  if (sidebarOverlayRef.value) {
    sidebarOverlayRef.value.removeEventListener('click', closeSidebar)
  }

  // 移除窗口 resize 事件
  window.removeEventListener('resize', () => {})

  // 恢复页面滚动
  document.body.style.overflow = ''
})

// 9. 初始化侧边栏状态
onMounted(() => {
  if (window.innerWidth >= 1024) {
    closeSidebar()
  }
})
</script>

<style scoped>
/* 侧边栏滚动样式 */
::v-deep(.sidebar-open) {
  overflow-y: auto;
}

::v-deep(.sidebar-open)::-webkit-scrollbar {
  width: 4px;
}
::v-deep(.sidebar-open)::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}
::v-deep(.sidebar-open)::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>