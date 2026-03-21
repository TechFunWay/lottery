# Frontend Build Skill

## 概述

前端构建 skill 用于编译和打包 Lottery Assistant 项目的 Vue 3 前端，生成生产环境可用的静态资源。

## 使用方法

### 方法 1: 直接使用构建脚本

在项目根目录下执行：

```bash
.codebuddy/skills/frontend-build/scripts/build.sh
```

### 方法 2: 使用 npm 命令

在前端目录下执行：

```bash
cd frontend
npm run build
```

## 构建脚本功能

构建脚本会自动完成以下操作：

1. **依赖检查** - 检查 `node_modules` 是否存在，如不存在则自动运行 `npm install`
2. **清理旧文件** - 删除旧的 `dist` 目录
3. **类型检查** - 运行 `vue-tsc` 进行 TypeScript 类型检查
4. **生产构建** - 运行 `vite build` 生成优化后的生产版本
5. **构建统计** - 显示构建结果，包括文件大小、文件数量等

## 构建产物

构建完成后，会在 `frontend/dist/` 目录生成以下文件：

```
dist/
├── index.html           # 主 HTML 入口文件
└── assets/
    ├── index-*.js       # 主 JavaScript bundle (173.78 kB)
    ├── StatisticsView-*.js  # 统计视图 (561.75 kB)
    ├── PurchaseView-*.js   # 购买视图 (24.81 kB)
    ├── DrawView-*.js       # 开奖视图 (17.95 kB)
    ├── UserManageView-*.js # 用户管理视图 (18.77 kB)
    ├── HomeView-*.js       # 主页视图 (13.84 kB)
    ├── HistoryHitView-*.js # 历史中奖视图 (10.23 kB)
    ├── LoginView-*.js      # 登录视图 (8.81 kB)
    ├── WinningsView-*.js   # 中奖视图 (6.41 kB)
    ├── NumberInput-*.js    # 数字输入组件 (6.35 kB)
    ├── index-*.css         # 主样式文件 (31.54 kB)
    └── ...                 # 其他资源和图标
```

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite 5
- **样式**: Tailwind CSS 3
- **图表**: ECharts 6
- **路由**: Vue Router 4
- **图标**: Lucide Vue Next

## 集成方式

### 1. 直接部署

将 `dist/` 目录部署到任何静态文件服务器（如 Nginx、Apache、CDN 等）

### 2. Go Embed 集成

使用 Go embed 将前端资源嵌入到后端二进制文件中，实现单文件部署：

```go
import "embed"

//go:embed dist/*
var frontendFS embed.FS
```

然后配合 `go-cross-platform-build` skill 生成跨平台可执行文件。

## 常见问题

### 1. 构建失败：TypeScript 类型错误

检查代码中的类型定义是否正确，确保所有 API 调用的类型匹配。

### 2. 构建失败：缺少依赖

如果提示缺少模块（如 autoprefixer），运行：

```bash
cd frontend
npm install -D autoprefixer
```

### 3. 大文件警告

构建时会提示某些 chunk 超过 500KB（如 StatisticsView），这是正常的，可以考虑：
- 使用动态 import() 进行代码分割
- 调整 `vite.config.ts` 中的 `build.chunkSizeWarningLimit`

### 4. 生产环境跨域问题

确保后端 CORS 配置允许生产环境的域名访问。

## 版本信息

- Skill 版本: 1.0.0
- 创建日期: 2026-03-21
- 适用项目: Lottery Assistant

## 相关 Skills

- `go-cross-platform-build`: 用于将前端打包到 Go 二进制文件中，实现跨平台部署

## 示例输出

```
========================================
  Frontend Build Script
========================================
🔨 Building frontend...

vite v5.4.21 building for production...
✓ 2397 modules transformed.
✓ built in 2.17s

✅ Build successful!
📦 Output directory: /Users/weiyi/develop/gitee/TechFunWay/lottery/frontend/dist
📊 Build Statistics:
   Total size: 920K
   File count: 21
📁 Main files:
   - index.html
   - assets/ (18 JS files, 1 CSS files)
🎉 Build completed successfully!
```
