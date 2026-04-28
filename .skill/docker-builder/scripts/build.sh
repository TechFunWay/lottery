#!/bin/bash

# Docker 镜像构建脚本（多架构，使用 Dockerfile）
# 自动编译前端，使用 Docker Buildx 构建多架构镜像

set -e

# 获取项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
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

# 构建前端
echo -e "${YELLOW}🔨 编译前端...${NC}"
if [ -f ".skill/frontend-build/scripts/build.sh" ]; then
    .skill/frontend-build/scripts/build.sh || {
        echo -e "${RED}❌ 前端编译失败${NC}"
        exit 1
    }
else
    echo -e "${RED}❌ 未找到前端构建脚本${NC}"
    exit 1
fi

# 创建 Dockerfile
echo -e "${YELLOW}📝 创建 Dockerfile...${NC}"
cat > Dockerfile << 'DOCKERFILE'
# 多阶段构建 - 优化镜像大小

# 阶段 1: 编译前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 阶段 2: 编译后端
FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -s -w -buildid=" \
    -o lottery .

# 阶段 3: 运行时镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

COPY --from=backend-builder --chown=appuser:appgroup /app/backend/lottery .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/index.html .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/lottery-web ./lottery-web
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/img ./img

RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app/data

USER appuser:appgroup

EXPOSE 8902

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8902/api/version || exit 1

ENTRYPOINT ["./lottery"]
CMD ["-data-dir", "/app/data", "-web-dir", "/app", "-device-type", "docker"]
DOCKERFILE

# 构建镜像
echo -e "${YELLOW}🔨 开始构建镜像...${NC}"
echo -e "${YELLOW}   平台: $PLATFORMS${NC}"
echo ""

BUILD_ARGS="--build-arg VERSION=$VERSION --build-arg BUILD_TIME=$BUILD_TIME --build-arg GIT_COMMIT=$GIT_COMMIT"

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

echo -e "${BLUE}镜像信息:${NC}"
docker images | grep "$FULL_IMAGE_NAME" | head -3
echo ""

echo -e "${BLUE}运行容器:${NC}"
echo -e "${GREEN}docker run -d --name $IMAGE_NAME -p ${PORT}:8902 -v \$(pwd)/data:/app/data $FULL_IMAGE_NAME:$VERSION${NC}"
echo ""
