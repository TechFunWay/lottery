#!/bin/bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_DIR"

VERSION_RAW=$(awk -F'"' '/^[[:space:]]*Version[[:space:]]*= / {print $2; exit}' backend/main.go)
[ -z "$VERSION_RAW" ] && echo "❌ 无法获取版本号" && exit 1
VERSION="${VERSION_RAW#v}"
APP_NAME="techfunway-lottery"
RELEASE_DIR="release/v${VERSION}"
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X main.Version=v${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w -buildid="
GOCACHE_DIR="${GOCACHE:-/tmp/lottery-fnos-go-cache}"
WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/lottery-fnpack.XXXXXX")
trap 'rm -rf "$WORK_DIR"' EXIT

command -v fnpack >/dev/null 2>&1 || echo "⚠️ 未找到 fnpack 命令，改用手工组装（结构与 fnpack build 产物一致）"

build_fpk_file() {
    # 优先用 fnpack CLI；没有时手工组装：app/ 打成 app.tgz（条目不带 app/ 前缀），
    # 再与其余文件按同顺序合成外层 tar.gz（fpk 格式逆向自 rental 真产物，参照 bookmarks 兜底）
    local package_dir=$1 out_file=$2
    if command -v fnpack >/dev/null 2>&1; then
        (cd "$package_dir" && fnpack build)
        mv "$package_dir/${APP_NAME}.fpk" "$out_file"
        return
    fi
    local app_entries
    app_entries=$(ls -A "$package_dir/app")
    (cd "$package_dir/app" && COPYFILE_DISABLE=1 tar czf ../app.tgz $app_entries)
    (cd "$package_dir" && COPYFILE_DISABLE=1 tar czf fpk-outer.tar.gz \
        app.tgz cmd config ICON.PNG ICON_256.PNG manifest wizard)
    rm -f "$package_dir/app.tgz"
    mv "$package_dir/fpk-outer.tar.gz" "$out_file"
}
mkdir -p "$RELEASE_DIR"

echo "============================================"
echo "  仅构建飞牛 fnOS 安装包 v${VERSION}"
echo "============================================"

echo "📦 构建飞牛前端..."
VITE_FNOS_APP=true npm --prefix frontend run build -- --base="/app/${APP_NAME}/"

build_fnpack() {
    local arch=$1 platform=$2 label=$3
    local binary="${WORK_DIR}/lottery-linux-${arch}"
    local package_dir="${WORK_DIR}/package-${arch}"

    echo "🔨 编译 linux/${arch}..."
    (
        cd backend
        GOCACHE="$GOCACHE_DIR" CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
            go build -trimpath -ldflags="$LDFLAGS" -o "$binary" .
    )

    mkdir -p "$package_dir"
    cp -R techfunway-lottery/. "$package_dir/"
    rm -rf "$package_dir/app/server" "$package_dir/app/www"
    mkdir -p "$package_dir/app/server" "$package_dir/app/www"
    cp "$binary" "$package_dir/app/server/lottery"
    chmod +x "$package_dir/app/server/lottery"
    cp -R frontend/dist/. "$package_dir/app/www/"
    sed -e "s/^version.*/version               = ${VERSION}/" \
        -e "s/^platform.*/platform              = ${platform}/" \
        "$package_dir/manifest" > "$package_dir/manifest.tmp"
    mv "$package_dir/manifest.tmp" "$package_dir/manifest"
    find "$package_dir" -name '.DS_Store' -delete
    printf 'Lottery Assistant - fnOS\nVersion: v%s\nBuild Time: %s\nGit Commit: %s\nPlatform: %s\n' \
        "$VERSION" "$BUILD_TIME" "$GIT_COMMIT" "$platform" > "$package_dir/VERSION.txt"

    build_fpk_file "$package_dir" "$RELEASE_DIR/${APP_NAME}-v${VERSION}-${label}.fpk"
    echo "✅ ${APP_NAME}-v${VERSION}-${label}.fpk"
}

build_fnpack amd64 x86 fnos-x86
build_fnpack arm64 arm fnos-arm64

echo "✅ 飞牛安装包已输出到 ${RELEASE_DIR}/"
ls -lh "$RELEASE_DIR"/*.fpk
