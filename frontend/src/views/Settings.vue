<template>
  <div class="text-gray-800 dark:text-gray-200" data-form-type="other">
    <!-- 隐藏的假表单，用于欺骗浏览器密码管理器，防止错误触发保存密码提示 -->
    <form style="display: none" aria-hidden="true">
      <input type="text" name="fake-username-field" />
      <input type="password" name="fake-password-field" />
    </form>
    <!-- 页面头部 -->
    <div class="settings-header container mx-auto px-4 py-4">
      <h1 class="page-title flex items-center text-2xl md:text-3xl font-bold">
        设置
      </h1>
      <p class="page-description text-gray-600 dark:text-gray-400 mt-2">
        管理您的系统设置
      </p>
    </div>

    <!-- 主要内容 -->
    <div class="container mx-auto px-4 pb-16">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <!-- 系统配置卡片 -->
        <div class="order-1 md:order-2 w-full p-0 mx-auto">
          <div
            class="panel-content p-6 md:p-8 bg-white dark:bg-gray-800 rounded-xl shadow-md"
          >
            <h2
              class="panel-title flex items-center text-xl font-semibold mb-8"
            >
              <span class="panel-icon mr-2 text-2xl">
                <i class="ri-list-settings-line"></i>
              </span>
              系统配置
            </h2>

            <div class="account-form space-y-6">
              <!-- 网站Logo设置：点击上传 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                  网站Logo
                </label>
                <div class="flex items-center gap-4">
                  <div
                    class="w-16 h-16 rounded-lg bg-gray-100 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 flex items-center justify-center overflow-hidden relative group cursor-pointer"
                    @click="triggerLogoUpload"
                  >
                    <img
                      v-if="systemSettings.site_logo"
                      :src="systemSettings.site_logo"
                      class="w-full h-full object-contain p-1"
                      alt="Site Logo"
                    />
                    <div
                      v-else
                      class="text-gray-400 dark:text-gray-500 text-xs text-center px-1"
                    >
                      点击上传
                    </div>
                    
                    <!-- 悬停遮罩 -->
                    <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                      <i class="ri-upload-2-line text-white text-xl"></i>
                    </div>
                  </div>
                  
                  <div class="flex-1">
                    <div class="flex gap-2 mb-2">
                        <button
                        type="button"
                        @click="triggerLogoUpload"
                        class="px-3 py-1.5 text-sm bg-primary text-white rounded hover:bg-primary-dark transition-colors"
                        >
                        上传Logo
                        </button>
                        <button
                        v-if="systemSettings.site_logo"
                        type="button"
                        @click="clearSiteLogo"
                        class="px-3 py-1.5 text-sm bg-red-50 text-red-600 rounded hover:bg-red-100 transition-colors"
                        >
                        清除
                        </button>
                    </div>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                      建议尺寸 128x128px，SVG 或 PNG 格式。上传后将应用于导航栏和浏览器标签页图标。
                    </p>
                  </div>
                  
                  <input
                    ref="logoInputRef"
                    type="file"
                    accept="image/*"
                    class="hidden"
                    @change="handleLogoSelect"
                  />
                </div>
              </div>

              <!-- 网站域名：失去焦点保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="site_domain"
                >
                  网站域名
                </label>
                <input
                  id="site_domain"
                  v-model="systemSettings.site_domain"
                  type="text"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  placeholder="如：example.com 或 https://example.com"
                  @blur="
                    handleFieldBlur('site_domain', systemSettings.site_domain)
                  "
                />
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  用于 Telegram Webhook 等功能，无需填写末尾斜杠
                </div>
              </div>

              <!-- TG Bot Token：失去焦点保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="tg_bot_token"
                >
                  TG Bot Token
                </label>
                <input
                  id="tg_bot_token"
                  v-model="systemSettings.tg_bot_token"
                  type="text"
                  autocomplete="off"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  placeholder="请输入TG Bot Token"
                  @blur="
                    handleFieldBlur('tg_bot_token', systemSettings.tg_bot_token)
                  "
                />
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  存储选择Telegram时必填
                </div>
              </div>

              <!-- TG 通知接收者：失去焦点保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="tg_receivers"
                >
                  TG 通知接收者
                </label>
                <input
                  id="tg_receivers"
                  v-model="systemSettings.tg_receivers"
                  type="text"
                  autocomplete="off"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  placeholder="接收通知的TG用户ID"
                  @blur="
                    handleFieldBlur('tg_receivers', systemSettings.tg_receivers)
                  "
                />
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  存储选择Telegram时必填
                </div>
              </div>

              <!-- TG 通知文本：失去焦点保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="tg_notice_text"
                >
                  TG 通知文本
                </label>
                <input
                  id="tg_notice_text"
                  v-model="systemSettings.tg_notice_text"
                  type="text"
                  autocomplete="off"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  placeholder="自定义TG通知文本"
                  @blur="
                    handleFieldBlur(
                      'tg_notice_text',
                      systemSettings.tg_notice_text
                    )
                  "
                />
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  默认模板：{username} {date} 上传了图片
                  {filename}，存储容器[{StorageType}]
                </div>
              </div>

              <!-- 最大文件大小：失去焦点保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="max_file_size"
                >
                  最大文件大小 (Bytes)
                </label>
                <input
                  id="max_file_size"
                  v-model="systemSettings.max_file_size"
                  type="number"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  placeholder="默认 10485760 (10MB)"
                  @blur="
                    handleFieldBlur(
                      'max_file_size',
                      systemSettings.max_file_size
                    )
                  "
                />
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  限制上传图片的最大体积，单位为字节 (Bytes)。例如 10MB = 10485760。
                </div>
              </div>

              <!-- 存储类型：下拉框变更保存 -->
              <div class="setting-group">
                <label
                  class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                  for="storage_type"
                >
                  存储类型
                </label>
                <select
                  id="storage_type"
                  v-model="systemSettings.storage_type"
                  class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                  @change="
                    handleSelectChange(
                      'storage_type',
                      systemSettings.storage_type
                    )
                  "
                >
                  <option value="" disabled>请选择存储类型</option>
                  <option value="default">本地存储</option>
                  <option value="s3">S3</option>
                  <option value="r2">R2</option>
                  <option value="webdav">WebDav</option>
                  <option value="ftp">FTP</option>

                  <option value="telegram">Telegram</option>
                  <option value="custom">NodeSeek / 自定义 API</option>
                </select>
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  <span v-if="systemSettings.storage_type === 'telegram'"
                    >选择Telegram存储必须使用海外服务器，否则无法正常上传、查看、删除图片</span
                  >
                  <span v-else-if="systemSettings.storage_type === 'custom'"
                    >支持 NodeSeek 图床 API 格式</span
                  >
                </div>
              </div>

              <!-- Custom API配置：失去焦点保存 -->
              <div
                v-if="systemSettings.storage_type === 'custom'"
                class="space-y-4 pt-2 border-t border-gray-200 dark:border-gray-700"
              >
                <h3 class="font-bold text-sm text-gray-800 dark:text-gray-200">
                  自定义 API 配置
                </h3>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="custom_api_url"
                  >
                    API 地址 (URL)
                  </label>
                  <input
                    id="custom_api_url"
                    v-model="systemSettings.custom_api_url"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="如 https://api.nodeimage.com"
                    @blur="
                      handleFieldBlur(
                        'custom_api_url',
                        systemSettings.custom_api_url
                      )
                    "
                  />
                  <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                    填写完整的 API 域名或根路径，无需包含 /api/upload
                  </div>
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="custom_api_key"
                  >
                    API Key
                  </label>
                  <input
                    id="custom_api_key"
                    v-model="systemSettings.custom_api_key"
                    type="text"
                    autocomplete="off"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请输入 API Key"
                    @blur="
                      handleFieldBlur(
                        'custom_api_key',
                        systemSettings.custom_api_key
                      )
                    "
                  />
                </div>
              </div>

              <!-- Watermark and Referer settings moved to left column -->

              <!-- S3/R2配置：失去焦点保存 -->
              <div
                v-if="['s3'].includes(systemSettings.storage_type)"
                class="space-y-4 pt-2 border-t border-gray-200 dark:border-gray-700"
              >
                <h3 class="font-bold text-sm text-gray-800 dark:text-gray-200">
                  S3 配置
                </h3>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="s3_endpoint"
                  >
                    S3 Endpoint
                  </label>
                  <input
                    id="s3_endpoint"
                    v-model="systemSettings.s3_endpoint"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="如：s3.us-west-004.backblazeb2.com"
                    @blur="
                      handleFieldBlur('s3_endpoint', systemSettings.s3_endpoint)
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="s3_access_key"
                  >
                    S3 AccessKey
                  </label>
                  <input
                    id="s3_access_key"
                    v-model="systemSettings.s3_access_key"
                    type="text"
                    autocomplete="off"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="S3访问密钥ID"
                    @blur="
                      handleFieldBlur(
                        's3_access_key',
                        systemSettings.s3_access_key
                      )
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="s3_secret_key"
                  >
                    S3 SecretKey
                  </label>
                  <input
                    id="s3_secret_key"
                    v-model="systemSettings.s3_secret_key"
                    type="password"
                    autocomplete="new-password"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="S3私有访问密钥"
                    @blur="
                      handleFieldBlur(
                        's3_secret_key',
                        systemSettings.s3_secret_key
                      )
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="S3Bucket"
                  >
                    S3 Bucket
                  </label>
                  <input
                    id="S3Bucket"
                    v-model="systemSettings.s3_bucket"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="存储桶名称"
                    @blur="
                      handleFieldBlur('s3_bucket', systemSettings.s3_bucket)
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="s3_custom_url"
                  >
                    自定义访问 URL（可选）
                  </label>
                  <input
                    id="s3_custom_url"
                    v-model="systemSettings.s3_custom_url"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="如：https://cdn.example.com"
                    @blur="
                      handleFieldBlur(
                        's3_custom_url',
                        systemSettings.s3_custom_url
                      )
                    "
                  />
                  <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                    留空则使用图床默认 URL，填写则使用 S3/R2 原始链接
                  </div>
                </div>

              <!-- R2配置：失去焦点保存 -->
              <div
                v-if="systemSettings.storage_type === 'r2'"
                class="space-y-4 pt-2 border-t border-gray-200 dark:border-gray-700"
              >
                <h3 class="font-bold text-sm text-gray-800 dark:text-gray-200">
                  R2 配置
                </h3>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="r2_endpoint"
                  >
                    R2 Endpoint
                  </label>
                  <input
                    id="r2_endpoint"
                    v-model="systemSettings.r2_endpoint"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="如：https://<ACCOUNT_ID>.r2.cloudflarestorage.com"
                    @blur="
                      handleFieldBlur('r2_endpoint', systemSettings.r2_endpoint)
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="r2_access_key"
                  >
                    R2 AccessKey
                  </label>
                  <input
                    id="r2_access_key"
                    v-model="systemSettings.r2_access_key"
                    type="text"
                    autocomplete="off"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="R2访问密钥ID"
                    @blur="
                      handleFieldBlur(
                        'r2_access_key',
                        systemSettings.r2_access_key
                      )
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="r2_secret_key"
                  >
                    R2 SecretKey
                  </label>
                  <input
                    id="r2_secret_key"
                    v-model="systemSettings.r2_secret_key"
                    type="password"
                    autocomplete="new-password"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="R2私有访问密钥"
                    @blur="
                      handleFieldBlur(
                        'r2_secret_key',
                        systemSettings.r2_secret_key
                      )
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="r2_bucket"
                  >
                    R2 Bucket
                  </label>
                  <input
                    id="r2_bucket"
                    v-model="systemSettings.r2_bucket"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="存储桶名称"
                    @blur="
                      handleFieldBlur('r2_bucket', systemSettings.r2_bucket)
                    "
                  />
                </div>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="r2_custom_url"
                  >
                    自定义访问 URL（可选）
                  </label>
                  <input
                    id="r2_custom_url"
                    v-model="systemSettings.r2_custom_url"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="如：https://img.example.com"
                    @blur="
                      handleFieldBlur(
                        'r2_custom_url',
                        systemSettings.r2_custom_url
                      )
                    "
                  />
                  <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                    留空则使用图床默认 URL，填写则使用 R2 原始链接（或绑定的自定义域名）
                  </div>
                </div>
              </div>


              <!-- WebDAV配置：失去焦点保存 -->
              <div
                v-if="systemSettings.storage_type === 'webdav'"
                class="space-y-4 pt-2 border-t border-gray-200 dark:border-gray-700"
              >
                <h3 class="font-bold text-sm text-gray-800 dark:text-gray-200">
                  WebDav 配置
                </h3>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="webdav_url"
                  >
                    WebDav URL
                  </label>
                  <input
                    id="webdav_url"
                    v-model="systemSettings.webdav_url"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 WebDav 地址"
                    @blur="
                      handleFieldBlur('webdav_url', systemSettings.webdav_url)
                    "
                  />
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="webdav_user"
                  >
                    WebDav 用户名
                  </label>
                  <input
                    id="webdav_user"
                    v-model="systemSettings.webdav_user"
                    type="text"
                    autocomplete="off"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 WebDav 用户名"
                    @blur="
                      handleFieldBlur('webdav_user', systemSettings.webdav_user)
                    "
                  />
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="webdav_pass"
                  >
                    WebDav 密码
                  </label>
                  <input
                    id="webdav_pass"
                    v-model="systemSettings.webdav_pass"
                    type="password"
                    autocomplete="new-password"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 WebDav 密码"
                    @blur="
                      handleFieldBlur('webdav_pass', systemSettings.webdav_pass)
                    "
                  />
                </div>
              </div>

              <!-- FTP配置：失去焦点保存 -->
              <div
                v-if="systemSettings.storage_type === 'ftp'"
                class="space-y-4 pt-2 border-t border-gray-200 dark:border-gray-700"
              >
                <h3 class="font-bold text-sm text-gray-800 dark:text-gray-200">
                  FTP 配置
                </h3>

                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="ftp_host"
                  >
                    FTP HOST
                  </label>
                  <input
                    id="ftp_host"
                    v-model="systemSettings.ftp_host"
                    type="text"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 FTP IP或域名"
                    @blur="handleFieldBlur('ftp_host', systemSettings.ftp_host)"
                  />
                </div>
                <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                  直接填写IP或域名，无需填写 ftp:// 或者 sftp://
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="ftp_port"
                  >
                    FTP 端口
                  </label>
                  <input
                    id="ftp_port"
                    v-model="systemSettings.ftp_port"
                    type="number"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="FTP 默认端口号 21"
                    @blur="handleFieldBlur('ftp_port', systemSettings.ftp_port)"
                  />
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="ftp_user"
                  >
                    FTP 用户名
                  </label>
                  <input
                    id="ftp_user"
                    v-model="systemSettings.ftp_user"
                    type="text"
                    autocomplete="username"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 FTP 用户名"
                    @blur="handleFieldBlur('ftp_user', systemSettings.ftp_user)"
                  />
                </div>
                <div class="setting-group">
                  <label
                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                    for="ftp_pass"
                  >
                    FTP 密码
                  </label>
                  <input
                    id="ftp_pass"
                    v-model="systemSettings.ftp_pass"
                    type="password"
                    autocomplete="new-password"
                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                    placeholder="请填写 FTP 登录密码"
                    @blur="handleFieldBlur('ftp_pass', systemSettings.ftp_pass)"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 系统设置卡片（开关部分不变） -->
        <div class="order-2 md:order-1 w-full p-0 mx-auto">
          <div
            class="panel-content p-6 md:p-8 bg-white dark:bg-gray-800 rounded-xl shadow-md"
          >
            <h2
              class="panel-title flex items-center text-xl font-semibold mb-8"
            >
              <span class="panel-icon mr-2 text-2xl">
                <i class="ri-settings-2-line"></i>
              </span>
              系统设置
            </h2>

            <div class="account-form space-y-6">
              <!-- 是否保存原图 -->
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  保存原图
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.original_image"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange(
                        'original_image',
                        systemSettings.original_image
                      )
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>

              <!-- 其他开关省略（和之前一致） -->
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  保存WEBP格式
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.save_webp"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange('save_webp', systemSettings.save_webp)
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  生成缩略图
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.thumbnail"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange('thumbnail', systemSettings.thumbnail)
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                生成缩略图，可提升后台预览速度，上传速度稍慢
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  允许游客登录
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.tourist"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange('tourist', systemSettings.tourist)
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                开启后访问登录页将自动以游客身份登录
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  启用TG通知
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.tg_notice"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange('tg_notice', systemSettings.tg_notice)
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-blue-500 dark:peer-checked:bg-blue-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                国内服务器不要开启TG通知
              </div>
              <!-- TG Webhook 上传开关 -->
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  启用TG Webhook上传
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.tg_webhook"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange(
                        'tg_webhook',
                        systemSettings.tg_webhook
                      )
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-blue-500 dark:peer-checked:bg-blue-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                开启后可向 Bot 发送图片 URL 来上传图片（需填写域名和 Bot Token）
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  Turnstile验证
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.turnstile"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange('turnstile', systemSettings.turnstile)
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-orange-500 dark:peer-checked:bg-orange-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                启用 Cloudflare Turnstile 人机验证，防止恶意登录
              </div>
              <!-- Turnstile 密钥配置 -->
              <div
                v-show="systemSettings.turnstile"
                class="pl-0 mt-4 space-y-4 border-t border-gray-100 dark:border-gray-700 pt-4"
              >
                <div class="setting-group">
                  <label
                    class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                    >站点密钥 (Site Key)</label
                  >
                  <input
                    v-model="systemSettings.turnstile_site_key"
                    type="text"
                    autocomplete="off"
                    class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                    placeholder="0x4AAAA..."
                    @blur="
                      handleFieldBlur(
                        'turnstile_site_key',
                        systemSettings.turnstile_site_key
                      )
                    "
                  />
                  <div class="mt-1 text-gray-400 text-[10px]">
                    从 Cloudflare Dashboard 获取，用于前端渲染验证组件
                  </div>
                </div>
                <div class="setting-group">
                  <label
                    class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                    >私密密钥 (Secret Key)</label
                  >
                  <input
                    v-model="systemSettings.turnstile_secret_key"
                    type="password"
                    autocomplete="new-password"
                    class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                    placeholder="0x4AAAA..."
                    @blur="
                      handleFieldBlur(
                        'turnstile_secret_key',
                        systemSettings.turnstile_secret_key
                      )
                    "
                  />
                  <div class="mt-1 text-gray-400 text-[10px]">
                    从 Cloudflare Dashboard 获取，用于后端验证 token
                  </div>
                </div>
                <div class="text-[10px] text-amber-600 dark:text-amber-400">
                  <i class="ri-information-line mr-1"></i>
                  也可通过环境变量 TURNSTILE_SITE_KEY 和 TURNSTILE_SECRET_KEY
                  配置
                </div>
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  开启图片水印
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.watermark_enable"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange(
                        'watermark_enable',
                        systemSettings.watermark_enable
                      )
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <!-- 水印详细配置 -->
              <div
                v-show="systemSettings.watermark_enable"
                class="pl-0 mt-4 space-y-4 border-t border-gray-100 dark:border-gray-700 pt-4"
              >
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div class="setting-group">
                    <label
                      class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                      >水印文本</label
                    >
                    <input
                      v-model="systemSettings.watermark_text"
                      type="text"
                      class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                      placeholder="水印文本"
                      @blur="
                        handleFieldBlur(
                          'watermark_text',
                          systemSettings.watermark_text
                        )
                      "
                    />
                  </div>
                  <div class="setting-group">
                    <label
                      class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                      >字体大小</label
                    >
                    <input
                      v-model="systemSettings.watermark_size"
                      type="text"
                      class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                      placeholder="如 20"
                      @blur="
                        handleFieldBlur(
                          'watermark_size',
                          systemSettings.watermark_size
                        )
                      "
                    />
                  </div>
                  <div class="setting-group">
                    <label
                      class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                      >字体颜色</label
                    >
                    <input
                      v-model="systemSettings.watermark_color"
                      type="text"
                      class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                      placeholder="#000000"
                      @blur="
                        handleFieldBlur(
                          'watermark_color',
                          systemSettings.watermark_color
                        )
                      "
                    />
                  </div>
                  <div class="setting-group">
                    <label
                      class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                      >透明度 (0-1)</label
                    >
                    <input
                      v-model="systemSettings.watermark_opac"
                      type="text"
                      class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                      placeholder="0.5"
                      @blur="
                        handleFieldBlur(
                          'watermark_opac',
                          systemSettings.watermark_opac
                        )
                      "
                    />
                  </div>
                  <div class="setting-group col-span-1 md:col-span-2">
                    <label
                      class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1"
                      >位置</label
                    >
                    <select
                      v-model="systemSettings.watermark_pos"
                      class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                      @change="
                        handleSelectChange(
                          'watermark_pos',
                          systemSettings.watermark_pos
                        )
                      "
                    >
                      <option value="top-left">左上角</option>
                      <option value="top-right">右上角</option>
                      <option value="bottom-left">左下角</option>
                      <option value="bottom-right">右下角</option>
                      <option value="center">居中</option>
                    </select>
                  </div>
                </div>
              </div>
              <div class="mt-1 text-gray-500 dark:text-gray-400 text-xs">
                新上传的图片自动添加水印，已上传的图片不会添加水印。
              </div>
              <div class="setting-group flex items-center justify-between py-2">
                <label
                  class="setting-label text-sm font-medium text-gray-700 dark:text-gray-300"
                >
                  开启来源白名单
                </label>
                <label class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    v-model="systemSettings.referer_white_enable"
                    class="sr-only peer"
                    @change="
                      handleSwitchChange(
                        'referer_white_enable',
                        systemSettings.referer_white_enable
                      )
                    "
                  />
                  <div
                    class="w-12 h-6 bg-gray-200 dark:bg-gray-700 rounded-full peer-checked:bg-green-500 dark:peer-checked:bg-green-600 switch-transition switch-antialias"
                  ></div>
                  <div
                    class="absolute left-1 top-1 bg-white dark:bg-gray-200 w-4 h-4 rounded-full switch-transition switch-antialias peer-checked:translate-x-6"
                  ></div>
                </label>
              </div>
              <!-- Referer 白名单配置 -->
              <div
                v-show="systemSettings.referer_white_enable"
                class="pl-0 mt-4 space-y-2 border-t border-gray-100 dark:border-gray-700 pt-4"
              >
                <label
                  class="block text-xs font-medium text-gray-500 dark:text-gray-400"
                  >允许的域名列表</label
                >
                <textarea
                  v-model="systemSettings.referer_white_list"
                  class="w-full px-3 py-2 text-sm border border-gray-200 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-700/50 focus:outline-none focus:border-primary"
                  placeholder="example.com, test.com"
                  rows="3"
                  @blur="
                    handleFieldBlur(
                      'referer_white_list',
                      systemSettings.referer_white_list
                    )
                  "
                ></textarea>
                <div class="text-[10px] text-gray-400">
                  仅需填写域名，多个用逗号分隔。无需http/端口。
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <!-- 图片裁剪模态框 -->
    <ImageCropper 
        v-model:visible="showCropper" 
        :image-src="cropperImage" 
        @cropped="handleCropConfirm" 
    />
</template>

<script setup>
import { ref, onMounted, reactive } from "vue";
import message from "@/utils/message.js";
import ImageCropper from '@/components/ImageCropper.vue';

const systemSettings = reactive({
  id: 1,
  original_image: false,
  save_webp: false,
  thumbnail: false,
  tourist: false,
  tg_notice: false,
  tg_webhook: false,
  turnstile: false,
  turnstile_site_key: "",
  turnstile_secret_key: "",
  site_domain: "",
  tg_bot_token: "",
  tg_receivers: "",
  tg_notice_text: "",
  storage_type: "",
  s3_endpoint: "",
  s3_access_key: "",
  s3_secret_key: "",
  s3_bucket: "",
  s3_custom_url: "",
  webdav_url: "",
  webdav_user: "",
  webdav_pass: "",
  ftp_host: "",
  ftp_port: 21,
  ftp_user: "",

  ftp_pass: "",
  custom_api_url: "",
  custom_api_key: "",
  watermark_enable: "",
  watermark_text: "",
  watermark_pos: "",
  watermark_size: "",
  watermark_color: "",
  watermark_opac: "",
  referer_white_list: "",
  referer_white_enable: false,
});

const updateSetting = reactive({});

// 加载状态
const isUpdating = ref(false);
let debounceTimer = null;

// 统一请求头配置（复用）
const getRequestHeaders = () => {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${localStorage.getItem("authToken")}`,
  };
};

// Logo Upload Logic
const logoInputRef = ref(null);
const showCropper = ref(false);
const cropperImage = ref('');

const triggerLogoUpload = () => {
  logoInputRef.value.click();
};

const handleLogoSelect = (event) => {
  const file = event.target.files[0];
  if (file) {
    if (file.size > 2 * 1024 * 1024) {
      message.error("图片大小不能超过 2MB");
      return;
    }
    const reader = new FileReader();
    reader.onload = (e) => {
      cropperImage.value = e.target.result;
      showCropper.value = true;
    };
    reader.readAsDataURL(file);
  }
  // Reset input
  event.target.value = '';
};

const handleCropConfirm = async (blob) => {
  const formData = new FormData();
  formData.append('images[]', blob, 'site_logo.png');
  
  // Custom upload generic handler or specific logo upload endpoint?
  // We can use the existing /api/upload/images?hidden=true but better to use a specific flow or just upload and get URL.
  // Wait, UpdateSettings expects a URL value. So we need to upload first.
  
  try {
     // Upload to get URL
     // We can reuse the upload endpoint. But we need to make sure it returns the URL.
     // Let's use /api/upload since it's an image.
     // But wait, user might want to hide it from gallery?
     // Yes, hidden=true.
     
     const uploadResponse = await fetch("/api/upload?hidden=true", {
        method: "POST",
        headers: {
            Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
        body: formData
     });
     
     const uploadResult = await uploadResponse.json();
     if (uploadResult.code === 200) {
        // API returns { data: { files: [...] } } structure
        const logoUrl = uploadResult.data.files && uploadResult.data.files.length > 0 ? uploadResult.data.files[0].url : '';
        if (logoUrl) {
            // Now save the setting
            saveSetting('site_logo', logoUrl);
            systemSettings.site_logo = logoUrl; // Immediate update
            localStorage.setItem("site_logo", logoUrl); // Cache it
        } else {
             message.error("Logo上传成功但未获取到URL");
        }
        
        // Also update favicon if possible (optional)
     } else {
        message.error("Logo上传失败: " + uploadResult.message);
     }
  } catch (error) {
     console.error(error);
     message.error("Logo上传出错");
  }
};

const clearSiteLogo = () => {
    saveSetting('site_logo', '');
    systemSettings.site_logo = '';
    localStorage.removeItem("site_logo");
};

const saveSetting = async (key, value) => {
  // 使用宽松比较，避免 undefined == '' 的情况触发更新
  // 同时检查 key 是否存在于 updateSetting 中
  if (key in updateSetting && updateSetting[key] == value) {
    return;
  }
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(async () => {
    try {
      if (isUpdating.value) return;
      isUpdating.value = true;

      const response = await fetch("/api/settings/update", {
        method: "POST",
        headers: getRequestHeaders(),
        body: JSON.stringify({
          key: key,
          value: value,
        }),
      });

      const result = await response.json();

      if (response.ok && result.code === 200) {
        message.success(`更新成功`);
        updateSetting[key] = value;
      } else {
        // 更新失败自动回滚
        if (updateSetting[key]) {
          systemSettings[key] = updateSetting[key];
        }
        message.error(`更新失败：${result.message || "未知错误"}`);
      }
    } catch (error) {
      console.error(`保存失败:`, error);
      message.error(`更新失败：网络异常`);
    } finally {
      isUpdating.value = false;
    }
  }, 300);
};

// 开关状态变更统一处理方法
const handleSwitchChange = (key, value) => {
  saveSetting(key, value);
};

// 输入框失去焦点处理
const handleFieldBlur = (key, value) => {
  if (key == "tg_bot_token" || key == "tg_receivers") {
    if (
      systemSettings.tg_bot_token == "" ||
      systemSettings.tg_receivers == ""
    ) {
      if (systemSettings.storage_type === "telegram") {
        // 未设置机器人令牌和接收者列表时，自动切回默认存储方式
        message.warning("配置不完整，切回默认存储方式");
        setTimeout(() => {
          systemSettings.storage_type = "default";
          saveSetting("storage_type", "default");
        }, 1500);

        return;
      }
    }
  }
  saveSetting(key, value);
};

// 下拉框变更处理
const handleSelectChange = async (key, value) => {
  // If Telegram storage is selected, you might want validation
  if (key == "storage_type" && value === "telegram") {
    if (!systemSettings.tg_bot_token || !systemSettings.tg_receivers) {
       message.warning("请先填写机器人令牌和接收者列表");
       return; 
    }
  }

  try {
    saveSetting(key, value);
  } catch (error) {
    message.error("更新失败: " + error.message);
  }
};

const fetchSettings = async () => {
    try {
        const response = await fetch("/api/settings/get", {
            method: "GET",
            headers: getRequestHeaders(),
        });
        const result = await response.json();
        if (result.code === 200) {
            Object.assign(systemSettings, result.data);
            // Sync tracker
            for (const k in result.data) {
                updateSetting[k] = result.data[k];
            }
            // Update local storage for logo
            if (result.data.site_logo) {
                localStorage.setItem("site_logo", result.data.site_logo);
            }
        }
    } catch (error) {
        console.error(error);
        message.error("加载设置失败");
    }
}

onMounted(() => {
  fetchSettings();
});
</script>

<style scoped>
.switch-transition {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.switch-antialias {
  transform: translateZ(0);
}
</style>
