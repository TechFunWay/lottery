#!/bin/bash

# 跨平台编译脚本
# 用于编译Go项目的所有平台版本

set -e

# 获取脚本所在目录的父目录（项目根目录）
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
COMPRESS="${1:-true}"
PROJECT_NAME="${PROJECT_NAME:-lottery-assistant}"
BINARY_NAME="${BINARY_NAME:-lottery}"

# 自动获取版本号
if [ -f "backend/main.go" ]; then
    # 编译临时程序获取版本
    TEMP_BUILD_DIR=$(mktemp -d)
    TEMP_VERSION_OUTPUT="$TEMP_BUILD_DIR/version_temp"
    cd backend
    go run main.go -version 2>/dev/null > "$TEMP_VERSION_OUTPUT" 2>&1 || true
    cd ..

    # 提取版本号
    if grep -q "Lottery Assistant" "$TEMP_VERSION_OUTPUT"; then
        VERSION=$(grep "Lottery Assistant" "$TEMP_VERSION_OUTPUT" | awk '{print $3}')
        BUILD_TIME=$(grep "Build Time:" "$TEMP_VERSION_OUTPUT" | awk '{print $3}')
    else
        echo -e "${RED}❌ 无法从 backend/main.go 获取版本号${NC}"
        echo -e "${YELLOW}请确保 backend/main.go 中的 Version 变量定义正确${NC}"
        rm -rf "$TEMP_BUILD_DIR"
        exit 1
    fi

    rm -rf "$TEMP_BUILD_DIR"
else
    echo -e "${RED}❌ 未找到 backend/main.go 文件${NC}"
    echo -e "${YELLOW}请确保在项目根目录下存在 backend/main.go 文件${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  跨平台编译${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  版本: $VERSION${NC}"
echo -e "${BLUE}  压缩: $COMPRESS${NC}"
echo -e "${BLUE}========================================${NC}"

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

# 检查前端
if [ ! -d "frontend/dist" ]; then
    echo -e "${YELLOW}⚠️  前端未编译，开始编译前端...${NC}"

    # 检查前端构建脚本
    if [ -f ".skill/frontend-build/scripts/build.sh" ]; then
        .skill/frontend-build/scripts/build.sh || {
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

# 准备前端资源
echo -e "${GREEN}✓ 前端资源将随编译产物一起发布${NC}"

# 编译参数
# -s: 去除符号表
# -w: 去除DWARF调试信息
# -trimpath: 去除文件系统路径（去除开发环境路径）
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

    # 压缩
    if command -v upx &> /dev/null; then
        echo -e "${YELLOW}📦 使用 UPX 压缩: $OUTPUT_NAME${NC}"
        upx --best --lzma "$OUTPUT_PATH/$OUTPUT_NAME" 2>/dev/null || true
    else
        echo -e "${YELLOW}ℹ️  UPX 未安装，使用 Go 编译优化${NC}"
        echo -e "${YELLOW}    编译参数已包含 -s -w -trimpath -buildid= 用于减小体积${NC}"
        echo -e "${YELLOW}    如需进一步压缩，请安装 UPX: https://upx.github.io/${NC}"
    fi

    # 创建版本信息文件
    cat > "$OUTPUT_PATH/VERSION.txt" << EOF
$PROJECT_NAME
Version: $VERSION
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT
Platform: $GOOS $GOARCH
EOF

    # 显示大小
    SIZE=$(du -h "$OUTPUT_PATH/$OUTPUT_NAME" | cut -f1)
    echo -e "${GREEN}✓ 完成: $OUTPUT_NAME ($SIZE)${NC}"
done

# 创建发行说明
cat > "$OUTPUT_DIR/RELEASE_NOTES.md" << EOF
# $PROJECT_NAME $VERSION

## 版本信息

- **版本号**: $VERSION
- **构建时间**: $BUILD_TIME
- **Git 提交**: $GIT_COMMIT

## 下载链接

根据您的操作系统下载对应的版本：

### macOS
- **Intel (x86_64)**: \`$PROJECT_NAME-$VERSION-darwin-amd64/$BINARY_NAME\`
- **Apple Silicon (M1/M2/M3)**: \`$PROJECT_NAME-$VERSION-darwin-arm64/$BINARY_NAME\`

### Linux
- **x86_64**: \`$PROJECT_NAME-$VERSION-linux-amd64/$BINARY_NAME\`
- **ARM64**: \`$PROJECT_NAME-$VERSION-linux-arm64/$BINARY_NAME\`

### Windows
- **x86_64**: \`$PROJECT_NAME-$VERSION-windows-amd64/$BINARY_NAME.exe\`
- **ARM64**: \`$PROJECT_NAME-$VERSION-windows-arm64/$BINARY_NAME.exe\`

## 使用方法

### 启动程序

\`\`\`bash
# macOS/Linux
cd $PROJECT_NAME-$VERSION-<platform>
chmod +x $BINARY_NAME
./$BINARY_NAME

# Windows
cd $PROJECT_NAME-$VERSION-<platform>
$BINARY_NAME.exe
\`\`\`

### 访问应用

程序启动后，在浏览器访问:

\`\`\`
http://localhost:8902
\`\`\`

### 查看版本

\`\`\`bash
# 命令行
./$BINARY_NAME

# API
curl http://localhost:8902/api/version
\`\`\`

## 注意事项

1. 发布包包含 lottery-web/ 文件夹，需保持目录结构完整
2. 首次运行会自动创建数据库文件
3. 默认端口为 8902，可通过环境变量 PORT 修改
4. 数据库默认路径为 ./data/database.db
5. 前端更新时只需替换 lottery-web/ 文件夹内容，无需重新编译后端

## 系统要求

- **macOS**: 10.15+ (Catalina 或更高版本)
- **Linux**: 任意支持 x86_64 或 ARM64 的发行版
- **Windows**: Windows 10 或更高版本
EOF

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有平台编译完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  输出目录: $OUTPUT_DIR/${NC}"
echo -e "${BLUE}  发行说明: $OUTPUT_DIR/RELEASE_NOTES.md${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo -e "${BLUE}编译的平台:${NC}"
for PLATFORM in "${PLATFORMS[@]}"; do
    echo -e "  ✓ $PLATFORM"
done
echo ""
