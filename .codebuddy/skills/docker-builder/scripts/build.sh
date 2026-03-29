#!/bin/bash

# Docker 镜像构建脚本
# 仅本地构建，不自动推送

set -e

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
cd "$PROJECT_ROOT"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
VERSION="${1:-v1.0.0}"
IMAGE_NAME="${IMAGE_NAME:-techfunways/lottery}"
PORT="${PORT:-8902}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"

# 镜像名称
FULL_IMAGE_NAME="$IMAGE_NAME"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Docker 镜像构建${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  版本: $VERSION${NC}"
echo -e "${BLUE}  镜像: $FULL_IMAGE_NAME${NC}"
echo -e "${BLUE}  平台: $PLATFORMS${NC}"
echo -e "${BLUE}========================================${NC}"

# 版本信息
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}📦 构建时间: $BUILD_TIME${NC}"
echo -e "${GREEN}🔖 Git 提交: $GIT_COMMIT${NC}"
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker 未安装${NC}"
    exit 1
fi

# 检查 Docker Buildx
if ! docker buildx version &> /dev/null; then
    echo -e "${RED}❌ Docker Buildx 未启用${NC}"
    echo -e "${YELLOW}请启用 Buildx: docker buildx create --use${NC}"
    exit 1
fi

# 检查前端
if [ ! -d "frontend/dist" ]; then
    echo -e "${YELLOW}⚠️  前端未编译，开始编译前端...${NC}"

    # 检查前端构建脚本
    if [ -f ".codebuddy/skills/frontend-build/scripts/build.sh" ]; then
        .codebuddy/skills/frontend-build/scripts/build.sh || {
            echo -e "${RED}❌ 前端编译失败${NC}"
            echo -e "${YELLOW}请先手动编译前端：cd frontend && npm run build${NC}"
            exit 1
        }
    else
        echo -e "${RED}❌ 未找到前端构建脚本${NC}"
        echo -e "${YELLOW}请先手动编译前端：cd frontend && npm run build${NC}"
        exit 1
    fi
fi

# 创建 Dockerfile
echo -e "${YELLOW}📝 创建 Dockerfile...${NC}"
cat > Dockerfile << 'EOF'
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
EOF

# 构建镜像
echo -e "${YELLOW}🔨 开始构建镜像...${NC}"
echo -e "${YELLOW}   标签: $TAGS${NC}"
echo -e "${YELLOW}   平台: $PLATFORMS${NC}"
echo ""

# 构建 args
BUILD_ARGS="--build-arg VERSION=$VERSION --build-arg BUILD_TIME=$BUILD_TIME --build-arg GIT_COMMIT=$GIT_COMMIT"

# 本地构建模式
echo -e "${YELLOW}📦 本地构建模式：仅在本地构建镜像${NC}"

docker buildx build \
    --platform "$PLATFORMS" \
    --tag "$FULL_IMAGE_NAME:$VERSION" \
    $BUILD_ARGS \
    --load \
    . || {
    echo -e "${RED}❌ 镜像构建失败${NC}"
    exit 1
}

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  镜像构建完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  镜像: $FULL_IMAGE_NAME:$VERSION${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 显示镜像信息
echo -e "${BLUE}镜像信息:${NC}"
docker images | grep "$FULL_IMAGE_NAME" | head -3
echo ""

# 提示运行命令
echo -e "${BLUE}运行容器:${NC}"
echo -e "${GREEN}docker run -d --name $IMAGE_NAME -p ${PORT}:8902 -v \$(pwd)/data:/app/data $FULL_IMAGE_NAME:$VERSION${NC}"
echo ""

# 提示推送命令
echo -e "${BLUE}手动推送到仓库:${NC}"
echo -e "${GREEN}docker tag $FULL_IMAGE_NAME:$VERSION yourusername/$FULL_IMAGE_NAME:$VERSION${NC}"
echo -e "${GREEN}docker push yourusername/$FULL_IMAGE_NAME:$VERSION${NC}"
echo ""

echo -e "${BLUE}使用环境变量自定义配置:${NC}"
echo -e "  IMAGE_NAME=myapp"
echo -e "  PORT=8080"
echo ""
