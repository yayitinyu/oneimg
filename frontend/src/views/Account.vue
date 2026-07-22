<template>
  <div class="account-page">
    <header>
      <p class="eyebrow">Personal space</p>
      <h1>账户设置</h1>
      <p>更新你的公开资料与登录凭据。</p>
    </header>

    <div class="account-grid">
      <section class="profile-card">
        <div class="avatar-editor" @click="avatarInput?.click()">
          <img v-if="profile.avatar" :src="profile.avatar" alt="当前头像" />
          <span v-else>{{ displayInitial }}</span>
          <button title="更换头像"><i class="mgc_camera_line"></i></button>
        </div>
        <input ref="avatarInput" class="hidden" type="file" accept="image/*" @change="handleAvatarSelect" />
        <h2>{{ profile.nickname || profile.username || '用户' }}</h2>
        <p>@{{ profile.username }}</p>
        <span class="role-badge"><i :class="isAdmin ? 'mgc_medal_line' : 'mgc_user_2_line'"></i>{{ isAdmin ? '管理员' : '普通用户' }}</span>

        <form class="profile-form" @submit.prevent="updateProfile">
          <label><span>昵称</span><input v-model.trim="nickname" maxlength="32" placeholder="你希望展示的名字" /></label>
          <button class="secondary-button" :disabled="updatingProfile"><i class="mgc_check_line"></i>保存资料</button>
        </form>
      </section>

      <section class="credentials-card">
        <div class="section-heading">
          <span class="icon-tile"><i class="mgc_key_2_line"></i></span>
          <div><h2>登录凭据</h2><p>用户名或密码变更后需要重新登录。</p></div>
        </div>
        <form @submit.prevent="updateAccount">
          <label><span>新用户名</span><input v-model.trim="account.newUsername" minlength="3" maxlength="32" placeholder="留空则不修改" /></label>
          <label><span>当前密码</span><input v-model="account.currentPassword" type="password" autocomplete="current-password" required placeholder="用于确认身份" /></label>
          <div class="two-column">
            <label><span>新密码</span><input v-model="account.newPassword" type="password" autocomplete="new-password" minlength="6" placeholder="留空则不修改" /></label>
            <label><span>确认新密码</span><input v-model="account.confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入" /></label>
          </div>
          <button class="primary-button" :disabled="updatingAccount"><i :class="updatingAccount ? 'mgc_loading_line animate-spin' : 'mgc_safe_lock_line'"></i>{{ updatingAccount ? '正在更新' : '更新登录信息' }}</button>
        </form>
      </section>

      <aside v-if="isAdmin" class="database-card">
        <div class="section-heading">
          <span class="icon-tile blue"><i class="mgc_storage_line"></i></span>
          <div><h2>数据库</h2><p>管理员可查看当前连接状态。</p></div>
        </div>
        <div class="db-metric"><span>数据库类型</span><strong>{{ formatDbType(dbStatus.type) }}</strong></div>
        <div class="db-metric"><span>连接状态</span><strong :class="dbStatus.connected ? 'online' : 'offline'"><i :class="dbStatus.connected ? 'mgc_check_circle_fill' : 'mgc_close_circle_fill'"></i>{{ dbStatus.connected ? '正常' : '异常' }}</strong></div>
      </aside>
    </div>

    <ImageCropper v-model:visible="showCropper" :image-src="cropperImage" @cropped="uploadAvatar" />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import ImageCropper from '@/components/ImageCropper.vue'
import message from '@/utils/message.js'

const router = useRouter()
const profile = reactive({ username: '', nickname: '', avatar: '', role: 2 })
const nickname = ref('')
const avatarInput = ref(null)
const cropperImage = ref('')
const showCropper = ref(false)
const updatingProfile = ref(false)
const updatingAccount = ref(false)
const account = reactive({ newUsername: '', currentPassword: '', newPassword: '', confirmPassword: '' })
const dbStatus = reactive({ type: 'loading', connected: false })

const isAdmin = computed(() => profile.role === 1)
const displayInitial = computed(() => (profile.nickname || profile.username || 'U').charAt(0).toUpperCase())
const authHeaders = (json = false) => ({ ...(json ? { 'Content-Type': 'application/json' } : {}), Authorization: `Bearer ${localStorage.getItem('authToken') || ''}` })

const loadProfile = async () => {
  const response = await fetch('/api/user/profile', { headers: authHeaders() })
  const result = await response.json()
  if (!response.ok || result.code !== 200) return message.error(result.message || '获取资料失败')
  Object.assign(profile, result.data)
  nickname.value = result.data.nickname || ''
  if (isAdmin.value) loadDatabaseStatus()
}

const updateProfile = async () => {
  if (!nickname.value) return message.warning('请输入昵称')
  updatingProfile.value = true
  try {
    const response = await fetch('/api/user/profile', { method: 'PUT', headers: authHeaders(true), body: JSON.stringify({ nickname: nickname.value }) })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '更新失败')
    profile.nickname = nickname.value
    message.success('资料已更新')
  } catch (error) { message.error(error.message) } finally { updatingProfile.value = false }
}

const handleAvatarSelect = (event) => {
  const file = event.target.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) return message.warning('请选择图片文件')
  const reader = new FileReader()
  reader.onload = () => { cropperImage.value = reader.result; showCropper.value = true }
  reader.readAsDataURL(file)
  event.target.value = ''
}

const uploadAvatar = async (blob) => {
  const form = new FormData()
  form.append('images[]', new File([blob], 'avatar.png', { type: 'image/png' }))
  try {
    const uploadResponse = await fetch('/api/upload/images?hidden=true', { method: 'POST', headers: authHeaders(), body: form })
    const uploadResult = await uploadResponse.json()
    const avatar = uploadResult.data?.files?.[0]?.url
    if (!uploadResponse.ok || uploadResult.code !== 200 || !avatar) throw new Error(uploadResult.message || '上传失败')
    const response = await fetch('/api/user/profile', { method: 'PUT', headers: authHeaders(true), body: JSON.stringify({ avatar }) })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '保存头像失败')
    profile.avatar = avatar
    localStorage.setItem('user_avatar', avatar)
    message.success('头像已更新')
  } catch (error) { message.error(error.message) }
}

const updateAccount = async () => {
  if (!account.newUsername && !account.newPassword) return message.warning('请输入新的用户名或密码')
  if (account.newPassword !== account.confirmPassword) return message.warning('两次输入的新密码不一致')
  updatingAccount.value = true
  try {
    const response = await fetch('/api/account/change', {
      method: 'POST', headers: authHeaders(true),
      body: JSON.stringify({ new_username: account.newUsername, current_password: account.currentPassword, new_password: account.newPassword }),
    })
    const result = await response.json()
    if (!response.ok || !result.success) throw new Error(result.message || '更新失败')
    localStorage.removeItem('authToken'); localStorage.removeItem('userInfo')
    message.success('登录信息已更新，请重新登录')
    setTimeout(() => router.push('/login'), 700)
  } catch (error) { message.error(error.message) } finally { updatingAccount.value = false }
}

const loadDatabaseStatus = async () => {
  try {
    const response = await fetch('/api/database/status', { headers: authHeaders() })
    const result = await response.json()
    if (response.ok && result.code === 200) Object.assign(dbStatus, result.data)
  } catch (error) { console.error('获取数据库状态失败:', error) }
}
const formatDbType = (type) => ({ postgresql: 'PostgreSQL', mysql: 'MySQL', sqlite: 'SQLite', loading: '读取中' }[type] || type)
onMounted(loadProfile)
</script>

<style scoped>
.account-page{width:min(1040px,100%);margin:0 auto;padding:1rem 0 4rem;color:#303a44}.dark .account-page{color:#edf1f4}.account-page>header{margin-bottom:1.5rem}.eyebrow{color:#c65b73;font-size:.7rem;font-weight:800;letter-spacing:.15em;text-transform:uppercase}.account-page>header h1{margin:.12rem 0 .35rem;font-size:clamp(2rem,4vw,3.3rem);line-height:1;letter-spacing:-.055em;font-weight:700}.account-page>header>p:last-child,.section-heading p{color:#88919b;font-size:.82rem}.account-grid{display:grid;grid-template-columns:minmax(0,.72fr) minmax(0,1.28fr);gap:1rem;align-items:start}.profile-card,.credentials-card,.database-card{padding:1.35rem;border:1px solid rgba(211,122,143,.14);border-radius:1.3rem;background:rgba(255,255,255,.82);box-shadow:0 18px 55px rgba(76,47,56,.07);backdrop-filter:blur(18px)}.dark .profile-card,.dark .credentials-card,.dark .database-card{background:rgba(30,35,43,.9);border-color:rgba(255,255,255,.065);box-shadow:0 20px 55px rgba(3,8,13,.32)}.profile-card{text-align:center}.avatar-editor{position:relative;width:7.2rem;height:7.2rem;margin:.3rem auto 1rem;border-radius:2rem;display:grid;place-items:center;overflow:hidden;color:#fff;background:linear-gradient(145deg,#e58ba0,#c85a74);font-size:2.5rem;font-weight:700;cursor:pointer;box-shadow:0 18px 38px rgba(190,75,101,.18)}.avatar-editor img{width:100%;height:100%;object-fit:cover}.avatar-editor button{position:absolute;right:.35rem;bottom:.35rem;display:grid;place-items:center;width:2rem;height:2rem;border-radius:.65rem;color:#b34a61;background:rgba(255,255,255,.9)}.profile-card h2{font-size:1.15rem;font-weight:700}.profile-card>p{color:#929aa3;font-size:.76rem}.role-badge{display:inline-flex;align-items:center;gap:.35rem;margin:.75rem 0 1.3rem;padding:.35rem .55rem;border-radius:.5rem;color:#b44b63;background:#fff0f3;font-size:.68rem;font-weight:800}.dark .role-badge{color:#ff9bb0;background:#3b3037}.profile-form{padding-top:1.1rem;border-top:1px solid rgba(125,108,113,.1);text-align:left}.section-heading{display:flex;align-items:center;gap:.7rem;margin-bottom:1.4rem}.section-heading h2{font-size:1.05rem;font-weight:700}.icon-tile{display:grid;place-items:center;width:2.65rem;height:2.65rem;border-radius:.8rem;color:#c4546d;background:#fff0f3;font-size:1.25rem}.icon-tile.blue{color:#567da5;background:#edf6ff}.dark .icon-tile{background:rgba(255,255,255,.07)}form{display:flex;flex-direction:column;gap:1rem}form label>span{display:block;margin-bottom:.4rem;color:#727c86;font-size:.75rem;font-weight:700}form input{width:100%;padding:.72rem .8rem;border:1px solid #e5dde0;border-radius:.8rem;background:rgba(255,255,255,.8);color:#303a44;outline:none;transition:.2s}.dark form input{background:#252b34;color:#f1f3f5;border-color:#3a424d}form input:focus{border-color:#df879b;box-shadow:0 0 0 3px rgba(223,135,155,.12)}.two-column{display:grid;grid-template-columns:1fr 1fr;gap:.8rem}.primary-button,.secondary-button{display:flex;align-items:center;justify-content:center;gap:.4rem;width:100%;padding:.75rem;border-radius:.8rem;font-size:.8rem;font-weight:800;transition:.2s}.primary-button{color:#fff;background:#d65e78;box-shadow:0 10px 24px rgba(190,75,101,.18)}.primary-button:hover{background:#c94d68}.secondary-button{color:#ad4960;background:#fff0f3}.dark .secondary-button{color:#ff9bb0;background:#3b3037}.database-card{grid-column:2}.db-metric{display:flex;align-items:center;justify-content:space-between;padding:.85rem 0;border-top:1px solid rgba(125,108,113,.1);font-size:.78rem}.db-metric span{color:#858e98}.db-metric strong{display:flex;align-items:center;gap:.3rem;font-variant-numeric:tabular-nums}.db-metric strong.online{color:#3c8b70}.db-metric strong.offline{color:#c9555b}@media(max-width:760px){.account-grid{grid-template-columns:1fr}.database-card{grid-column:1}.two-column{grid-template-columns:1fr}}
</style>
