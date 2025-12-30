# ============================================
# 阶段1：构建前端
# ============================================
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# 安装 pnpm
RUN npm install -g pnpm

# 先复制依赖文件，利用 Docker 缓存
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制源码并构建
COPY frontend/ ./
RUN pnpm run build

# ============================================
# 阶段2：构建后端
# ============================================
FROM golang:1.24-alpine AS backend-builder

# 安装 CGO 编译依赖
RUN apk add --no-cache gcc g++ musl-dev libwebp-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

COPY backend/ ./backend/
COPY main.go ./

# 复制前端构建结果
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY --from=frontend-builder /app/frontend/src/assets/fonts/ ./frontend/src/assets/fonts/

# 编译（启用 CGO 支持 webp）
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o main ./main.go

# ============================================
# 阶段3：运行时镜像
# ============================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    libwebp \
    wget

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件和前端资源
COPY --from=backend-builder /app/main ./
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 创建必要的目录
RUN mkdir -p /app/data /app/uploads

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
    CMD wget --spider -q http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./main"]
