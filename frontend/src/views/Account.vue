<template>
    <div class="text-gray-800 dark:text-gray-200">
        <!-- 页面头部 -->
        <div class="settings-header container mx-auto px-4 py-4">
            <h1 class="page-title flex items-center text-2xl md:text-3xl font-bold">
                设置
            </h1>
            <p class="page-description text-gray-600 dark:text-gray-400 mt-2">管理您的系统账户</p>
        </div>

        <!-- 主要内容 - 双栏布局 -->
        <div class="container mx-auto px-4 pb-16">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">

                <!-- 左侧：账户设置 -->
                <div class="bg-white dark:bg-gray-800 rounded-xl shadow-md overflow-hidden w-full">
                    <div class="panel-content p-6 md:p-8">
                        <h2 class="panel-title flex items-center text-xl font-semibold mb-6">
                            <span class="panel-icon mr-2 text-2xl">
                                <i class="ri-user-3-line"></i>
                            </span>
                            个人资料
                        </h2>
                        
                        <!-- 头像设置 -->
                        <div class="flex items-center gap-4 mb-6">
                            <div class="relative">
                                <div class="w-20 h-20 rounded-full bg-gradient-to-br from-primary to-blue-400 flex items-center justify-center text-white text-2xl font-bold overflow-hidden">
                                    <img v-if="userProfile.avatar" :src="userProfile.avatar" class="w-full h-full object-cover" alt="头像" />
                                    <span v-else>{{ (userProfile.nickname || userProfile.username || 'U').charAt(0).toUpperCase() }}</span>
                                </div>
                                <button 
                                    type="button"
                                    @click="triggerAvatarUpload"
                                    class="absolute -bottom-1 -right-1 w-7 h-7 bg-primary hover:bg-primary-dark text-white rounded-full flex items-center justify-center shadow-md transition-colors"
                                    title="更换头像"
                                >
                                    <i class="ri-camera-line text-sm"></i>
                                </button>
                                <input 
                                    ref="avatarInput"
                                    type="file"
                                    accept="image/*"
                                    @change="handleAvatarSelect"
                                    class="hidden"
                                />
                            </div>
                            <div>
                                <p class="font-medium text-gray-900 dark:text-white">{{ userProfile.nickname || userProfile.username || '未设置' }}</p>
                                <p class="text-sm text-gray-500 dark:text-gray-400">点击更换头像</p>
                            </div>
                        </div>

                        <!-- 昵称设置 -->
                        <div class="setting-group mb-6">
                            <label class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" for="nickname">
                                昵称
                            </label>
                            <input 
                                id="nickname"
                                v-model="profileForm.nickname"
                                type="text" 
                                class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                                placeholder="设置您的昵称"
                                maxlength="20"
                            />
                            <button 
                                type="button"
                                @click="updateProfile"
                                :disabled="isUpdatingProfile"
                                class="mt-3 w-full py-2.5 bg-primary hover:bg-primary-dark text-white rounded-lg transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                            >
                                <i v-if="isUpdatingProfile" class="ri-loader-4-line animate-spin"></i>
                                <span>保存昵称</span>
                            </button>
                        </div>

                        <hr class="border-gray-200 dark:border-gray-700 my-6" />

                        <h3 class="flex items-center text-lg font-semibold mb-4">
                            <span class="mr-2 text-xl">
                                <i class="ri-shield-user-line"></i>
                            </span>
                            安全设置
                        </h3>
                        
                        <!-- 账户修改表单 -->
                        <form @submit.prevent="updateAccount" class="account-form space-y-6">
                            <!-- 新用户名 -->
                            <div class="setting-group">
                                <label 
                                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" 
                                    for="newUsername"
                                >
                                    新用户名（留空则不修改）
                                </label>
                                <input 
                                    id="newUsername"
                                    v-model="accountForm.newUsername"
                                    type="text" 
                                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                                    placeholder="留空则不修改用户名"
                                    minlength="3"
                                    maxlength="20"
                                />
                            </div>
                            
                            <!-- 当前密码 -->
                            <div class="setting-group">
                                <label 
                                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" 
                                    for="currentPassword"
                                >
                                    当前密码 <span class="text-red-500">*</span>
                                </label>
                                <input 
                                    id="currentPassword"
                                    v-model="accountForm.currentPassword"
                                    type="password" 
                                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                                    placeholder="请输入当前密码以确认修改"
                                    required
                                />
                            </div>
                            
                            <!-- 新密码 -->
                            <div class="setting-group">
                                <label 
                                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" 
                                    for="newPassword"
                                >
                                    新密码（留空则不修改）
                                </label>
                                <input 
                                    id="newPassword"
                                    v-model="accountForm.newPassword"
                                    type="password" 
                                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                                    placeholder="留空则不修改密码（至少6位）"
                                    minlength="6"
                                />
                            </div>
                            
                            <!-- 确认新密码 -->
                            <div class="setting-group">
                                <label 
                                    class="setting-label block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1" 
                                    for="confirmPassword"
                                >
                                    确认新密码
                                </label>
                                <input 
                                    id="confirmPassword"
                                    v-model="accountForm.confirmPassword"
                                    type="password" 
                                    class="setting-input w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 focus:ring-2 focus:ring-primary focus:border-primary dark:focus:ring-primary/70 dark:focus:border-primary/70 transition-colors outline-none"
                                    placeholder="请再次输入新密码"
                                />
                            </div>
                            
                            <!-- 提交按钮 -->
                            <div class="setting-group pt-2">
                                <button 
                                    type="submit" 
                                    :disabled="isUpdatingAccount"
                                    class="setting-btn accent w-full py-3 px-6 bg-primary hover:bg-primary/90 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2 focus:ring-2 focus:ring-primary/50 focus:outline-none"
                                >
                                    <span v-if="isUpdatingAccount" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                                    <span>保存修改</span>
                                </button>
                            </div>
                        </form>
                    </div>
                </div>

                <!-- 右侧：数据库状态 -->
                <div class="bg-white dark:bg-gray-800 rounded-xl shadow-md overflow-hidden w-full">
                    <div class="panel-content p-6 md:p-8">
                        <h2 class="panel-title flex items-center text-xl font-semibold mb-8">
                            <span class="panel-icon mr-2 text-2xl">
                                <i class="ri-database-2-line"></i>
                            </span>
                            数据库状态
                        </h2>
                        
                        <div class="space-y-6">
                            <!-- 数据库类型 -->
                            <div class="flex items-center justify-between py-3 border-b border-gray-200 dark:border-gray-700">
                                <div class="flex items-center gap-3">
                                    <div class="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                                        <i class="ri-stack-line text-blue-600 dark:text-blue-400 text-xl"></i>
                                    </div>
                                    <span class="text-gray-700 dark:text-gray-300 font-medium">数据库类型</span>
                                </div>
                                <span class="text-gray-900 dark:text-white font-semibold">
                                    {{ formatDbType(dbStatus.type) }}
                                </span>
                            </div>
                            
                            <!-- 连接状态 -->
                            <div class="flex items-center justify-between py-3 border-b border-gray-200 dark:border-gray-700">
                                <div class="flex items-center gap-3">
                                    <div class="w-10 h-10 rounded-lg flex items-center justify-center"
                                         :class="dbStatus.connected ? 'bg-green-100 dark:bg-green-900/30' : 'bg-red-100 dark:bg-red-900/30'">
                                        <i class="text-xl" 
                                           :class="dbStatus.connected 
                                               ? 'ri-checkbox-circle-line text-green-600 dark:text-green-400' 
                                               : 'ri-close-circle-line text-red-600 dark:text-red-400'"></i>
                                    </div>
                                    <span class="text-gray-700 dark:text-gray-300 font-medium">连接状态</span>
                                </div>
                                <span class="px-3 py-1 rounded-full text-sm font-medium"
                                      :class="dbStatus.connected 
                                          ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400' 
                                          : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'">
                                    {{ dbStatus.connected ? '已连接' : '未连接' }}
                                </span>
                            </div>
                            
                            <!-- 提示信息 -->
                            <div class="mt-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                <p class="text-sm text-gray-600 dark:text-gray-400">
                                    <i class="ri-information-line mr-1"></i>
                                    数据库配置通过环境变量设置，修改后需重启服务生效。
                                </p>
                                <p class="text-xs text-gray-500 dark:text-gray-500 mt-2">
                                    优先级：PostgreSQL &gt; MySQL &gt; SQLite
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import message from '@/utils/message.js'

const router = useRouter()

// 表单数据
const accountForm = ref({
    newUsername: '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
})

// 用户资料
const userProfile = ref({
    username: '',
    nickname: '',
    avatar: ''
})

const profileForm = ref({
    nickname: ''
})

const avatarInput = ref(null)
const isUpdatingProfile = ref(false)

// 数据库状态
const dbStatus = ref({
    type: 'loading',
    connected: false
})

// 加载状态
const isUpdatingAccount = ref(false)

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
                userProfile.value = result.data
                profileForm.value.nickname = result.data.nickname || ''
            }
        }
    } catch (error) {
        console.error('获取用户资料失败:', error)
    }
}

// 更新资料
const updateProfile = async () => {
    if (!profileForm.value.nickname.trim()) {
        message.warning('请输入昵称')
        return
    }
    
    isUpdatingProfile.value = true
    try {
        const response = await fetch('/api/user/profile', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            },
            body: JSON.stringify({
                nickname: profileForm.value.nickname
            })
        })
        
        const result = await response.json()
        if (response.ok && result.code === 200) {
            userProfile.value.nickname = profileForm.value.nickname
            message.success('昵称更新成功')
        } else {
            throw new Error(result.message || '更新失败')
        }
    } catch (error) {
        message.error(error.message || '更新失败')
    } finally {
        isUpdatingProfile.value = false
    }
}

// 头像上传
const triggerAvatarUpload = () => {
    avatarInput.value?.click()
}

const handleAvatarSelect = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    
    // 验证文件类型
    if (!file.type.startsWith('image/')) {
        message.error('请选择图片文件')
        return
    }
    
    // 上传头像
    const formData = new FormData()
    formData.append('images[]', file)
    
    try {
        const response = await fetch('/api/upload/images?hidden=true', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            },
            body: formData
        })
        
        const result = await response.json()
        if (response.ok && result.code === 200 && result.data?.files?.length > 0) {
            const avatarUrl = result.data.files[0].url
            
            // 更新头像URL
            const updateResponse = await fetch('/api/user/profile', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('authToken')}`
                },
                body: JSON.stringify({ avatar: avatarUrl })
            })
            
            if (updateResponse.ok) {
                userProfile.value.avatar = avatarUrl
                message.success('头像更新成功')
            }
        } else {
            throw new Error(result.message || '上传失败')
        }
    } catch (error) {
        message.error('头像上传失败: ' + error.message)
    }
    
    e.target.value = ''
}

// 格式化数据库类型显示
const formatDbType = (type) => {
    const typeMap = {
        'postgresql': 'PostgreSQL',
        'mysql': 'MySQL',
        'sqlite': 'SQLite',
        'loading': '加载中...'
    }
    return typeMap[type] || type
}

// 获取数据库状态
const fetchDbStatus = async () => {
    try {
        const response = await fetch('/api/database/status', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })
        
        if (response.ok) {
            const result = await response.json()
            if (result.code === 200 && result.data) {
                dbStatus.value = result.data
            }
        }
    } catch (error) {
        console.error('获取数据库状态失败:', error)
    }
}

// 更新账户信息
const updateAccount = async () => {
    const { newUsername, currentPassword, newPassword, confirmPassword } = accountForm.value
    
    // 检查是否有任何修改
    const hasUsernameChange = newUsername && newUsername.trim() !== ''
    const hasPasswordChange = newPassword && newPassword.trim() !== ''
    
    if (!hasUsernameChange && !hasPasswordChange) {
        message.error('请输入要修改的用户名或密码')
        return
    }
    
    // 验证用户名（如果要修改）
    if (hasUsernameChange) {
        if (newUsername.length < 3) {
            message.error('用户名长度至少为3位')
            return
        }
        
        if (newUsername.length > 20) {
            message.error('用户名长度不能超过20位')
            return
        }
    }
    
    // 验证密码（如果要修改）
    if (hasPasswordChange) {
        if (newPassword.length < 6) {
            message.error('新密码长度至少为6位')
            return
        }
        
        if (newPassword !== confirmPassword) {
            message.error('两次输入的新密码不一致')
            return
        }
    }
    
    // 验证当前密码
    if (!currentPassword) {
        message.error('请输入当前密码以确认修改')
        return
    }
    
    try {
        isUpdatingAccount.value = true
        
        const response = await fetch('/api/account/change', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            },
            body: JSON.stringify({
                new_username: newUsername,
                current_password: currentPassword,
                new_password: newPassword
            })
        })
        
        const result = await response.json()
        
        if (!response.ok || !result.success) {
            // 未授权处理
            if (response.status === 401) {
                localStorage.removeItem('authToken')
                router.push('/login')
                return message.error('登录已过期，请重新登录')
            }
            throw new Error(result.message || '修改失败')
        }
        
        message.success('修改成功')

        // 清空表单
        accountForm.value = {
            newUsername: '',
            currentPassword: '',
            newPassword: '',
            confirmPassword: ''
        }
        
        // 刷新页面
        setTimeout(() => {
            window.location.reload();
        }, 1000)

    } catch (error) {
        message.error(error.message || '更新失败')
    } finally {
        isUpdatingAccount.value = false
    }
}

onMounted(() => {
    fetchDbStatus()
    fetchUserProfile()
})
</script>