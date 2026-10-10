# Lottery Assistant 构建系统

本项目提供了完整的自动化构建系统，支持开发环境和发布环境的构建。

## 快速开始

### 1. 开发环境（热重载）

```bash
# 方法1: 使用 Makefile
make dev

# 方法2: 使用启动脚本
./start.sh dev
```

这将：
1. 构建前端（Vue 3 + TypeScript）
2. 构建后端（Go）
3. 启动服务（端口 8902）

### 2. 发布环境（多平台）

```bash
# 方法1: 使用 Makefile
make release

# 方法2: 使用启动脚本
./start.sh release
```

这将：
1. 构建前端
2. 跨平台编译（macOS, Linux, Windows）
3. 打包飞牛 NAS 应用（.fpk 文件）

### 3. 清理构建产物

```bash
# 方法1: 使用 Makefile
make clean

# 方法2: 使用启动脚本
./start.sh clean
```

## 详细命令

### Makefile 命令

```bash
# 显示帮助信息
make help

# 仅构建前端
make dev-frontend

# 仅构建后端
make dev-backend
make backend-build

# 仅运行服务（不构建）
make run

# 启动前端开发服务器（热重载）
make frontend-dev

# 快速发布（使用现有脚本）
make release-quick

# 检查依赖
make check-deps

# 显示版本信息
make version

# 清理构建产物
make clean
```

### 启动脚本命令

```bash
# 显示帮助信息
./start.sh help

# 开发环境
./start.sh dev

# 发布环境
./start.sh release

# 清理构建产物
./start.sh clean
```

## 环境变量

可以通过环境变量自定义配置：

```bash
# 设置服务端口
PORT=8080 make dev

# 设置数据目录
DATA_DIR=/path/to/data make dev

# 设置前端目录
WEB_DIR=./frontend/dist make dev

# 组合使用
PORT=8080 DATA_DIR=/data WEB_DIR=./dist make dev
```

## 构建产物

### 开发环境
1. `frontend/dist/` - 前端构建产物
2. `./lottery` - 后端二进制文件

### 发布环境
1. `release/{version}/lottery-assistant-{version}-{platform}/` - 各平台编译产物
2. `release/{version}/*.fpk` - 飞牛 NAS 应用包
3. `release/{version}/RELEASE_NOTES.md` - 发行说明

## 支持的平台

### 后端编译
- macOS (Intel x86_64)
- macOS (Apple Silicon ARM64)
- Linux (x86_64)
- Linux (ARM64)
- Windows (x86_64)
- Windows (ARM64)

### 飞牛 NAS 应用
- Linux ARM64
- Linux x86_64

## 依赖要求

### 必需依赖
1. **Node.js** (v16+)
2. **npm** 或 **yarn**
3. **Go** (v1.20+)

### 可选依赖
1. **fnpack** - 飞牛应用打包工具
2. **upx** - 二进制压缩工具

### 检查依赖
```bash
make check-deps
```

## 项目结构

```
lottery/
├── frontend/          # Vue 3 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── backend/           # Go 后端
│   ├── main.go
│   ├── handlers/
│   ├── services/
│   └── migrations/
├── techfunway-lottery/ # 飞牛应用配置
│   └── manifest
├── release/           # 发布产物
├── Makefile          # 构建脚本
├── start.sh          # 启动脚本
├── pack.sh           # 完整打包脚本
└── BUILD_README.md   # 本文档
```

## 常见问题

### 1. 前端构建失败
```bash
# 清理并重新安装依赖
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

### 2. 后端构建失败
```bash
# 更新 Go 模块
cd backend
go mod tidy
go build
```

### 3. 飞牛应用打包失败
确保已安装 `fnpack` 工具：
```bash
# 安装 fnpack
# 请参考飞牛官方文档
```

### 4. 端口被占用
```bash
# 使用其他端口
PORT=8903 make dev
```

## 版本更新

版本号定义在 `backend/main.go` 中：
```go
var (
    Version   = "v1.0.0"
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

构建时会自动注入构建时间和 Git 提交信息。

## 升级系统

项目包含完整的升级系统，支持从任意旧版本升级到最新版本：
- 升级脚本：`backend/migrations/upgrade_scripts.go`
- 升级服务：`backend/services/upgrade_service.go`
- 升级 API：`backend/handlers/version_handler.go`

启动时会自动检查并执行必要的升级脚本。

## 故障排除

### 查看日志
```bash
# 运行服务时查看详细日志
DEBUG=true make dev
```

### 重置数据库
```bash
# 删除数据目录重新开始
rm -rf data/
make dev
```

### 检查版本
```bash
# 查看当前版本
make version
```

## 联系支持

如有问题，请参考：
1. [项目 README.md](./README.md)
2. [升级文档 UPGRADE.md](./UPGRADE.md)
3. [前端功能文档 FRONTEND_VERSION_FEATURE.md](./FRONTEND_VERSION_FEATURE.md)