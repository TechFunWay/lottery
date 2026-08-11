#!/bin/bash

# 彩彩助手发版打包脚本
# 一键完成：前端构建 → 跨平台编译 → zip打包 → 飞牛fpk打包
# 用法: bash scripts/build-release.sh [版本号]

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
PROJECT_NAME="${PROJECT_NAME:-lottery-assistant}"
BINARY_NAME="${BINARY_NAME:-lottery}"

# 获取版本号
if [ -n "$1" ]; then
    VERSION="$1"
else
    # 从 main.go 自动读取版本号
    if [ -f "backend/main.go" ]; then
        VERSION=$(grep -o 'Version = "v[^"]*"' backend/main.go | head -1 | cut -d'"' -f2)
        if [ -z "$VERSION" ]; then
            echo -e "${RED}❌ 无法从 backend/main.go 获取版本号${NC}"
            echo -e "${YELLOW}请确保 backend/main.go 中的 Version 变量定义正确${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ 未找到 backend/main.go 文件${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  彩彩助手发版打包${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  版本: $VERSION${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 版本信息
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}📦 构建时间: $BUILD_TIME${NC}"
echo -e "${GREEN}🔖 Git 提交: $GIT_COMMIT${NC}"
echo ""

# 清理旧的 release 产物
if [ -d "release/$VERSION" ]; then
    echo -e "${YELLOW}🧹 清理旧编译产物 release/$VERSION...${NC}"
    rm -rf "release/$VERSION"
    echo -e "${GREEN}✅ 清理完成${NC}"
    echo ""
fi

# ============================================
# 步骤 1: 前端构建
# ============================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  步骤 1/4: 前端构建${NC}"
echo -e "${BLUE}========================================${NC}"

if [ -d "frontend/dist" ]; then
    echo -e "${YELLOW}⚠️  前端已编译，跳过构建${NC}"
else
    echo -e "${GREEN}🔨 构建前端...${NC}"
    cd frontend
    npm run build
    cd ..
    echo -e "${GREEN}✅ 前端构建完成${NC}"
fi
echo ""

# ============================================
# 步骤 2: 跨平台编译
# ============================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  步骤 2/4: 跨平台编译${NC}"
echo -e "${BLUE}========================================${NC}"

# 编译参数
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -s -w -buildid="

# 支持的平台
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

# 输出目录
OUTPUT_DIR="release/$VERSION"
mkdir -p "$OUTPUT_DIR"

# 编译每个平台
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"

    echo -e "${YELLOW}🔨 编译: $PLATFORM${NC}"

    OUTPUT_NAME="$BINARY_NAME"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="$OUTPUT_NAME.exe"
    fi

    OUTPUT_PATH="$OUTPUT_DIR/$PROJECT_NAME-$VERSION-$GOOS-$GOARCH"
    mkdir -p "$OUTPUT_PATH"

    # 编译（ENV=production 关闭 Gin 调试日志）
    cd backend
    CGO_ENABLED=0 ENV=production GOOS="$GOOS" GOARCH="$GOARCH" \
        go build -trimpath -ldflags="$LDFLAGS" -o "../$OUTPUT_PATH/$OUTPUT_NAME" .
    cd ..

    # 复制前端资源到每个平台
    cp frontend/dist/index.html "$OUTPUT_PATH/"
    if [ -d "frontend/dist/lottery-web" ]; then
        cp -r frontend/dist/lottery-web "$OUTPUT_PATH/"
    fi
    if [ -d "frontend/dist/img" ]; then
        cp -r frontend/dist/img "$OUTPUT_PATH/"
    fi

    # 生成 VERSION.txt
    cat > "$OUTPUT_PATH/VERSION.txt" << EOF
Lottery Assistant
Version: $VERSION
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT
Platform: $GOOS $GOARCH
EOF

    echo -e "${GREEN}✅ $PLATFORM 编译完成${NC}"
done
echo ""

# ============================================
# 步骤 3: zip 打包
# ============================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  步骤 3/4: zip 打包${NC}"
echo -e "${BLUE}========================================${NC}"

cd "$OUTPUT_DIR"
for dir in $PROJECT_NAME-$VERSION-*; do
    if [ -d "$dir" ]; then
        echo -e "${YELLOW}📦 打包: $dir.zip${NC}"
        zip -r "$dir.zip" "$dir" > /dev/null
        echo -e "${GREEN}✅ $dir.zip 打包完成${NC}"
    fi
done
cd "$PROJECT_ROOT"
echo ""

# ============================================
# 步骤 4: 飞牛 fpk 打包
# ============================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  步骤 4/4: 飞牛 fpk 打包${NC}"
echo -e "${BLUE}========================================${NC}"

if command -v fnpack &> /dev/null; then
    echo -e "${GREEN}📦 开始飞牛 fpk 打包...${NC}"

    # 调用 fnnas-packager 脚本
    if [ -f "scripts/package-multiplatform.sh" ]; then
        bash scripts/package-multiplatform.sh "$VERSION" || {
            echo -e "${YELLOW}⚠️  飞牛 fpk 打包失败，跳过${NC}"
        }
    else
        echo -e "${YELLOW}⚠️  未找到 fnnas-packager 脚本，跳过 fpk 打包${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  fnpack 未安装，跳过飞牛 fpk 打包${NC}"
    echo -e "${YELLOW}   安装方法: https://www.fnnas.com/docs/dev/fnpack${NC}"
fi
echo ""

# ============================================
# 生成 RELEASE_NOTES.md
# ============================================
echo -e "${GREEN}📝 生成 RELEASE_NOTES.md...${NC}"

cat > "$OUTPUT_DIR/RELEASE_NOTES.md" << EOF
# Lottery Assistant v$VERSION

## 下载

### 通用平台包
- [macOS Intel (amd64)](./$PROJECT_NAME-$VERSION-darwin-amd64.zip)
- [macOS Apple Silicon (arm64)](./$PROJECT_NAME-$VERSION-darwin-arm64.zip)
- [Linux amd64](./$PROJECT_NAME-$VERSION-linux-amd64.zip)
- [Linux arm64](./$PROJECT_NAME-$VERSION-linux-arm64.zip)
- [Windows amd64](./$PROJECT_NAME-$VERSION-windows-amd64.zip)
- [Windows arm64](./$PROJECT_NAME-$VERSION-windows-arm64.zip)

### 飞牛 NAS
- [amd64 (.fpk)](./techfunway.lottery-$VERSION-x86.fpk)
- [arm64 (.fpk)](./techfunway.lottery-$VERSION-arm64.fpk)

## 使用方法

### macOS/Linux
\`\`\`bash
cd $PROJECT_NAME-$VERSION-<platform>
chmod +x $BINARY_NAME
./$BINARY_NAME
\`\`\`

### Windows
\`\`\`cmd
cd $PROJECT_NAME-$VERSION-<platform>
$BINARY_NAME.exe
\`\`\`

访问 http://localhost:8902 使用应用

## 系统要求

- macOS 10.15+ / Linux / Windows 10+
- 无需安装 Go 或 Node.js（已内置）

## 版本信息

- 版本: $VERSION
- 构建时间: $BUILD_TIME
- Git 提交: $GIT_COMMIT
EOF

echo -e "${GREEN}✅ RELEASE_NOTES.md 生成完成${NC}"
echo ""

# ============================================
# 完成
# ============================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}🎉 发版打包完成！${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}📦 产物目录: release/$VERSION/${NC}"
echo -e "${GREEN}📋 发布说明: release/$VERSION/RELEASE_NOTES.md${NC}"
echo ""
echo -e "${YELLOW}下一步:${NC}"
echo -e "  1. 检查 release/$VERSION/ 目录中的产物"
echo -e "  2. 上传到 Gitee/GitHub Release"
echo -e "  3. 上传 fpk 到飞牛应用商店（如有）"
echo -e "  4. 推送 Docker 镜像（如有）"
