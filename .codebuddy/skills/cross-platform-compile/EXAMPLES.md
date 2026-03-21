# 使用示例

## 示例 1: 基本编译

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
```

输出:
```
========================================
  跨平台编译
========================================
  版本: v1.0.0
  压缩: false
========================================

📦 构建时间: 2026-03-20T16:30:33Z
🔖 Git 提交: d80a8e9

✓ 前端资源已复制到 embed/dist/
🔨 编译: darwin/amd64
✓ 完成: lottery-assistant ( 12M)
🔨 编译: darwin/arm64
✓ 完成: lottery-assistant ( 11M)
...
========================================
  所有平台编译完成！
========================================
  输出目录: release/v1.0.0/
  发行说明: release/v1.0.0/RELEASE_NOTES.md
========================================

编译的平台:
  ✓ darwin/amd64
  ✓ darwin/arm64
  ✓ linux/amd64
  ✓ linux/arm64
  ✓ windows/amd64
  ✓ windows/arm64
```

## 示例 2: 发布新版本

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.1.0
```

在Git创建标签后编译：

```bash
# 创建版本标签
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# 编译
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.1.0
```

## 示例 3: 编译并压缩

```bash
# 先安装UPX
brew install upx  # macOS
# 或 sudo apt-get install upx  # Linux

# 编译并压缩
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.0.0 true
```

压缩效果:
```
darwin-amd64: 12M → 6M
darwin-arm64: 11M → 5M
linux-amd64:   11M → 5M
windows-amd64: 12M → 6M
```

## 示例 4: 仅修改版本号重新编译

```bash
# 查看当前Git commit
git log -1 --oneline

# 使用当前commit作为版本
COMMIT_SHA=$(git rev-parse --short HEAD)
.codebuddy/skills/cross-platform-compile/scripts/compile.sh "v1.0.0-$COMMIT_SHA"
```

生成文件名: `lottery-assistant-v1.0.0-d80a8e9-linux-amd64`

## 示例 5: 检查编译产物

```bash
# 查看编译产物
ls -lh release/v1.0.0/

# 查看版本信息
cat release/v1.0.0/lottery-assistant-v1.0.0-darwin-amd64/VERSION.txt

# 测试运行
cd release/v1.0.0/lottery-assistant-v1.0.0-darwin-amd64
./lottery-assistant
```

## 示例 6: 部署到服务器

```bash
# 编译Linux版本
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.0.0

# 上传到服务器
scp -r release/v1.0.0/lottery-assistant-v1.0.0-linux-amd64 user@server:/opt/

# SSH登录并运行
ssh user@server
cd /opt/lottery-assistant-v1.0.0-linux-amd64
chmod +x lottery-assistant
./lottery-assistant
```

## 示例 7: 在Docker中使用

```dockerfile
# Dockerfile
FROM alpine:latest

COPY release/v1.0.0/lottery-v1.0.0-linux-amd64/lottery /app/

EXPOSE 8902

WORKDIR /app
CMD ["./lottery"]
```

构建镜像:
```bash
docker build -t lottery:v1.0.0 .
```

## 示例 8: 创建发布包

```bash
# 编译
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.0.0 true

# 为每个平台创建zip包
cd release/v1.0.0
for dir in lottery-assistant-v1.0.0-*; do
    zip -r "${dir}.zip" "$dir"
done

# 查看生成的zip包
ls -lh *.zip
```

## 示例 9: 集成到CI/CD

GitHub Actions 示例:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Compile all platforms
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          .codebuddy/skills/cross-platform-compile/scripts/compile.sh $VERSION true
      
      - name: Upload Release
        uses: softprops/action-gh-release@v1
        with:
          files: release/${{ github.ref_name }}/*
```

## 示例 10: 对比版本大小

```bash
# 编译两个版本
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.0.0 false
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.0.0 true

# 对比大小
echo "未压缩:"
du -h release/v1.0.0/lottery-assistant-v1.0.0-darwin-amd64/lottery-assistant

echo "压缩后:"
du -h release/v1.0.0-compressed/lottery-assistant-v1.0.0-darwin-amd64/lottery-assistant
```
