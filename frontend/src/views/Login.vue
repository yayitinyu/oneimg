<template>
  <div class="auth-shell">
    <section class="auth-story" aria-label="OneImg 介绍">
      <div class="story-mark">
        <img :src="logoImg" alt="OneImg" />
      </div>
      <div>
        <p class="eyebrow">A small place for your images</p>
        <h1>把每一张图片，<br />安放在初春里。</h1>
        <p class="story-copy">上传、整理与分享都保持简单。登录后，你只会看到属于自己的图片空间。</p>
      </div>
      <div class="story-pills">
        <span><i class="mgc_pic_2_line"></i>独立画廊</span>
        <span><i class="mgc_time_duration_line"></i>可选生命周期</span>
        <span><i class="mgc_safe_shield_2_line"></i>账号隔离</span>
      </div>
    </section>

    <section class="auth-card">
      <div class="mobile-logo"><img :src="logoImg" alt="OneImg" /></div>
      <div v-if="loginConfig.registrationMode !== 'closed'" class="auth-tabs" role="tablist">
        <button type="button" :class="{ active: mode === 'login' }" role="tab" :aria-selected="mode === 'login'" @click="mode = 'login'">登录</button>
        <button
          type="button"
          :class="{ active: mode === 'register' }"
          role="tab"
          :aria-selected="mode === 'register'"
          @click="mode = 'register'"
        >注册</button>
      </div>

      <header>
        <p class="eyebrow">{{ mode === 'login' ? 'Welcome back' : 'Create your space' }}</p>
        <h2>{{ mode === 'login' ? '欢迎回来' : '创建普通账号' }}</h2>
        <p>{{ mode === 'login' ? '继续管理你的图片与链接。' : '注册后只能管理自己上传的图片。' }}</p>
      </header>

      <form @submit.prevent="mode === 'login' ? handleLogin() : handleRegister()">
        <label>
          <span>用户名</span>
          <div class="input-wrap">
            <i class="mgc_user_2_line"></i>
            <input v-model.trim="form.username" autocomplete="username" minlength="3" maxlength="32" placeholder="输入用户名" required />
          </div>
        </label>
        <label>
          <span>密码</span>
          <div class="input-wrap">
            <i class="mgc_lock_line"></i>
            <input v-model="form.password" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" type="password" minlength="6" maxlength="72" placeholder="至少 6 位" required />
          </div>
        </label>
        <label v-if="mode === 'register'">
          <span>确认密码</span>
          <div class="input-wrap">
            <i class="mgc_key_2_line"></i>
            <input v-model="form.confirmPassword" autocomplete="new-password" type="password" minlength="6" maxlength="72" placeholder="再次输入密码" required />
          </div>
        </label>
        <label v-if="mode === 'register' && loginConfig.registrationMode === 'invite'">
          <span>邀请码</span>
          <div class="input-wrap">
            <i class="mgc_ticket_line"></i>
            <input v-model.trim="form.invitationCode" autocomplete="off" placeholder="XXXX-XXXX-XXXX-XXXX" required />
          </div>
        </label>

        <div v-if="loginConfig.turnstile" class="turnstile-wrap"><div id="turnstile-container"></div></div>

        <button class="submit-button" :disabled="isLoading" type="submit">
          <i :class="isLoading ? 'mgc_loading_line animate-spin' : mode === 'login' ? 'mgc_enter_door_line' : 'mgc_user_add_line'"></i>
          {{ isLoading ? loadingText : mode === 'login' ? '登录' : '创建账号' }}
        </button>
      </form>

      <p v-if="mode === 'register' && loginConfig.registrationMode === 'invite'" class="form-note">
        <i class="mgc_information_line"></i>当前仅开放邀请码注册
      </p>
    </section>
  </div>
</template>

<script setup>
import { nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import message from '@/utils/message.js'
import defaultLogo from '@/assets/logo-v2.png'

const mode = ref('login')
const isLoading = ref(false)
const loadingText = ref('请稍候')
const turnstileToken = ref('')
const logoImg = ref(localStorage.getItem('site_logo') || defaultLogo)
let turnstileWidgetId = null

const form = reactive({ username: '', password: '', confirmPassword: '', invitationCode: '' })
const loginConfig = reactive({ turnstile: false, turnstileSiteKey: '', registrationMode: 'closed' })

const ensureTurnstile = () => {
  if (!loginConfig.turnstile) return true
  if (!turnstileToken.value) {
    message.warning('请完成人机验证')
    return false
  }
  return true
}

const submitAuth = async (endpoint, payload) => {
  isLoading.value = true
  loadingText.value = endpoint === '/api/register' ? '正在创建' : '正在验证'
  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '请求失败')
    const user = result.data?.user || {}
    localStorage.setItem('authToken', result.data?.token || 'session')
    localStorage.setItem('userInfo', JSON.stringify({
      username: user.username || payload.username,
      nickname: user.nickname || '',
      avatar: user.avatar || '',
      role: user.role,
      isAdmin: user.role === 1,
    }))
    message.success(endpoint === '/api/register' ? '注册成功' : '登录成功')
    window.location.assign('/')
  } catch (error) {
    message.error(error.message)
    resetTurnstile()
  } finally {
    isLoading.value = false
  }
}

const handleLogin = () => {
  if (!ensureTurnstile()) return
  submitAuth('/api/login', {
    username: form.username,
    password: form.password,
    turnstileToken: turnstileToken.value,
  })
}

const handleRegister = () => {
  if (form.password !== form.confirmPassword) return message.warning('两次输入的密码不一致')
  if (!ensureTurnstile()) return
  submitAuth('/api/register', {
    username: form.username,
    password: form.password,
    invitation_code: form.invitationCode,
    turnstileToken: turnstileToken.value,
  })
}

const initTurnstile = async () => {
  if (!loginConfig.turnstile || !window.turnstile) return
  await nextTick()
  const container = document.getElementById('turnstile-container')
  if (!container) return
  if (turnstileWidgetId !== null) window.turnstile.remove(turnstileWidgetId)
  turnstileWidgetId = window.turnstile.render(container, {
    sitekey: loginConfig.turnstileSiteKey,
    callback: (token) => { turnstileToken.value = token },
    'expired-callback': () => { turnstileToken.value = '' },
    'error-callback': () => { turnstileToken.value = '' },
    theme: document.documentElement.classList.contains('dark') ? 'dark' : 'light',
    size: 'flexible',
  })
}

const resetTurnstile = () => {
  if (turnstileWidgetId !== null && window.turnstile) window.turnstile.reset(turnstileWidgetId)
  turnstileToken.value = ''
}

const loadTurnstileScript = () => {
  if (window.turnstile) return initTurnstile()
  let script = document.querySelector('script[data-oneimg-turnstile]')
  if (!script) {
    script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.dataset.oneimgTurnstile = 'true'
    document.head.appendChild(script)
  }
  script.addEventListener('load', initTurnstile, { once: true })
}

const getLoginSettings = async () => {
  try {
    const response = await fetch('/api/settings/login')
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error('获取登录配置失败')
    loginConfig.turnstile = Boolean(result.data.turnstile)
    loginConfig.turnstileSiteKey = result.data.turnstile_site_key || ''
    loginConfig.registrationMode = result.data.registration_mode || 'open'
    logoImg.value = result.data.site_logo || defaultLogo
    if (result.data.site_logo) localStorage.setItem('site_logo', result.data.site_logo)
    else localStorage.removeItem('site_logo')
    if (loginConfig.turnstile) loadTurnstileScript()
  } catch (error) {
    message.error(error.message)
  }
}

watch(mode, () => { turnstileToken.value = ''; if (loginConfig.turnstile) setTimeout(initTurnstile, 0) })
onMounted(getLoginSettings)
onUnmounted(() => { if (turnstileWidgetId !== null && window.turnstile) window.turnstile.remove(turnstileWidgetId) })
</script>

<style scoped>
.auth-shell {
  width: min(1040px, 100%);
  min-height: min(720px, calc(100dvh - 3rem));
  display: grid;
  grid-template-columns: 1.08fr .92fr;
  overflow: hidden;
  border: 1px solid rgba(214, 133, 151, .16);
  border-radius: 1.8rem;
  background: rgba(255, 255, 255, .78);
  box-shadow: 0 30px 90px rgba(88, 55, 65, .12);
  backdrop-filter: blur(22px);
}

.dark .auth-shell {
  background: rgba(29, 34, 41, .9);
  border-color: rgba(255, 255, 255, .07);
  box-shadow: 0 32px 90px rgba(3, 8, 13, .42);
}

.auth-story {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 3.2rem;
  overflow: hidden;
  background:
    radial-gradient(circle at 82% 20%, rgba(255, 255, 255, .78), transparent 25%),
    radial-gradient(circle at 10% 90%, rgba(218, 115, 139, .22), transparent 34%),
    linear-gradient(145deg, #fff4f6 0%, #f7e9e8 56%, #e9ece6 100%);
}

.auth-story::after {
  content: '';
  position: absolute;
  right: -8rem;
  bottom: -8rem;
  width: 19rem;
  height: 19rem;
  border: 1px solid rgba(185, 91, 112, .15);
  border-radius: 42% 58% 66% 34%;
  transform: rotate(18deg);
}

.dark .auth-story {
  background:
    radial-gradient(circle at 80% 20%, rgba(255, 255, 255, .07), transparent 27%),
    linear-gradient(145deg, #3b2d34, #293239);
}

.story-mark {
  width: 4rem;
  height: 4rem;
  display: grid;
  place-items: center;
  border-radius: 1.1rem;
  background: rgba(255, 255, 255, .72);
  box-shadow: 0 12px 30px rgba(144, 72, 88, .11);
}

.story-mark img,
.mobile-logo img {
  width: 100%;
  height: 100%;
  padding: .52rem;
  object-fit: contain;
}

.eyebrow {
  color: #c75a73;
  font-size: .7rem;
  font-weight: 800;
  letter-spacing: .16em;
  text-transform: uppercase;
}

.auth-story h1 {
  margin: .55rem 0 1.2rem;
  color: #3c3236;
  font-size: clamp(2.45rem, 5vw, 4rem);
  font-weight: 700;
  line-height: 1.08;
  letter-spacing: -.065em;
  text-wrap: balance;
}

.dark .auth-story h1 { color: #fff5f7; }
.story-copy { max-width: 30rem; color: #756b6f; font-size: .95rem; line-height: 1.8; }
.dark .story-copy { color: #c4bbc0; }

.story-pills {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: .55rem;
}

.story-pills span {
  display: flex;
  align-items: center;
  gap: .35rem;
  padding: .48rem .68rem;
  border: 1px solid rgba(182, 96, 115, .14);
  border-radius: .7rem;
  background: rgba(255, 255, 255, .45);
  color: #7d646b;
  font-size: .7rem;
  font-weight: 700;
}

.dark .story-pills span {
  border-color: rgba(255, 255, 255, .08);
  background: rgba(255, 255, 255, .05);
  color: #d7cbd0;
}

.auth-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: clamp(1.5rem, 5vw, 3.5rem);
  background: rgba(255, 255, 255, .64);
}

.dark .auth-card { background: rgba(25, 30, 37, .5); }
.mobile-logo { display: none; }

.auth-tabs {
  display: flex;
  gap: .25rem;
  align-self: flex-start;
  padding: .25rem;
  margin-bottom: 2.3rem;
  border-radius: .75rem;
  background: #f4eeef;
}

.dark .auth-tabs { background: #252b33; }

.auth-tabs button {
  min-width: 4.7rem;
  padding: .5rem .8rem;
  border-radius: .58rem;
  color: #8c8286;
  font-size: .78rem;
  font-weight: 700;
  transition: .2s;
}

.auth-tabs button.active {
  background: white;
  color: #ad455d;
  box-shadow: 0 5px 14px rgba(93, 55, 64, .08);
}

.dark .auth-tabs button.active { background: #3b3339; color: #ff9cb1; }
.auth-card header { margin-bottom: 1.6rem; }
.auth-card h2 { margin: .2rem 0 .3rem; font-size: 1.75rem; font-weight: 700; letter-spacing: -.04em; }
.auth-card header > p:last-child { color: #8a939d; font-size: .82rem; }

form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

form label > span {
  display: block;
  margin-bottom: .42rem;
  color: #6f7882;
  font-size: .76rem;
  font-weight: 700;
}

.input-wrap {
  display: flex;
  align-items: center;
  overflow: hidden;
  border: 1px solid #e4dcdf;
  border-radius: .82rem;
  background: rgba(255, 255, 255, .82);
  transition: border-color .2s, box-shadow .2s, background-color .2s;
}

.input-wrap:focus-within {
  border-color: #dc8397;
  box-shadow: 0 0 0 3px rgba(220, 131, 151, .12);
}

.input-wrap i { flex: 0 0 auto; margin-left: .82rem; color: #bd8793; }

.input-wrap input {
  width: 100%;
  min-width: 0;
  padding: .75rem .82rem;
  outline: none;
  background: transparent;
  color: #313a44;
  font-size: .88rem;
}

.dark .input-wrap { border-color: #3a424d; background: #242a32; }
.dark .input-wrap input { color: #f5f6f7; }

/* Keep browser-managed credentials visually integrated with the whole field. */
.input-wrap:has(input:-webkit-autofill) { background: #f8f2f4; }
.dark .input-wrap:has(input:-webkit-autofill) { background: #292e37; }

.input-wrap input:-webkit-autofill,
.input-wrap input:-webkit-autofill:hover,
.input-wrap input:-webkit-autofill:focus {
  -webkit-text-fill-color: #313a44 !important;
  caret-color: #313a44;
  -webkit-box-shadow: 0 0 0 1000px #f8f2f4 inset !important;
  box-shadow: 0 0 0 1000px #f8f2f4 inset !important;
  transition: background-color 9999s ease-out;
}

.dark .input-wrap input:-webkit-autofill,
.dark .input-wrap input:-webkit-autofill:hover,
.dark .input-wrap input:-webkit-autofill:focus {
  -webkit-text-fill-color: #f5f6f7 !important;
  caret-color: #f5f6f7;
  -webkit-box-shadow: 0 0 0 1000px #292e37 inset !important;
  box-shadow: 0 0 0 1000px #292e37 inset !important;
}

.submit-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: .45rem;
  width: 100%;
  border-radius: .82rem;
  font-size: .85rem;
  font-weight: 800;
  transition: .2s;
}

.submit-button {
  margin-top: .25rem;
  padding: .82rem;
  background: #d85f79;
  color: white;
  box-shadow: 0 12px 26px rgba(190, 75, 101, .2);
}

.submit-button:hover:not(:disabled) { background: #c94e69; transform: translateY(-1px); }
.submit-button:active { transform: scale(.985); }
.submit-button:disabled { opacity: .58; cursor: not-allowed; }

.turnstile-wrap {
  width: 100%;
  min-height: 65px;
  display: flex;
  align-items: center;
  justify-content: center;
}

#turnstile-container {
  width: 100%;
  min-width: 300px;
}

.form-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: .35rem;
  margin-top: 1rem;
  color: #958a8e;
  font-size: .72rem;
}

@media (max-width: 760px) {
  .auth-shell {
    width: 100%;
    min-height: 100dvh;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: visible;
    border: none;
    border-radius: 0;
    box-shadow: none;
    background: transparent;
    padding: max(0.5rem, env(safe-area-inset-top)) max(0.5rem, env(safe-area-inset-right)) max(0.5rem, env(safe-area-inset-bottom)) max(0.5rem, env(safe-area-inset-left));
    box-sizing: border-box;
  }

  .auth-story { display: none; }

  .auth-card {
    width: min(100%, 23.5rem);
    min-height: auto;
    margin: auto;
    padding: 1.25rem 1.15rem;
    border-radius: 1.25rem;
    border: 1px solid rgba(214, 133, 151, .2);
    background: rgba(255, 255, 255, 0.88);
    box-shadow: 0 15px 40px rgba(88, 55, 65, 0.08);
    backdrop-filter: blur(20px);
    box-sizing: border-box;
  }

  .dark .auth-card {
    background: rgba(29, 34, 41, 0.92);
    border-color: rgba(255, 255, 255, 0.08);
    box-shadow: 0 18px 45px rgba(3, 8, 13, 0.38);
  }

  .mobile-logo {
    width: 2.6rem;
    height: 2.6rem;
    display: grid;
    place-items: center;
    margin: 0 auto .45rem;
    border-radius: .75rem;
    background: #fff4f5;
    box-shadow: 0 8px 18px rgba(144, 72, 88, .08);
  }

  .dark .mobile-logo { background: #3c3037; }

  .auth-tabs {
    width: min(100%, 18rem);
    align-self: center;
    margin: 0 auto .5rem;
  }

  .auth-tabs button { flex: 1; min-width: 0; padding-block: .35rem; font-size: .8rem; }
  .auth-card header { margin-bottom: .65rem; text-align: center; }
  .auth-card header .eyebrow { display: none; }
  .auth-card h2 { margin-top: 0; font-size: 1.35rem; }
  .auth-card header > p:last-child { font-size: .75rem; margin-top: .15rem; }
  form { gap: .5rem; }
  form label > span { margin-bottom: .2rem; font-size: .75rem; }
  .input-wrap input { padding-block: .55rem; font-size: .88rem; }
  .turnstile-wrap { min-height: 52px; margin-block: .2rem; }
  .submit-button { padding: .6rem; margin-top: .25rem; font-size: .9rem; }
  .form-note { margin-top: .45rem; font-size: .7rem; }
}

@media (max-width: 360px) {
  .auth-card {
    padding: 1rem .85rem;
  }

  #turnstile-container { min-width: 0; }
}

@media (max-width: 760px) and (max-height: 700px) {
  .mobile-logo { width: 2.4rem; height: 2.4rem; margin-bottom: .35rem; }
  .auth-tabs { margin-bottom: .4rem; }
  .auth-card header { margin-bottom: .45rem; }
  .auth-card h2 { font-size: 1.25rem; }
  form { gap: .45rem; }
  .input-wrap input { padding-block: .48rem; }
  .turnstile-wrap { min-height: 50px; }
  .submit-button { padding: .55rem; }
}
</style>
