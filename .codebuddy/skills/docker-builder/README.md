# Docker Builder Skill

## 快速开始

### 本地构建

构建 Docker 镜像（默认版本 v1.0.0）：

```bash
.codebuddy/skills/docker-builder/scripts/build.sh
```

### 指定版本

```bash
.codebuddy/skills/docker-builder/scripts/build.sh v1.2.0
```

### 手动推送镜像

构建后手动推送到仓库：

```bash
# 推送到 Docker Hub
docker login
docker tag lottery-assistant:v1.2.0 yourusername/lottery-assistant:v1.2.0
docker push yourusername/lottery-assistant:v1.2.0
```

## 参数说明

| 参数 | 说明 | 默认值 |
|-----|------|-------|
| 第1个 | 版本号 | v1.0.0 |

## 运行容器

### 基本运行

```bash
docker run -d \
  --name lottery-assistant \
  -p 8902:8902 \
  -v /path/to/data:/app/data \
  lottery-assistant:v1.2.0
```

### 使用 Docker Compose

```bash
docker-compose up -d
```

## 环境变量

| 变量 | 说明 | 默认值 |
|-----|------|-------|
| IMAGE_NAME | 镜像名称 | lottery-assistant |
| PLATFORMS | 构建平台 | linux/amd64,linux/arm64 |

## 详细文档

查看 [SKILL.md](SKILL.md) 了解更多详情。
