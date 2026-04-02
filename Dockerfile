# 多阶段构建 - 优化镜像大小

# 阶段 1: 编译前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端源代码
COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 阶段 2: 编译后端
FROM golang:1.25-alpine AS backend-builder

# 安装必要的工具
RUN apk add --no-cache git

WORKDIR /app/backend

# 复制 go.mod 和 go.sum
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源代码
COPY backend/ ./

# 编译参数
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -s -w -buildid=" \
    -o lottery .

# 阶段 3: 运行时镜像
FROM alpine:latest

# 安装 ca-certificates
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# 复制二进制文件
COPY --from=backend-builder --chown=appuser:appgroup /app/backend/lottery .

# 复制前端文件
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/index.html .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/lottery-web ./lottery-web

# 创建数据目录
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app/data

# 切换到非 root 用户
USER appuser:appgroup

# 暴露端口
EXPOSE 8902

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8902/api/version || exit 1

# 运行应用
ENTRYPOINT ["./lottery"]
