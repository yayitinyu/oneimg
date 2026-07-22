<template>
  <div class="gallery-page text-gray-800 dark:text-gray-200">
    <!-- 主要内容 -->
    <div class="gallery-content container mx-auto px-2 md:px-4 py-8 md:py-12">
      <header class="gallery-heading">
        <div>
          <p class="eyebrow">Visual archive</p>
          <h1>{{ isAdmin ? "全站画廊" : "我的画廊" }}</h1>
          <p>{{ gallerySubtitle }}</p>
        </div>
        <div class="gallery-count" aria-live="polite">
          <i class="mgc_pic_2_line"></i>
          <span><strong>{{ totalCount }}</strong> 张图片</span>
        </div>
      </header>

      <!-- 顶部筛选栏 -->
      <section class="gallery-toolbar glass-card" aria-label="画廊筛选与视图">
        <form class="gallery-search" role="search" @submit.prevent="submitSearch">
          <i class="mgc_search_2_line"></i>
          <input
            v-model="searchInput"
            type="search"
            autocomplete="off"
            placeholder="搜索文件名"
            aria-label="搜索图片文件名"
          />
          <button v-if="searchInput" type="button" title="清空搜索" @click="clearSearch">
            <i class="mgc_close_circle_fill"></i>
          </button>
        </form>

        <div v-if="isAdmin" class="owner-filter" aria-label="图片归属">
          <button
            v-for="option in ownerOptions"
            :key="option.value"
            type="button"
            :class="{ active: ownerScope === option.value }"
            @click="changeOwner(option.value)"
          >
            {{ option.label }}
          </button>
        </div>

        <!-- 视图切换 -->
        <div class="view-toggle flex items-center gap-1">
          <button
            type="button"
            @click="viewMode = 'grid'"
            class="toolbar-icon"
            :class="{ active: viewMode === 'grid' }"
            title="网格视图"
          >
            <i class="mgc_grid_fill"></i>
          </button>
          <button
            type="button"
            @click="viewMode = 'masonry'"
            class="toolbar-icon"
            :class="{ active: viewMode === 'masonry' }"
            title="瀑布流视图"
          >
            <i class="mgc_layout_grid_line"></i>
          </button>

          <!-- 批量管理按钮 -->
          <button
            v-if="canManageImages && !batchMode"
            type="button"
            @click="enterBatchMode"
            class="batch-entry"
          >
            <i class="mgc_checkbox_line"></i>
            <span class="hidden sm:inline">批量管理</span>
            <span class="sm:hidden">批量</span>
          </button>
          <span v-else-if="batchMode" class="batch-active">批量模式</span>
        </div>
      </section>

      <!-- 加载状态 -->
      <div
        v-if="loading"
        class="loading-container flex flex-col items-center justify-center py-20"
      >
        <div
          class="spinner w-10 h-10 border-4 border-gray-200 dark:border-gray-700 border-t-primary dark:border-t-primary rounded-full animate-spin mb-4"
        ></div>
        <p class="text-gray-600 dark:text-gray-400">加载中...</p>
      </div>

      <!-- 图片网格/列表 -->
      <div v-else-if="images.length > 0" class="images-container">
        <!-- 网格视图 -->
        <div
          v-if="viewMode === 'grid'"
          class="images-grid grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4"
        >
          <div
            v-for="image in images"
            :key="image.id"
            class="image-card bg-white/80 dark:bg-gray-800/80 glass-card rounded-2xl shadow-md overflow-hidden hover:shadow-xl transition-all duration-300 cursor-pointer border border-white/50 dark:border-gray-700/60 relative"
            :class="{
              'ring-2 ring-primary': batchMode && isSelected(image.id),
            }"
            @click="batchMode ? toggleSelect(image.id) : openPreview(image)"
          >
            <!-- 批量选择复选框 -->
            <div
              v-if="batchMode"
              class="absolute top-2 right-2 z-10"
              @click.stop="toggleSelect(image.id)"
            >
              <div
                class="w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all"
                :class="
                  isSelected(image.id)
                    ? 'bg-primary border-primary text-white'
                    : 'bg-white/90 dark:bg-gray-800/90 border-gray-300 dark:border-gray-600'
                "
              >
                <i
                  v-if="isSelected(image.id)"
                  class="mgc_check_line text-sm"
                ></i>
              </div>
            </div>
            <div
              class="image-wrapper relative aspect-video overflow-hidden bg-gray-100 dark:bg-gray-900"
            >
              <p
                v-if="isAdmin"
                class="owner-badge"
                :class="ownerClass(image.owner_type)"
                :title="image.owner_name"
              >
                {{ ownerLabel(image) }}
              </p>
              <div
                class="loading absolute inset-0 flex items-center justify-center z-0 text-slate-300"
              >
                <svg
                  class="w-8 h-8 animate-spin"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  style="transform: scaleX(-1) scaleY(-1)"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </div>
              <img
                :src="getFullUrl(image.thumbnail || image.url)"
                :alt="image.filename"
                class="image-thumbnail w-full h-full object-cover transition-transform duration-500 hover:scale-105 opacity-0"
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
                @error="
                  (e) => {
                    const svg = `
                                    <svg width='100%' height='100%' viewBox='0 0 200 150' fill='none' xmlns='http://www.w3.org/2000/svg' preserveAspectRatio='none'>
                                      <rect width='200' height='150' fill='%23f3f4f6'/>
                                      <path d='M92.5 67.5L80 80L60 60L30 90H170L125 45L92.5 77.5' stroke='%239ca3af' stroke-width='4' stroke-linecap='round' stroke-linejoin='round'/>
                                      <circle cx='140' cy='50' r='10' stroke='%239ca3af' stroke-width='4'/>
                                      <path d='M100 75L120 75' stroke='%239ca3af' stroke-width='4' stroke-linecap='round'/>
                                      <text x='100' y='120' font-family='Arial, sans-serif' font-size='14' fill='%239ca3af' text-anchor='middle'>加载失败</text>
                                    </svg>`;
                    e.target.src = `data:image/svg+xml;base64,${btoa(
                      unescape(encodeURIComponent(svg))
                    )}`;
                    e.target.classList.add(
                      'object-contain',
                      'p-4',
                      'bg-gray-50',
                      'dark:bg-gray-800'
                    );
                    e.target.parentElement
                      .querySelector('.loading')
                      .classList.add('hidden');
                    e.target.classList.remove('opacity-0');
                  }
                "
              />
              <span v-if="image.expires_at" class="card-expiry">
                <i class="mgc_time_line"></i>{{ formatExpiration(image.expires_at) }}
              </span>
            </div>
            <div class="image-info p-3">
              <p
                class="image-filename font-medium text-sm truncate whitespace-nowrap overflow-hidden"
              >
                {{ image.filename }}
              </p>
              <p
                class="image-meta text-xs text-gray-500 dark:text-gray-400 mt-1"
              >
                {{ formatFileSize(image.file_size) }} • {{ image.width }}×{{
                  image.height
                }}
              </p>
              <p
                class="image-date text-xs text-gray-500 dark:text-gray-400 mt-1 truncate"
              >
                {{ formatDate(image.created_at) }} · {{ formatStorageType(image.storage) }}
              </p>
            </div>
          </div>
        </div>

        <!-- 瀑布流视图 -->
        <div
          v-else-if="viewMode === 'masonry'"
          class="columns-2 sm:columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4"
        >
          <div
            v-for="image in images"
            :key="image.id"
            class="masonry-card break-inside-avoid overflow-hidden rounded-2xl shadow-md hover:shadow-xl transition-all duration-300 cursor-pointer relative"
            :class="{
              'ring-2 ring-primary': batchMode && isSelected(image.id),
            }"
            @click="batchMode ? toggleSelect(image.id) : openPreview(image)"
          >
            <!-- 批量选择复选框 -->
            <div
              v-if="batchMode"
              class="absolute top-2 right-2 z-10"
              @click.stop="toggleSelect(image.id)"
            >
              <div
                class="w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all"
                :class="
                  isSelected(image.id)
                    ? 'bg-primary border-primary text-white'
                    : 'bg-white/90 dark:bg-gray-800/90 border-gray-300 dark:border-gray-600'
                "
              >
                <i
                  v-if="isSelected(image.id)"
                  class="mgc_check_line text-sm"
                ></i>
              </div>
            </div>
            <div
              class="relative overflow-hidden bg-gray-100 dark:bg-gray-900 rounded-2xl"
            >
              <p
                v-if="isAdmin"
                class="owner-badge"
                :class="ownerClass(image.owner_type)"
                :title="image.owner_name"
              >
                {{ ownerLabel(image) }}
              </p>
              <div
                class="loading absolute inset-0 flex items-center justify-center z-0 text-slate-300"
              >
                <svg
                  class="w-8 h-8 animate-spin"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  style="transform: scaleX(-1) scaleY(-1)"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
              </div>
              <img
                :src="getFullUrl(image.thumbnail || image.url)"
                :alt="image.filename"
                class="w-full h-auto object-cover opacity-0 transition-all duration-500"
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
                @error="
                  (e) => {
                    const svg = `
                                    <svg width='100%' height='100%' viewBox='0 0 200 150' fill='none' xmlns='http://www.w3.org/2000/svg' preserveAspectRatio='none'>
                                      <rect width='200' height='150' fill='%23f3f4f6'/>
                                      <path d='M92.5 67.5L80 80L60 60L30 90H170L125 45L92.5 77.5' stroke='%239ca3af' stroke-width='4' stroke-linecap='round' stroke-linejoin='round'/>
                                      <circle cx='140' cy='50' r='10' stroke='%239ca3af' stroke-width='4'/>
                                      <path d='M100 75L120 75' stroke='%239ca3af' stroke-width='4' stroke-linecap='round'/>
                                      <text x='100' y='120' font-family='Arial, sans-serif' font-size='14' fill='%239ca3af' text-anchor='middle'>加载失败</text>
                                    </svg>`;
                    e.target.src = `data:image/svg+xml;base64,${btoa(
                      unescape(encodeURIComponent(svg))
                    )}`;
                    e.target.classList.add(
                      'object-contain',
                      'p-4',
                      'bg-gray-50',
                      'dark:bg-gray-800'
                    );
                    e.target.parentElement
                      .querySelector('.loading')
                      .classList.add('hidden');
                    e.target.classList.remove('opacity-0');
                  }
                "
              />
              <span v-if="image.expires_at" class="card-expiry">
                <i class="mgc_time_line"></i>{{ formatExpiration(image.expires_at) }}
              </span>
            </div>
          </div>
        </div>

        <!-- 瀑布流无限滚动加载触发器 -->
        <div
          v-if="viewMode === 'masonry' && !loading && hasMore"
          ref="loadMoreTrigger"
          class="w-full h-20 flex items-center justify-center"
        >
          <div
            v-if="isLoadingMore"
            class="flex items-center gap-2 text-gray-500"
          >
            <i class="mgc_loading_line animate-spin"></i>
            <span>加载中...</span>
          </div>
          <div v-else class="text-gray-400 text-sm">向下滚动加载更多</div>
        </div>

        <div
          v-if="viewMode === 'masonry' && !hasMore && images.length > 0"
          class="w-full py-8 text-center text-gray-400 text-sm"
        >
          已加载全部图片
        </div>

        <!-- 分页 (仅网格视图) -->
        <div
          v-if="viewMode === 'grid' && totalPages > 1"
          class="pagination flex flex-wrap items-center justify-center gap-2 py-8"
        >
          <button
            @click="changePage(currentPage - 1)"
            :disabled="currentPage <= 1"
            class="page-btn px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 transition-all text-sm"
            :class="{ 'opacity-50 cursor-not-allowed': currentPage <= 1 }"
          >
            上一页
          </button>

          <div class="page-numbers flex gap-1">
            <button
              v-for="page in visiblePages"
              :key="page"
              @click="changePage(page)"
              class="w-9 h-9 flex items-center justify-center rounded-lg border transition-all text-sm"
              :class="[
                page === currentPage
                  ? 'bg-primary text-white border-primary'
                  : 'border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700',
              ]"
            >
              {{ page }}
            </button>
          </div>

          <button
            @click="changePage(currentPage + 1)"
            :disabled="currentPage >= totalPages"
            class="page-btn px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 transition-all text-sm"
            :class="{
              'opacity-50 cursor-not-allowed': currentPage >= totalPages,
            }"
          >
            下一页
          </button>
        </div>
      </div>

      <!-- 空状态 -->
      <div
        v-else
        class="empty-state flex flex-col items-center justify-center py-20 text-center"
      >
        <div class="empty-icon text-6xl mb-4 text-gray-400 dark:text-gray-600">
          <i class="mgc_pic_ai_line"></i>
        </div>
        <h3 class="text-xl font-bold mb-2">{{ emptyTitle }}</h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ activeSearch ? "试试其他关键词，或清空搜索条件。" : "这里还没有图片，" }}
          <router-link to="/" class="text-primary hover:underline"
            v-if="!activeSearch">去上传一些吧</router-link
          >
        </p>
      </div>
    </div>

    <!-- 批量操作悬浮菜单 -->
    <Transition name="float-menu">
      <div
        v-if="batchMode"
        class="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3"
      >
        <!-- 已选计数 -->
        <div
          class="floating-menu-badge bg-white dark:bg-gray-800 px-4 py-2 rounded-full shadow-lg text-sm font-medium text-center"
        >
          已选 {{ selectedImages.length }} 项
        </div>

        <!-- 操作按钮组 -->
        <div class="floating-menu-buttons flex flex-col items-end gap-2">
          <button
            @click="toggleSelectAll"
            class="floating-btn halo-button w-12 h-12 rounded-full flex items-center justify-center text-lg"
            :title="isAllSelected ? '取消全选' : '全选'"
          >
            <i
              :class="
                isAllSelected
                  ? 'mgc_minimize_line'
                  : 'mgc_checkbox_line'
              "
            ></i>
          </button>
          <button
            @click="batchCopyLinks"
            :disabled="selectedImages.length === 0"
            class="floating-btn halo-button halo-button-primary w-12 h-12 rounded-full flex items-center justify-center text-lg disabled:opacity-50"
            title="复制链接"
          >
            <i class="mgc_link_2_line"></i>
          </button>
          <button
            @click="batchDeleteImages"
            :disabled="selectedImages.length === 0"
            class="floating-btn halo-button halo-button-danger w-12 h-12 rounded-full flex items-center justify-center text-lg disabled:opacity-50"
            title="删除"
          >
            <i class="mgc_delete_2_line"></i>
          </button>
          <button
            @click="exitBatchMode"
            class="floating-btn halo-button w-12 h-12 rounded-full flex items-center justify-center text-lg"
            title="取消"
          >
            <i class="mgc_close_line"></i>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onUnmounted, nextTick, watch } from "vue";
import { useRouter } from "vue-router";
import { escapeHtml } from "@/utils/escapeHtml.js";

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

// 格式化存储类型显示（首字母大写）
const formatStorageType = (storage) => {
  if (!storage) return "未知";
  if (storage === "default") return "本地";
  // 首字母大写
  return storage.charAt(0).toUpperCase() + storage.slice(1);
};

// 响应式数据（仅保留必要项）
const images = ref([]);
const loading = ref(false);
const viewMode = ref("grid");
const currentPage = ref(1);
const totalPages = ref(1);
const totalCount = ref(0);
const pageSize = ref(20);
const userInfo = JSON.parse(localStorage.getItem("userInfo") || "{}");
const isAdmin = ref(userInfo.role === 1 || userInfo.isAdmin === true);
const isGuest = ref(userInfo.role === 3 || userInfo.isTourist === true);
const canManageImages = computed(() => !isGuest.value);
const ownerScope = ref(isAdmin.value ? "all" : "mine");
const searchInput = ref("");
const activeSearch = ref("");
const ownerOptions = [
  { value: "all", label: "全部" },
  { value: "mine", label: "我的" },
  { value: "admins", label: "管理员" },
  { value: "users", label: "普通用户" },
  { value: "guests", label: "游客" },
];

const gallerySubtitle = computed(() => {
  if (!isAdmin.value) return "只显示你上传的图片，其他账号无法查看或管理。";
  const current = ownerOptions.find((option) => option.value === ownerScope.value);
  return `正在浏览${current?.label || "全部"}图片，可搜索、预览或批量整理。`;
});

const emptyTitle = computed(() => {
  if (activeSearch.value) return `没有找到“${activeSearch.value}”`;
  if (!isAdmin.value) return "你的画廊还是空的";
  const current = ownerOptions.find((option) => option.value === ownerScope.value);
  return `${current?.label || "当前范围"}暂无图片`;
});

// 无限滚动相关
const hasMore = ref(true);
const isLoadingMore = ref(false);
const loadMoreTrigger = ref(null);
let intersectionObserver = null;

// 批量选择相关
const batchMode = ref(false);
const selectedImages = ref([]);

// 计算属性：是否全选
const isAllSelected = computed(() => {
  return (
    images.value.length > 0 &&
    selectedImages.value.length === images.value.length
  );
});

// 当前预览的图片
const currentPreviewImage = ref(null);

// 批量选择操作
const enterBatchMode = () => {
  batchMode.value = true;
  selectedImages.value = [];
};

const exitBatchMode = () => {
  batchMode.value = false;
  selectedImages.value = [];
};

const toggleSelect = (imageId) => {
  const index = selectedImages.value.indexOf(imageId);
  if (index === -1) {
    selectedImages.value.push(imageId);
  } else {
    selectedImages.value.splice(index, 1);
  }
};

const isSelected = (imageId) => {
  return selectedImages.value.includes(imageId);
};

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedImages.value = [];
  } else {
    selectedImages.value = images.value.map((img) => img.id);
  }
};

// 批量删除图片
const batchDeleteImages = async () => {
  if (selectedImages.value.length === 0) return;

  const modal = new PopupModal({
    title: "确认批量删除",
    content: `
            <div class="flex gap-3">
                <i class="mgc_warning_line text-red-500 text-xl mt-1"></i>
                <div>
                    <p>确定要删除选中的 <strong>${selectedImages.value.length}</strong> 张图片吗？</p>
                    <p class="mt-1 text-secondary text-sm">图片将从存储中永久删除，无法恢复</p>
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
        text: "确认删除",
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

const executeBatchDelete = async () => {
  const loading = Loading.show({
    text: `正在删除 ${selectedImages.value.length} 张图片...`,
    color: "#ff4d4f",
    mask: true,
  });

  let successCount = 0;
  let failCount = 0;

  for (const imageId of selectedImages.value) {
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
      console.error("删除图片错误:", error);
      failCount++;
    }
  }

  await loading.hide();

  if (failCount === 0) {
    Message.success(`成功删除 ${successCount} 张图片`);
  } else {
    Message.warning(`删除完成：成功 ${successCount} 张，失败 ${failCount} 张`);
  }

  exitBatchMode();
  loadImages();
};

// 批量复制链接
const batchCopyLinks = async () => {
  if (selectedImages.value.length === 0) return;

  const selectedImgs = images.value.filter((img) =>
    selectedImages.value.includes(img.id)
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

// 计算属性（分页显示）
const visiblePages = computed(() => {
  const pages = [];
  const start = Math.max(1, currentPage.value - 2);
  const end = Math.min(totalPages.value, currentPage.value + 2);

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});

// 路由实例
const router = useRouter();

// 切换图片归属
const changeOwner = (owner) => {
  if (ownerScope.value !== owner) {
    ownerScope.value = owner;
    currentPage.value = 1;
    exitBatchMode();
    loadImages();
  }
};

const submitSearch = () => {
  const nextSearch = searchInput.value.trim();
  if (activeSearch.value === nextSearch && currentPage.value === 1) return;
  activeSearch.value = nextSearch;
  currentPage.value = 1;
  exitBatchMode();
  loadImages();
};

const clearSearch = () => {
  searchInput.value = "";
  submitSearch();
};

let searchTimer = null;
watch(searchInput, () => {
  window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(submitSearch, 350);
});

let loadRequestId = 0;

// 加载图片列表（核心功能）
const loadImages = async () => {
  const requestId = ++loadRequestId;
  loading.value = true;

  try {
    const params = new URLSearchParams({
      page: currentPage.value,
      limit: pageSize.value,
      sort_by: "created_at", // 固定默认排序
      sort_order: "desc",
    });
    if (isAdmin.value) params.set("owner", ownerScope.value);
    if (activeSearch.value) params.set("search", activeSearch.value);

    const response = await fetch(`/api/images?${params}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
    });

    if (response.ok) {
      const result = await response.json();
      if (requestId !== loadRequestId) return;
      images.value = result.data.images || [];
      totalPages.value = result.data.total_pages || 1;
      totalCount.value = result.data.total || 0;
      // 更新 hasMore 状态
      hasMore.value = currentPage.value < totalPages.value;
    } else {
      // 未授权跳转登录页
      if (response.status === 401) {
        localStorage.removeItem("authToken");
        router.push("/login");
        Message.error("登录已过期，请重新登录");
        return;
      }
      throw new Error("加载图片失败");
    }
  } catch (error) {
    if (requestId !== loadRequestId) return;
    console.error("加载图片错误:", error);
    Message.error("加载图片失败: " + error.message);
  } finally {
    if (requestId !== loadRequestId) return;
    loading.value = false;
    // 重新设置无限滚动观察器
    if (viewMode.value === "masonry") {
      nextTick(() => {
        setupInfiniteScroll();
      });
    }
  }
};

// 分页处理
const changePage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
    loadImages();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
};

// 加载更多图片（瀑布流无限滚动）
const loadMoreImages = async () => {
  if (isLoadingMore.value || !hasMore.value || viewMode.value !== "masonry")
    return;

  isLoadingMore.value = true;

  try {
    const nextPage = currentPage.value + 1;
    const params = new URLSearchParams({
      page: nextPage,
      limit: pageSize.value,
      sort_by: "created_at",
      sort_order: "desc",
    });
    if (isAdmin.value) params.set("owner", ownerScope.value);
    if (activeSearch.value) params.set("search", activeSearch.value);

    const response = await fetch(`/api/images?${params}`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
    });

    if (response.ok) {
      const result = await response.json();
      const newImages = result.data.images || [];

      if (newImages.length > 0) {
        images.value = [...images.value, ...newImages];
        currentPage.value = nextPage;
        totalPages.value = result.data.total_pages || 1;
      }

      // 检查是否还有更多
      hasMore.value = nextPage < totalPages.value;
    }
  } catch (error) {
    console.error("加载更多图片错误:", error);
  } finally {
    isLoadingMore.value = false;
  }
};

// 设置无限滚动观察器
const setupInfiniteScroll = () => {
  if (intersectionObserver) {
    intersectionObserver.disconnect();
  }

  intersectionObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting && !isLoadingMore.value && hasMore.value) {
          loadMoreImages();
        }
      });
    },
    {
      rootMargin: "100px",
    }
  );

  nextTick(() => {
    if (loadMoreTrigger.value) {
      intersectionObserver.observe(loadMoreTrigger.value);
    }
  });
};

// 图片预览（核心功能）
const openPreview = (image) => {
  currentPreviewImage.value = image;
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
  const errorBase64 = `data:image/svg+xml;base64,${btoa(
    unescape(encodeURIComponent(errorSvg))
  )}`;

  const customModal = new PopupModal({
    title: "图片预览",
    content: `
            <div class="image-preview-popup w-full max-w-[96vw] sm:max-w-5xl max-h-[85vh] flex flex-col overflow-hidden bg-white/85 dark:bg-dark-200/85 glass-card rounded-2xl">
                <!-- 顶部操作栏 -->
                <div class="preview-header bg-light-50/70 dark:bg-dark-300/70 pb-2 flex flex-col gap-2 px-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
                    <div class="flex flex-col min-w-0 gap-1">
                        <h3 class="text-xs sm:text-sm font-medium truncate">${safeFileName}</h3>
                        <p class="text-[11px] text-secondary truncate">${formatDate(
                          image.created_at
                        )}</p>
                    </div>
                    <div class="flex gap-2 flex-wrap justify-end sm:justify-end w-full sm:w-auto">
                        <!-- 复制按钮 -->
                        <div class="relative z-100">
                            <button
                                class="halo-button-copy h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                                onclick="event.stopPropagation(); window.togglePreviewCopyMenu()"
                                title="复制链接"
                            >
                                <i class="mgc_code_line text-xs"></i>
                                <span>复制</span>
                            </button>
                            <!-- 复制下拉框 -->
                            <div
                                class="absolute left-1/2 sm:left-auto sm:right-0 top-full mt-1 w-32 bg-white/90 dark:bg-dark-200/90 rounded-xl shadow-2xl border border-white/40 dark:border-dark-100/60 backdrop-blur-xl z-101 transition-all duration-200 hidden opacity-0 translate-y-[-5px] -translate-x-1/2 sm:translate-x-0 z-[999]"
                                id="previewCopyDropdown"
                            >
                                <div class="p-1.5 space-y-1">
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('url')"
                                    >
                                        <i class="mgc_link_2_line text-primary"></i>
                                        <span class="font-semibold">URL</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('html')"
                                    >
                                        <i class="mgc_code_line text-orange-500"></i>
                                        <span class="font-semibold">HTML</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('markdown')"
                                    >
                                        <i class="mgc_markdown_line text-blue-500"></i>
                                        <span class="font-semibold">MD</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('bbcode')"
                                    >
                                        <i class="mgc_brackets_line text-purple-500"></i>
                                        <span class="font-semibold">BBCode</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                        <!-- 下载按钮 -->
                        <button
                            class="halo-button halo-button-primary h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                            onclick="event.stopPropagation(); window.downloadPreviewImage()"
                        >
                            <i class="mgc_download_2_fill text-xs"></i>
                            下载
                        </button>
                        <!-- 删除按钮 -->
                        ${
                          canManageImages.value
                            ? `
                        <button
                            class="halo-button text-danger h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                            onclick="event.stopPropagation(); window.deletePreviewImage('${image.id}')"
                        >
                            <i class="mgc_delete_2_fill text-xs"></i>
                            删除
                        </button>
                        `
                            : ""
                        }
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
                        <i class="mgc_storage_line"></i>
                        存储: ${
                          image.storage === "telegram"
                            ? "Telegram"
                            : formatStorageType(image.storage)
                        }
                    </div>
                </div>
            </div>
        `,
    type: "default",
    buttons: [
      {
        text: "确定",
        type: "default",
        callback: (modal) => {
          modal.close();
          // 清理全局函数和DOM
          delete window.togglePreviewCopyMenu;
          delete window.copyPreviewImageLink;
          delete window.downloadPreviewImage;
          delete window.deletePreviewImage;
        },
      },
    ],
    maskClose: true,
    zIndex: 10000,
    maxHeight: "90vh",
  });

  // 注册弹窗内操作函数（避免全局污染，关闭时清理）
  window.togglePreviewCopyMenu = () => {
    const dropdown = document.getElementById("previewCopyDropdown");
    if (dropdown) {
      const isHidden = dropdown.classList.contains("hidden");
      if (isHidden) {
        dropdown.classList.remove("hidden", "opacity-0", "translate-y-[-5px]");
        dropdown.classList.add("block", "opacity-100", "translate-y-0");
      } else {
        dropdown.classList.add("hidden", "opacity-0", "translate-y-[-5px]");
        dropdown.classList.remove("block", "opacity-100", "translate-y-0");
      }
    }
  };

  window.copyPreviewImageLink = (type) => copyImageLink(type);
  window.downloadPreviewImage = () => downloadImage();
  window.deletePreviewImage = async (id) => {
    customModal.close();
    await deleteImage(id);
  };

  customModal.open();
};

// 复制图片链接
const copyImageLink = async (type) => {
  if (!currentPreviewImage.value) return;
  const image = currentPreviewImage.value;
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
    Message.success(`已复制${type.toUpperCase()}格式链接`, {
      position: "top-center",
      zIndex: 20000,
    });
  } catch (error) {
    // 降级处理
    const textArea = document.createElement("textarea");
    textArea.value = copyText;
    document.body.appendChild(textArea);
    textArea.select();
    document.execCommand("copy");
    document.body.removeChild(textArea);
    Message.success(`已复制${type.toUpperCase()}格式链接`, {
      position: "top-center",
      zIndex: 20000,
    });
  } finally {
    // 关闭下拉框
    nextTick(() => {
      const dropdown = document.getElementById("previewCopyDropdown");
      if (dropdown) {
        dropdown.classList.add("hidden", "opacity-0", "translate-y-[-5px]");
        dropdown.classList.remove("block", "opacity-100", "translate-y-0");
      }
    });
  }
};

// 下载图片
const downloadImage = () => {
  if (!currentPreviewImage.value) return;
  const image = currentPreviewImage.value;
  const link = document.createElement("a");
  link.href = image.url;
  link.download = image.filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  Message.success("下载已开始");
};

// 快捷删除图片功能
const deleteImage = async (imageId) => {
  const modal = new PopupModal({
    title: "确认删除",
    content: `
      <div class="flex gap-3">
        <i class="mgc_warning_line text-warning text-xl mt-1"></i>
        <div>
          <p>确定要删除这张图片吗？</p>
          <p class="mt-1 text-secondary text-sm">删除后无法恢复，请谨慎操作</p>
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
        text: "确认删除",
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

// 删除图片
const deleteAsync = async (id) => {
  const loading = Loading.show({
    text: "删除中...",
    color: "#ff4d4f",
    mask: true,
  });
  try {
    const response = await fetch(`/api/images/${id}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("authToken")}`,
      },
    });

    if (response.ok) {
      Message.success("图片删除成功");
      loadImages(); // 重新加载列表
      return true;
    } else {
      const result = await response.json();
      throw new Error(result.message || "删除失败");
    }
  } catch (error) {
    console.error("删除图片错误:", error);
    Message.error("删除图片失败: " + error.message);
    return false;
  } finally {
    await loading.hide();
  }
};

// 图片加载错误处理
const handleImageError = (event) => {
  // 占位图（灰色背景+问号）
  event.target.src =
    "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPuWbvueJh+WKoOi9veWksei0pTwvdGV4dD48L3N2Zz4=";
};

// 工具函数
const formatFileSize = (bytes) => {
  if (!bytes) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return "";
  const date = new Date(dateString);
  return date.toLocaleString("zh-CN");
};

const formatExpiration = (dateString) => {
  if (!dateString) return "永久";
  const milliseconds = new Date(dateString).getTime() - Date.now();
  if (milliseconds <= 0) return "即将清理";
  const minutes = Math.ceil(milliseconds / 60000);
  if (minutes < 60) return `${minutes} 分钟后`;
  const hours = Math.ceil(minutes / 60);
  if (hours < 24) return `${hours} 小时后`;
  return `${Math.ceil(hours / 24)} 天后`;
};

const ownerLabel = (image) => image.owner_name || {
  admin: "管理员",
  user: "普通用户",
  external: "外部",
  guest: "游客",
}[image.owner_type] || "未知";

const ownerClass = (ownerType) => ({
  admin: "admin",
  user: "user",
  external: "external",
  guest: "guest",
}[ownerType] || "guest");

// 生命周期
onMounted(() => {
  loadImages();
  setupInfiniteScroll();
});

// 监听视图切换，重置 hasMore 并重新加载
watch(viewMode, (newMode) => {
  currentPage.value = 1;
  hasMore.value = true;
  loadImages();
  if (newMode === "masonry") {
    nextTick(() => {
      setupInfiniteScroll();
    });
  }
});

// 清理资源
onUnmounted(() => {
  window.clearTimeout(searchTimer);
  // 清理可能的全局函数
  delete window.togglePreviewCopyMenu;
  delete window.copyPreviewImageLink;
  delete window.downloadPreviewImage;
  delete window.deletePreviewImage;
  // 清理无限滚动观察器
  if (intersectionObserver) {
    intersectionObserver.disconnect();
  }
});
</script>

<style scoped>
.gallery-page {
  min-height: calc(100vh - 7rem);
}

.gallery-content {
  max-width: 1180px;
}

.gallery-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 2rem;
  margin-bottom: 1.35rem;
}

.eyebrow {
  color: #c25a72;
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.gallery-heading h1 {
  margin: 0.1rem 0 0.35rem;
  color: #303941;
  font-size: clamp(2rem, 4vw, 3.4rem);
  font-weight: 750;
  letter-spacing: -0.055em;
  line-height: 1;
}

.dark .gallery-heading h1 {
  color: #f0f2f4;
}

.gallery-heading > div:first-child > p:last-child {
  color: #858d95;
  font-size: 0.82rem;
}

.gallery-count {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 8.5rem;
  padding: 0.72rem 0.85rem;
  border: 1px solid rgba(203, 103, 127, 0.14);
  border-radius: 0.95rem;
  color: #7b7377;
  background: rgba(255, 255, 255, 0.75);
  box-shadow: 0 10px 28px rgba(78, 48, 57, 0.06);
  font-size: 0.72rem;
}

.gallery-count i {
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border-radius: 0.65rem;
  color: #c4556d;
  background: #fff0f3;
  font-size: 1.05rem;
}

.gallery-count strong {
  color: #3f474f;
  font-size: 1rem;
}

.dark .gallery-count {
  color: #b9b1b5;
  background: rgba(31, 36, 43, 0.86);
  border-color: rgba(255, 255, 255, 0.07);
}

.dark .gallery-count i {
  color: #ff9bb0;
  background: #3b3037;
}

.dark .gallery-count strong {
  color: #eef1f3;
}

.gallery-toolbar {
  display: grid;
  grid-template-columns: minmax(13rem, 1fr) auto auto;
  align-items: center;
  gap: 0.7rem;
  margin-bottom: 1.25rem;
  padding: 0.65rem;
  border-radius: 1.05rem;
}

.gallery-search {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  height: 2.55rem;
  padding: 0 0.75rem;
  border: 1px solid rgba(128, 116, 120, 0.13);
  border-radius: 0.75rem;
  color: #a35a6b;
  background: rgba(249, 246, 246, 0.88);
}

.gallery-search:focus-within {
  border-color: rgba(204, 92, 118, 0.42);
  box-shadow: 0 0 0 3px rgba(207, 92, 119, 0.08);
}

.gallery-search input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  color: #414850;
  background: transparent;
  font-size: 0.76rem;
}

.gallery-search button {
  color: #b5a9ac;
}

.dark .gallery-search {
  color: #f193a7;
  background: rgba(18, 22, 28, 0.66);
  border-color: rgba(255, 255, 255, 0.06);
}

.dark .gallery-search input {
  color: #e9edef;
}

.owner-filter {
  display: flex;
  gap: 0.2rem;
  padding: 0.22rem;
  border-radius: 0.72rem;
  background: rgba(245, 240, 241, 0.88);
}

.owner-filter button {
  padding: 0.48rem 0.62rem;
  border-radius: 0.58rem;
  color: #898084;
  font-size: 0.68rem;
  font-weight: 700;
  white-space: nowrap;
  transition: 0.18s ease;
}

.owner-filter button:hover,
.owner-filter button.active {
  color: #b74a63;
  background: white;
  box-shadow: 0 3px 10px rgba(76, 47, 55, 0.06);
}

.dark .owner-filter {
  background: rgba(20, 24, 30, 0.68);
}

.dark .owner-filter button.active,
.dark .owner-filter button:hover {
  color: #ff9bb0;
  background: #333039;
}

.view-toggle {
  justify-content: flex-end;
}

.toolbar-icon,
.batch-entry {
  height: 2.55rem;
  border-radius: 0.7rem;
  color: #8b8387;
  background: rgba(248, 244, 245, 0.78);
  transition: 0.18s ease;
}

.toolbar-icon {
  width: 2.55rem;
}

.toolbar-icon:hover,
.toolbar-icon.active {
  color: #bd5068;
  background: #fff0f3;
}

.batch-entry {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-left: 0.2rem;
  padding: 0 0.72rem;
  font-size: 0.7rem;
  font-weight: 700;
}

.batch-active {
  padding: 0 0.45rem;
  color: #bd5068;
  font-size: 0.68rem;
  font-weight: 750;
}

.dark .toolbar-icon,
.dark .batch-entry {
  color: #bcb4b8;
  background: rgba(20, 24, 30, 0.72);
}

.dark .toolbar-icon:hover,
.dark .toolbar-icon.active,
.dark .batch-entry:hover {
  color: #ff9bb0;
  background: #3a3036;
}

.image-card,
.masonry-card {
  transform: translateY(0);
}

.image-card:hover,
.masonry-card:hover {
  transform: translateY(-3px);
}

.owner-badge,
.card-expiry {
  position: absolute;
  z-index: 5;
  display: inline-flex;
  align-items: center;
  max-width: calc(100% - 1.25rem);
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.34);
  border-radius: 0.55rem;
  box-shadow: 0 4px 14px rgba(28, 25, 27, 0.1);
  backdrop-filter: blur(10px);
  font-size: 0.6rem;
  font-weight: 750;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.owner-badge {
  top: 0.55rem;
  left: 0.55rem;
  padding: 0.38rem 0.48rem;
}

.owner-badge.admin {
  color: #a9435b;
  background: rgba(255, 236, 241, 0.88);
}

.owner-badge.user {
  color: #397763;
  background: rgba(231, 249, 242, 0.88);
}

.owner-badge.guest,
.owner-badge.external {
  color: #6d697f;
  background: rgba(240, 239, 249, 0.88);
}

.card-expiry {
  right: 0.55rem;
  bottom: 0.55rem;
  gap: 0.25rem;
  padding: 0.38rem 0.5rem;
  color: #7a505a;
  background: rgba(255, 251, 251, 0.88);
}

.dark .card-expiry {
  color: #ffd3dc;
  background: rgba(40, 33, 38, 0.86);
  border-color: rgba(255, 255, 255, 0.08);
}

@media (max-width: 980px) {
  .gallery-toolbar {
    grid-template-columns: minmax(12rem, 1fr) auto;
  }

  .owner-filter {
    grid-column: 1 / -1;
    grid-row: 2;
    overflow-x: auto;
  }
}

@media (max-width: 640px) {
  .gallery-heading {
    align-items: stretch;
    flex-direction: column;
    gap: 0.9rem;
  }

  .gallery-heading h1 {
    font-size: 2.25rem;
  }

  .gallery-count {
    align-self: flex-start;
  }

  .gallery-toolbar {
    grid-template-columns: 1fr;
    padding: 0.55rem;
  }

  .gallery-search,
  .owner-filter,
  .view-toggle {
    grid-column: 1;
  }

  .view-toggle {
    justify-content: stretch;
  }

  .toolbar-icon {
    flex: 0 0 2.55rem;
  }

  .batch-entry {
    flex: 1;
    justify-content: center;
  }
}

@media (prefers-reduced-motion: reduce) {
  .image-card,
  .masonry-card,
  .toolbar-icon,
  .batch-entry,
  .owner-filter button {
    transition: none;
  }

  .image-card:hover,
  .masonry-card:hover {
    transform: none;
  }
}
</style>
