# 初春图床系统

一个功能完整的现代化图床管理系统，基于 Vue.js 3 + Go 构建，支持多种存储方式、POW验证、剪贴板上传等高级功能。

## 🐳 Docker 部署

### 环境要求
- Docker 20.10.0 或更高版本
- Docker Compose v2.0.0 或更高版本

### 使用 Docker Compose 部署

1. **克隆项目**
```bash
git clone https://github.com/onexru/oneimg.git
cd oneimg
```

2. **启动服务**
```bash
docker compose up -d
```

3. **访问系统**
- `http://localhost:8080`

4. **停止服务**
```bash
docker compose down
```

### 数据持久化
系统数据和上传的图片通过 Docker 数据卷保持持久化：
- 上传的图片存储在 `./uploads` 目录
- 数据库文件存储在 `./data` 目录

## 环境变量配置

通过 `.env` 文件或环境变量配置系统：

```env
# 服务器配置
SERVER_PORT=8080

# 数据库配置（优先级：PostgreSQL > MySQL > SQLite）
SQLITE_PATH=./data/data.db

# MySQL配置
IS_MYSQL=false
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=oneimgxru

# PostgreSQL配置
IS_POSTGRES=false
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=
POSTGRES_DB=oneimg

# 文件上传配置
MAX_FILE_SIZE=10485760
ALLOWED_TYPES=image/jpeg,image/png,image/gif,image/webp

# 默认用户配置
DEFAULT_USER=admin
DEFAULT_PASS=123456

# Session配置
SESSION_SECRET=your_secret_key
```

## 功能特性

### 🗄️ 多数据库支持
- **SQLite** - 默认，轻量级
- **MySQL** - 适合生产环境
- **PostgreSQL** - 企业级数据库支持

### 📦 多存储支持
- **本地存储** - 默认存储方式
- **S3/R2** - 兼容 S3 协议的对象存储（支持自定义访问 URL）
- **WebDAV** - WebDAV 协议存储
- **FTP** - FTP 服务器存储
- **Telegram** - Telegram Bot 存储

### 🔐 安全认证
- POW (工作量证明) 验证登录
- Session 会话管理
- 密码加密存储
- 会话超时保护
- Referer 来源白名单

### 📤 图片上传
- **剪贴板粘贴直接上传** - 支持 Ctrl+V 粘贴上传
- **URL 直链上传** - 通过图片 URL 直接上传
- 拖拽上传支持
- 批量文件选择上传
- 支持多种图片格式 (JPEG, PNG, GIF, WebP, SVG, BMP)
- 自动压缩和格式转换
- 可选 WebP 格式输出
- 文件大小限制和格式验证
- 上传进度显示

### 🖼️ 图片管理
- 图片预览和详情查看
- 多种复制链接格式（URL、Markdown、HTML、BBCode）
- 图片信息展示（尺寸、大小、存储类型）
- 批量删除功能
- 缩略图生成

### 🎨 图片水印
- 自定义水印文本
- 可调整大小、颜色、透明度
- 多种位置选择（四角、居中）
- 新上传自动添加水印

### 👤 用户系统
- 管理员账户
- 游客登录模式（可配置）
- 个人资料设置（昵称、头像）
- 密码修改

### 📊 数据统计
- 仪表板概览
- 存储空间统计
- 图片数量统计
- 实时数据更新

### 🎯 用户界面
- 现代化设计风格
- 响应式布局 (支持移动端)
- 深色/浅色主题切换
- 流畅的动画效果
- 直观的操作体验
- Telegram 通知支持

## 技术栈

### 前端
- Vue.js 3
- Vite
- Tailwind CSS
- Vue Router
- Pinia

### 后端
- Go (Gin Framework)
- GORM
- SQLite / MySQL / PostgreSQL

## License

MIT License