<template>
  <!-- 主要内容区域 -->
  <div class="home-page">
    <div class="home-tools">
      <div class="quick-switches" aria-label="快捷开关">
        <button class="quick-toggle" :aria-pressed="isDarkMode" @click="toggleTheme">
          <span class="quick-icon"><i :class="isDarkMode ? 'mgc_moon_stars_fill' : 'mgc_sun_2_line'"></i></span>
          <span class="quick-copy"><b>{{ isDarkMode ? '夜间' : '日间' }}</b></span>
        </button>
        <button class="quick-toggle" :class="{ active: saveWebp }" :aria-pressed="saveWebp" @click="toggleSaveWebp">
          <span class="quick-icon"><i class="mgc_pic_ai_line"></i></span>
          <span class="quick-copy"><b>WebP</b><em>{{ saveWebp ? '开' : '关' }}</em></span>
        </button>
      </div>
    </div>

    <!-- 上传区域 -->
    <section class="upload-section mb-6">
      <div
        class="upload-panel"
      >
        <div class="upload-toolbar">
          <div class="lifetime-select">
            <span><i class="mgc_time_duration_line"></i>保存时间</span>
            <AppSelect v-model="expiresIn" :options="lifetimeOptions" aria-label="选择图片保存时间" />
          </div>
          <!-- 上传模式切换 -->
          <div
            class="mode-switch"
          >
            <button
              @click="uploadMode = 'file'"
              class="px-3 py-1.5 text-sm rounded-md transition-all duration-200"
              :class="
                uploadMode === 'file'
                  ? 'bg-white dark:bg-dark-200 text-primary shadow-sm'
                  : 'text-secondary hover:text-primary'
              "
            >
              <i class="mgc_pic_line"></i><span>文件</span>
            </button>
            <button
              @click="uploadMode = 'url'"
              class="px-3 py-1.5 text-sm rounded-md transition-all duration-200"
              :class="
                uploadMode === 'url'
                  ? 'bg-white dark:bg-dark-200 text-primary shadow-sm'
                  : 'text-secondary hover:text-primary'
              "
            >
              <i class="mgc_link_2_line"></i><span>URL</span>
            </button>
          </div>
        </div>

        <!-- 文件上传模式 -->
        <div v-if="uploadMode === 'file'">
          <!-- 拖拽上传区域 -->
          <div
            class="upload-area group relative rounded-2xl border-2 border-dashed transition-all duration-300 cursor-pointer overflow-hidden hover:border-primary/50 hover:shadow-[0_0_15px_rgba(var(--primary-rgb),0.1)] dark:hover:shadow-[0_0_15px_rgba(var(--primary-rgb),0.2)]"
            :class="{
              'border-primary/30 bg-primary/5 dark:bg-primary/5': isDragOver,
              'border-light-300 dark:border-dark-100 bg-white dark:bg-dark-200/60':
                !isDragOver && !isUploading,
              'border-primary/50 bg-primary/10 dark:bg-primary/10': isUploading,
            }"
            @drop="handleDrop"
            @dragover="handleDragOver"
            @dragenter="handleDragEnter"
            @dragleave="handleDragLeave"
            @click="triggerFileInput"
          >
            <!-- 未上传状态 -->
            <div
              v-if="!isUploading"
              class="upload-content py-12 px-4 text-center"
            >
              <div class="upload-icon mb-4 flex justify-center">
                <i class="upload-arrow mgc_arrow_up_line"></i>
              </div>
              <h3
                class="text-xl font-bold text-gray-700 dark:text-gray-200 mb-2"
              >
                Click, Paste or Drop
              </h3>
              <p class="text-gray-400 text-sm mb-6 font-medium">
                JPG, PNG, GIF, WEBP, HEIC
              </p>
              <!-- Hidden button if visual is enough, or keep it minimal -->
              <button
                class="hidden bg-primary/10 text-primary px-6 py-2 rounded-full font-medium hover:bg-primary/20 transition-colors duration-200 items-center justify-center gap-2 mx-auto"
              >
                <i class="mgc_add_line"></i>
                选择图片
              </button>
              <p
                class="paste-tip text-sm text-secondary flex items-center justify-center gap-2 mt-3"
              >
                支持 Ctrl+V 粘贴剪贴板图片，或直接拖入图片
              </p>
            </div>

            <!-- 上传进度状态 -->
            <div v-else class="upload-progress py-16 px-4 text-center">
              <div
                class="spinner w-10 h-10 border-4 border-primary/30 border-t-primary rounded-full animate-spin mx-auto mb-3"
              ></div>
              <p class="text-secondary text-sm mb-3">
                正在上传 {{ uploadingCount }} 个文件（{{
                  Math.round(uploadProgress)
                }}%）
              </p>
              <div
                class="progress-bar w-full max-w-md mx-auto h-2 bg-light-200 dark:bg-dark-100 rounded-full overflow-hidden"
              >
                <div
                  class="progress-fill h-full bg-primary transition-all duration-300 ease-out"
                  :style="{ width: uploadProgress + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- URL上传模式 -->
        <div v-else class="url-upload-area">
          <div class="flex flex-col gap-4 py-8 px-4">
            <div class="text-center mb-2">
              <div class="text-4xl text-primary mb-2">
                <i class="mgc_link_2_line"></i>
              </div>
              <p class="text-secondary text-sm">
                粘贴图片直链地址，支持 http/https 协议
              </p>
            </div>
            <div
              class="flex flex-col sm:flex-row gap-3 max-w-xl mx-auto w-full"
            >
              <input
                v-model="urlInput"
                type="url"
                placeholder="请输入图片URL"
                class="flex-1 px-4 py-3 rounded-xl border border-light-300 dark:border-dark-100 bg-white dark:bg-dark-200 focus:ring-2 focus:ring-primary focus:border-primary outline-none transition-all w-full"
                @keydown.enter="uploadByUrl"
                :disabled="isUploadingUrl"
              />
              <button
                @click="uploadByUrl"
                :disabled="!urlInput.trim() || isUploadingUrl"
                class="px-6 py-3 bg-primary hover:bg-primary-dark text-white rounded-xl transition-colors duration-200 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed w-full sm:w-auto shrink-0"
              >
                <i
                  v-if="isUploadingUrl"
                  class="mgc_loading_line animate-spin"
                ></i>
                <i v-else class="mgc_upload_2_line"></i>
                <span>{{ isUploadingUrl ? "上传中..." : "上传" }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- 隐藏的文件输入 -->
        <input
          ref="fileInput"
          type="file"
          multiple
          accept="image/*,.heic,.heif"
          @change="handleFileSelect"
          class="hidden"
        />
      </div>
    </section>

    <!-- 最近上传的图片 -->
    <section class="recent-section">
      <div class="flex justify-between items-center mb-3">
        <h2 class="section-title text-base font-extrabold tracking-wider text-primary flex items-center gap-2">
          <span>LATEST</span>
          <span class="text-xs font-medium text-secondary opacity-80 tracking-normal">最近上传</span>
        </h2>
      </div>

      <!-- 图片网格 -->
      <!-- 骨架屏加载状态 -->
      <div
        v-if="isLoadingRecent"
        class="recent-skeleton grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <div
          v-for="i in 6"
          :key="i"
          class="rounded-2xl bg-white dark:bg-dark-100 border border-light-200/80 dark:border-dark-100/80 overflow-hidden"
        >
          <div
            class="aspect-video bg-gray-200 dark:bg-dark-200/50 animate-pulse"
          ></div>
          <div
            class="px-3 py-2 bg-white/95 dark:bg-dark-200/90 border-t border-light-200/50 dark:border-dark-100/50 flex justify-between items-center"
          >
            <div
              class="w-1/2 h-4 bg-gray-200 dark:bg-dark-300 rounded animate-pulse"
            ></div>
            <div class="flex gap-2">
              <div
                class="w-8 h-8 rounded bg-gray-200 dark:bg-dark-300 animate-pulse"
              ></div>
              <div
                class="w-8 h-8 rounded bg-gray-200 dark:bg-dark-300 animate-pulse"
              ></div>
              <div
                class="w-8 h-8 rounded bg-gray-200 dark:bg-dark-300 animate-pulse"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-else-if="recentImages.length > 0"
        class="recent-grid grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <div
          v-for="image in recentImages"
          :key="image.id"
          class="recent-item rounded-2xl bg-white dark:bg-dark-100 transition-all duration-300 hover:shadow-xl dark:hover:shadow-dark-md group relative overflow-visible flex flex-col border border-light-200/80 dark:border-dark-100/80"
        >
          <span v-if="image.expires_at" class="expiry-badge">
            <i class="mgc_time_line"></i>{{ formatExpiration(image.expires_at) }}
          </span>
          <!-- 图片区域 -->
          <div
            class="aspect-video overflow-hidden cursor-pointer rounded-t-2xl"
            @click.stop="previewImage(image)"
          >
            <div
              class="loading absolute inset-0 flex items-center justify-center pointer-events-none"
            >
              <i
                class="mgc_loading_line text-3xl animate-spin text-gray-300 dark:text-gray-600"
              ></i>
            </div>
            <img
              :src="getFullUrl(image.thumbnail || image.url)"
              :alt="image.filename || '图片预览'"
              class="recent-image w-full h-full object-cover transition-all duration-500 group-hover:scale-110 opacity-0"
              loading="lazy"
              referrerpolicy="no-referrer"
              @load="
                (e) => {
                  e.target.classList.remove('opacity-0');
                  e.target.parentElement
                    .querySelector('.loading')
                    .classList.add('hidden');
                }
              "
              @error="handleImageError"
            />
          </div>
          <!-- 底部操作栏（移动端可见） -->
          <div
            class="flex items-center gap-3 justify-between px-3 py-2 bg-white/95 dark:bg-dark-200/90 rounded-b-2xl shadow-inner"
          >
            <div class="flex flex-col min-w-0 flex-1">
              <p
                class="recent-filename text-xs font-medium text-gray-800 dark:text-light-100 break-all leading-snug"
              >
                {{ image.filename }}
              </p>
              <p class="text-[11px] text-secondary leading-tight mt-0.5">
                {{ image.expires_at ? formatExpiration(image.expires_at) : formatDate(image.created_at) }}
              </p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <div
                class="relative"
                :class="{ 'z-50': activeCopyMenu === image.id }"
              >
                <button
                  class="halo-button-copy h-8 w-8 flex items-center justify-center"
                  title="复制链接"
                  @click.stop="toggleCardCopyMenu(image.id)"
                >
                  <i class="mgc_code_line text-sm"></i>
                </button>
                <div
                  v-show="activeCopyMenu === image.id"
                  class="copy-dropdown absolute right-0 top-full mt-1 w-36 bg-white/95 dark:bg-dark-200/95 rounded-2xl shadow-2xl border border-light-200/80 dark:border-dark-100/80 backdrop-blur-xl"
                >
                  <div class="p-1.5 grid grid-cols-2 gap-1.5">
                    <button
                      @click.stop="copyImageLink(image, 'url')"
                      class="w-full text-left px-2 py-1.5 text-[11px] text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded-lg transition-colors duration-200 flex items-center gap-1.5"
                    >
                      <i class="mgc_link_2_line text-primary"></i>
                      <span class="font-semibold">URL</span>
                    </button>
                    <button
                      @click.stop="copyImageLink(image, 'markdown')"
                      class="w-full text-left px-2 py-1.5 text-[11px] text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded-lg transition-colors duration-200 flex items-center gap-1.5"
                    >
                      <i class="mgc_markdown_line text-blue-500"></i>
                      <span class="font-semibold">MD</span>
                    </button>
                    <button
                      @click.stop="copyImageLink(image, 'html')"
                      class="w-full text-left px-2 py-1.5 text-[11px] text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded-lg transition-colors duration-200 flex items-center gap-1.5"
                    >
                      <i class="mgc_code_line text-orange-500"></i>
                      <span class="font-semibold">HTML</span>
                    </button>
                    <button
                      @click.stop="copyImageLink(image, 'bbcode')"
                      class="w-full text-left px-2 py-1.5 text-[11px] text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded-lg transition-colors duration-200 flex items-center gap-1.5"
                    >
                      <i class="mgc_brackets_line text-purple-500"></i>
                      <span class="font-semibold">BB</span>
                    </button>
                  </div>
                </div>
              </div>
              <button
                @click.stop="downloadImage(image)"
                class="halo-button halo-button-primary h-8 w-8 flex items-center justify-center bg-white dark:bg-dark-200"
                title="下载图片"
              >
                <i class="mgc_download_2_fill text-sm"></i>
              </button>
              <button
                @click.stop="deleteImage(image.id)"
                class="halo-button text-danger h-8 w-8 flex items-center justify-center bg-white dark:bg-dark-200"
                title="删除图片"
              >
                <i class="mgc_delete_2_fill text-sm"></i>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 无图片状态 -->
      <div
        v-else
        class="no-images bg-white dark:bg-dark-200 rounded-xl shadow-md dark:shadow-dark-md p-8 text-center"
      >
        <div class="text-5xl text-light-300 dark:text-dark-100 mb-3">
          <i class="mgc_pic_line"></i>
        </div>
        <p class="text-secondary text-base mb-4">暂无上传的图片</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from "vue";
import AppSelect from "@/components/AppSelect.vue";
import { escapeHtml } from "@/utils/escapeHtml.js";

const svgDataUrl = (svg) =>
  `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;

// 获取完整URL的函数
const getFullUrl = (path) => {
  if (!path) return "";
  if (
    path.startsWith("http://") ||
    path.startsWith("https://") ||
    path.startsWith("//")
  ) {
    return path;
  }
  if (typeof window !== "undefined") {
    return window.location.origin + path;
  }
  return path;
};

const lifetimeOptions = [
  { value: "never", label: "永久保存" },
  { value: "1h", label: "1 小时" },
  { value: "1d", label: "1 天" },
  { value: "7d", label: "7 天" },
  { value: "30d", label: "30 天" },
  { value: "90d", label: "90 天" },
];
const webpStorageKey = "upload-save-webp";
const lifetimeStorageKey = "upload-lifetime";
const themeStorageKey = "theme-preference";
const storedWebp = localStorage.getItem(webpStorageKey);
const saveWebp = ref(storedWebp === null ? true : storedWebp === "true");
const expiresIn = ref(localStorage.getItem(lifetimeStorageKey) || "never");
const isDarkMode = ref(document.documentElement.classList.contains("dark"));
const currentUser = JSON.parse(localStorage.getItem("userInfo") || "{}");

const toggleTheme = () => {
  isDarkMode.value = !isDarkMode.value;
  document.documentElement.classList.toggle("dark", isDarkMode.value);
  localStorage.setItem(themeStorageKey, isDarkMode.value ? "dark" : "light");
};

const toggleSaveWebp = async () => {
  const previous = saveWebp.value;
  saveWebp.value = !previous;
  localStorage.setItem(webpStorageKey, String(saveWebp.value));
  if (currentUser.role !== 1 && currentUser.isAdmin !== true) return;
  try {
    const response = await fetch("/api/settings/update", {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${localStorage.getItem("authToken")}` },
      body: JSON.stringify({ key: "save_webp", value: saveWebp.value }),
    });
    const result = await response.json();
    if (saveWebp.value) {
      Message.success("WebP 转换已开启");
    } else {
      Message.error("WebP 转换已关闭");
    }
  } catch (error) {
    saveWebp.value = previous;
    localStorage.setItem(webpStorageKey, String(previous));
    Message.error(error.message);
  }
};

const loadUploadPreferences = async () => {
  try {
    const response = await fetch("/api/settings/login");
    const result = await response.json();
    if (response.ok && result.code === 200 && storedWebp === null) {
      saveWebp.value = result.data.save_webp !== false;
      localStorage.setItem(webpStorageKey, String(saveWebp.value));
    }
  } catch (error) {
    console.error("获取上传偏好失败:", error);
  }
};

// 响应式数据
const isDragOver = ref(false);
const isUploading = ref(false);
const uploadingCount = ref(0);
const uploadProgress = ref(0);
const recentImages = ref([]);
const isLoadingRecent = ref(true);
const fileInput = ref(null);

// URL上传相关
const uploadMode = ref("file"); // 'file' or 'url'
const urlInput = ref("");
const isUploadingUrl = ref(false);

// 下拉框控制变量
const activeCopyMenu = ref(null); // 卡片复制菜单
let currentPreviewImage = null; // 当前预览的图片
let previewModalInstance = null; // 预览弹窗实例（用于关闭控制）

// 批量选择相关
const batchMode = ref(false);
const selectedRecords = ref([]);

// 计算属性：是否全选
const isAllSelected = computed(() => {
  return (
    recentImages.value.length > 0 &&
    selectedRecords.value.length === recentImages.value.length
  );
});

// 批量选择操作
const enterBatchMode = () => {
  batchMode.value = true;
  selectedRecords.value = [];
};

const exitBatchMode = () => {
  batchMode.value = false;
  selectedRecords.value = [];
};

const toggleRecordSelect = (imageId) => {
  const index = selectedRecords.value.indexOf(imageId);
  if (index === -1) {
    selectedRecords.value.push(imageId);
  } else {
    selectedRecords.value.splice(index, 1);
  }
};

const isRecordSelected = (imageId) => {
  return selectedRecords.value.includes(imageId);
};

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedRecords.value = [];
  } else {
    selectedRecords.value = recentImages.value.map((img) => img.id);
  }
};

// 批量管理记录
const batchDeleteRecords = async () => {
  if (selectedRecords.value.length === 0) return;

  const modal = new PopupModal({
    title: "批量操作",
    content: `
      <div class="flex gap-3">
        <i class="mgc_question_line text-blue-500 text-xl mt-1"></i>
        <div>
          <p>请选择要对选中的 <strong>${selectedRecords.value.length}</strong> 条记录执行的操作：</p>
           <ul class="mt-2 text-sm text-secondary list-disc pl-4 space-y-1">
             <li><strong>仅移除记录</strong>：仅从"最近上传"列表中移除，图片仍保留在画廊和存储中。</li>
             <li><strong>彻底删除</strong>：彻底删除图片文件和数据库记录，无法恢复。</li>
           </ul>
        </div>
      </div>
    `,
    buttons: [
      {
        text: "取消",
        type: "default",
        callback: (modal) => modal.close(),
      },
      {
        text: "仅移除记录",
        type: "primary",
        callback: async (modal) => {
          modal.close();
          await executeBatchDismiss();
        },
      },
      {
        text: "彻底删除",
        type: "danger",
        callback: async (modal) => {
          modal.close();
          await executeBatchDelete();
        },
      },
    ],
    maskClose: true,
  });
  modal.open();
};

const executeBatchDismiss = async () => {
  const loading = Loading.show({
    text: `正在移除 ${selectedRecords.value.length} 条记录...`,
    color: "#1890ff",
    mask: true,
  });

  let successCount = 0;
  let failCount = 0;

  for (const imageId of selectedRecords.value) {
    try {
      const response = await fetch(`/api/images/${imageId}/recent`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      if (response.ok) {
        successCount++;
      } else {
        failCount++;
      }
    } catch (error) {
      console.error("移除记录错误:", error);
      failCount++;
    }
  }

  await loading.hide();

  if (failCount === 0) {
    Message.success(`成功移除 ${successCount} 条记录`);
  } else {
    Message.warning(`移除完成：成功 ${successCount} 条，失败 ${failCount} 条`);
  }

  exitBatchMode();
  loadRecentImages();
};

const executeBatchDelete = async () => {
  const loading = Loading.show({
    text: `正在删除 ${selectedRecords.value.length} 条记录...`,
    color: "#ff4d4f",
    mask: true,
  });

  let successCount = 0;
  let failCount = 0;

  for (const imageId of selectedRecords.value) {
    try {
      const response = await fetch(`/api/images/${imageId}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      if (response.ok) {
        successCount++;
      } else {
        failCount++;
      }
    } catch (error) {
      console.error("删除记录错误:", error);
      failCount++;
    }
  }

  await loading.hide();

  if (failCount === 0) {
    Message.success(`成功删除 ${successCount} 条记录`);
  } else {
    Message.warning(`删除完成：成功 ${successCount} 条，失败 ${failCount} 条`);
  }

  exitBatchMode();
  loadRecentImages();
};

// 批量复制链接
const batchCopyRecordLinks = async () => {
  if (selectedRecords.value.length === 0) return;

  const selectedImgs = recentImages.value.filter((img) =>
    selectedRecords.value.includes(img.id)
  );
  const urls = selectedImgs.map((img) => getFullUrl(img.url)).join("\n");

  try {
    await navigator.clipboard.writeText(urls);
    Message.success(`已复制 ${selectedImgs.length} 个链接到剪贴板`);
  } catch (error) {
    console.error("复制失败:", error);
    Message.error("复制失败");
  }
};

// 卡片复制菜单切换
const toggleCardCopyMenu = (imageId) => {
  if (activeCopyMenu.value === imageId) {
    activeCopyMenu.value = null;
  } else {
    activeCopyMenu.value = imageId;
  }
};

// 全局点击关闭下拉框
const handleGlobalClick = (e) => {
  if (activeCopyMenu.value !== null) {
    const cardCopyMenus = document.querySelectorAll(
      ".recent-item .relative.z-50"
    );
    let isClickInside = false;
    cardCopyMenus.forEach((menu) => {
      if (menu.contains(e.target)) {
        isClickInside = true;
      }
    });
    if (!isClickInside) {
      activeCopyMenu.value = null;
    }
  }
};

// 拖拽处理
const handleDragOver = (e) => {
  e.preventDefault();
  isDragOver.value = true;
};

const handleDragEnter = (e) => {
  e.preventDefault();
  isDragOver.value = true;
};

const handleDragLeave = (e) => {
  e.preventDefault();
  if (!e.currentTarget.contains(e.relatedTarget)) {
    isDragOver.value = false;
  }
};

const handleDrop = (e) => {
  e.preventDefault();
  isDragOver.value = false;

  const files = Array.from(e.dataTransfer.files);
  const imageFiles = files.filter(isImageFile);

  if (imageFiles.length > 0) {
    uploadFiles(imageFiles);
  } else {
    // 替换为 Message 错误提示
    Message.error("请拖拽图片文件", {
      duration: 3000,
      position: "top-right",
    });
  }
};

const isImageFile = (file) => {
  if (file.type.startsWith("image/")) return true;
  return /\.(?:heic|heif)$/i.test(file.name);
};

// 文件选择处理
const triggerFileInput = () => {
  if (!isUploading.value && fileInput.value) {
    fileInput.value.click();
  }
};

const handleFileSelect = (e) => {
  const files = Array.from(e.target.files);
  if (files.length > 0) {
    uploadFiles(files);
  }
  e.target.value = "";
};

// 剪贴板粘贴处理
const handlePaste = async (e) => {
  const items = e.clipboardData.items;
  const imageFiles = [];

  for (let item of items) {
    if (item.type.startsWith("image/")) {
      const file = item.getAsFile();
      if (file) {
        const timestamp = new Date().getTime();
        const extension = item.type.split("/")[1] || "png";
        const newFile = new File([file], `paste-${timestamp}.${extension}`, {
          type: item.type,
        });
        imageFiles.push(newFile);
      }
    }
  }

  if (imageFiles.length > 0) {
    e.preventDefault();
    uploadFiles(imageFiles);
    Message.success(`从剪贴板粘贴了 ${imageFiles.length} 个图片`, {
      duration: 2000,
      position: "top-right",
    });
  }
};

// 文件上传
const uploadFiles = async (files) => {
  if (isUploading.value) return;

  isUploading.value = true;
  uploadingCount.value = files.length;
  uploadProgress.value = 0;

  const formData = new FormData();
  files.forEach((file) => {
    formData.append("images[]", file);
  });
  formData.append("expires_in", expiresIn.value);
  formData.append("save_webp", String(saveWebp.value));
  localStorage.setItem(lifetimeStorageKey, expiresIn.value);

  try {
    const progressInterval = setInterval(() => {
      if (uploadProgress.value < 95) {
        uploadProgress.value += Math.random() * 5;
      }
    }, 150);

    const response = await fetch("/api/upload/images", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
      body: formData,
    });

    clearInterval(progressInterval);
    uploadProgress.value = 100;

    const result = await response.json();

    if (response.ok && result.code === 200) {
      await loadRecentImages();
      Message.success(`上传成功`, {
        duration: 2000,
        position: "top-right",
      });
    } else {
      throw new Error(result.message || "上传失败");
    }
  } catch (error) {
    console.error("上传错误:", error);
    Message.error(`上传失败: ${error.message}`, {
      duration: 3000,
      position: "top-right",
      showClose: true,
    });
  } finally {
    isUploading.value = false;
    uploadingCount.value = 0;
    uploadProgress.value = 0;
  }
};

// URL上传
const uploadByUrl = async () => {
  const url = urlInput.value.trim();
  if (!url) {
    Message.warning("请输入图片URL");
    return;
  }

  // 简单验证URL格式
  if (!url.startsWith("http://") && !url.startsWith("https://")) {
    Message.error("URL必须以 http:// 或 https:// 开头");
    return;
  }

  isUploadingUrl.value = true;
  localStorage.setItem(lifetimeStorageKey, expiresIn.value);

  try {
    const response = await fetch("/api/upload/url", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
      body: JSON.stringify({ url, expires_in: expiresIn.value, save_webp: saveWebp.value }),
    });

    const result = await response.json();

    if (response.ok && result.code === 200) {
      urlInput.value = "";
      await loadRecentImages();
      Message.success("URL图片上传成功", {
        duration: 2000,
        position: "top-right",
      });
    } else {
      throw new Error(result.message || "URL上传失败");
    }
  } catch (error) {
    console.error("URL上传错误:", error);
    Message.error(`上传失败: ${error.message}`, {
      duration: 3000,
      position: "top-right",
      showClose: true,
    });
  } finally {
    isUploadingUrl.value = false;
  }
};

// 加载最近上传的图片
const loadRecentImages = async () => {
  isLoadingRecent.value = true;
  try {
    const response = await fetch(
      "/api/images?limit=12&visibility=visible&recent=true",
      {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      }
    );

    if (response.ok) {
      const result = await response.json();
      recentImages.value = Array.isArray(result.data?.images)
        ? result.data.images
        : [];
    }
  } catch (error) {
    console.error("加载图片失败:", error);
    recentImages.value = [];
    Message.error(`加载图片失败: ${error.message}`, {
      duration: 3000,
      position: "top-right",
      showClose: true,
    });
  } finally {
    isLoadingRecent.value = false;
  }
};

// 多格式复制功能
const copyImageLink = async (image, type) => {
  if (!image) return;
  const fullUrl = getFullUrl(image.url);
  let copyText = "";

  switch (type) {
    case "url":
      copyText = fullUrl;
      break;
    case "html":
      copyText = `<img src="${escapeHtml(fullUrl)}" alt="${escapeHtml(image.filename)}" width="${
        image.width || ""
      }" height="${image.height || ""}">`;
      break;
    case "markdown":
      copyText = `![img](${fullUrl})`;
      break;
    case "bbcode":
      copyText = `[img]${fullUrl}[/img]`;
      break;
    default:
      copyText = fullUrl;
  }

  try {
    await navigator.clipboard.writeText(copyText);
    Message.success(`已复制${getTypeText(type)}格式`, {
      duration: 1500,
      position: "top-center",
      zIndex: 20000,
    });
  } catch (error) {
    const textArea = document.createElement("textarea");
    textArea.value = copyText;
    document.body.appendChild(textArea);
    textArea.select();
    document.execCommand("copy");
    document.body.removeChild(textArea);
    Message.success(`已复制${getTypeText(type)}格式`, {
      duration: 1500,
      position: "top-center",
      zIndex: 20000,
    });
  } finally {
    // 复制后强制关闭所有下拉框
    nextTick(() => {
      activeCopyMenu.value = null;
    });
  }
};

// 辅助函数：获取复制类型文本
const getTypeText = (type) => {
  switch (type) {
    case "url":
      return "URL";
    case "html":
      return "HTML";
    case "markdown":
      return "Markdown";
    case "bbcode":
      return "BBCode";
    default:
      return "";
  }
};

const formatDate = (dateString) => {
  if (!dateString) return "";
  const date = new Date(dateString);
  return date.toLocaleString("zh-CN");
};

const formatExpiration = (dateString) => {
  if (!dateString) return "永久保存";
  const milliseconds = new Date(dateString).getTime() - Date.now();
  if (milliseconds <= 0) return "已过期";
  const hours = Math.ceil(milliseconds / 3600000);
  if (hours < 24) return `${hours} 小时后过期`;
  const days = Math.ceil(hours / 24);
  return `${days} 天后过期`;
};

// 删除上传记录（仅删除记录，不删除存储文件）
// 删除/移除图片
const deleteImage = async (imageId) => {
  const modal = new PopupModal({
    title: "移除或是删除",
    content: `
      <div class="flex gap-3">
        <i class="mgc_question_line text-blue-500 text-xl mt-1"></i>
        <div>
          <p>请选择要执行的操作：</p>
           <ul class="mt-2 text-sm text-secondary list-disc pl-4 space-y-1">
             <li><strong>仅移除记录</strong>：仅从"最近上传"列表中移除，图片仍保留在画廊和存储中。</li>
             <li><strong>彻底删除</strong>：彻底删除图片文件和数据库记录，无法恢复。</li>
           </ul>
        </div>
      </div>
    `,
    buttons: [
      {
        text: "取消",
        type: "default",
        callback: (modal) => modal.close(),
      },
      {
        text: "仅移除记录",
        type: "primary",
        callback: async (modal) => {
          modal.close();
          await dismissAsync(imageId);
        },
      },
      {
        text: "彻底删除",
        type: "danger",
        callback: async (modal) => {
          modal.close();
          await deleteAsync(imageId);
        },
      },
    ],
    maskClose: true,
  });
  modal.open();
};

const dismissAsync = async (imageId) => {
  const loading = Loading.show({
    text: "移除中...",
    color: "#1890ff",
    mask: true,
  });
  try {
    const response = await fetch(`/api/images/${imageId}/recent`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
    });

    if (response.ok) {
      Message.success("已从最近上传中移除");
      // 如果删除的是当前预览的图片，关闭预览弹窗
      if (currentPreviewImage?.id === imageId && previewModalInstance) {
        previewModalInstance.close();
        currentPreviewImage = null;
        previewModalInstance = null;
      }
      await loadRecentImages();
    } else {
      const result = await response.json();
      throw new Error(result.message || "移除失败");
    }
  } catch (error) {
    console.error(error);
    Message.error(error.message);
  } finally {
    loading.hide();
  }
};

const deleteAsync = async (imageId) => {
  const loading = Loading.show({
    text: "删除中...",
    color: "#ff4d4f",
    mask: true,
  });
  try {
    // 彻底删除图片文件和数据库记录
    const response = await fetch(`/api/images/${imageId}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        "Content-Type": "application/json",
      },
    });

    if (response.ok) {
      Message.success("删除成功", {
        duration: 1500,
        position: "top-right",
      });
      // 如果删除的是当前预览的图片，关闭预览弹窗
      if (currentPreviewImage?.id === imageId && previewModalInstance) {
        previewModalInstance.close();
        currentPreviewImage = null;
        previewModalInstance = null;
      }
      activeCopyMenu.value = null;
      await loadRecentImages();
    } else {
      const result = await response.json();
      throw new Error(result.message || "删除失败");
    }
  } catch (error) {
    console.error("删除图片错误:", error);
    Message.error(`删除失败: ${error.message}`, {
      duration: 3000,
      position: "top-right",
      showClose: true,
    });
  } finally {
    await loading.hide();
  }
};

// 工具函数
const formatFileSize = (bytes) => {
  if (!bytes || bytes < 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

// 核心：图片预览
const previewImage = (image) => {
  if (!image || !image.url) {
    Message.error("图片信息不完整，无法预览", {
      duration: 2000,
      position: "top-right",
    });
    return;
  }

  currentPreviewImage = image;
  const safeFileName = escapeHtml(image.filename || "图片");
  const safeImageURL = escapeHtml(getFullUrl(image.url));

  // 预先生成错误占位图的 Base64
  const errorSvg = `
    <svg width='100%' height='100%' viewBox='0 0 200 150' fill='none' xmlns='http://www.w3.org/2000/svg' preserveAspectRatio='none'>
      <rect width='200' height='150' fill='%23f3f4f6'/>
      <path d='M92.5 67.5L80 80L60 60L30 90H170L125 45L92.5 77.5' stroke='%239ca3af' stroke-width='4' stroke-linecap='round' stroke-linejoin='round'/>
      <circle cx='140' cy='50' r='10' stroke='%239ca3af' stroke-width='4'/>
      <path d='M100 75L120 75' stroke='%239ca3af' stroke-width='4' stroke-linecap='round'/>
      <text x='100' y='120' font-family='Arial, sans-serif' font-size='14' fill='%239ca3af' text-anchor='middle'>加载失败</text>
    </svg>`;
  const errorBase64 = svgDataUrl(errorSvg);

  // 构建预览弹窗内容
  const previewContent = `
    <div class="image-preview-popup w-full max-w-4xl max-h-[85vh] flex flex-col overflow-hidden bg-white/90 dark:bg-dark-200/90 glass-card rounded-2xl p-2 sm:p-4">
      <!-- 顶部文件名栏：完整显示文件名 -->
      <div class="preview-filename-bar mb-2 pb-2 border-b border-light-200/80 dark:border-dark-100/80 px-1">
        <h3 class="text-xs sm:text-sm font-semibold text-gray-800 dark:text-gray-100 break-all leading-normal select-text">${safeFileName}</h3>
      </div>
      
      <!-- 操作按钮栏 -->
      <div class="preview-header pb-2 flex flex-col items-center gap-2 px-1">
        <div class="flex gap-1.5 flex-wrap items-center justify-center w-full mx-auto">
          <button class="px-2.5 py-1 text-xs rounded-lg bg-light-200/80 dark:bg-dark-300/80 text-secondary hover:text-primary hover:bg-light-100 dark:hover:bg-dark-200 transition-colors flex items-center gap-1 font-medium shrink-0" onclick="event.stopPropagation(); window.copyPreviewImageLink('url')">
            <i class="mgc_link_2_line text-primary"></i>
            <span>URL</span>
          </button>
          <button class="px-2.5 py-1 text-xs rounded-lg bg-light-200/80 dark:bg-dark-300/80 text-secondary hover:text-primary hover:bg-light-100 dark:hover:bg-dark-200 transition-colors flex items-center gap-1 font-medium shrink-0" onclick="event.stopPropagation(); window.copyPreviewImageLink('markdown')">
            <i class="mgc_markdown_line text-blue-500"></i>
            <span>MD</span>
          </button>
          <button class="px-2.5 py-1 text-xs rounded-lg bg-light-200/80 dark:bg-dark-300/80 text-secondary hover:text-primary hover:bg-light-100 dark:hover:bg-dark-200 transition-colors flex items-center gap-1 font-medium shrink-0" onclick="event.stopPropagation(); window.copyPreviewImageLink('html')">
            <i class="mgc_code_line text-orange-500"></i>
            <span>HTML</span>
          </button>
          <button class="px-2.5 py-1 text-xs rounded-lg bg-light-200/80 dark:bg-dark-300/80 text-secondary hover:text-primary hover:bg-light-100 dark:hover:bg-dark-200 transition-colors flex items-center gap-1 font-medium shrink-0" onclick="event.stopPropagation(); window.copyPreviewImageLink('bbcode')">
            <i class="mgc_brackets_line text-purple-500"></i>
            <span>BB</span>
          </button>
        </div>
        <div class="flex gap-2 items-center justify-center w-full">
          <button
            class="halo-button halo-button-primary px-3 py-1 text-xs font-semibold flex items-center gap-1"
            onclick="event.stopPropagation(); window.downloadPreviewImage()"
          >
            <i class="mgc_download_2_fill text-xs"></i>
            下载
          </button>
          <button
            class="halo-button halo-button-danger px-3 py-1 text-xs font-semibold flex items-center gap-1"
            onclick="event.stopPropagation(); window.deletePreviewImage()"
          >
            <i class="mgc_delete_2_fill text-xs"></i>
            删除
          </button>
        </div>
      </div>
      
      <!-- 预览图片区域 -->
      <div class="max-h-[360px] flex-1 overflow-auto flex items-center justify-center">
        <a 
            class="spotlight min-w-full max-w-full min-h-[260px] block" 
            href="${safeImageURL}"
            data-description="尺寸: ${image.width || "未知"}×${
    image.height || "未知"
  } | 大小: ${formatFileSize(image.file_size || 0)} | 上传日期：${formatDate(
    image.created_at
  )}"
        >
            <div class="relative max-w-full w-fill max-h-[360px] min-h-[260px] rounded-lg overflow-hidden image-skeleton flex items-center justify-center">
                <img 
                    src="${safeImageURL}"
                    alt="${safeFileName}"
                    class="max-w-full w-fill max-h-[360px] min-h-[260px] object-contain rounded-lg relative z-10 opacity-0"
                    onload="this.classList.add('image-fade-in'); this.classList.remove('opacity-0'); this.parentElement.classList.remove('image-skeleton')"
                    onerror="this.parentElement.classList.remove('image-skeleton'); this.classList.remove('opacity-0'); this.src='${errorBase64}'; this.classList.add('object-contain', 'p-4', 'bg-gray-50', 'dark:bg-gray-800');"
                />
            </div>
        </a>
      </div>
      
      <!-- 底部信息栏 -->
      <!-- 底部信息栏 -->
      <div class="pt-2 flex flex-wrap gap-2 text-xs text-secondary ml-1 px-1">
        <div class="flex items-center gap-1.5">
          <i class="mgc_ruler_line w-3.5 text-center"></i>
          尺寸: ${image.width || "未知"}×${image.height || "未知"}
        </div>
        <div class="flex items-center gap-1.5">
          <i class="mgc_pic_line w-3.5 text-center"></i>
          大小: ${formatFileSize(image.file_size || 0)}
        </div>
        <div class="flex items-center gap-1.5">
          <i class="mgc_storage_line w-3.5 text-center"></i>
          存储: ${
            image.storage === "telegram"
              ? "Telegram"
              : (image.storage === "default" ? "本地" : image.storage) || "未知"
          }
        </div>
        <div class="flex items-center gap-1.5">
          <i class="mgc_calendar_line w-3.5 text-center"></i>
          上传日期：${formatDate(image.created_at)}
        </div>
      </div>
    </div>
  `;

  // 全局注册预览相关函数（供弹窗内 DOM 调用）
  window.copyPreviewImageLink = (type) => {
    copyImageLink(currentPreviewImage, type);
  };

  window.downloadPreviewImage = () => {
    downloadImage(currentPreviewImage);
  };

  window.deletePreviewImage = () => {
    deleteImage(currentPreviewImage.id);
    closePreviewModal();
  };

  window.closePreviewModal = () => {
    if (previewModalInstance) {
      previewModalInstance.close();
      currentPreviewImage = null;
      previewModalInstance = null;
    }
  };

  // 创建预览弹窗实例
  previewModalInstance = new PopupModal({
    title: "图片预览",
    content: previewContent,
    type: "default",
    buttons: [
      {
        text: "确定",
        type: "default",
        callback: (modal) => modal.close(),
      },
    ],
    maskClose: true,
    zIndex: 10000,
    maxHeight: "90vh",
    onClose: () => {
      window.copyPreviewImageLink = null;
      window.downloadPreviewImage = null;
      window.deletePreviewImage = null;
      window.closePreviewModal = null;
      currentPreviewImage = null;
      previewModalInstance = null;
    },
  });

  // 打开弹窗
  previewModalInstance.open();

  // 处理弹窗内点击事件
  nextTick(() => {
    const previewContent = document.querySelector(".image-preview-popup");
    if (previewContent) {
      previewContent.addEventListener("click", (e) => {
        e.stopPropagation();
      });
    }
  });
};

const downloadImage = (image) => {
  if (!image || !image.url) {
    Message.error("图片信息不完整，无法下载", {
      duration: 2000,
      position: "top-right",
    });
    return;
  }
  const fullUrl = getFullUrl(image.url);
  const link = document.createElement("a");
  link.href = fullUrl;
  link.download = image.filename || `image-${Date.now()}.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  Message.info("开始下载图片", {
    duration: 1500,
    position: "top-right",
  });
  activeCopyMenu.value = null;
};

const handleImageError = (event) => {
  // 更加美观的占位图（SVG Base64）
  // 包含一个简单的图标和灰色背景
  const svg = `
  <svg width="100%" height="100%" viewBox="0 0 200 150" fill="none" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="none">
    <rect width="200" height="150" fill="#f3f4f6"/>
    <path d="M92.5 67.5L80 80L60 60L30 90H170L125 45L92.5 77.5" stroke="#9ca3af" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
    <circle cx="140" cy="50" r="10" stroke="#9ca3af" stroke-width="4"/>
    <path d="M100 75L120 75" stroke="#9ca3af" stroke-width="4" stroke-linecap="round"/>
    <text x="100" y="120" font-family="Arial, sans-serif" font-size="14" fill="#9ca3af" text-anchor="middle">加载失败</text>
  </svg>`;

  event.target.src = svgDataUrl(svg);
  event.target.classList.add(
    "object-contain",
    "p-4",
    "bg-gray-50",
    "dark:bg-gray-800"
  );
};

// 生命周期
onMounted(() => {
  loadUploadPreferences();
  document.addEventListener("paste", handlePaste);
  document.addEventListener("click", handleGlobalClick);
  setTimeout(() => {
    loadRecentImages();
  }, 100);
});

onUnmounted(() => {
  document.removeEventListener("paste", handlePaste);
  document.removeEventListener("click", handleGlobalClick);
  window.copyPreviewImageLink = null;
  window.downloadPreviewImage = null;
  window.deletePreviewImage = null;
  window.closePreviewModal = null;
  // 关闭预览弹窗
  if (previewModalInstance) {
    previewModalInstance.close();
  }
  // 关闭所有通知
  if (window.onmessage) {
    Message.closeAll();
  }
});
</script>

<style scoped>
.home-page { width:min(1180px,100%); margin:0 auto; padding:1rem 0 4rem; color:#303a44; }
.dark .home-page { color:#edf1f4; }
.home-tools { display:flex; justify-content:flex-end; margin-bottom:.8rem; }
.quick-switches { display:flex; justify-content:flex-end; gap:.5rem; }
.quick-toggle { min-width:5.25rem; height:3.15rem; display:flex; align-items:center; gap:.48rem; padding:.42rem .65rem; border:1px solid rgba(210,119,141,.15); border-radius:.9rem; color:#6f7882; background:rgba(255,255,255,.78); box-shadow:0 10px 28px rgba(77,46,55,.06); transition:.2s ease; }
.quick-toggle:hover { transform:translateY(-1px); border-color:rgba(210,98,124,.35); }
.quick-toggle .quick-icon { flex:0 0 auto; display:grid; place-items:center; width:2rem; height:2rem; border-radius:.65rem; color:#c75a73; background:#fff0f3; font-size:1.05rem; }
.quick-copy { display:flex; min-width:0; flex-direction:column; align-items:flex-start; gap:.16rem; }.quick-toggle b { text-align:left; font-size:.71rem; line-height:1; }.quick-toggle em { color:#aaa1a4; text-align:left; font-size:.58rem; line-height:1; font-style:normal; }
.quick-toggle.active em { color:#4c917b; }.dark .quick-toggle{color:#d7dce1;background:rgba(31,36,44,.86);border-color:rgba(255,255,255,.065);box-shadow:0 12px 30px rgba(3,8,13,.25)}.dark .quick-toggle .quick-icon{color:#ff9db1;background:#3b3037}
.upload-panel { padding:1.15rem; border:1px solid rgba(210,119,141,.14); border-radius:1.35rem; background:rgba(255,255,255,.82); box-shadow:0 20px 60px rgba(76,47,56,.075); backdrop-filter:blur(18px); transition:.25s ease; }
.dark .upload-panel { background:rgba(30,35,43,.9); border-color:rgba(255,255,255,.065); box-shadow:0 22px 58px rgba(3,8,13,.32); }
.upload-toolbar { display:flex; align-items:center; justify-content:space-between; gap:1rem; margin-bottom:1rem; }
.lifetime-select { display:flex; align-items:center; gap:.65rem; color:#7e8791; font-size:.73rem; font-weight:700; }.lifetime-select>span{display:flex;align-items:center;gap:.35rem;white-space:nowrap}.lifetime-select>span i{color:#ce647b;font-size:1rem}.lifetime-select :deep(.app-select){width:8.15rem}
.mode-switch { display:flex; align-items:center; gap:.2rem; padding:.25rem; border-radius:.75rem; background:#f5eff0; }.dark .mode-switch{background:#252b33}.mode-switch button{display:inline-flex;min-height:2.25rem;align-items:center;justify-content:center;gap:.42rem;line-height:1}.mode-switch button i{display:grid;place-items:center;margin:0!important;font-size:1rem;line-height:1}.mode-switch button span{line-height:1}
.upload-area { min-height:18.5rem; border-color:#eadde0; background:radial-gradient(circle at 50% 22%,rgba(232,129,153,.08),transparent 28%),rgba(255,252,252,.45)!important; }.dark .upload-area{border-color:#414955;background:radial-gradient(circle at 50% 22%,rgba(232,129,153,.1),transparent 28%),rgba(29,34,41,.3)!important}
.upload-arrow{color:#d2657d;font-size:2.7rem;line-height:1;transition:transform .25s ease,color .25s ease}.upload-area:hover .upload-arrow{color:#bd4e68;transform:translateY(-3px)}.dark .upload-arrow{color:#f095a9}
.recent-section { margin-top:1.7rem; }.recent-item { border-color:rgba(210,119,141,.12)!important; box-shadow:0 12px 34px rgba(76,47,56,.055); }.dark .recent-item{box-shadow:0 14px 35px rgba(3,8,13,.25)}
.expiry-badge { position:absolute; z-index:8; top:.55rem; left:.55rem; display:flex; align-items:center; gap:.28rem; padding:.32rem .48rem; border:1px solid rgba(255,255,255,.5); border-radius:.55rem; color:#fff; background:rgba(55,46,49,.58); backdrop-filter:blur(10px); font-size:.62rem; font-weight:700; }
@media(max-width:700px){.quick-switches{align-self:flex-end;width:auto;gap:.35rem}.quick-toggle{min-width:4.15rem;height:2.45rem;gap:.34rem;padding:.28rem .42rem;border-radius:.72rem;box-shadow:0 7px 20px rgba(77,46,55,.055)}.quick-toggle .quick-icon{width:1.7rem;height:1.7rem;border-radius:.52rem;font-size:.9rem}.quick-toggle b{font-size:.65rem}.quick-toggle em{font-size:.53rem}.upload-toolbar{align-items:stretch;flex-direction:column}.lifetime-select{justify-content:space-between}.lifetime-select :deep(.app-select){width:min(8.15rem,48vw)}.mode-switch{align-self:stretch}.mode-switch button{flex:1}.upload-panel{padding:.85rem;border-radius:1rem}.upload-area{min-height:16rem}}
</style>
