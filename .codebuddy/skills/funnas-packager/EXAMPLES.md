# 飞牛 NAS 应用打包示例

## 示例 1: 基本打包

打包应用到默认输出目录：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh
```

输出：
```
输出目录: ./release/funnas/
✓ 找到: backend/lottery
✓ 找到: frontend/dist/
✓ 找到: techfunway-lottery/
✓ 创建应用目录结构
✓ 复制后端文件
✓ 复制前端资源
✓ 复制图标和资源
✓ 生成配置文件
✓ 生成启动脚本
✓ 生成版本文件
✓ 打包完成
```

## 示例 2: 打包并压缩

打包应用并创建压缩包：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --compress
```

输出：
```
✓ 创建压缩包: techfunway-lottery-v1.0.0.tar.gz
✓ 生成校验文件: techfunway-lottery-v1.0.0.tar.gz.sha256
```

## 示例 3: 指定输出目录

打包应用到指定目录：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --output /tmp/funnas-app
```

输出：
```
输出目录: /tmp/funnas-app
应用已打包到: /tmp/funnas-app/techfunway-lottery
```

## 示例 4: 指定版本打包

打包指定版本：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --version v1.2.0
```

VERSION.txt 内容：
```
Lottery Assistant - FunNAS App
Version: v1.2.0
Build Time: 2026-03-21T12:00:00Z
Git Commit: abc1234
Platform: FunNAS
Author: TechFunWay
```

## 示例 5: 打包并包含数据库

打包应用并包含数据库文件：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --with-data
```

输出：
```
✓ 复制数据库文件
```

## 示例 6: 打包并清理临时文件

打包应用并清理旧的输出：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --clean
```

输出：
```
清理输出目录...
✓ 创建应用目录结构
```

## 示例 7: 完整打包流程

完整打包流程（清理 + 压缩 + 指定版本 + 包含数据）：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh \
  --clean \
  --compress \
  --version v1.0.0 \
  --with-data \
  --output ./release
```

## 示例 8: 部署到 NAS

打包并上传到 NAS：

```bash
# 1. 打包应用
.codebuddy/skills/funnas-packager/scripts/package.sh --compress

# 2. 上传到 NAS
scp release/funnas/techfunway-lottery-v1.0.0.tar.gz admin@nas:/tmp/

# 3. SSH 登录并安装
ssh admin@nas
cd /tmp
tar -xzf techfunway-lottery-v1.0.0.tar.gz
cd techfunway-lottery

# 4. 安装应用（根据 FunNAS 实际安装方式）
sudo cp -r techfunway-lottery /opt/
sudo /opt/techfunway-lottery/app/scripts/start.sh
```

## 示例 9: 检查打包结果

检查打包后的应用结构：

```bash
# 查看应用目录
ls -la release/funnas/techfunway-lottery/

# 查看应用配置
cat release/funnas/techfunway-lottery/app/conf/app.conf

# 查看版本信息
cat release/funnas/techfunway-lottery/VERSION.txt

# 查看压缩包大小
ls -lh release/funnas/techfunway-lottery-v1.0.0.tar.gz

# 查看校验文件
cat release/funnas/techfunway-lottery-v1.0.0.tar.gz.sha256
```

## 示例 10: 验证压缩包完整性

验证压缩包的完整性：

```bash
# 生成新的校验和
sha256sum release/funnas/techfunway-lottery-v1.0.0.tar.gz

# 对比校验和
cat release/funnas/techfunway-lottery-v1.0.0.tar.gz.sha256
```

## 示例 11: 查看帮助信息

查看打包脚本的帮助信息：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --help
```

输出：
```
用法: ./package.sh [选项]

选项:
  --output DIR    输出目录 (默认: ./release/funnas/)
  --version VER   指定版本号 (默认: v1.0.0)
  --clean         清理临时文件
  --compress      创建压缩包
  --with-data     包含数据库文件
  --help         显示帮助信息
```

## 示例 12: 调试打包过程

调试打包过程中的问题：

```bash
# 启用 bash 调试模式
bash -x .codebuddy/skills/funnas-packager/scripts/package.sh

# 查看详细输出
.codebuddy/skills/funnas-packager/scripts/package.sh 2>&1 | tee package.log
```

## 示例 13: 自动化打包脚本

创建自动化打包脚本：

```bash
#!/bin/bash
# auto-package.sh

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "用法: $0 <version>"
    exit 1
fi

echo "开始打包版本 $VERSION..."

# 编译后端
echo "编译后端..."
cd backend
go build -o lottery
cd ..

# 构建前端
echo "构建前端..."
cd frontend
npm run build
cd ..

# 生成图标
echo "生成图标..."
python3 generate_icons.py

# 打包应用
echo "打包应用..."
.codebuddy/skills/funnas-packager/scripts/package.sh \
  --clean \
  --compress \
  --version "$VERSION" \
  --with-data

echo "打包完成！"
```

使用：
```bash
./auto-package.sh v1.0.0
```

## 示例 14: 批量打包多个版本

批量打包多个版本：

```bash
for version in v1.0.0 v1.1.0 v1.2.0; do
    echo "打包版本: $version"
    .codebuddy/skills/funnas-packager/scripts/package.sh \
      --clean \
      --compress \
      --version "$version"
done
```

## 示例 15: 创建版本发布说明

创建版本发布说明：

```bash
# 打包应用
.codebuddy/skills/funnas-packager/scripts/package.sh --compress --version v1.0.0

# 创建发布说明
cat > release/funnas/RELEASE_NOTES_v1.0.0.md <<EOF
# 彩票助手 v1.0.0 - FunNAS 版

## 发布日期

2026-03-21

## 新功能

- 智能选号功能
- 购买记录管理
- 历史查询功能
- 数据统计分析

## 安装说明

1. 下载压缩包: techfunway-lottery-v1.0.0.tar.gz
2. 校验文件完整性
3. 上传到 NAS 并安装
4. 访问 http://nas-ip:8902

## 系统要求

- FunNAS 系统
- 最低 512MB 内存
- 至少 100MB 可用磁盘空间

## 更新日志

### v1.0.0 (2026-03-21)

- 初始发布
- 完整功能实现
EOF
```
