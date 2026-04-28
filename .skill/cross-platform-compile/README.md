# Cross-Platform Compile Skill

## 快速开始

### 基本编译

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
```

编译默认版本 v1.0.0，生成所有平台的二进制文件。

### 指定版本号

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.2.0
```

### 启用压缩

```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh v1.2.0 true
```

压缩可减小30-50%的二进制文件体积（需要安装UPX）。

## 参数说明

| 参数 | 说明 | 默认值 |
|-----|------|-------|
| 第1个 | 版本号 | v1.0.0 |
| 第2个 | 是否压缩 | false |

## 输出位置

编译产物位于 `release/<version>/` 目录，例如：

```
release/v1.2.0/
├── lottery-v1.2.0-darwin-amd64/
├── lottery-v1.2.0-darwin-arm64/
├── lottery-v1.2.0-linux-amd64/
├── lottery-v1.2.0-linux-arm64/
├── lottery-v1.2.0-windows-amd64/
├── lottery-v1.2.0-windows-arm64/
└── RELEASE_NOTES.md
```

## 支持的平台

- **macOS**: Intel (amd64), Apple Silicon (arm64)
- **Linux**: x86_64 (amd64), ARM64
- **Windows**: x86_64 (amd64), ARM64

## 详细文档

查看 [SKILL.md](SKILL.md) 了解更多详情。
