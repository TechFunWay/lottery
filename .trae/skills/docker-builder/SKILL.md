---
name: "docker-builder"
description: "Docker 多架构构建工具，用于为 techfunway.lottery 应用创建 amd64 和 arm64 架构的 Docker 镜像"
---

# Docker 多架构构建技能

## 功能描述

该技能用于为 Linux 系统的 amd64 和 arm64 平台构建 Docker 多架构镜像，使用 Docker Buildx 实现完全离线的本地构建（不推送）。

## 构建方式

使用 release 目录中已编译好的 Linux 可执行文件进行构建，无需重新编译，构建速度快，镜像体积小。

**镜像内容：**
- `lottery` - 主程序二进制文件
- `index.html` + `lottery-web/` - 前端静态文件

## 支持的架构

- linux/amd64 (x86_64)
- linux/arm64 (ARM64)

## 前置条件

1. 已安装 Docker，且版本支持 Buildx
2. release 目录中存在对应版本的可执行文件（linux-amd64 + linux-arm64）

## 使用方法

在项目根目录执行：

```bash
bash .trae/skills/docker-builder/scripts/docker_builder.sh
```

## Dockerfile

`Dockerfile.release` 使用预编译二进制文件构建，支持 `ARG TARGETARCH` 和 `ARG VERSION`。

## 构建结果

- `techfunways/lottery:v1.0.0` (多架构镜像，包含 amd64 + arm64)
- `techfunways/lottery:latest` (同上)

## 使用示例

```bash
docker run -d -p 8902:8902 techfunways/lottery:latest
```

## 注意事项

1. 构建过程不需要网络连接（完全本地构建）
2. 镜像基于 scratch，体积小巧
3. 如遇到 docker buildx 问题，请检查 Docker 版本
