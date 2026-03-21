#!/bin/bash

# 飞牛 NAS 应用多平台打包脚本
# 支持 ARM 和 x86 平台，从 Go 程序获取版本号

set -e

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
APP_NAME="techfunway-lottery"
FUNNAS_DIR="$PROJECT_ROOT/$APP_NAME"
RELEASE_DIR="$PROJECT_ROOT/release"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  飞牛 NAS 应用多平台打包工具${NC}"
echo -e "${BLUE}========================================${NC}"

# 步骤 1: 从 Go 程序获取版本号
echo ""
echo -e "${YELLOW}📦 步骤 1: 获取版本信息${NC}"

# 编译临时程序获取版本
TEMP_BUILD_DIR=$(mktemp -d)
cd "$PROJECT_ROOT/backend"

# 获取当前平台的版本信息
TEMP_VERSION_OUTPUT="$TEMP_BUILD_DIR/version_temp"
go run main.go -version 2>/dev/null > "$TEMP_VERSION_OUTPUT" 2>&1 || true

# 提取版本号
if grep -q "Lottery Assistant" "$TEMP_VERSION_OUTPUT"; then
    VERSION=$(grep "Lottery Assistant" "$TEMP_VERSION_OUTPUT" | awk '{print $3}')
    BUILD_TIME=$(grep "Build Time:" "$TEMP_VERSION_OUTPUT" | awk '{print $3}')
    # 从编译产物中获取 git commit
    GIT_COMMIT=$(cd "$PROJECT_ROOT" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
else
    VERSION="v1.0.0"
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    GIT_COMMIT=$(cd "$PROJECT_ROOT" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
fi

echo -e "${GREEN}✓ 版本号: $VERSION${NC}"
echo -e "${GREEN}✓ 构建时间: $BUILD_TIME${NC}"
echo -e "${GREEN}✓ Git 提交: $GIT_COMMIT${NC}"

rm -rf "$TEMP_BUILD_DIR"

# 步骤 2: 检查编译产物
echo ""
echo -e "${YELLOW}📦 步骤 2: 检查编译产物${NC}"

PLATFORMS=("linux-arm64" "linux-amd64")
BINARY_FILES=()

for PLATFORM in "${PLATFORMS[@]}"; do
    BINARY_PATH="$PROJECT_ROOT/release/$VERSION/lottery-assistant-$VERSION-$PLATFORM/lottery"
    if [ "$PLATFORM" = "linux-amd64" ]; then
        BINARY_NAME="lottery"
    else
        BINARY_NAME="lottery"
    fi
    
    FULL_PATH="$PROJECT_ROOT/release/$VERSION/lottery-assistant-$VERSION-$PLATFORM/$BINARY_NAME"
    
    if [ ! -f "$FULL_PATH" ]; then
        echo -e "${RED}❌ 错误: 未找到 $PLATFORM 的编译产物${NC}"
        echo -e "${YELLOW}请先运行跨平台编译: .codebuddy/skills/cross-platform-compile/scripts/compile.sh${NC}"
        exit 1
    fi
    
    BINARY_FILES+=("$FULL_PATH:$PLATFORM")
    echo -e "${GREEN}✓ 找到: $PLATFORM${NC}"
done

# 步骤 3: 检查前端资源
echo ""
echo -e "${YELLOW}📦 步骤 3: 检查前端资源${NC}"

FRONTEND_DIST="$PROJECT_ROOT/frontend/dist/lottery-web"
if [ ! -d "$FRONTEND_DIST" ]; then
    echo -e "${RED}❌ 错误: 前端资源不存在: $FRONTEND_DIST${NC}"
    echo -e "${YELLOW}请先编译前端: cd frontend && npm run build${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 前端资源: $FRONTEND_DIST${NC}"

# 步骤 4: 检查飞牛应用目录
echo ""
echo -e "${YELLOW}📦 步骤 4: 检查飞牛应用目录${NC}"

if [ ! -d "$FUNNAS_DIR" ]; then
    echo -e "${RED}❌ 错误: 飞牛应用目录不存在: $FUNNAS_DIR${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 飞牛应用目录: $FUNNAS_DIR${NC}"

# 步骤 5: 打包每个平台
echo ""
echo -e "${YELLOW}📦 步骤 5: 打包各平台应用${NC}"

mkdir -p "$RELEASE_DIR"

for BINARY_FILE in "${BINARY_FILES[@]}"; do
    IFS=':' read -r BINARY_PATH PLATFORM <<< "$BINARY_FILE"

    # 确定平台类型
    if [[ "$PLATFORM" == *"arm"* ]]; then
        PLATFORM_TYPE="arm"
    else
        PLATFORM_TYPE="x86"
    fi

    OUTPUT_DIR="$PROJECT_ROOT/.temp-$APP_NAME-$PLATFORM_TYPE"
    PLATFORM_DIR="$PROJECT_ROOT/release/$VERSION/lottery-assistant-$VERSION-$PLATFORM"

    echo ""
    echo -e "${BLUE}  正在打包: $PLATFORM_TYPE ($PLATFORM)${NC}"

    # 清理并创建临时输出目录
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"

    # 复制基础结构
    cp -r "$FUNNAS_DIR/"* "$OUTPUT_DIR/"

    # 创建 app/server 目录
    mkdir -p "$OUTPUT_DIR/app/server"

    # 复制后端二进制到 app/server
    cp "$BINARY_PATH" "$OUTPUT_DIR/app/server/lottery"
    chmod +x "$OUTPUT_DIR/app/server/lottery"

    # 创建 app/www 目录
    mkdir -p "$OUTPUT_DIR/app/www"

    # 复制前端资源到 app/www
    # index.html 和 lottery-web 从编译产物复制
    if [ -f "$PLATFORM_DIR/index.html" ]; then
        cp "$PLATFORM_DIR/index.html" "$OUTPUT_DIR/app/www/"
    fi
    if [ -d "$PLATFORM_DIR/lottery-web" ]; then
        cp -r "$PLATFORM_DIR/lottery-web" "$OUTPUT_DIR/app/www/"
    fi

    # 修改 manifest 中的 platform
    MANIFEST="$OUTPUT_DIR/manifest"
    if [ -f "$MANIFEST" ]; then
        sed "s/platform.*/platform=$PLATFORM_TYPE/" "$MANIFEST" > "$MANIFEST.tmp"
        mv "$MANIFEST.tmp" "$MANIFEST"
    fi

    # 生成 VERSION.txt
    cat > "$OUTPUT_DIR/VERSION.txt" << EOF
Lottery Assistant - FunNAS App
Version: $VERSION
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT
Platform: $PLATFORM_TYPE ($PLATFORM)
Author: TechFunWay
EOF

    # 打包成 .fpk
    cd "$OUTPUT_DIR"
    fnpack build
    FPK_FILE="$APP_NAME.fpk"

    if [ -f "$FPK_FILE" ]; then
        FINAL_NAME="$APP_NAME-$VERSION-$PLATFORM_TYPE.fpk"
        mv "$FPK_FILE" "$FINAL_NAME"
        mkdir -p "$RELEASE_DIR/$VERSION"
        cp "$FINAL_NAME" "$RELEASE_DIR/$VERSION/"

        SIZE=$(du -h "$RELEASE_DIR/$VERSION/$FINAL_NAME" | cut -f1)
        echo -e "${GREEN}  ✓ 完成: $FINAL_NAME ($SIZE)${NC}"
    else
        echo -e "${RED}  ❌ 打包失败: $PLATFORM_TYPE${NC}"
    fi

    # 清理临时目录
    cd "$PROJECT_ROOT"
    rm -rf "$OUTPUT_DIR"
done

# 完成
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有平台打包完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}输出目录: $RELEASE_DIR/$VERSION${NC}"
echo ""
echo -e "${BLUE}生成的 .fpk 文件:${NC}"
ls -lh "$RELEASE_DIR/$VERSION"/*.fpk 2>/dev/null || echo "  (未找到 .fpk 文件)"
echo ""
echo -e "${GREEN}========================================${NC}"
