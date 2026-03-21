# Cross-Platform Compile Skill

用于跨平台编译Go项目的Skill，支持自动编译前端、嵌入资源并生成多平台二进制文件。

## 功能特性

- ✅ 自动编译前端项目
- ✅ 将前端资源嵌入到Go二进制文件中
- ✅ 支持多平台交叉编译（macOS/Linux/Windows，amd64/arm64）
- ✅ 生成版本信息和发布说明
- ✅ 可选的二进制文件压缩（使用UPX）
- ✅ 统一的输出目录结构，便于分发

## 支持的平台

| 操作系统 | 架构 | 目标设备 |
|---------|------|---------|
| macOS | amd64 | Intel芯片 |
| macOS | arm64 | Apple Silicon (M1/M2/M3) |
| Linux | amd64 | x86_64服务器 |
| Linux | arm64 | ARM64服务器/树莓派 |
| Windows | amd64 | x86_64桌面 |
| Windows | arm64 | Surface Pro X等ARM设备 |

## 前置要求

- Go 1.25+ 
- Node.js 18+ (用于编译前端)
- Git (用于获取commit信息)
- UPX (可选，用于压缩)

## 使用方法

### 基本用法

编译默认版本 v1.0.0：

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
```

### 指定版本号

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.2.0
```

### 启用压缩

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.2.0 true
```

参数说明：
- 第1个参数：版本号（默认 v1.0.0）
- 第2个参数：是否压缩（默认 false）

## 输出目录结构

编译完成后，产物位于 `release/<version>/` 目录：

```
release/v1.0.0/
├── lottery-assistant-v1.0.0-darwin-amd64/
│   ├── lottery-assistant
│   └── VERSION.txt
├── lottery-assistant-v1.0.0-darwin-arm64/
│   ├── lottery-assistant
│   └── VERSION.txt
├── lottery-assistant-v1.0.0-linux-amd64/
│   ├── lottery-assistant
│   └── VERSION.txt
├── lottery-assistant-v1.0.0-linux-arm64/
│   ├── lottery-assistant
│   └── VERSION.txt
├── lottery-assistant-v1.0.0-windows-amd64/
│   ├── lottery-assistant.exe
│   └── VERSION.txt
├── lottery-assistant-v1.0.0-windows-arm64/
│   ├── lottery-assistant.exe
│   └── VERSION.txt
└── RELEASE_NOTES.md
```

## VERSION.txt 格式

每个平台文件夹都包含一个 `VERSION.txt` 文件，记录版本信息：

```
Lottery
Version: v1.0.0
Build Time: 2026-03-20T16:30:33Z
Git Commit: d80a8e9
Platform: darwin amd64
```

## RELEASE_NOTES.md

自动生成的发行说明，包含：
- 版本信息
- 各平台下载链接
- 使用方法
- 系统要求

## 二进制压缩（可选）

启用压缩后，会使用UPX压缩二进制文件，可以减小30-50%的体积：

- macOS amd64: 12M → 6M
- macOS arm64: 11M → 5M
- Linux amd64: 11M → 5M
- Windows amd64: 12M → 6M

安装UPX：

```bash
# macOS
brew install upx

# Ubuntu/Debian
sudo apt-get install upx

# Fedora
sudo dnf install upx
```

## 版本嵌入

编译时会将以下信息嵌入到二进制文件中：

```go
var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

可以通过以下方式查看版本：

```bash
# 直接运行
./lottery

# API接口
curl http://localhost:8902/api/version
```

## 部署使用

### macOS/Linux

```bash
cd lottery-v1.0.0-<platform>
chmod +x lottery
./lottery
```

### Windows

```cmd
cd lottery-v1.0.0-<platform>
lottery.exe
```

访问 http://localhost:8902 使用应用

## 环境变量

可通过环境变量自定义配置：

```bash
# 修改端口
PORT=9000 ./lottery

# 修改数据库路径
DB_PATH=/data/lottery.db ./lottery
```

## 故障排查

### 前端编译失败

确保已安装Node.js依赖：

```bash
cd frontend
npm install
```

### 跨平台编译失败

确保Go环境正确：

```bash
go version  # 应该是 Go 1.25+
```

### UPX压缩失败

检查UPX是否安装：

```bash
upx --version
```

如果未安装，脚本会跳过压缩步骤继续执行。

## 注意事项

1. 编译产物包含嵌入的前端资源，无需单独部署
2. 首次运行会自动创建数据库文件
3. 默认端口为 8902
4. 数据库默认路径为 `./data/lottery-assistant.db`

## 项目集成

要为其他项目使用此skill，需要：

1. 确保项目结构符合要求：
   - `backend/` - Go后端代码
   - `frontend/` - 前端项目
   - `frontend/dist/` - 前端编译输出

2. 修改 `scripts/compile.sh` 中的配置：
   - `BINARY_NAME` - 二进制文件名
   - `PROJECT_NAME` - 项目名称

3. 确保backend/main.go中定义了版本变量：
   ```go
   var (
       Version   = "dev"
       BuildTime = "unknown"
       GitCommit = "unknown"
   )
   ```
