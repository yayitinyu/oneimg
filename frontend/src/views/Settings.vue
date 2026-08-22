<template>
  <div class="settings-page">
    <PageHeading
      eyebrow="Control room"
      title="系统设置"
      description="管理注册、存储与安全策略。改动会在保存后立即生效。"
    >
      <template #action>
      <button class="primary-button" :disabled="saving || !dirty" @click="saveAll">
        <i :class="saving ? 'mgc_loading_line animate-spin' : 'mgc_check_circle_line'"></i>
        {{ saving ? '正在保存' : dirty ? '保存更改' : '已保存' }}
      </button>
      </template>
    </PageHeading>

    <div v-if="loading" class="panel loading-panel">
      <span class="skeleton-line wide"></span>
      <span class="skeleton-line"></span>
      <span class="skeleton-card"></span>
    </div>

    <template v-else>
      <nav class="settings-tabs" aria-label="设置分类">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          <i :class="tab.icon"></i>
          <span>{{ tab.label }}</span>
        </button>
      </nav>

      <section v-if="activeTab === 'accounts'" class="settings-grid">
        <article class="panel span-7">
          <div class="panel-heading">
            <span class="icon-tile coral"><i class="mgc_user_security_line"></i></span>
            <div>
              <h2>账号与注册</h2>
              <p>普通账号只能查看和管理自己上传的图片。</p>
            </div>
          </div>

          <div class="field-block">
            <label>注册方式</label>
            <div class="choice-grid">
              <label
                v-for="mode in registrationModes"
                :key="mode.value"
                class="choice-card"
                :class="{ selected: settings.registration_mode === mode.value }"
              >
                <input v-model="settings.registration_mode" type="radio" :value="mode.value" />
                <i :class="mode.icon"></i>
                <strong>{{ mode.title }}</strong>
                <span>{{ mode.description }}</span>
              </label>
            </div>
          </div>

          <div class="quota-settings">
            <div class="quota-copy">
              <span class="icon-tile quota-icon"><i class="mgc_storage_line"></i></span>
              <div>
                <strong>普通用户空间上限</strong>
                <span>每个普通账号独立计算；设为 0 时不限制。</span>
              </div>
            </div>
            <label class="field-block quota-input">
              <span>每个账号</span>
              <div class="input-suffix">
                <input v-model.number="userStorageQuotaGB" type="number" min="0" max="1048576" step="0.1" inputmode="decimal" />
                <b>GB</b>
              </div>
            </label>
          </div>

          <div v-if="settings.registration_mode === 'invite'" class="invite-box">
            <div class="subheading-row">
              <div>
                <h3>邀请码</h3>
                <p>每个邀请码只能使用一次，明文只在生成时显示。</p>
              </div>
              <div class="invite-actions">
                <input v-model.number="inviteCount" type="number" min="1" max="20" aria-label="生成数量" />
                <button class="secondary-button" :disabled="generatingInvites" @click="generateInvites">
                  <i class="mgc_coupon_line"></i>
                  生成
                </button>
              </div>
            </div>

            <div v-if="generatedCodes.length" class="generated-codes">
              <div>
                <strong>请现在保存这些邀请码</strong>
                <span>关闭或刷新页面后将无法再次查看完整内容。</span>
              </div>
              <code>{{ generatedCodes.join('\n') }}</code>
              <button class="text-button" @click="copyGeneratedCodes">
                <i class="mgc_copy_2_line"></i>复制全部
              </button>
            </div>

            <div class="invite-list">
              <div v-if="!invitations.length" class="mini-empty">
                <i class="mgc_ticket_line"></i>
                <span>还没有邀请码</span>
              </div>
              <div v-for="invite in invitations" :key="invite.id" class="invite-row">
                <span class="invite-hint">••••-{{ invite.hint }}</span>
                <span :class="invite.used_at ? 'status used' : 'status ready'">
                  {{ invite.used_at ? '已使用' : '可使用' }}
                </span>
                <time>{{ formatDate(invite.created_at) }}</time>
                <button
                  v-if="!invite.used_at"
                  class="icon-button danger"
                  title="删除邀请码"
                  @click="deleteInvite(invite.id)"
                >
                  <i class="mgc_delete_2_line"></i>
                </button>
              </div>
            </div>
          </div>
        </article>

        <article class="panel span-5">
          <div class="panel-heading">
            <span class="icon-tile pink"><i class="mgc_flower_2_line"></i></span>
            <div>
              <h2>站点信息</h2>
              <p>用于品牌展示和 Telegram Webhook。</p>
            </div>
          </div>

          <label class="field-block">
            <span>网站域名</span>
            <input v-model.trim="settings.site_domain" type="text" placeholder="img.example.com" />
          </label>

          <div class="field-block">
            <span>站点图标</span>
            <div class="logo-editor">
              <div class="logo-preview">
                <img :src="settings.site_logo || defaultLogo" :alt="settings.site_logo ? '当前站点图标' : '默认站点图标'" />
              </div>
              <div>
                <button class="secondary-button" type="button" @click="logoInput?.click()">
                  <i class="mgc_upload_2_line"></i>上传图标
                </button>
                <button v-if="settings.site_logo" class="text-button danger-text" type="button" @click="settings.site_logo = ''">
                  恢复默认
                </button>
              </div>
              <input ref="logoInput" class="hidden" type="file" accept="image/*" @change="handleLogoSelect" />
            </div>
          </div>
        </article>
      </section>

      <section v-else-if="activeTab === 'storage'" class="settings-grid">
        <article class="panel span-5">
          <div class="panel-heading">
            <span class="icon-tile blue"><i class="mgc_storage_line"></i></span>
            <div>
              <h2>存储渠道</h2>
              <p>所有访问统一经过本站图片代理。</p>
            </div>
          </div>
          <div class="storage-options">
            <label
              v-for="storage in storageOptions"
              :key="storage.value"
              :class="{ selected: settings.storage_type === storage.value }"
            >
              <input v-model="settings.storage_type" type="radio" :value="storage.value" />
              <i :class="storage.icon"></i>
              <span>{{ storage.label }}</span>
            </label>
          </div>
          <p class="inline-note">
            <i class="mgc_information_line"></i>
            S3 与 R2 的图片链接由 OneImg 统一生成并安全代理。
          </p>
        </article>

        <article class="panel span-7">
          <div class="panel-heading compact">
            <div>
              <p class="eyebrow">{{ currentStorageLabel }}</p>
              <h2>连接配置</h2>
            </div>
          </div>

          <label v-if="settings.storage_type === 'default'" class="field-block">
            <span>本地存储路径</span>
            <input v-model.trim="settings.storage_path" type="text" placeholder="/uploads" />
          </label>

          <div v-if="settings.storage_type === 's3'" class="form-grid">
            <FieldInput v-model="settings.s3_endpoint" label="Endpoint" placeholder="https://s3.example.com" />
            <FieldInput v-model="settings.s3_bucket" label="Bucket" />
			<FieldInput v-model="settings.s3_region" label="Region" placeholder="us-east-1" />
            <FieldInput v-model="settings.s3_access_key" label="Access Key" />
			<FieldInput v-model="settings.s3_secret_key" label="Secret Key" type="password" />
			<div class="switch-row span-2 compact-switch">
			  <div><strong>Path-style addressing</strong></div>
			  <label class="toggle"><input v-model="settings.s3_path_style" type="checkbox" /><span></span></label>
			</div>
          </div>

          <div v-if="settings.storage_type === 'r2'" class="form-grid">
            <FieldInput v-model="settings.r2_endpoint" label="Endpoint" placeholder="https://ACCOUNT.r2.cloudflarestorage.com" />
            <FieldInput v-model="settings.r2_bucket" label="Bucket" />
            <FieldInput v-model="settings.r2_access_key" label="Access Key" />
            <FieldInput v-model="settings.r2_secret_key" label="Secret Key" type="password" />
          </div>

          <div v-if="settings.storage_type === 'webdav'" class="form-grid">
            <FieldInput v-model="settings.webdav_url" label="WebDAV URL" />
            <FieldInput v-model="settings.webdav_user" label="用户名" />
            <FieldInput v-model="settings.webdav_pass" class="span-2" label="密码" type="password" />
          </div>

          <div v-if="settings.storage_type === 'ftp'" class="form-grid">
            <FieldInput v-model="settings.ftp_host" label="主机" />
            <FieldInput v-model.number="settings.ftp_port" label="端口" type="number" />
            <FieldInput v-model="settings.ftp_user" label="用户名" />
            <FieldInput v-model="settings.ftp_pass" label="密码" type="password" />
          </div>

          <div v-if="settings.storage_type === 'telegram'" class="form-grid">
            <FieldInput v-model="settings.tg_bot_token" class="span-2" label="Bot Token" type="password" />
            <FieldInput v-model="settings.tg_receivers" class="span-2" label="接收者 ID" placeholder="多个 ID 用逗号分隔" />
            <p class="inline-note span-2">
              <i class="mgc_information_line"></i>
              第一个接收者用于保存图片；启用通知后，其余接收者也会收到通知。
            </p>
          </div>
        </article>

        <article v-if="migrationAvailable || migration" class="panel span-12 migration-panel">
          <div class="panel-heading migration-heading">
            <span class="icon-tile green"><i class="mgc_transfer_4_line"></i></span>
            <div>
              <h2>迁移对象存储</h2>
              <p>S3 / R2 之间迁移</p>
            </div>
            <button class="icon-button" :disabled="migrationLoading" title="刷新进度" @click="refreshMigration">
              <i :class="migrationLoading ? 'mgc_loading_line animate-spin' : 'mgc_refresh_2_line'"></i>
            </button>
          </div>

          <div v-if="migration" class="migration-status-card">
            <div class="migration-status-row">
              <div>
                <span :class="['migration-badge', migration.status]">{{ migrationStatusLabel }}</span>
                <strong>{{ migration.source_type.toUpperCase() }} → {{ migration.target_type.toUpperCase() }}</strong>
              </div>
              <b>{{ migration.progress_percent }}%</b>
            </div>
            <div class="migration-progress"><i :style="{ width: `${migration.progress_percent}%` }"></i></div>
            <div class="migration-metrics">
              <span>{{ migration.copied_objects }} / {{ migration.total_objects }} 个对象</span>
              <span>{{ formatFileSize(migration.copied_bytes) }}</span>
            </div>
            <p v-if="migration.error" class="migration-error">{{ migration.error }}</p>
            <details v-if="migration.failed_items?.length" class="migration-details">
              <summary>失败详情</summary>
              <div v-for="item in migration.failed_items" :key="item.id">
                <code>{{ item.object_key }}</code>
                <span>{{ item.error }}</span>
              </div>
            </details>
            <button v-if="migration.status === 'failed'" class="secondary-button retry-button" :disabled="migrationStarting" @click="retryMigration">
              <i class="mgc_refresh_2_line"></i>重试失败项
            </button>
          </div>

          <form v-if="migrationAvailable && !migrationActive" class="migration-form" @submit.prevent="confirmStartMigration">
            <p class="migration-form-label">源存储</p>
            <div class="storage-options migration-targets">
              <label
                v-for="source in availableMigrationSources"
                :key="source.type"
                :class="{ selected: migrationTarget.source_type === source.type }"
              >
                <input v-model="migrationTarget.source_type" type="radio" :value="source.type" />
                <i :class="source.type === 'r2' ? 'mgc_cloud_2_line' : 'mgc_cloud_line'"></i>
                <span>{{ source.type.toUpperCase() }} <small>{{ source.image_count }} 张</small></span>
              </label>
            </div>
            <p class="migration-form-label">目标存储</p>
            <div class="storage-options migration-targets">
              <label :class="{ selected: migrationTarget.target_type === 's3' }">
                <input v-model="migrationTarget.target_type" type="radio" value="s3" />
                <i class="mgc_cloud_line"></i><span>S3</span>
              </label>
              <label :class="{ selected: migrationTarget.target_type === 'r2' }">
                <input v-model="migrationTarget.target_type" type="radio" value="r2" />
                <i class="mgc_cloud_2_line"></i><span>R2</span>
              </label>
            </div>
            <div class="form-grid">
              <FieldInput v-model="migrationTarget.endpoint" label="目标 Endpoint" :placeholder="migrationTarget.target_type === 'r2' ? 'https://ACCOUNT.r2.cloudflarestorage.com' : 'https://s3.example.com'" />
              <FieldInput v-model="migrationTarget.bucket" label="目标 Bucket" />
              <FieldInput v-if="migrationTarget.target_type === 's3'" v-model="migrationTarget.region" label="目标 Region" placeholder="us-east-1" />
              <FieldInput v-model="migrationTarget.access_key" label="目标 Access Key" />
              <FieldInput v-model="migrationTarget.secret_key" label="目标 Secret Key" type="password" />
              <div v-if="migrationTarget.target_type === 's3'" class="switch-row compact-switch">
                <div><strong>Path-style addressing</strong></div>
                <label class="toggle"><input v-model="migrationTarget.force_path_style" type="checkbox" /><span></span></label>
              </div>
            </div>
            <div class="migration-actions">
              <p><i class="mgc_information_line"></i>迁移期间暂停上传和删除；旧存储对象不会被删除。</p>
              <button class="primary-button" type="submit" :disabled="migrationStarting || storageSettingsDirty">
                <i :class="migrationStarting ? 'mgc_loading_line animate-spin' : 'mgc_transfer_4_line'"></i>
                {{ migrationStarting ? '正在连接' : '开始迁移' }}
              </button>
            </div>
          </form>
        </article>
      </section>

      <section v-else-if="activeTab === 'images'" class="settings-grid">
        <article class="panel span-7">
          <div class="panel-heading">
            <span class="icon-tile green"><i class="mgc_magic_2_line"></i></span>
            <div>
              <h2>图片处理</h2>
              <p>WebP 快捷开关已移到首页右上角。</p>
            </div>
          </div>
          <div class="form-grid">
            <label class="field-block">
              <span>WebP 质量 · {{ settings.webp_quality }}</span>
              <input v-model.number="settings.webp_quality" class="range" type="range" min="1" max="100" />
            </label>
            <label class="field-block">
              <span>单张上限</span>
              <div class="input-suffix"><input v-model.number="maxFileSizeMB" type="number" min="1" /><b>MB</b></div>
            </label>
          </div>
        </article>

        <article class="panel span-5 soft-panel">
          <i class="mgc_time_duration_line lifecycle-illustration"></i>
          <p class="eyebrow">Image lifetime</p>
          <h2>生命周期由上传者选择</h2>
          <p>首页上传时可选 1 小时、1 天、7 天、30 天、90 天或永久保存。过期链接会返回 410，并由后台清理存储文件。</p>
        </article>
      </section>

      <section v-else class="settings-grid">
        <article class="panel span-6">
          <div class="panel-heading">
            <span class="icon-tile amber"><i class="mgc_shield_line"></i></span>
            <div><h2>访问安全</h2><p>控制人机验证与外链来源。</p></div>
          </div>
          <div class="switch-row">
            <div><strong>Cloudflare Turnstile</strong><span>登录与注册都会要求完成人机验证。</span></div>
            <label class="toggle"><input v-model="settings.turnstile" type="checkbox" /><span></span></label>
          </div>
          <div v-if="settings.turnstile" class="form-grid nested-fields">
            <FieldInput v-model="settings.turnstile_site_key" label="Site Key" />
            <FieldInput v-model="settings.turnstile_secret_key" label="Secret Key" type="password" />
          </div>
          <div class="switch-row">
            <div><strong>来源白名单</strong><span>只允许本站和指定域名引用图片。</span></div>
            <label class="toggle"><input v-model="settings.referer_white_enable" type="checkbox" /><span></span></label>
          </div>
          <label v-if="settings.referer_white_enable" class="field-block nested-fields">
            <span>允许的域名</span>
            <textarea v-model="settings.referer_white_list" rows="3" placeholder="example.com, blog.example.com"></textarea>
          </label>
        </article>

        <article class="panel span-6">
          <div class="panel-heading">
            <span class="icon-tile blue"><i class="mgc_telegram_line"></i></span>
            <div><h2>Telegram 自动化</h2><p>通知和 URL Webhook 上传可独立开启。</p></div>
          </div>
          <div class="switch-row">
            <div><strong>上传通知</strong><span>上传成功后向接收者推送消息。</span></div>
            <label class="toggle"><input v-model="settings.tg_notice" type="checkbox" /><span></span></label>
          </div>
          <label v-if="settings.tg_notice" class="field-block nested-fields">
            <span>通知文本</span>
            <textarea v-model="settings.tg_notice_text" rows="4" placeholder="支持 {username}、{filename}、{url}" />
          </label>
          <div class="switch-row">
            <div><strong>Webhook 上传</strong><span>向机器人发送图片 URL 即可保存。</span></div>
            <label class="toggle"><input v-model="settings.tg_webhook" type="checkbox" /><span></span></label>
          </div>
        </article>
      </section>
    </template>

    <ImageCropper v-model:visible="showCropper" :image-src="cropperImage" @cropped="uploadLogo" />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import ImageCropper from '@/components/ImageCropper.vue'
import PageHeading from '@/components/PageHeading.vue'
import defaultLogo from '@/assets/logo-v2.png'
import message from '@/utils/message.js'
import '@/utils/popupModal.js'

const FieldInput = defineComponent({
  name: 'FieldInput',
  props: { modelValue: [String, Number], label: String, type: { type: String, default: 'text' }, placeholder: String },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () => h('label', { class: ['field-block', attrs.class] }, [
      h('span', props.label),
      h('input', {
        value: props.modelValue,
        type: props.type,
        placeholder: props.placeholder,
        onInput: (event) => emit('update:modelValue', props.type === 'number' ? Number(event.target.value) : event.target.value),
      }),
    ])
  },
})

const tabs = [
  { id: 'accounts', label: '账号', icon: 'mgc_user_2_line' },
  { id: 'storage', label: '存储', icon: 'mgc_storage_line' },
  { id: 'images', label: '图片', icon: 'mgc_pic_ai_line' },
  { id: 'security', label: '安全', icon: 'mgc_safe_shield_2_line' },
]
const registrationModes = [
  { value: 'open', title: '开放注册', description: '任何人都可创建账号', icon: 'mgc_user_add_line' },
  { value: 'invite', title: '邀请码注册', description: '持有一次性邀请码才可注册', icon: 'mgc_ticket_line' },
  { value: 'closed', title: '关闭注册', description: '仅保留现有账号登录', icon: 'mgc_lock_line' },
]
const storageOptions = [
  { value: 'default', label: '本地', icon: 'mgc_folder_line' },
  { value: 's3', label: 'S3', icon: 'mgc_cloud_line' },
  { value: 'r2', label: 'R2', icon: 'mgc_cloud_2_line' },
  { value: 'webdav', label: 'WebDAV', icon: 'mgc_server_2_line' },
  { value: 'ftp', label: 'FTP', icon: 'mgc_transfer_4_line' },
  { value: 'telegram', label: 'Telegram', icon: 'mgc_telegram_line' },
]

const activeTab = ref('accounts')
const loading = ref(true)
const saving = ref(false)
const originalSettings = ref('{}')
const invitations = ref([])
const generatedCodes = ref([])
const inviteCount = ref(1)
const generatingInvites = ref(false)
const logoInput = ref(null)
const showCropper = ref(false)
const cropperImage = ref('')
const migration = ref(null)
const migrationSources = ref([])
const migrationLoading = ref(false)
const migrationStarting = ref(false)
let migrationPollTimer = null

const settings = reactive({
  site_domain: '', site_logo: '', registration_mode: 'open', user_storage_quota: 0,
  storage_type: 'default', storage_path: '/uploads', max_file_size: 10485760,
  original_image: false, thumbnail: true, webp_quality: 95,
  s3_endpoint: '', s3_region: 'us-east-1', s3_access_key: '', s3_secret_key: '', s3_bucket: '', s3_path_style: true,
  r2_endpoint: '', r2_access_key: '', r2_secret_key: '', r2_bucket: '',
  webdav_url: '', webdav_user: '', webdav_pass: '',
  ftp_host: '', ftp_user: '', ftp_pass: '', ftp_port: 21,
  tg_bot_token: '', tg_receivers: '', tg_notice: false, tg_notice_text: '', tg_webhook: false,
  turnstile: false, turnstile_site_key: '', turnstile_secret_key: '',
  referer_white_enable: false, referer_white_list: '',
})

const migrationTarget = reactive({
  source_type: '', target_type: 'r2', endpoint: '', region: 'us-east-1', bucket: '',
  access_key: '', secret_key: '', force_path_style: true,
})

const saveableKeys = Object.keys(settings)
const dirty = computed(() => JSON.stringify(pickSettings()) !== originalSettings.value)
const maxFileSizeMB = computed({
  get: () => Math.max(1, Math.round(settings.max_file_size / 1024 / 1024)),
  set: (value) => { settings.max_file_size = Math.max(1, Number(value) || 1) * 1024 * 1024 },
})
const userStorageQuotaGB = computed({
  get: () => Number((Math.max(0, settings.user_storage_quota) / 1024 / 1024 / 1024).toFixed(2)),
  set: (value) => {
    const gigabytes = Math.min(1048576, Math.max(0, Number(value) || 0))
    settings.user_storage_quota = Math.round(gigabytes * 1024 * 1024 * 1024)
  },
})
const currentStorageLabel = computed(() => storageOptions.find((item) => item.value === settings.storage_type)?.label || '存储')
const availableMigrationSources = computed(() => migrationSources.value.filter((source) => source.configured && source.image_count > 0))
const migrationAvailable = computed(() => availableMigrationSources.value.length > 0)
const migrationActive = computed(() => ['pending', 'running'].includes(migration.value?.status))
const migrationStatusLabel = computed(() => ({
  pending: '等待中', running: '迁移中', failed: '迁移失败', completed: '已完成',
}[migration.value?.status] || '未知'))
const storageSettingKeys = ['storage_type', 's3_endpoint', 's3_region', 's3_access_key', 's3_secret_key', 's3_bucket', 's3_path_style', 'r2_endpoint', 'r2_access_key', 'r2_secret_key', 'r2_bucket']
const storageSettingsDirty = computed(() => {
  try {
    const original = JSON.parse(originalSettings.value)
    return storageSettingKeys.some((key) => JSON.stringify(original[key]) !== JSON.stringify(settings[key]))
  } catch { return true }
})

const pickSettings = () => Object.fromEntries(saveableKeys.map((key) => [key, settings[key]]))

const authHeaders = (json = false) => ({
  ...(json ? { 'Content-Type': 'application/json' } : {}),
  Authorization: `Bearer ${localStorage.getItem('authToken') || ''}`,
})

const fetchSettings = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/settings/get', { headers: authHeaders() })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '获取设置失败')
    for (const key of saveableKeys) {
      if (Object.prototype.hasOwnProperty.call(result.data, key)) settings[key] = result.data[key]
    }
    originalSettings.value = JSON.stringify(pickSettings())
    if (settings.registration_mode === 'invite') await fetchInvitations()
  } catch (error) {
    message.error(error.message)
  } finally {
    loading.value = false
  }
}

const stopMigrationPolling = () => {
  if (migrationPollTimer) clearInterval(migrationPollTimer)
  migrationPollTimer = null
}

const startMigrationPolling = () => {
  if (migrationPollTimer) return
  migrationPollTimer = setInterval(() => fetchLatestMigration(true), 2500)
}

const fetchMigrationSources = async (silent = false) => {
  try {
    const response = await fetch('/api/storage/migrations/sources', { headers: authHeaders() })
    const body = await response.json()
    if (!response.ok || body.code !== 200) throw new Error(body.message || '获取源存储失败')
    migrationSources.value = body.data || []
    const sources = availableMigrationSources.value
    if (!sources.some((source) => source.type === migrationTarget.source_type)) {
      migrationTarget.source_type = [...sources].sort((a, b) => b.image_count - a.image_count)[0]?.type || 's3'
    }
  } catch (error) {
    if (!silent) message.error(error.message)
  }
}

const refreshMigration = async () => {
  migrationLoading.value = true
  try {
    await Promise.all([fetchLatestMigration(), fetchMigrationSources()])
  } finally {
    migrationLoading.value = false
  }
}

const fetchLatestMigration = async (silent = false) => {
  if (!silent) migrationLoading.value = true
  try {
    const response = await fetch('/api/storage/migrations/latest', { headers: authHeaders() })
    const body = await response.json()
    if (!response.ok || body.code !== 200) throw new Error(body.message || '获取迁移进度失败')
    const previousStatus = migration.value?.status
    migration.value = body.data
    if (migrationActive.value) {
      startMigrationPolling()
    } else {
      stopMigrationPolling()
      if (previousStatus && previousStatus !== migration.value?.status && migration.value?.status === 'completed') {
        message.success('对象存储迁移完成')
        await fetchSettings()
        await fetchMigrationSources(true)
      }
    }
  } catch (error) {
    if (!silent) message.error(error.message)
  } finally {
    if (!silent) migrationLoading.value = false
  }
}

const createMigration = async () => {
  migrationStarting.value = true
  try {
    const response = await fetch('/api/storage/migrations', {
      method: 'POST', headers: authHeaders(true), body: JSON.stringify(migrationTarget),
    })
    const body = await response.json()
    if (!response.ok || body.code !== 200) throw new Error(body.message || '创建迁移任务失败')
    migration.value = body.data
    migrationTarget.secret_key = ''
    message.success('迁移任务已开始')
    startMigrationPolling()
  } catch (error) {
    message.error(error.message)
  } finally {
    migrationStarting.value = false
  }
}

const confirmStartMigration = () => {
  if (!migrationTarget.endpoint.trim() || !migrationTarget.bucket.trim() || !migrationTarget.access_key.trim() || !migrationTarget.secret_key.trim()) {
    message.error('请填写完整的目标存储配置')
    return
  }
  window.showConfirm(
    '<p>程序会先测试目标存储，再迁移数据库中记录的原图和缩略图。迁移完成前不会切换当前存储。</p>',
    '开始迁移',
    createMigration,
  )
}

const retryMigration = async () => {
  if (!migration.value?.id) return
  migrationStarting.value = true
  try {
    const response = await fetch(`/api/storage/migrations/${migration.value.id}/retry`, { method: 'POST', headers: authHeaders() })
    const body = await response.json()
    if (!response.ok || body.code !== 200) throw new Error(body.message || '重试迁移失败')
    migration.value = body.data
    message.success('已重新开始迁移')
    startMigrationPolling()
  } catch (error) {
    message.error(error.message)
  } finally {
    migrationStarting.value = false
  }
}

const saveAll = async () => {
  const original = JSON.parse(originalSettings.value)
  const changed = saveableKeys.filter((key) => JSON.stringify(original[key]) !== JSON.stringify(settings[key]))
  if (!changed.length) return
  saving.value = true
  try {
    for (const key of changed) {
      const response = await fetch('/api/settings/update', {
        method: 'POST', headers: authHeaders(true), body: JSON.stringify({ key, value: settings[key] }),
      })
      const result = await response.json()
      if (!response.ok || result.code !== 200) throw new Error(result.message || `保存 ${key} 失败`)
    }
    originalSettings.value = JSON.stringify(pickSettings())
    message.success('设置已保存')
  } catch (error) {
    message.error(error.message)
  } finally {
    saving.value = false
  }
}

const fetchInvitations = async () => {
  const response = await fetch('/api/invitations', { headers: authHeaders() })
  const result = await response.json()
  if (response.ok && result.code === 200) invitations.value = result.data || []
}

const generateInvites = async () => {
  generatingInvites.value = true
  try {
    const response = await fetch('/api/invitations', {
      method: 'POST', headers: authHeaders(true), body: JSON.stringify({ count: Math.min(20, Math.max(1, inviteCount.value || 1)) }),
    })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '生成失败')
    generatedCodes.value = result.data?.codes || []
    await fetchInvitations()
  } catch (error) {
    message.error(error.message)
  } finally {
    generatingInvites.value = false
  }
}

const deleteInvite = async (id) => {
  const response = await fetch(`/api/invitations/${id}`, { method: 'DELETE', headers: authHeaders() })
  const result = await response.json()
  if (!response.ok || result.code !== 200) return message.error(result.message || '删除失败')
  invitations.value = invitations.value.filter((invite) => invite.id !== id)
}

const copyGeneratedCodes = async () => {
  await navigator.clipboard.writeText(generatedCodes.value.join('\n'))
  message.success('邀请码已复制')
}

const handleLogoSelect = (event) => {
  const file = event.target.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => { cropperImage.value = reader.result; showCropper.value = true }
  reader.readAsDataURL(file)
  event.target.value = ''
}

const uploadLogo = async (blob) => {
  const form = new FormData()
  form.append('images[]', new File([blob], 'site-logo.png', { type: 'image/png' }))
  try {
    const response = await fetch('/api/upload/images?hidden=true', { method: 'POST', headers: authHeaders(), body: form })
    const result = await response.json()
    if (!response.ok || result.code !== 200 || !result.data?.files?.[0]?.url) throw new Error(result.message || '上传失败')
    settings.site_logo = result.data.files[0].url
    message.success('图标已上传，请保存设置')
  } catch (error) {
    message.error(error.message)
  }
}

const formatDate = (value) => new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
const formatFileSize = (bytes = 0) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

watch(() => settings.registration_mode, (mode) => { if (mode === 'invite') fetchInvitations() })
onMounted(async () => {
  await fetchSettings()
  await Promise.all([fetchLatestMigration(true), fetchMigrationSources()])
})
onBeforeUnmount(stopMigrationPolling)
</script>

<style scoped>
.settings-page { width: min(1180px, 100%); margin: 0 auto; padding: 1rem 0 4rem; color: #29323d; }
.dark .settings-page { color: #edf1f5; }
.page-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 2rem; margin-bottom: 1.5rem; }
.page-heading h1 { margin: .1rem 0 .35rem; font-size: clamp(2rem, 4vw, 3.4rem); line-height: 1; letter-spacing: -.055em; font-weight: 700; }
.page-heading p:not(.eyebrow), .panel-heading p, .soft-panel > p:last-child { color: #74808d; line-height: 1.65; }
.eyebrow { color: #d9637c; font-size: .72rem; font-weight: 700; letter-spacing: .15em; text-transform: uppercase; }
.settings-tabs { display: inline-flex; gap: .35rem; padding: .35rem; margin-bottom: 1.15rem; border: 1px solid rgba(210, 113, 135, .14); border-radius: 1rem; background: rgba(255,255,255,.66); box-shadow: 0 12px 34px rgba(89, 57, 66, .06); backdrop-filter: blur(18px); }
.dark .settings-tabs { background: rgba(34, 39, 47, .8); border-color: rgba(255,255,255,.07); }
.settings-tabs button { display:flex; align-items:center; gap:.45rem; padding:.65rem .9rem; border-radius:.75rem; color:#7c8793; transition:.2s ease; }
.settings-tabs button:hover { color:#d9637c; }
.settings-tabs button.active { color:#b94964; background:#fff5f6; box-shadow:0 4px 14px rgba(178,70,95,.09); }
.dark .settings-tabs button.active { color:#ff9eb2; background:#3b3038; }
.settings-grid { display:grid; grid-template-columns:repeat(12,minmax(0,1fr)); gap:1rem; align-items:start; }
.span-7 { grid-column:span 7; } .span-6 { grid-column:span 6; } .span-5 { grid-column:span 5; } .span-2 { grid-column:span 2; }
.panel { border:1px solid rgba(210,113,135,.13); border-radius:1.35rem; padding:1.35rem; background:rgba(255,255,255,.82); box-shadow:0 18px 55px rgba(77,51,58,.07); backdrop-filter:blur(18px); }
.dark .panel { background:rgba(31,36,44,.9); border-color:rgba(255,255,255,.065); box-shadow:0 22px 60px rgba(8,12,17,.32); }
.panel-heading { display:flex; gap:.8rem; align-items:center; margin-bottom:1.35rem; }
.panel-heading.compact { align-items:flex-start; }
.panel-heading h2, .soft-panel h2 { font-size:1.14rem; font-weight:700; letter-spacing:-.02em; }
.panel-heading p { margin-top:.12rem; font-size:.82rem; }
.icon-tile { display:grid; place-items:center; width:2.7rem; height:2.7rem; border-radius:.85rem; font-size:1.3rem; }
.icon-tile.coral { color:#c6506b; background:#fff0f2; } .icon-tile.pink { color:#ca6c9f;background:#fff1f8; } .icon-tile.blue { color:#527aa4;background:#edf6ff; } .icon-tile.green { color:#478775;background:#edf9f3; } .icon-tile.amber { color:#a77a3e;background:#fff7e8; }
.dark .icon-tile { background:rgba(255,255,255,.07); }
.field-block { display:flex; flex-direction:column; gap:.45rem; margin-bottom:1rem; color:#66717d; font-size:.8rem; font-weight:600; }
.field-block input:not([type='range']), .field-block textarea, .invite-actions input { width:100%; border:1px solid #e6dfe1; border-radius:.8rem; padding:.72rem .82rem; background:rgba(255,255,255,.86); color:#29323d; outline:none; font-weight:500; transition:.2s ease; }
.field-block input:focus, .field-block textarea:focus, .invite-actions input:focus { border-color:#e48ca0; box-shadow:0 0 0 3px rgba(228,140,160,.13); }
.dark .field-block input:not([type='range']), .dark .field-block textarea, .dark .invite-actions input { background:#252b34; color:#f2f4f7; border-color:#3b424d; }
.choice-grid { display:grid; grid-template-columns:repeat(3,1fr); gap:.65rem; }
.choice-card { position:relative; display:flex; min-height:8.2rem; flex-direction:column; gap:.3rem; padding:1rem; border:1px solid #eee5e7; border-radius:1rem; cursor:pointer; transition:.2s ease; }
.choice-card input, .storage-options input { position:absolute; opacity:0; pointer-events:none; }
.choice-card > i { color:#bd7887; font-size:1.3rem; margin-bottom:.35rem; }
.choice-card strong { color:#35404b; font-size:.9rem; }.choice-card span { color:#89929d; font-size:.75rem; line-height:1.45; }
.choice-card:hover, .choice-card.selected { border-color:#e293a5; background:#fff7f8; transform:translateY(-1px); }
.choice-card.selected::after { content:'✓'; position:absolute; right:.7rem; top:.6rem; color:#cb536d; font-weight:700; }
.dark .choice-card { border-color:#3a414b; }.dark .choice-card strong { color:#eff2f5; }.dark .choice-card.selected { background:#3b3038; }
.quota-settings{display:flex;align-items:flex-end;justify-content:space-between;gap:1rem;margin-top:1rem;padding:1rem 0 0;border-top:1px solid rgba(128,110,115,.1)}.quota-copy{display:flex;align-items:center;gap:.7rem;min-width:0}.quota-copy>div{display:flex;flex-direction:column;gap:.2rem}.quota-copy strong{font-size:.88rem}.quota-copy>div>span{color:#89929d;font-size:.76rem}.quota-icon{width:2.45rem;height:2.45rem;font-size:1.1rem}.quota-input{flex:0 0 9rem;margin:0}.quota-input>span{text-align:right;font-size:.7rem}
.switch-row { display:flex; justify-content:space-between; align-items:center; gap:1rem; padding:1rem 0; border-top:1px solid rgba(128,110,115,.1); }
.switch-row > div { display:flex; flex-direction:column; gap:.2rem; }.switch-row strong { font-size:.88rem; }.switch-row span { color:#89929d; font-size:.76rem; }
.toggle { position:relative; flex:0 0 auto; width:2.65rem; height:1.5rem; cursor:pointer; }.toggle input { position:absolute; opacity:0; }
.toggle span { position:absolute; inset:0; border-radius:999px; background:#d9dce0; transition:.2s ease; }.toggle span::after { content:''; position:absolute; width:1.12rem; height:1.12rem; left:.19rem; top:.19rem; border-radius:50%; background:white; box-shadow:0 2px 5px rgba(0,0,0,.15); transition:.2s ease; }
.toggle input:checked + span { background:#df7189; }.toggle input:checked + span::after { transform:translateX(1.15rem); }
.invite-box { margin-top:1rem; padding:1rem; border-radius:1rem; background:#fff9f4; border:1px solid #f3e3d6; }.dark .invite-box { background:#302d2b; border-color:#49413c; }
.subheading-row { display:flex; justify-content:space-between; gap:1rem; align-items:flex-start; }.subheading-row h3 { font-weight:700; }.subheading-row p { color:#8c837c; font-size:.74rem; margin-top:.2rem; }
.invite-actions { display:flex; gap:.45rem; }.invite-actions input { width:4rem; padding:.5rem; }
.generated-codes { display:grid; gap:.65rem; margin-top:1rem; padding:.9rem; border-radius:.8rem; color:#815261; background:#fff; border:1px dashed #e6a5b4; }.dark .generated-codes { background:#27242a; }
.generated-codes > div { display:flex; flex-direction:column; }.generated-codes span { color:#8a7d81; font-size:.72rem; }.generated-codes code { white-space:pre-wrap; user-select:text; font-family:ui-monospace,SFMono-Regular,Consolas,monospace; font-size:.82rem; line-height:1.65; }
.invite-list { margin-top:.8rem; }.invite-row { display:grid; grid-template-columns:1fr auto auto auto; gap:.7rem; align-items:center; min-height:2.65rem; border-top:1px solid rgba(139,112,119,.1); font-size:.76rem; }
.invite-hint { font-family:ui-monospace,SFMono-Regular,Consolas,monospace; }.invite-row time { color:#9a9194; font-variant-numeric:tabular-nums; }.status { padding:.2rem .45rem; border-radius:.4rem; font-size:.68rem; }.status.ready{color:#36816b;background:#eaf8f2}.status.used{color:#8b8587;background:#eee9ea}.dark .status.ready{background:#253f37}.dark .status.used{background:#3b393a}
.mini-empty { display:flex; align-items:center; justify-content:center; gap:.4rem; padding:1rem; color:#9a9294; }
.primary-button,.secondary-button { display:inline-flex; align-items:center; justify-content:center; gap:.45rem; border-radius:.8rem; font-weight:700; transition:.2s ease; }
.primary-button { padding:.72rem 1rem; color:#fff; background:#d85f79; box-shadow:0 10px 24px rgba(190,75,101,.2); }.primary-button:hover:not(:disabled){background:#c94e69;transform:translateY(-1px)}.primary-button:disabled{opacity:.55;cursor:not-allowed}
.secondary-button { padding:.58rem .75rem; color:#a4495e; background:#fff0f3; }.secondary-button:hover{background:#ffe4e9}.dark .secondary-button{background:#432f37;color:#ff9ab0}
.text-button { display:inline-flex; align-items:center; gap:.35rem; color:#bf5069; font-size:.76rem; font-weight:700; }.danger-text{color:#c55757;margin-left:.7rem}.icon-button{display:grid;place-items:center;width:2rem;height:2rem;border-radius:.6rem}.icon-button.danger{color:#c55757}.icon-button.danger:hover{background:#fff0f0}
.logo-editor { display:flex; align-items:center; gap:.8rem; }.logo-preview { display:grid; place-items:center; width:4.3rem;height:4.3rem;border-radius:1rem;background:#fff4f5;border:1px solid #f0dde1;color:#d47a8e;font-size:1.5rem;overflow:hidden}.logo-preview img{width:100%;height:100%;padding:.52rem;object-fit:contain}
.storage-options { display:grid; grid-template-columns:repeat(2,1fr); gap:.55rem; }.storage-options label { position:relative; display:flex; align-items:center; gap:.55rem; padding:.72rem; border:1px solid #ebe3e5; border-radius:.8rem; cursor:pointer; transition:.2s ease; }.storage-options label:hover,.storage-options label.selected{color:#bc536b;border-color:#e8a0b0;background:#fff6f7}.dark .storage-options label{border-color:#3a414b}.dark .storage-options label.selected{background:#3b3038}
.inline-note { display:flex; gap:.45rem; align-items:flex-start; margin-top:1rem; padding:.7rem; border-radius:.75rem; color:#7b7770; background:#faf7f1; font-size:.73rem; line-height:1.5; }.dark .inline-note{background:#2e2f31;color:#aaa39d}
.form-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:0 1rem; }.form-grid .span-2{grid-column:span 2}.nested-fields{margin-top:.7rem}
.span-12{grid-column:1/-1}.compact-switch{min-height:3rem;padding:.55rem 0;margin-bottom:1rem}.migration-panel{display:grid;gap:1rem}.migration-heading{margin-bottom:0}.migration-heading>div:nth-child(2){flex:1}.migration-heading .icon-button:disabled{opacity:.5}.migration-status-card{display:grid;gap:.75rem;padding:1rem;border:1px solid rgba(128,110,115,.1);border-radius:1rem;background:#faf8f8}.dark .migration-status-card{background:#292f37;border-color:rgba(255,255,255,.07)}.migration-status-row,.migration-status-row>div,.migration-metrics,.migration-actions{display:flex;align-items:center}.migration-status-row{justify-content:space-between;gap:1rem}.migration-status-row>div{gap:.65rem}.migration-status-row>b{font-size:1.1rem;font-variant-numeric:tabular-nums}.migration-badge{padding:.25rem .5rem;border-radius:.45rem;font-size:.68rem;font-weight:800}.migration-badge.pending,.migration-badge.running{color:#376f91;background:#e9f5fc}.migration-badge.failed{color:#a74747;background:#fff0f0}.migration-badge.completed{color:#357864;background:#eaf8f2}.dark .migration-badge.pending,.dark .migration-badge.running{background:#253c4a}.dark .migration-badge.failed{background:#482d31}.dark .migration-badge.completed{background:#253f37}.migration-progress{height:.48rem;border-radius:999px;background:#e9e2e4;overflow:hidden}.migration-progress i{display:block;height:100%;border-radius:inherit;background:linear-gradient(90deg,#da7188,#76a8c7);transition:width .35s ease}.dark .migration-progress{background:#3a414a}.migration-metrics{justify-content:space-between;color:#858e98;font-size:.72rem;font-variant-numeric:tabular-nums}.migration-error{color:#bd5050;font-size:.76rem}.migration-details{font-size:.73rem}.migration-details summary{cursor:pointer;color:#b65369;font-weight:700}.migration-details>div{display:grid;gap:.2rem;padding:.55rem 0;border-top:1px solid rgba(128,110,115,.1)}.migration-details code{overflow-wrap:anywhere}.migration-details span{color:#a64e4e}.retry-button{justify-self:start}.migration-form{display:grid;gap:1rem;padding-top:.25rem}.migration-form-label{color:#776f73;font-size:.73rem;font-weight:700}.migration-targets{grid-template-columns:repeat(2,minmax(0,9rem));}.migration-targets small{display:block;color:#939ba4;font-size:.65rem;font-weight:500}.migration-actions{justify-content:space-between;gap:1rem;padding-top:.8rem;border-top:1px solid rgba(128,110,115,.1)}.migration-actions p{display:flex;align-items:center;gap:.4rem;color:#837c7f;font-size:.73rem}.migration-actions .primary-button{flex:0 0 auto}
.input-suffix{display:flex;align-items:center}.input-suffix input{border-radius:.8rem 0 0 .8rem!important}.input-suffix b{align-self:stretch;display:grid;place-items:center;padding:0 .7rem;border:1px solid #e6dfe1;border-left:0;border-radius:0 .8rem .8rem 0;color:#968c8f;background:#faf7f8}.range{accent-color:#d85f79}
.soft-panel { min-height:19rem; display:flex; flex-direction:column; justify-content:flex-end; overflow:hidden; background:radial-gradient(circle at 80% 12%,rgba(234,146,165,.2),transparent 36%),rgba(255,255,255,.82); }.lifecycle-illustration{align-self:flex-end;margin:auto 1rem 1rem 0;font-size:4.5rem;color:#e497a8;transform:rotate(8deg)}
.loading-panel{display:grid;gap:.8rem}.skeleton-line,.skeleton-card{display:block;border-radius:.6rem;background:linear-gradient(90deg,#f2eaec,#faf7f8,#f2eaec);background-size:200% 100%;animation:shimmer 1.4s infinite}.skeleton-line{height:1rem;width:40%}.skeleton-line.wide{width:70%;height:2rem}.skeleton-card{height:14rem}@keyframes shimmer{to{background-position:-200% 0}}
@media(max-width:900px){.span-7,.span-6,.span-5{grid-column:1/-1}.settings-grid{gap:.8rem}}
@media(max-width:640px){.primary-button{width:100%}.settings-tabs{display:grid;grid-template-columns:repeat(4,1fr);width:100%}.settings-tabs button{justify-content:center;padding:.6rem .35rem}.settings-tabs button span{font-size:.72rem}.panel{padding:1rem;border-radius:1rem}.choice-grid{grid-template-columns:1fr}.choice-card{min-height:auto}.quota-settings{align-items:stretch;flex-direction:column}.quota-input{width:100%;flex-basis:auto}.quota-input>span{text-align:left}.form-grid{grid-template-columns:1fr}.form-grid .span-2{grid-column:span 1}.subheading-row{flex-direction:column}.invite-actions{width:100%}.invite-actions input{flex:1}.invite-row{grid-template-columns:1fr auto auto}.invite-row time{display:none}.migration-targets{grid-template-columns:1fr 1fr}.migration-status-row>div{align-items:flex-start;flex-direction:column}.migration-actions{align-items:stretch;flex-direction:column}.migration-actions .primary-button{width:100%}}
@media(max-width:640px){.primary-button{width:auto}}
</style>
