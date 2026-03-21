#!/bin/bash

# 飞牛 NAS 应用打包脚本
# 用于将彩票助手项目打包成飞牛 NAS 应用

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-$PROJECT_ROOT/release/funnas}"
APP_NAME="techfunway-lottery"
VERSION="${VERSION:-v1.0.0}"
CLEAN=false
COMPRESS=false
WITH_DATA=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --compress)
            COMPRESS=true
            shift
            ;;
        --with-data)
            WITH_DATA=true
            shift
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --output DIR    输出目录 (默认: ./release/funnas/)"
            echo "  --version VER   指定版本号 (默认: v1.0.0)"
            echo "  --clean         清理临时文件"
            echo "  --compress      创建压缩包"
            echo "  --with-data     包含数据库文件"
            echo "  --help         显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用 --help 查看帮助"
            exit 1
            ;;
    esac
done

echo "========================================="
echo "  飞牛 NAS 应用打包工具"
echo "========================================="
echo ""
echo "项目目录: $PROJECT_ROOT"
echo "输出目录: $OUTPUT_DIR"
echo "版本: $VERSION"
echo ""

# 函数：打印步骤
print_step() {
    echo ""
    echo "📦 $1"
}

# 函数：检查文件存在
check_file() {
    if [ ! -f "$1" ]; then
        echo "❌ 错误: 文件不存在: $1"
        exit 1
    fi
    echo "✓ 找到: $1"
}

# 函数：检查目录存在
check_dir() {
    if [ ! -d "$1" ]; then
        echo "❌ 错误: 目录不存在: $1"
        exit 1
    fi
    echo "✓ 找到: $1"
}

print_step "步骤 1: 验证环境"
echo "检查必要文件和目录..."

# 检查后端
check_file "$PROJECT_ROOT/backend/lottery"

# 检查前端
check_dir "$PROJECT_ROOT/frontend/dist"

# 检查应用目录
check_dir "$PROJECT_ROOT/techfunway-lottery"

# 检查图标
check_file "$PROJECT_ROOT/techfunway-lottery/ICON.PNG"
check_file "$PROJECT_ROOT/techfunway-lottery/ICON_256.PNG"

echo ""
print_step "步骤 2: 准备应用目录"

# 清理并创建输出目录
if [ "$CLEAN" = true ] && [ -d "$OUTPUT_DIR/$APP_NAME" ]; then
    echo "清理输出目录..."
    rm -rf "$OUTPUT_DIR/$APP_NAME"
fi

APP_DIR="$OUTPUT_DIR/$APP_NAME"
mkdir -p "$APP_DIR"

# 创建应用子目录
mkdir -p "$APP_DIR/app/bin"
mkdir -p "$APP_DIR/app/conf"
mkdir -p "$APP_DIR/app/data/db"
mkdir -p "$APP_DIR/app/ui/images"
mkdir -p "$APP_DIR/app/scripts"

echo "✓ 创建应用目录结构"

print_step "步骤 3: 打包后端"

# 复制后端二进制文件
echo "复制后端二进制文件..."
cp "$PROJECT_ROOT/backend/lottery" "$APP_DIR/app/bin/lottery"
chmod +x "$APP_DIR/app/bin/lottery"
echo "✓ 复制后端文件"

# 复制数据库（如果指定）
if [ "$WITH_DATA" = true ] && [ -f "$PROJECT_ROOT/backend/data/db/database.db" ]; then
    echo "复制数据库文件..."
    cp "$PROJECT_ROOT/backend/data/db/database.db" "$APP_DIR/app/data/db/database.db"
    echo "✓ 复制数据库文件"
fi

print_step "步骤 4: 打包前端"

# 复制前端资源
echo "复制前端构建产物..."
if [ -d "$PROJECT_ROOT/frontend/dist/lottery-web" ]; then
    cp -r "$PROJECT_ROOT/frontend/dist/lottery-web" "$APP_DIR/app/ui/lottery-web"
elif [ -d "$PROJECT_ROOT/frontend/dist" ]; then
    cp -r "$PROJECT_ROOT/frontend/dist/"* "$APP_DIR/app/ui/"
else
    echo "⚠️  警告: 前端资源目录不存在"
fi

echo "✓ 复制前端资源"

print_step "步骤 5: 处理图标和资源"

# 复制应用图标
echo "复制应用图标..."
cp "$PROJECT_ROOT/techfunway-lottery/ICON.PNG" "$APP_DIR/"
cp "$PROJECT_ROOT/techfunway-lottery/ICON_256.PNG" "$APP_DIR/"

# 复制 UI 图标
if [ -d "$PROJECT_ROOT/techfunway-lottery/app/ui/images" ]; then
    cp "$PROJECT_ROOT/techfunway-lottery/app/ui/images/"*.png "$APP_DIR/app/ui/images/"
fi

echo "✓ 复制图标和资源"

print_step "步骤 6: 生成配置文件"

# 复制 manifest
if [ -f "$PROJECT_ROOT/techfunway-lottery/manifest" ]; then
    cp "$PROJECT_ROOT/techfunway-lottery/manifest" "$APP_DIR/app/"
    echo "✓ 复制 manifest"
fi

# 生成 app.conf
cat > "$APP_DIR/app/conf/app.conf" <<EOF
[app]
name=lottery-assistant
version=${VERSION#v}
port=8902

[service]
autostart=true
restart=true
max_retries=3

[database]
path=./data/db/database.db
backup_enabled=true
backup_retention=7
EOF

echo "✓ 生成 app.conf"

# 生成 service.conf
cat > "$APP_DIR/app/conf/service.conf" <<EOF
[Unit]
Description=Lottery Assistant
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/$APP_NAME/app
ExecStart=/opt/$APP_NAME/app/bin/lottery -port 8902 -dataUrl /opt/$APP_NAME/app/data
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "✓ 生成 service.conf"

print_step "步骤 7: 生成启动脚本"

# 生成启动脚本
cat > "$APP_DIR/app/scripts/start.sh" <<EOF
#!/bin/bash
# 彩票助手启动脚本

APP_DIR="\$(cd "\$(dirname "\$0")" && pwd)"
cd "\$APP_DIR/.."

echo "启动彩票助手..."
echo "工作目录: \$PWD"
echo "端口: 8902"

# 检查是否已经在运行
if pgrep -f "lottery.*8902" > /dev/null; then
    echo "彩票助手已在运行"
    exit 0
fi

# 启动服务
./bin/lottery -port 8902 -dataUrl ./data

echo "彩票助手已启动"
echo "访问地址: http://localhost:8902"
EOF

chmod +x "$APP_DIR/app/scripts/start.sh"
echo "✓ 生成 start.sh"

# 生成停止脚本
cat > "$APP_DIR/app/scripts/stop.sh" <<EOF
#!/bin/bash
# 彩票助手停止脚本

echo "停止彩票助手..."

# 停止服务
pkill -f "lottery.*8902" || echo "彩票助手未运行"

echo "彩票助手已停止"
EOF

chmod +x "$APP_DIR/app/scripts/stop.sh"
echo "✓ 生成 stop.sh"

print_step "步骤 8: 生成版本文件"

# 获取 git 信息
GIT_COMMIT=$(cd "$PROJECT_ROOT" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat > "$APP_DIR/VERSION.txt" <<EOF
Lottery Assistant - FunNAS App
Version: $VERSION
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT
Platform: FunNAS
Author: TechFunWay
EOF

echo "✓ 生成 VERSION.txt"

print_step "步骤 9: 生成 README.md"

cat > "$APP_DIR/README.md" <<EOF
# 彩票助手 - FunNAS 应用

## 版本

$VERSION

## 简介

彩票助手是一款功能完整的彩票管理工具，提供选号、购买、历史查询等功能。

## 功能特性

- 🎰 智能选号
- 📊 购买记录
- 📈 统计分析
- 🔍 历史查询
- 💾 数据备份

## 安装

1. 将应用包上传到飞牛 NAS
2. 在应用中心安装
3. 启动服务

## 使用

启动后访问: http://nas-ip:8902

## 配置

配置文件位置: \`/opt/$APP_NAME/app/conf/app.conf\`

数据库位置: \`/opt/$APP_NAME/app/data/db/database.db\`

## 脚本

- \`./scripts/start.sh\` - 启动服务
- \`./scripts/stop.sh\` - 停止服务

## 技术支持

- 作者: TechFunWay
- 许可证: MIT

## 更新日志

### $VERSION

- 初始发布版本
EOF

echo "✓ 生成 README.md"

print_step "步骤 10: 创建发布包"

if [ "$COMPRESS" = true ]; then
    echo "创建压缩包..."
    cd "$OUTPUT_DIR"
    tar -czf "${APP_NAME}-${VERSION}.tar.gz" "$APP_NAME"
    echo "✓ 创建压缩包: ${APP_NAME}-${VERSION}.tar.gz"

    # 生成校验文件
    if command -v sha256sum &> /dev/null; then
        sha256sum "${APP_NAME}-${VERSION}.tar.gz" > "${APP_NAME}-${VERSION}.tar.gz.sha256"
        echo "✓ 生成校验文件"
    fi
fi

print_step "打包完成"
echo ""
echo "应用已打包到: $APP_DIR"
echo ""
echo "目录结构:"
ls -la "$APP_DIR"
echo ""

if [ "$COMPRESS" = true ]; then
    echo "压缩包:"
    ls -lh "$OUTPUT_DIR/${APP_NAME}-${VERSION}.tar.gz"
    echo ""
fi

echo "========================================="
echo "  打包成功！"
echo "========================================="
