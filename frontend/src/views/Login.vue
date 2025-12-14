<template>
    <div class="login flex items-center justify-center p-4">
        <!-- 全局加载遮罩 -->
        <div v-if="isLoading" class="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50">
            <div class="loading-card bg-white dark:bg-gray-800 rounded-xl shadow-2xl p-6 max-w-md w-full m-[15px] flex flex-col items-center justify-center">
                <!-- 加载动画 -->
                <div class="loading-spinner mb-4">
                    <svg class="animate-spin h-10 w-10 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                </div>
                <h3 class="loading-title text-lg font-bold text-center text-gray-800 dark:text-white mb-2">{{ loadingTitle }}</h3>
                <p class="loading-text text-center text-gray-600 dark:text-gray-300 mb-4">{{ loadingText }}</p>
                <!-- 进度条 -->
                <!-- 进度条已移除 -->
            </div>
        </div>

        <!-- 登录卡片 -->
        <div class="card bg-white dark:bg-gray-800 rounded-xl shadow-lg w-full max-w-md transition-all duration-300" :class="{ 'opacity-50 pointer-events-none': isLoading }">
            <div class="card-body p-6">
                <h5 class="card-title text-2xl font-bold text-center text-gray-800 dark:text-white mb-8">登录</h5>
                <!-- 用户名输入 -->
                <div class="form-group mb-6">
                    <label for="username" class="form-label block text-gray-700 dark:text-gray-300 mb-2">用户名</label>
                    <input 
                        type="text" 
                        v-model="username" 
                        class="form-input w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary dark:bg-gray-700 dark:text-white transition-all outline-none"
                        placeholder="用户名"
                        :disabled="isLoading"
                        @keyup.enter="handleLogin"
                    />
                </div>
                <!-- 密码输入 -->
                <div class="form-group mb-6">
                    <label for="password" class="form-label block text-gray-700 dark:text-gray-300 mb-2">密码</label>
                    <input 
                        type="password" 
                        v-model="password" 
                        class="form-input w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary dark:bg-gray-700 dark:text-white transition-all outline-none"
                        placeholder="密码"
                        :disabled="isLoading"
                        @keyup.enter="handleLogin"
                    />
                </div>
                
                <!-- Cloudflare Turnstile 验证 -->
                <div v-if="loginConfig.turnstile" class="form-group mb-6">
                    <div class="flex justify-center transform scale-90 sm:scale-100 origin-center">
                        <div id="turnstile-container"></div>
                    </div>
                </div>
                
                <!-- 登录按钮 -->
                <div class="form-group">
                    <button 
                        @click="handleLogin" 
                        class="login-btn w-full py-3 bg-primary hover:bg-primary/90 text-white font-medium rounded-lg transition-all duration-200 flex items-center justify-center"
                        :class="{ 'opacity-70 cursor-not-allowed': isLoading }"
                        :disabled="isLoading"
                    >
                        登录
                    </button>
                </div>
                <!-- 游客登录按钮 -->
                <div v-if="loginConfig.tourist" class="form-group mt-4">
                    <button 
                        @click="handleTouristLogin" 
                        class="tourist-login-btn w-full py-3 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 font-medium rounded-lg transition-all duration-200 flex items-center justify-center gap-2"
                        :class="{ 'opacity-70 cursor-not-allowed': isLoading }"
                        :disabled="isLoading"
                    >
                        <i class="ri-user-line"></i>
                        游客访问
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive, watch } from 'vue';
import message from '@/utils/message.js';

// 响应式数据
const username = ref('');
const password = ref('');
const isLoading = ref(false);
const loadingTitle = ref('');
const loadingText = ref('');
const loadingProgress = ref(0);
const turnstileToken = ref('');
let turnstileWidgetId = null;

// 登录配置
const loginConfig = reactive({
    turnstile: false,
    turnstileSiteKey: '', // 从后端 API 获取
    tourist: false
})

// 加载状态管理
const setLoadingState = (title, text, progress = 0) => {
    isLoading.value = true;
    loadingTitle.value = title;
    loadingText.value = text;
    loadingProgress.value = progress;
};

const clearLoadingState = () => {
    isLoading.value = false;
    loadingTitle.value = '';
    loadingText.value = '';
    loadingProgress.value = 0;
};

// 生成游客指纹
const generateTouristFingerprint = async () => {
    try {
        const fingerprintParams = await window.GuestFingerprint.getRequestParams();
        return fingerprintParams.guest_uuid;
    } catch (e) {
        console.error('生成游客指纹失败:', e);
        // 降级方案：生成临时标识
        return 'guest_' + Math.random().toString(36).substr(2, 16);
    }
};

// 游客登录处理 - 跳过 Turnstile 验证
// 游客登录处理 - 跳过 Turnstile 验证
const handleTouristLogin = async () => {
    if (isLoading.value) return;

    // 如果开启了 Turnstile 验证且没有 token
    if (loginConfig.turnstile && !turnstileToken.value) {
        message.warning('请完成人机验证');
        return;
    }

    setLoadingState('游客登录', '正在生成游客身份...', 30);

    // 生成游客唯一标识
    const touristId = await generateTouristFingerprint();
    username.value = touristId;
    password.value = 'tourist_' + touristId.substr(0, 8);

    // 游客登录跳过验证
    setLoadingState('游客登录', '正在登录...', 60);
    putLogin(turnstileToken.value, touristId);
};

// 登录处理
const handleLogin = () => {
    if (isLoading.value) return;
    
    if (!username.value || !password.value) {
        message.warning('请输入用户名和密码');
        return;
    }
    
    // 如果开启了 Turnstile 验证且没有 token
    if (loginConfig.turnstile && !turnstileToken.value) {
        message.warning('请完成人机验证');
        return;
    }
    
    putLogin(turnstileToken.value);
};

// 提交登录请求
const putLogin = async (token, touristId = '') => {
    setLoadingState('登录中', '正在验证用户信息...', 90);
    
    try {
        // 组装登录参数
        const loginData = {
            username: username.value,
            password: password.value,
            turnstileToken: token
        };

        // 游客登录时补充指纹信息
        if (touristId) {
            loginData.touristFingerprint = touristId;
            try {
                const fingerprintParams = await window.GuestFingerprint.getRequestParams();
                loginData.fusionHash = fingerprintParams.fusion_hash;
                loginData.stableFeatures = fingerprintParams.stable_features;
            } catch (e) {
                console.error('获取指纹参数失败:', e);
            }
        }

        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(loginData)
        });
        
        const result = await response.json();
        
        if (response.ok && result.code === 200) {
            // 保存 Auth Token
            if (result.data && result.data.token) {
                localStorage.setItem('authToken', result.data.token);
            }

            // 保存用户信息
            const savedUser = (result.data && result.data.user) || {};
            const userInfo = {
                username: savedUser.username || username.value,
                nickname: savedUser.nickname || '',
                avatar: savedUser.avatar || '',
                role: savedUser.role,
                isTourist: !!touristId,
                touristFingerprint: touristId || ''
            };
            localStorage.setItem('userInfo', JSON.stringify(userInfo));
            
            setLoadingState((result.message || '登录成功'), '即将跳转到主页...', 100);
            
            setTimeout(() => {
                clearLoadingState();
                // 跳转到主页
                window.location.href = '/';
            }, 1500);
        } else {
            clearLoadingState();
            message.error('登录失败: ' + (result.message || '未知错误'));
            // 重置 Turnstile
            resetTurnstile();
        }
    } catch (error) {
        clearLoadingState();
        message.error('登录请求失败，请检查网络连接: ' + error.message);
        resetTurnstile();
    }
};

// 初始化 Turnstile
const initTurnstile = () => {
    if (!loginConfig.turnstile) return;
    
    const container = document.getElementById('turnstile-container');
    if (!container || !window.turnstile) {
        setTimeout(initTurnstile, 500);
        return;
    }
    
    // 清空容器
    container.innerHTML = '';
    
    // 渲染 Turnstile
    turnstileWidgetId = window.turnstile.render('#turnstile-container', {
        sitekey: loginConfig.turnstileSiteKey,
        callback: (token) => {
            turnstileToken.value = token;
        },
        'expired-callback': () => {
            turnstileToken.value = '';
        },
        'error-callback': () => {
            message.error('验证组件加载失败');
            turnstileToken.value = '';
        },
        theme: document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    });
};

// 重置 Turnstile
const resetTurnstile = () => {
    if (turnstileWidgetId && window.turnstile) {
        window.turnstile.reset(turnstileWidgetId);
        turnstileToken.value = '';
    }
};

// 获取登录配置
const getLoginSettings = async () => { 
    try {
        const response = await fetch('/api/settings/login', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        const result = await response.json();
        if (response.ok && result.code === 200) {
            // 映射字段名
            loginConfig.turnstile = result.data.turnstile || false;
            loginConfig.turnstileSiteKey = result.data.turnstile_site_key || '';
            loginConfig.tourist = result.data.tourist || false;
        } else {
            console.error('获取登录配置失败');
        }
    } catch (error) {
        console.error('获取登录配置失败:', error);
    }
};

// 监听 turnstile 配置变化
watch(() => loginConfig.turnstile, (newVal) => {
    if (newVal) {
        setTimeout(initTurnstile, 100);
    }
});

// 加载 Turnstile 脚本
onMounted(async () => {
    // 修复URL方法兼容问题
    if (!URL.revokeObjectUrl && URL.revokeObjectURL) {
        URL.revokeObjectUrl = URL.revokeObjectURL;
    }

    // 获取登录配置
    await getLoginSettings();
    
    // 加载 Turnstile 脚本
    if (!document.querySelector('script[src*="turnstile"]')) {
        const script = document.createElement('script');
        script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js';
        script.async = true;
        script.defer = true;
        script.onload = () => {
            console.log('Turnstile 脚本加载完成');
            if (loginConfig.turnstile) {
                initTurnstile();
            }
        };
        script.onerror = () => {
            console.error('Turnstile 脚本加载失败');
        };
        document.head.appendChild(script);
    } else if (loginConfig.turnstile) {
        initTurnstile();
    }
});

// 清理资源
onUnmounted(() => {
    if (turnstileWidgetId && window.turnstile) {
        window.turnstile.remove(turnstileWidgetId);
    }
});
</script>