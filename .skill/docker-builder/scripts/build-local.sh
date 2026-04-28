#!/bin/bash

# Docker 本地镜像构建脚本（使用 scratch 镜像，不依赖外部镜像）
# 需要先运行跨平台编译

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$PROJECT_ROOT"

VERSION="${1:-v1.0.0}"
IMAGE_NAME="techfunways/lottery"

echo "========================================"
echo "  Docker 本地镜像构建（scratch）"
echo "========================================"
echo "  版本: $VERSION"
echo "  镜像: $IMAGE_NAME"
echo "========================================"

# 检查编译产物
BINARY_PATH="release/$VERSION/lottery-assistant-$VERSION-linux-amd64/lottery"
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ 错误: 编译产物不存在: $BINARY_PATH"
    echo "请先运行跨平台编译：.skill/cross-platform-compile/scripts/compile.sh"
    exit 1
fi

if [ ! -f "release/$VERSION/lottery-assistant-$VERSION-linux-amd64/index.html" ]; then
    echo "❌ 错误: 前端资源不存在"
    exit 1
fi

# 创建临时构建目录
BUILD_DIR=$(mktemp -d)
mkdir -p "$BUILD_DIR/app"

# 复制二进制文件
cp "$BINARY_PATH" "$BUILD_DIR/app/"
chmod +x "$BUILD_DIR/app/lottery"

# 复制前端文件
cp "release/$VERSION/lottery-assistant-$VERSION-linux-amd64/index.html" "$BUILD_DIR/app/"
cp -r "release/$VERSION/lottery-assistant-$VERSION-linux-amd64/lottery-web" "$BUILD_DIR/app/"
cp -r "release/$VERSION/lottery-assistant-$VERSION-linux-amd64/img" "$BUILD_DIR/app/"

# 创建 Dockerfile
cat > "$BUILD_DIR/Dockerfile" << 'EOF'
FROM scratch

COPY app /app

WORKDIR /app

EXPOSE 8902

CMD ["./lottery", "-data-dir", "/app/data", "-web-dir", "/app", "-device-type", "docker"]
EOF

# 构建镜像
echo "🔨 构建镜像..."
docker build -t "$IMAGE_NAME:$VERSION" "$BUILD_DIR"

# 清理
rm -rf "$BUILD_DIR"

echo ""
echo "========================================"
echo "  ✅ 镜像构建完成！"
echo "========================================"
echo "  镜像: $IMAGE_NAME:$VERSION"
echo "========================================"
echo ""
echo "运行容器:"
echo "  docker run -d --name lottery -p 8902:8902 -v \$(pwd)/data:/app/data $IMAGE_NAME:$VERSION"
echo ""
