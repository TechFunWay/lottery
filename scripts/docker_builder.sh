#!/bin/bash

set -e

# 获取版本号
VERSION=$(grep -E '^\s+Version\s*=\s*"' backend/main.go | head -1 | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: 无法从 backend/main.go 中获取版本号"
    exit 1
fi

echo "使用版本号: $VERSION"

# 检查 amd64 可执行文件
AMD64_DIR="release/${VERSION}/lottery-assistant-${VERSION}-linux-amd64"
if [ ! -f "$AMD64_DIR/lottery" ]; then
    echo "Error: 缺少 linux-amd64 可执行文件: $AMD64_DIR/lottery"
    echo "请先运行 make release"
    exit 1
fi

# 检查 arm64 可执行文件
ARM64_DIR="release/${VERSION}/lottery-assistant-${VERSION}-linux-arm64"
if [ ! -f "$ARM64_DIR/lottery" ]; then
    echo "Error: 缺少 linux-arm64 可执行文件: $ARM64_DIR/lottery"
    echo "请先运行 make release"
    exit 1
fi

# 检查前端文件
if [ ! -f "$AMD64_DIR/index.html" ] || [ ! -d "$AMD64_DIR/lottery-web" ]; then
    echo "Error: 前端文件不完整 (index.html 或 lottery-web 缺失)"
    exit 1
fi

echo "可执行文件和前端文件检查通过"

# 镜像名称
REPO_NAME="techfunways/lottery"

echo ""
echo "================================================"
echo " 构建模式: 多架构本地构建（不推送）"
echo " 版本: $VERSION"
echo " 架构: linux/amd64, linux/arm64"
echo " 镜像: ${REPO_NAME}"
echo "================================================"
echo ""

# 使用默认 docker driver (docker buildx 自动识别 containerd)
docker buildx use default

# 多架构构建（--load 加载到本地 docker images）
echo "使用 Docker Buildx 构建多架构镜像..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --build-arg VERSION="${VERSION}" \
    --load \
    -t "${REPO_NAME}:${VERSION}" \
    -t "${REPO_NAME}:latest" \
    -f Dockerfile.release \
    .

echo ""
echo "================================================"
echo " 构建完成！"
echo "================================================"
echo ""
echo "镜像:"
docker images --filter "reference=${REPO_NAME}" --format "  - {{.Repository}}:{{.Tag}} ({{.Size}})"
echo ""

echo "使用方法:"
echo "  docker run -d -p 8902:8902 ${REPO_NAME}:latest"
echo ""
echo "注意: Docker 会自动选择适合当前架构的镜像版本"
