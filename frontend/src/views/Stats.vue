<template>
  <div class="stats-page">
    <header class="page-heading">
      <div>
        <p class="eyebrow">{{ isAdmin ? 'System pulse' : 'Your image pulse' }}</p>
        <h1>{{ isAdmin ? '全站统计' : '我的统计' }}</h1>
        <p>{{ isAdmin ? '从上传趋势到存储分布，快速了解整个图床。' : '只统计你自己上传且仍在生命周期内的图片。' }}</p>
      </div>
      <button class="refresh-button" :disabled="loading" @click="loadDashboard">
        <i class="mgc_refresh_2_line" :class="{ 'animate-spin': loading }"></i>刷新
      </button>
    </header>

    <div v-if="loading && !loaded" class="summary-grid">
      <div v-for="index in 4" :key="index" class="summary-card skeleton"></div>
    </div>

    <template v-else>
      <section class="summary-grid">
        <article class="summary-card hero-card">
          <div><span>图片总数</span><strong>{{ formatNumber(stats.total_images) }}</strong></div>
          <i class="mgc_pic_2_line"></i>
          <small>今日 +{{ formatNumber(stats.today_uploads) }}</small>
        </article>
        <article class="summary-card">
          <span class="metric-icon"><i class="mgc_storage_line"></i></span>
          <div><span>占用空间</span><strong>{{ formatFileSize(stats.total_size) }}</strong><small>平均 {{ formatFileSize(stats.average_size) }}</small></div>
        </article>
        <article class="summary-card">
          <span class="metric-icon"><i class="mgc_calendar_month_line"></i></span>
          <div><span>本月上传</span><strong>{{ formatNumber(stats.month_uploads) }}</strong><small>最大 {{ formatFileSize(stats.largest_size) }}</small></div>
        </article>
        <article class="summary-card">
          <span class="metric-icon"><i class="mgc_time_duration_line"></i></span>
          <div><span>即将过期</span><strong>{{ formatNumber(stats.expiring_soon) }}</strong><small>{{ formatNumber(stats.permanent_images) }} 张永久保存</small></div>
        </article>
      </section>

      <section class="dashboard-grid">
        <article class="panel trend-panel">
          <div class="panel-heading">
            <div><p class="eyebrow">Upload rhythm</p><h2>上传趋势</h2></div>
            <div class="chart-controls">
              <div class="metric-switch">
                <button :class="{ active: chartMetric === 'count' }" @click="chartMetric = 'count'">数量</button>
                <button :class="{ active: chartMetric === 'size' }" @click="chartMetric = 'size'">空间</button>
              </div>
              <select v-model="period" @change="loadPeriod">
                <option value="dashboard">近 14 天</option>
                <option value="day">近 30 天</option>
                <option value="week">近 12 周</option>
                <option value="month">近 12 月</option>
                <option value="year">近 5 年</option>
              </select>
            </div>
          </div>

          <div class="trend-summary">
            <div><span>{{ activeTrendLabel }}</span><strong>{{ chartMetric === 'count' ? `${formatNumber(activeTrendValue)} 张` : formatFileSize(activeTrendValue) }}</strong></div>
            <small>峰值 {{ chartMetric === 'count' ? `${formatNumber(maxTrendValue)} 张` : formatFileSize(maxTrendValue) }}</small>
          </div>

          <div class="chart-wrap" @mouseleave="activePoint = trendData.length - 1">
            <div class="grid-lines"><span v-for="i in 4" :key="i"></span></div>
            <svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="上传趋势折线图">
              <defs><linearGradient id="areaFill" x1="0" x2="0" y1="0" y2="1"><stop offset="0" stop-color="#db6d84" stop-opacity=".28"/><stop offset="1" stop-color="#db6d84" stop-opacity="0"/></linearGradient></defs>
              <polygon :points="areaPoints" fill="url(#areaFill)" />
              <polyline :points="chartPoints" fill="none" stroke="#d45d77" stroke-width="1.8" vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" />
              <circle v-for="point in plottedTrend" :key="point.index" :cx="point.x" :cy="point.y" :r="activePoint === point.index ? 2.2 : 1.2" :class="{ active: activePoint === point.index }" @mouseenter="activePoint = point.index" />
            </svg>
            <div class="axis-labels"><span>{{ shortDate(trendData[0]?.date) }}</span><span>{{ shortDate(trendData.at(-1)?.date) }}</span></div>
          </div>
        </article>

        <article class="panel format-panel">
          <div class="panel-heading"><div><p class="eyebrow">Formats</p><h2>图片格式</h2></div><i class="mgc_chart_pie_2_line heading-icon"></i></div>
          <div v-if="stats.format_stats.length" class="distribution-list">
            <div v-for="item in stats.format_stats" :key="item.format" class="distribution-row">
              <div><strong>{{ formatMime(item.format) }}</strong><span>{{ item.count }} 张 · {{ formatFileSize(item.size) }}</span></div>
              <span>{{ percentage(item.count, stats.total_images) }}%</span>
              <div class="bar"><i :style="{ width: `${percentage(item.count, stats.total_images)}%` }"></i></div>
            </div>
          </div>
          <div v-else class="empty-mini"><i class="mgc_pic_line"></i>暂无格式数据</div>
        </article>

        <article class="panel storage-panel">
          <div class="panel-heading"><div><p class="eyebrow">Destinations</p><h2>存储分布</h2></div><span v-if="isAdmin" class="owner-count">{{ stats.active_owners }} 位上传者</span></div>
          <div class="storage-list">
            <div v-for="item in stats.storage_stats" :key="item.storage" class="storage-row">
              <span class="storage-icon"><i :class="storageIcon(item.storage)"></i></span>
              <div><strong>{{ storageLabel(item.storage) }}</strong><small>{{ item.count }} 张图片</small></div>
              <b>{{ formatFileSize(item.size) }}</b>
            </div>
            <div v-if="!stats.storage_stats.length" class="empty-mini"><i class="mgc_storage_line"></i>暂无存储数据</div>
          </div>
        </article>

        <article class="panel size-panel">
          <div class="panel-heading"><div><p class="eyebrow">Weight</p><h2>文件大小</h2></div><i class="mgc_chart_bar_2_line heading-icon"></i></div>
          <div class="size-bars">
            <div v-for="item in stats.size_distribution" :key="item.range">
              <span>{{ item.range }}</span>
              <div><i :style="{ width: `${sizeBarWidth(item.count)}%` }"></i></div>
              <strong>{{ item.count }}</strong>
            </div>
          </div>
        </article>

        <article class="panel recent-panel">
          <div class="panel-heading"><div><p class="eyebrow">Latest</p><h2>最近上传</h2></div><router-link to="/gallery">查看画廊<i class="mgc_right_line"></i></router-link></div>
          <div v-if="stats.recent_images.length" class="recent-strip">
            <a v-for="image in stats.recent_images" :key="image.id" :href="fullUrl(image.url)" target="_blank" rel="noopener">
              <img :src="fullUrl(image.thumbnail || image.url)" :alt="image.filename" loading="lazy" />
              <span><strong>{{ image.filename }}</strong><small>{{ formatFileSize(image.file_size) }}</small></span>
            </a>
          </div>
          <div v-else class="empty-mini large"><i class="mgc_pic_line"></i><span>还没有可统计的图片</span></div>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import message from '@/utils/message.js'

const currentUser = JSON.parse(localStorage.getItem('userInfo') || '{}')
const isAdmin = currentUser.role === 1 || currentUser.isAdmin === true
const loading = ref(false)
const loaded = ref(false)
const chartMetric = ref('count')
const period = ref('dashboard')
const activePoint = ref(0)
const trendData = ref([])
const stats = ref({
  total_images: 0, total_size: 0, average_size: 0, largest_size: 0, today_uploads: 0, month_uploads: 0,
  permanent_images: 0, expiring_soon: 0, active_owners: 0, recent_images: [], upload_trend: [], format_stats: [], storage_stats: [], size_distribution: [],
})

const metricValue = (item) => Number(item?.[chartMetric.value] || 0)
const maxTrendValue = computed(() => Math.max(0, ...trendData.value.map(metricValue)))
const plottedTrend = computed(() => {
  const max = maxTrendValue.value || 1
  const lastIndex = Math.max(1, trendData.value.length - 1)
  return trendData.value.map((item, index) => ({ item, index, x: (index / lastIndex) * 100, y: 88 - (metricValue(item) / max) * 70 }))
})
const chartPoints = computed(() => plottedTrend.value.map((point) => `${point.x},${point.y}`).join(' '))
const areaPoints = computed(() => plottedTrend.value.length ? `0,90 ${chartPoints.value} 100,90` : '')
const activeTrend = computed(() => trendData.value[activePoint.value] || trendData.value.at(-1) || {})
const activeTrendValue = computed(() => metricValue(activeTrend.value))
const activeTrendLabel = computed(() => activeTrend.value.date ? new Date(normalizeDate(activeTrend.value.date)).toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: activeTrend.value.date.length > 7 ? 'numeric' : undefined }) : '暂无数据')
const maxSizeCount = computed(() => Math.max(1, ...stats.value.size_distribution.map((item) => item.count)))

const authHeaders = () => ({ Authorization: `Bearer ${localStorage.getItem('authToken') || ''}` })
const loadDashboard = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/stats/dashboard', { headers: authHeaders() })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '加载统计失败')
    stats.value = { ...stats.value, ...(result.data || {}) }
    if (period.value === 'dashboard') {
      trendData.value = result.data?.upload_trend || []
      activePoint.value = Math.max(0, trendData.value.length - 1)
    }
    loaded.value = true
  } catch (error) { message.error(error.message) } finally { loading.value = false }
}

const loadPeriod = async () => {
  if (period.value === 'dashboard') {
    trendData.value = stats.value.upload_trend || []
    activePoint.value = Math.max(0, trendData.value.length - 1)
    return
  }
  try {
    const response = await fetch(`/api/stats/images?period=${period.value}`, { headers: authHeaders() })
    const result = await response.json()
    if (!response.ok || result.code !== 200) throw new Error(result.message || '加载趋势失败')
    trendData.value = result.data || []
    activePoint.value = Math.max(0, trendData.value.length - 1)
  } catch (error) { message.error(error.message) }
}

const formatNumber = (value) => Number(value || 0).toLocaleString('zh-CN')
const formatFileSize = (bytes) => {
  const value = Number(bytes || 0)
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(units.length - 1, Math.floor(Math.log(value) / Math.log(1024)))
  return `${(value / 1024 ** index).toFixed(index > 1 ? 2 : 0)} ${units[index]}`
}
const percentage = (part, total) => total ? Math.round((part / total) * 100) : 0
const sizeBarWidth = (count) => Math.max(count ? 7 : 0, Math.round((count / maxSizeCount.value) * 100))
const formatMime = (mime) => ({ 'image/jpeg': 'JPEG', 'image/png': 'PNG', 'image/webp': 'WebP', 'image/gif': 'GIF', 'image/avif': 'AVIF' }[mime] || (mime || '未知').replace('image/', '').toUpperCase())
const storageLabel = (storage) => ({ default: '本地存储', s3: 'Amazon S3', r2: 'Cloudflare R2', webdav: 'WebDAV', ftp: 'FTP', telegram: 'Telegram' }[storage] || storage || '未知')
const storageIcon = (storage) => ({ default: 'mgc_folder_line', s3: 'mgc_cloud_line', r2: 'mgc_cloud_2_line', webdav: 'mgc_server_2_line', ftp: 'mgc_transfer_4_line', telegram: 'mgc_telegram_line' }[storage] || 'mgc_storage_line')
const fullUrl = (path) => !path || /^https?:\/\//.test(path) ? path : `${window.location.origin}${path}`
const normalizeDate = (date) => date.length === 4 ? `${date}-01-01` : date.length === 7 ? `${date}-01` : date
const shortDate = (date) => date ? date.slice(date.length === 4 ? 0 : 5) : ''
onMounted(loadDashboard)
</script>

<style scoped>
.stats-page{width:min(1180px,100%);margin:0 auto;padding:1rem 0 4rem;color:#303a44}.dark .stats-page{color:#edf1f4}.page-heading{display:flex;align-items:flex-end;justify-content:space-between;gap:2rem;margin-bottom:1.4rem}.eyebrow{color:#c65a73;font-size:.68rem;font-weight:800;letter-spacing:.15em;text-transform:uppercase}.page-heading h1{margin:.12rem 0 .35rem;font-size:clamp(2rem,4.3vw,3.55rem);line-height:1;letter-spacing:-.06em;font-weight:700}.page-heading>div>p:last-child{color:#858f99;font-size:.84rem}.refresh-button{display:flex;align-items:center;gap:.38rem;padding:.62rem .8rem;border:1px solid #eadfe2;border-radius:.75rem;color:#ac5367;background:rgba(255,255,255,.72);font-size:.75rem;font-weight:700;transition:.2s}.refresh-button:hover{background:#fff1f4}.dark .refresh-button{background:#252b33;border-color:#3d4550;color:#ff9bb0}.summary-grid{display:grid;grid-template-columns:1.2fr repeat(3,1fr);gap:.8rem;margin-bottom:.8rem}.summary-card{min-height:8.6rem;padding:1.15rem;display:flex;align-items:center;gap:.8rem;border:1px solid rgba(211,122,143,.13);border-radius:1.15rem;background:rgba(255,255,255,.8);box-shadow:0 15px 42px rgba(76,47,56,.06);backdrop-filter:blur(16px)}.dark .summary-card{background:rgba(30,35,43,.88);border-color:rgba(255,255,255,.06);box-shadow:0 16px 40px rgba(3,8,13,.28)}.summary-card>div{display:flex;flex-direction:column;min-width:0}.summary-card span{color:#858e98;font-size:.72rem;font-weight:700}.summary-card strong{margin:.22rem 0 .1rem;font-size:1.55rem;letter-spacing:-.04em;font-variant-numeric:tabular-nums}.summary-card small{color:#a29a9d;font-size:.66rem}.metric-icon{display:grid!important;place-items:center;flex:0 0 auto;width:2.65rem;height:2.65rem;border-radius:.82rem;color:#c65a73!important;background:#fff0f3;font-size:1.2rem!important}.dark .metric-icon{background:#3b3037}.hero-card{position:relative;align-items:flex-end;justify-content:space-between;overflow:hidden;color:#fff;background:radial-gradient(circle at 85% 12%,rgba(255,255,255,.28),transparent 26%),linear-gradient(145deg,#db7188,#bd536c)}.dark .hero-card{background:radial-gradient(circle at 85% 12%,rgba(255,255,255,.12),transparent 26%),linear-gradient(145deg,#a74c62,#713747)}.hero-card span,.hero-card small{color:rgba(255,255,255,.76)}.hero-card>i{position:absolute;right:.9rem;top:.5rem;font-size:4.4rem;opacity:.16;transform:rotate(8deg)}.hero-card>small{position:absolute;right:1rem;bottom:1rem}.dashboard-grid{display:grid;grid-template-columns:1.25fr .75fr;gap:.8rem}.panel{padding:1.2rem;border:1px solid rgba(211,122,143,.13);border-radius:1.2rem;background:rgba(255,255,255,.82);box-shadow:0 17px 48px rgba(76,47,56,.065);backdrop-filter:blur(18px)}.dark .panel{background:rgba(30,35,43,.9);border-color:rgba(255,255,255,.06);box-shadow:0 18px 45px rgba(3,8,13,.3)}.panel-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:1rem;margin-bottom:1rem}.panel-heading h2{font-size:1.05rem;font-weight:700;letter-spacing:-.02em}.heading-icon{color:#c76b80;font-size:1.35rem}.trend-panel{grid-column:span 1;min-height:24rem}.chart-controls{display:flex;gap:.45rem}.metric-switch{display:flex;padding:.2rem;border-radius:.58rem;background:#f5eff0}.metric-switch button{padding:.35rem .5rem;border-radius:.43rem;color:#92888c;font-size:.65rem;font-weight:700}.metric-switch button.active{color:#ae4960;background:#fff}.dark .metric-switch{background:#252b33}.dark .metric-switch button.active{color:#ff9bb0;background:#3b3037}.chart-controls select{padding:.4rem .55rem;border:1px solid #e8dfe2;border-radius:.58rem;color:#747d86;background:#fff;font-size:.66rem;outline:none}.dark .chart-controls select{color:#dce1e5;background:#252b33;border-color:#3b434e}.trend-summary{display:flex;align-items:flex-end;justify-content:space-between;margin:1.2rem 0 .5rem}.trend-summary>div{display:flex;flex-direction:column}.trend-summary span,.trend-summary small{color:#999195;font-size:.68rem}.trend-summary strong{font-size:1.65rem;letter-spacing:-.04em;font-variant-numeric:tabular-nums}.chart-wrap{position:relative;height:13rem;padding-bottom:1.1rem}.chart-wrap svg{position:absolute;inset:0 0 1.1rem;width:100%;height:calc(100% - 1.1rem);overflow:visible}.chart-wrap circle{fill:#fff;stroke:#d45d77;stroke-width:1;vector-effect:non-scaling-stroke;cursor:pointer;transition:r .15s}.dark .chart-wrap circle{fill:#232932}.chart-wrap circle.active{fill:#d45d77}.grid-lines{position:absolute;inset:0 0 1.1rem;display:flex;flex-direction:column;justify-content:space-between}.grid-lines span{display:block;border-top:1px dashed rgba(130,108,114,.12)}.axis-labels{position:absolute;inset:auto 0 0;display:flex;justify-content:space-between;color:#aaa2a5;font-size:.61rem}.format-panel{min-height:24rem}.distribution-list{display:flex;flex-direction:column;gap:1rem}.distribution-row{display:grid;grid-template-columns:1fr auto;gap:.45rem}.distribution-row>div:first-child{display:flex;flex-direction:column}.distribution-row strong{font-size:.78rem}.distribution-row span{color:#999195;font-size:.64rem}.distribution-row>span{font-variant-numeric:tabular-nums}.bar{grid-column:1/-1;height:.35rem;border-radius:999px;background:#f0e9eb;overflow:hidden}.bar i{display:block;height:100%;border-radius:inherit;background:#d76c83}.dark .bar{background:#3a3f48}.storage-panel,.size-panel{min-height:17rem}.owner-count{padding:.32rem .5rem;border-radius:.45rem;color:#a84a60;background:#fff0f3;font-size:.62rem;font-weight:700}.dark .owner-count{color:#ff9bb0;background:#3b3037}.storage-list{display:flex;flex-direction:column}.storage-row{display:grid;grid-template-columns:auto 1fr auto;align-items:center;gap:.65rem;padding:.72rem 0;border-top:1px solid rgba(125,108,113,.09)}.storage-icon{display:grid;place-items:center;width:2.25rem;height:2.25rem;border-radius:.7rem;color:#bf5d73;background:#fff2f4}.dark .storage-icon{background:#3b3037}.storage-row>div{display:flex;flex-direction:column}.storage-row strong{font-size:.76rem}.storage-row small{color:#9b9396;font-size:.62rem}.storage-row>b{font-size:.7rem;font-variant-numeric:tabular-nums}.size-bars{display:flex;flex-direction:column;gap:.75rem}.size-bars>div{display:grid;grid-template-columns:5.7rem 1fr 1.6rem;align-items:center;gap:.55rem}.size-bars span,.size-bars strong{font-size:.64rem}.size-bars strong{text-align:right;font-variant-numeric:tabular-nums}.size-bars>div>div{height:.45rem;border-radius:999px;background:#f0e9eb;overflow:hidden}.size-bars i{display:block;height:100%;border-radius:inherit;background:#d76c83}.dark .size-bars>div>div{background:#3a3f48}.recent-panel{grid-column:1/-1}.recent-panel .panel-heading a{display:flex;align-items:center;gap:.25rem;color:#b65369;font-size:.7rem;font-weight:700}.recent-strip{display:grid;grid-template-columns:repeat(4,1fr);gap:.65rem}.recent-strip a{position:relative;aspect-ratio:16/10;border-radius:.8rem;overflow:hidden;background:#f1ebed}.recent-strip img{width:100%;height:100%;object-fit:cover;transition:.3s ease}.recent-strip a:hover img{transform:scale(1.04)}.recent-strip a>span{position:absolute;inset:auto .4rem .4rem;display:flex;justify-content:space-between;align-items:center;gap:.4rem;padding:.38rem .45rem;border:1px solid rgba(255,255,255,.2);border-radius:.5rem;color:#fff;background:rgba(34,29,31,.58);backdrop-filter:blur(8px)}.recent-strip strong{overflow:hidden;text-overflow:ellipsis;font-size:.61rem;white-space:nowrap}.recent-strip small{flex:0 0 auto;font-size:.56rem}.empty-mini{min-height:9rem;display:flex;align-items:center;justify-content:center;gap:.4rem;color:#a49b9e;font-size:.72rem}.empty-mini.large{flex-direction:column;font-size:.78rem}.empty-mini.large i{font-size:2rem}.skeleton{background:linear-gradient(90deg,#f2eaec,#faf7f8,#f2eaec);background-size:200% 100%;animation:shimmer 1.4s infinite}@keyframes shimmer{to{background-position:-200% 0}}@media(max-width:900px){.summary-grid{grid-template-columns:repeat(2,1fr)}.dashboard-grid{grid-template-columns:1fr}.recent-strip{grid-template-columns:repeat(3,1fr)}}@media(max-width:640px){.page-heading{align-items:flex-start;flex-direction:column;gap:.8rem}.refresh-button{width:100%;justify-content:center}.summary-grid{grid-template-columns:1fr 1fr}.summary-card{min-height:7.6rem;align-items:flex-start;flex-direction:column;padding:1rem}.hero-card{grid-column:1/-1}.metric-icon{width:2.25rem;height:2.25rem}.summary-card strong{font-size:1.25rem}.chart-controls{align-items:flex-end;flex-direction:column}.trend-panel{min-height:22rem}.recent-strip{display:flex;overflow-x:auto;scroll-snap-type:x mandatory}.recent-strip a{flex:0 0 75%;scroll-snap-align:start}.panel{padding:1rem}}
</style>
