# 阶段1：构建前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# 安装pnpm并构建前端
RUN npm install -g pnpm
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build


# 阶段2：构建后端
FROM golang:1.24-alpine AS backend-builder

# 安装CGO编译依赖
RUN apk add --no-cache gcc g++ musl-dev libwebp-dev

# 设置工作目录
WORKDIR /app

# 复制Go依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制后端源代码
COPY backend/ ./backend/
COPY main.go ./

# 复制前端构建结果到后端可访问的路径
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY --from=frontend-builder /app/frontend/src/assets/fonts/ ./frontend/src/assets/fonts/

# 编译后端应用（启用CGO支持webp）
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./main.go


# 阶段3：最终运行环境
FROM alpine:3.18

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    libwebp

WORKDIR /app

COPY --from=backend-builder /app/main ./
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY .env ./

EXPOSE 8080

# 🌸 启动前修权限，再启动 Go
CMD sh -c "chmod -R 755 /app/data /app/uploads || true && ./main"
