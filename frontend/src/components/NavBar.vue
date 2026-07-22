<template>
  <header class="topbar">
    <div class="nav-shell">
      <router-link to="/" class="brand" aria-label="返回首页">
        <span class="brand-mark"><img :src="logoImg" alt="" /></span>
        <span><strong>初春图床</strong><small>OneImg</small></span>
      </router-link>

      <nav class="desktop-nav" aria-label="主导航">
        <router-link v-for="item in navItems" :key="item.path" :to="item.path" :class="{ active: isActive(item.path) }">
          <i :class="item.icon"></i>{{ item.name }}
        </router-link>
      </nav>

      <div class="nav-actions">
        <router-link v-if="!isGuest" to="/account" class="profile-button" title="账户设置">
          <span class="avatar">
            <img v-if="profile.avatar" :src="profile.avatar" alt="用户头像" />
            <b v-else>{{ displayName.charAt(0).toUpperCase() }}</b>
          </span>
          <span class="profile-copy"><strong>{{ displayName }}</strong><small>{{ isAdmin ? '管理员' : '普通用户' }}</small></span>
        </router-link>
        <span v-else class="guest-badge"><i class="mgc_incognito_mode_line"></i>游客</span>
        <button class="menu-button" :aria-expanded="mobileOpen" aria-label="打开导航" @click="mobileOpen = !mobileOpen">
          <i :class="mobileOpen ? 'mgc_close_line' : 'mgc_menu_line'"></i>
        </button>
      </div>
    </div>

    <transition name="menu">
      <div v-if="mobileOpen" class="mobile-panel">
        <nav>
          <router-link v-for="item in navItems" :key="item.path" :to="item.path" :class="{ active: isActive(item.path) }" @click="mobileOpen = false">
            <i :class="item.icon"></i><span>{{ item.name }}</span><i class="mgc_right_line tail"></i>
          </router-link>
          <router-link v-if="!isGuest" to="/account" :class="{ active: isActive('/account') }" @click="mobileOpen = false">
            <i class="mgc_user_edit_line"></i><span>账户</span><i class="mgc_right_line tail"></i>
          </router-link>
        </nav>
        <button class="logout-button" @click="handleLogout"><i class="mgc_exit_door_line"></i>退出登录</button>
      </div>
    </transition>
  </header>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import defaultLogo from '@/assets/logo.png'

const route = useRoute()
const router = useRouter()
const mobileOpen = ref(false)
const logoImg = ref(localStorage.getItem('site_logo') || defaultLogo)
const userInfo = reactive(JSON.parse(localStorage.getItem('userInfo') || '{}'))
const profile = reactive({ username: userInfo.username || '', nickname: userInfo.nickname || '', avatar: localStorage.getItem('user_avatar') || userInfo.avatar || '' })

const isAdmin = computed(() => userInfo.role === 1 || userInfo.isAdmin === true)
const isGuest = computed(() => userInfo.role === 3 || userInfo.isTourist === true)
const displayName = computed(() => profile.nickname || profile.username || (isGuest.value ? '游客' : '用户'))
const navItems = computed(() => [
  { path: '/', name: '上传', icon: 'mgc_home_3_line' },
  { path: '/gallery', name: '画廊', icon: 'mgc_pic_2_line' },
  { path: '/stats', name: '统计', icon: 'mgc_chart_bar_line' },
  ...(isAdmin.value ? [{ path: '/settings', name: '设置', icon: 'mgc_settings_3_line' }] : []),
])

const isActive = (path) => path === '/' ? route.path === '/' : route.path.startsWith(path)

const fetchProfile = async () => {
  if (isGuest.value) return
  try {
    const response = await fetch('/api/user/profile')
    const result = await response.json()
    if (response.ok && result.code === 200) {
      Object.assign(profile, result.data)
      if (result.data.avatar) localStorage.setItem('user_avatar', result.data.avatar)
    }
  } catch (error) {
    console.error('获取用户资料失败:', error)
  }
}

const fetchBrand = async () => {
  try {
    const response = await fetch('/api/settings/login')
    const result = await response.json()
    if (response.ok && result.data?.site_logo) {
      logoImg.value = result.data.site_logo
      localStorage.setItem('site_logo', result.data.site_logo)
    }
  } catch (error) {
    console.error('获取站点图标失败:', error)
  }
}

const handleLogout = async () => {
  try { await fetch('/api/logout', { method: 'POST' }) } catch (error) { console.error('退出登录失败:', error) }
  localStorage.removeItem('authToken')
  localStorage.removeItem('userInfo')
  localStorage.removeItem('user_avatar')
  mobileOpen.value = false
  router.push('/login')
}

onMounted(() => { fetchProfile(); fetchBrand() })
</script>

<style scoped>
.topbar { position:fixed; inset:0 0 auto; z-index:40; padding:.7rem 1rem; pointer-events:none; }
.nav-shell { width:min(1180px,100%); min-height:3.75rem; margin:0 auto; padding:.45rem .55rem; display:flex; align-items:center; gap:1rem; border:1px solid rgba(212,121,143,.16); border-radius:1.15rem; background:rgba(255,255,255,.82); box-shadow:0 12px 42px rgba(84,53,62,.08); backdrop-filter:blur(22px); pointer-events:auto; }
.dark .nav-shell { background:rgba(28,33,40,.88); border-color:rgba(255,255,255,.07); box-shadow:0 15px 45px rgba(2,7,12,.32); }
.brand { display:flex; align-items:center; gap:.65rem; padding:.15rem .45rem .15rem .15rem; color:#38414b; }.dark .brand{color:#f2f4f6}.brand-mark{display:grid;place-items:center;width:2.65rem;height:2.65rem;border-radius:.82rem;background:#fff3f5;overflow:hidden}.dark .brand-mark{background:#3c3037}.brand-mark img{width:1.95rem;height:1.95rem;object-fit:contain}.brand>span:last-child{display:flex;flex-direction:column;line-height:1.08}.brand strong{font-size:.92rem;letter-spacing:-.02em}.brand small{margin-top:.18rem;color:#b07381;font-size:.58rem;font-weight:800;letter-spacing:.14em;text-transform:uppercase}
.desktop-nav { display:flex; align-items:center; gap:.2rem; margin:auto; }.desktop-nav a{display:flex;align-items:center;gap:.4rem;padding:.58rem .78rem;border-radius:.72rem;color:#7a838d;font-size:.78rem;font-weight:700;transition:.2s ease}.desktop-nav a:hover{color:#bd536b;background:#fff6f7}.desktop-nav a.active{color:#ad455e;background:#fff0f3}.dark .desktop-nav a:hover,.dark .desktop-nav a.active{color:#ff9bb0;background:#3c3037}
.nav-actions{display:flex;align-items:center;gap:.45rem}.profile-button{display:flex;align-items:center;gap:.55rem;padding:.28rem .52rem .28rem .3rem;border-radius:.78rem;transition:.2s}.profile-button:hover{background:#fff5f6}.dark .profile-button:hover{background:#333039}.avatar{display:grid;place-items:center;width:2.25rem;height:2.25rem;border-radius:.7rem;overflow:hidden;color:white;background:#d76b82}.avatar img{width:100%;height:100%;object-fit:cover}.profile-copy{display:flex;flex-direction:column;max-width:7rem;line-height:1.1}.profile-copy strong{overflow:hidden;text-overflow:ellipsis;font-size:.72rem;white-space:nowrap}.profile-copy small{margin-top:.18rem;color:#9b9295;font-size:.6rem}.guest-badge{display:flex;align-items:center;gap:.35rem;padding:.55rem .65rem;border-radius:.7rem;color:#8a7e82;background:#f5eff0;font-size:.72rem;font-weight:700}.dark .guest-badge{background:#292f37;color:#ccc2c6}
.menu-button{display:none;width:2.45rem;height:2.45rem;border-radius:.72rem;color:#9a5263;background:#fff1f4;font-size:1.15rem}.dark .menu-button{color:#ffa1b4;background:#3b3037}
.mobile-panel{display:none;pointer-events:auto}.logout-button{display:flex;align-items:center;justify-content:center;gap:.4rem;width:100%;padding:.72rem;border-radius:.75rem;color:#bb4e57;background:#fff0f0;font-size:.78rem;font-weight:700}.dark .logout-button{background:#3d2c31;color:#ff9ba2}
.menu-enter-active,.menu-leave-active{transition:.2s ease}.menu-enter-from,.menu-leave-to{opacity:0;transform:translateY(-8px)}
@media(max-width:760px){.topbar{padding:.55rem .65rem}.nav-shell{min-height:3.45rem;border-radius:1rem}.desktop-nav,.profile-button,.guest-badge{display:none}.brand{margin-right:auto}.menu-button{display:grid;place-items:center}.mobile-panel{display:block;width:calc(100% - 1.3rem);margin:.45rem auto 0;padding:.55rem;border:1px solid rgba(212,121,143,.14);border-radius:1rem;background:rgba(255,255,255,.96);box-shadow:0 18px 48px rgba(77,46,56,.14);backdrop-filter:blur(20px)}.dark .mobile-panel{background:rgba(29,34,41,.98);border-color:rgba(255,255,255,.07)}.mobile-panel nav{display:grid;gap:.15rem;margin-bottom:.45rem}.mobile-panel nav a{display:grid;grid-template-columns:1.5rem 1fr auto;align-items:center;gap:.45rem;padding:.68rem;border-radius:.72rem;color:#747d87;font-size:.78rem;font-weight:700}.mobile-panel nav a.active{color:#b64b63;background:#fff1f4}.dark .mobile-panel nav a.active{color:#ff9bb0;background:#3b3037}.mobile-panel .tail{font-size:.8rem;color:#b7afb2}}
</style>
